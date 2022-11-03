import os
import requests
import json
import datetime
from pprint import pprint

LIVE_URL = "https://api-fxtrade.oanda.com/v3"
DEMO_URL = "https://api-fxpractice.oanda.com/v3"

# api status判定用 res.status_code = 2XX -> OK 以外はERROR
OK = 1
ERROR = 0


def log(msg, at=""):
    now = datetime.datetime.now()
    nowstr = now.strftime("%Y-%m-%d %H:%M:%S")
    print(nowstr + ' ' + msg + "--at:" + at)

# checks start
# check_responseで利用 直接呼ばない


def check_http(res):
    '''
    Description
        resにstatus_codeが存在するかチェック
    Args
        res -> api response
    Returns
        True | None
    '''
    if res is None:
        return
    if not hasattr(res, "status_code"):
        return
    return True

# check_responseで利用 直接呼ばない


def check_content_type(res, ctype="application/json"):
    '''
    Description
        res.headersが存在するかチェック
        res.headers["Content-Type"]が存在するかチェック
        res.headers["Content-Type"]がctypeと一致するかチェック
        * 実質res.json()が実行可能か判定するための関数
    Args
        res -> api response
    Returns
        True | False | None
    '''
    if res is None:
        return
    if not hasattr(res, "headers"):
        return
    if not res.headers.get("Content-Type"):
        return
    return res.headers["Content-Type"] == ctype

# add_statusで利用　直接呼ばない


def check_status_code(code):
    return code >= 200 and code <= 299

# eval_json_responseで利用　直接呼ばない


def add_status(resj, code):
    '''
    Description
        res.json()にstatus_codeに応じた独自status{"ok":OK | ERROR}を追加
    Args
        resj[Dict] -> res.json()
        code[Int]-> res.status_code
    Returns
        resj[Dict]
    '''
    ok = OK if check_status_code(code) else ERROR
    resj["ok"] = ok
    return resj


def check_response(res):
    '''
    Description
        api実行結果のチェック
    Args
        res -> api response
    Returns
        True | None
    '''
    if res is None:
        return
    if not check_http(res):
        return
    if not check_content_type(res):
        return
    return True


def eval_json_response(res, caller=""):
    '''
    Description
        res.json()を実行して独自status_code付与
        status_codeがエラーならエラー出力してNoneを返す
    Args
        res-> res.jsonが可能であること(check_content_type が真)
    Returns
        Dict | None
    '''
    resj = res.json()
    resj_wstat = add_status(resj, res.status_code)
    if resj_wstat["ok"] == ERROR:
        ecode = resj_wstat["errorCode"] if resj_wstat.get("errorCode") else ""
        emsg = resj_wstat["errorMessage"] if resj_wstat.get(
            "errorMessage") else ""
        log(f"ERROR_CODE:{ecode}--MSG:{emsg}", caller)
        return
    return resj_wstat
# checks end

# transactions start


def parse_tactions(res):
    '''
    Args
        res[dict] -> tactionsResponse["pages"]の1要素。配列なので１つずつ呼ぶ
    '''
    if not res:
        return
    query = res.split("?")
    if len(query) == 1:
        return  # ?が無い場合リターン
    query = query[1]  # url~?key=val&key2=val2... の?以降
    params = query.split("&")

    start = None
    end = None
    for param in params:
        _query = param.split("=")
        if len(_query) == 2:  # = が存在する場合のみ
            if _query[0] == "from":
                start = _query[1]
            elif _query[0] == "to":
                end = _query[1]

    return start, end

# transactions end


