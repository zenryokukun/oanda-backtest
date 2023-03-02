import json
import matplotlib.pyplot as plt
import datetime

ACTUAL_F = "./testdata.json"
BKTEST_F = "./bal.json"
# CANDLES_F = "./candles.json"
UNIT = 10000
PAIR = "USD_JPY"


def load(fpath):
    with open(fpath) as f:
        return json.load(f)


# testdata成形
def mold(data):
    ret = {
        "X": [],
        "Y": [],
    }
    for d in data:
        # t = datetime.datetime.strptime(
        #     d["time"], "%Y-%m-%dT%H:%M:%S.000000000Z")

        ret["X"].append(d["Unix"])
        ret["Y"].append(d["mid"]["c"])
    return ret


def slice_data(obj, start):
    i = 0
    for i, j in enumerate(obj["X"]):
        if j >= start:
            break

    for k, v in obj.items():
        obj[k] = v[i:]


# backtest用
def graph():
    ac = mold(load(ACTUAL_F))
    bk = load(BKTEST_F)
    pd = load("./pos.json")

    unit = 1
    bk["X"] = [datetime.datetime.fromtimestamp(v) for v in bk["X"][::unit]]
    bk["Y"] = [float(v) for v in bk["Y"][::unit]]
    ac["X"] = [datetime.datetime.fromtimestamp(v) for v in ac["X"][::unit]]
    ac["Y"] = [float(v) for v in ac["Y"][::unit]]
    # ポジションデータなので[::unit]の間引きは不要
    pd["X"] = [datetime.datetime.fromtimestamp(v) for v in pd["X"]]

    openbuy_x = []
    openbuy_y = []
    opensell_x = []
    opensell_y = []
    close_x = []
    close_y = []
    for i in range(len(pd["X"])):
        _x = pd["X"][i]
        _y = pd["Y"][i]
        if pd["Action"][i] == "OPEN":
            if pd["Side"][i] == "BUY":
                openbuy_x.append(_x)
                openbuy_y.append(_y)
            else:
                opensell_x.append(_x)
                opensell_y.append(_y)
        else:
            close_x.append(_x)
            close_y.append(_y)

    fig = plt.figure()
    # plt.rcParams["axes.facecolor"] = (1, 1, 1, 0)
    # ******左グラフ
    # 利益
    ax = fig.add_subplot(111)
    ax.set_ylabel("balance")
    ax.plot(bk["X"], bk["Y"], label="backtest", color="orange")

    # ******右グラフ
    # 実際の価格
    ax2 = ax.twinx()
    ax2.set_ylabel("price")
    ax2.plot(ac["X"], ac["Y"], label="USD_JPY")
    # 取引箇所
    ax2.scatter(openbuy_x, openbuy_y, label="@backtest_openBuy", color="red")
    ax2.scatter(opensell_x, opensell_y,
                label="@backtest_openSell", color="lime")
    ax2.scatter(close_x, close_y, label="@backtest_close",
                facecolors="none", edgecolors="black", s=80)

    plt.title(f"Oanda BackTest @{PAIR}:{UNIT}")
    plt.xlabel("TIME")
    ax.legend(loc=2)
    ax2.legend(loc=3)
    plt.gcf().autofmt_xdate()
    plt.tight_layout()
    plt.grid(True)
    plt.show()

    # plt.title("Oanda BackTest")
    # plt.plot(ac["X"], ac["Y"], label="backtest")
    # plt.xlabel("TIME")
    # plt.ylabel("JPY")
    # plt.legend()
    # plt.xticks(rotation=30)
    # plt.tight_layout()
    # plt.grid(True)
    # plt.show()


