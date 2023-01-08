import json
import matplotlib.pyplot as plt
import datetime

ACTUAL_F = "./testdata.json"
BKTEST_F = "./bal.json"
CANDLES_F = "./candles.json"
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


if __name__ == "__main__":
    graph()