class Oanda:
    def __init__(self, live=False):
        '''
        Description
            constructor
        Args
            live[bool] -> True->本番 False->デモ
        Returns
            oanda
        '''
        here = os.path.dirname(__file__)
        key_file = "./key.json"
        key_path = os.path.join(here, key_file)

        with open(key_path, "r") as file:
            data = json.load(file)

        self.token = data["live"]["token"] if live else data["demo"]["token"]
        self.accountId = data["live"]["id"] if live else data["demo"]["id"]
        self.url = LIVE_URL if live else DEMO_URL

    def headers(self, mime="application/json"):
        ret = {
            "Content-Type": mime,
            "Authorization": "Bearer " + self.token
        }
        return ret

    def get(self, url, params=None, caller=""):
        '''
        Description
            requests.get実行
        Args
            url[str] -> full
            params[Dict | None] -> getのparams指定するやつ
        Returns
            dict | None
        '''
        res = None
        headers = self.headers()
        try:
            res = requests.get(url, params=params, headers=headers)
        except requests.exceptions.ConnectionError as e:
            log(f"connection error at {url}:{e}", caller)
            return
        if not check_response(res):
            log(
                f"http error at {url}:check_http or check_content_type \
                    returned None", caller
            )
            return

        return eval_json_response(res, caller)

    def exec(self, url, params=None, body=None, caller="", method="POST"):
        headers = self.headers()
        data = json.dumps(body) if body else ""
        if method == "POST":
            func = requests.post
        elif method == "PUT":
            func = requests.put
        else:
            return
        try:
            res = func(url, headers=headers, params=params, data=data)
        except requests.exceptions.ConnectionError as e:
            log(f"connection error at {url}:{e}", caller)
            return

        if not check_response(res):
            log(f"http error at {url}:check_http or \
                check_content_type returned None", caller)
            return

        return eval_json_response(res, caller)

    def price(self, instruments="USD_JPY"):
        path = f"/accounts/{self.accountId}/pricing"
        uri = self.url + path
        params = {"instruments": instruments}
        res = self.get(uri, params, caller="price")
        return res

    def candles(self, instruments="USD_JPY", granularity="H1",
                count=20, start=None, end=None):
        path = f"/instruments/{instruments}/candles"
        uri = self.url + path
        params = {
            "instruments": instruments,
            "granularity": granularity,
            "count": count,
            "from": start,
            "to": end
        }
        res = self.get(uri, params, caller="candles")
        return res

    def positions(self, instruments=None):
        '''
        Description
            instrumentで指定された通貨のポジションを一覧を取得。無い場合は全て。
            指定あり->res["position"]
            指定なし->res["positions"]となるので注意
            ポジション単位でなく通貨単位に集約されるので注意
        '''
        path = f"/accounts/{self.accountId}/positions"
        if instruments:
            path += "/" + instruments
        uri = self.url + path
        res = self.get(uri, caller="positions")
        return res

    # 注文一覧取得
    def get_orders(self, state="PENDING", instrument=None,
                   ids=None, count=50, beforeID=None):
        '''
        Description
            注文一覧をゲット。state,instrument,ids,beforeIDで抽出可能。
            デフォルトでは未決済一覧。
        '''
        path = f"/accounts/{self.accountId}/orders"
        uri = self.url + path
        params = {
            "state": state, "instrument": instrument, "ids": ids
        }
        res = self.get(uri, params=params, caller="get_orders")
        return res

    # 新規取引
    def post_order(self, type="MARKET", instrument="USD_JPY",
                   units=None, timeInForce="FOK", **opt):

        path = f"/accounts/{self.accountId}/orders"
        uri = self.url + path

        order = {
            "type": type,
            "instrument": instrument,
            "units": units,
            "timeInForce": timeInForce
        }
        order = {**order, **opt}
        body = {"order": order}

        res = self.exec(uri, body=body, caller="order")
        return res

    # 建玉に決済注文（指値、逆差し、とれ～る）
    def post_closing_order(self, **opt):
        '''
        想定keys:type,id,price,distance,timeInForce,gtdTime
        '''
        path = f"/accounts/{self.accountId}/orders"
        uri = self.url + path
        body = {"order": opt}
        res = self.exec(uri, body=body, caller="post_closing_order")
        return res

    # 注文キャンセル
    def cancel_order(self, orderId):
        '''
        Description
            orderIdの注文をキャンセルし、optの注文で差し替え
        Args
            orderId[str] -> 注文した時の'id'だと思う
        Returns
            Dict | None
        '''
        path = f"/accounts/{self.accountId}/orders/{orderId}/cancel"
        uri = self.url + path
        res = self.exec(uri, method="PUT", caller="cancel_orders")
        return res

    # 注文変更
    def replace_order(self, orderId, **opt):
        '''
        Description
            orderIdの注文をキャンセルし、optの注文で差し替え
        Args
            orderId[str] -> 注文した時の'id'だと思う
            opt[OrderRequest] -> ドキュメントのOrderRequest
        Returns
            Dict | None
        '''
        path = f"/accounts/{self.accountId}/orders/{orderId}"
        uri = self.url + path
        data = {"order": opt}
        res = self.exec(uri, body=data, method="PUT", caller="replace_order")
        return res

    # MARKET close
    def close_order(self, instrument="USD_JPY",
                    longUnits="NONE", shortUnits="NONE"):
        '''
        Description
            instrumentで指定したペアを成行で決済する
        Args
            longUnits[int] :決済するロンポジ数量
            shortUnits[int]:決済するショートポジ数量
        Returns
            Dict | None
        '''
        if longUnits == "NONE" and shortUnits == "NONE":
            return
        path = f"/accounts/{self.accountId}/positions/{instrument}/close"
        uri = self.url + path
        body = {
            "longUnits": str(longUnits),
            "shortUnits": str(shortUnits)
        }
        res = self.exec(uri, body=body, method="PUT", caller="close_order")
        return res

    def tactions(self, pageSize=None, start=None, end=None):
        '''
        Args
            pageSize[int]
            start[str] -> date
            end[str] -> date
        '''
        path = f"/accounts/{self.accountId}/transactions"
        uri = self.url + path
        params = {
            "pageSize": pageSize,
            "from": start,
            "to": end
        }
        res = self.get(uri, params, caller="transactions")
        return res

    def tactions_idrange(self, id_from=None, id_end=None):
        '''
        Args
            id_from[str],id_end[str]
        '''
        path = f"/accounts/{self.accountId}/transactions/idrange"
        uri = self.url + path
        params = {"from": id_from, "to": id_end}
        res = self.get(uri, params, caller="tacitions_idrange")
        return res

    def account(self, accountId=None):
        path = f"/accounts/{accountId or self.accountId}"
        uri = self.url + path
        res = self.get(uri, caller="account")
        return res

    def position_book(self, instruments="USD_JPY"):
        path = f"/instruments/{instruments}/positionBook"
        uri = self.url + path
        res = self.get(uri, caller="pBook")
        return res