# 実取引とのコンペア用
def graph_compare():
    ac = mold(load(ACTUAL_F))  # ロウソク
    bk = load(BKTEST_F)  # バックテスト結果
    pd = load("./pos.json")  # バックテストポジション
    bl = load("../oanda-bot/balance.json")  # リアル結果
    rp = load("../oanda-bot/trade.json")  # リアル　ポジション

    unit = 1
    bk["X"] = [datetime.datetime.fromtimestamp(v) for v in bk["X"][::unit]]
    bk["Y"] = [float(v) for v in bk["Y"][::unit]]
    ac["X"] = [datetime.datetime.fromtimestamp(v) for v in ac["X"][::unit]]
    ac["Y"] = [float(v) for v in ac["Y"][::unit]]
    bl["X"] = [datetime.datetime.fromtimestamp(v) for v in bl["X"][::unit]]
    bl["TotalPL"] = [float(v) for v in bl["TotalPL"][::unit]]
    # ポジションデータなので[::unit]の間引きは不要
    pd["X"] = [datetime.datetime.fromtimestamp(v) for v in pd["X"]]
    rp["X"] = [datetime.datetime.fromtimestamp(v) for v in rp["X"]]

    # リアル結果はデータ量が少ないので、他のデータをリアルの最初のXに揃える、
    start = bl["X"][0]
    slice_data(ac, start)
    slice_data(bk, start)
    slice_data(pd, start)
    slice_data(rp, start)

    openbuy_x = []
    openbuy_y = []
    opensell_x = []
    opensell_y = []
    close_x = []
    close_y = []
    for i in range(len(pd["X"])):
        _x = pd["X"][i]
        _y = pd["Y"][i]
        if pd["Action"][i] == "OPEN":
            if pd["Side"][i] == "BUY":
                openbuy_x.append(_x)
                openbuy_y.append(_y)
            else:
                opensell_x.append(_x)
                opensell_y.append(_y)
        else:
            close_x.append(_x)
            close_y.append(_y)

    real_openbuy_x = []
    real_openbuy_y = []
    real_opensell_x = []
    real_opensell_y = []
    real_close_x = []
    real_close_y = []
    for i in range(len(rp["X"])):
        _x = rp["X"][i]
        _y = rp["Y"][i]
        if rp["Action"][i] == "OPEN":
            if rp["Side"][i] == "BUY":
                real_openbuy_x.append(_x)
                real_openbuy_y.append(_y)
            else:
                real_opensell_x.append(_x)
                real_opensell_y.append(_y)
        else:
            real_close_x.append(_x)
            real_close_y.append(_y)

    fig = plt.figure()
    # plt.rcParams["axes.facecolor"] = (1, 1, 1, 0)
    # ******左グラフ
    # 利益
    ax = fig.add_subplot(111)
    ax.set_ylabel("balance")
    ax.plot(bk["X"], bk["Y"], label="backtest", color="orange")
    ax.plot(bl["X"], bl["TotalPL"], label="real", color="turquoise")

    # ******右グラフ
    # 実際の価格
    ax2 = ax.twinx()
    ax2.set_ylabel("price")
    ax2.plot(ac["X"], ac["Y"], label="USD_JPY")
    # 取引箇所
    ax2.scatter(openbuy_x, openbuy_y, label="@backtest_openBuy", color="red")
    ax2.scatter(opensell_x, opensell_y,
                label="@backtest_openSell", color="lime")
    ax2.scatter(close_x, close_y, label="@backtest_close",
                facecolors="none", edgecolors="black", s=80)

    ax2.scatter(real_openbuy_x, real_openbuy_y, label="@real_openBuy",
                facecolors="none", edgecolors="blue", s=160)
    ax2.scatter(real_opensell_x, real_opensell_y, label="@real_openSell",
                facecolors="none", edgecolors="yellow", s=160)
    ax2.scatter(real_close_x, real_close_y, label="@real_close",
                facecolors="none", edgecolors="brown", s=200)

    plt.title(f"Comparing Oanda @{PAIR}:{UNIT}")
    plt.xlabel("TIME")
    ax.legend(loc=2)
    ax2.legend(loc=3)
    plt.gcf().autofmt_xdate()
    plt.tight_layout()
    plt.grid(True)
    # plt.show()
    plt.savefig("result.png")


if __name__ == "__main__":
    graph_compare()