if __name__ == "__main__":

    o = Oanda(True)
    c = o.position_book()
    pprint(c)
    with open("test.json", mode="w") as f:
        json.dump(c, f, indent=2)
    # c = o.candles(granularity="H4", count=50, end="2021-11-01T05:00:00")
    # if c is not None:
    #     pprint(c)
    '''
    t = o.tactions()
    page = t["pages"][0]
    pprint(t["pages"])
    s,e = parse_tactions(page)
    res = o.tactions_idrange(s,e)["transactions"]
    ts = [data for data in res if data.get("pl")]
    prof = sum([float(d["pl"]) for d in ts])
    pprint(f"prof:{prof}")
    '''
    # a = o.account()
    # pprint(a)

    # order = rorder.limit(units=-1000,price=116.50)
    # res = o.post_order(**order)

    # param = rorder.add_stop_loss_by_id('114',116.8)
    # res = o.post_closing_order(**param)
    # print(res)

    '''
    order = rorder.limit(units=-1000,price=116)
    order = rorder.add_take_profit(order,115.50)
    order = rorder.add_stop_loss(order,116.50)
    r = o.post_order(**order)
    print(r)
    pprint(o.get_orders())
    '''
    # param = rorder.stop(units=1000,price=116)
    # r = o.post_order(**param)
    # print(r)
    # param = rorder.limit(units=-1000,price=116)
    # r = o.post_order(**param)
    # oo = rorder.market(units=1000)
    # r = o.post_order(**oo)
    # c = o.cancel_order('75')
    # print(c)
    '''
    order = {
        "type":"LIMIT",
        "instrument":"USD_JPY",
        "units":"-1000",
        "price":"116.00",
        "timeInForce":"GTC",
    }
    r = o.replace_order("71",**order)
    '''
    # c = o.cancel_order("72")
    # print(c)
    # r = o.get_orders()
    # pprint(r)
    # r = o.price("USD_JPY,EUR_USD")
    # print(r)
    # c = o.candles()
    # print(c)
    # p = o.positions()
    # u = o.positions("USdD_JPY")
    # p = o.price("USD_JPdY")
    # c = o.candles("USD_JPY")
    # r =o.post_order(units="1000")
    # c = o.close_order(shortUnits="1000")
    # print(c)
    '''
    order = {
        "orderType":"LIMIT",
        "instrument":"USD_JPY",
        "units":"-1000",
        "price":"115.50",
        "timeInForce":"GTC",
        "takeProfitOnFill":{
            "price":"116.50",
            "timeInForce":"GTC"
        },
        "stopLossOnFill":{
            "price":"116.50",
            "timeInForce":"GTC"
        }
    }
    l = o.post_order(**order)
    pprint(l)
    '''
