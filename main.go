package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/zenryokukun/surfergopher/minmax"
)

const (
	POS_FILE = "./pos.json"
	BAL_FILE = "./bal.json"
)

type TestCandleStick struct {
	// Pricesがmid ask bidで可変のため、
	// testdataからロード時（mid）にフォーマットできるように
	// field tagをつけたもの。
	Complete bool
	Time     string
	Unix     int64
	Prices   *Hloc `json:"mid"`
}

type TestCandleSticks []*TestCandleStick

func (s TestCandleSticks) Extract(hloc string) []float64 {

	vals := []float64{}
	for _, s := range s {
		val := 0.0
		switch hloc {
		case "L":
			val = s.Prices.L
		case "H":
			val = s.Prices.H
		case "O":
			val = s.Prices.O
		case "C":
			val = s.Prices.C
		default:
			val = 0.0
		}
		vals = append(vals, val)
	}
	return vals
}

type Chart struct {
	X      []int64
	Y      []float64
	Side   []string //"BUY"" or "S"ELL"
	Action []string //"OPEN" or "CLOSE"
}

type Balance struct {
	X []int64
	Y []float64
}

type TestPosition struct {
	size  float64
	price float64
	side  string
	time  int64
}

type Summary struct {
	TestPosition
	chart Chart
	pl    float64 //total profit
	//profR  float64 // profit ratio
	//lossR  float64 //loss ratio MUST BE NEGATIVE
	spread float64
	cnt    int // count of trades
}

//methods

func (b *Balance) add(ot int64, v float64) {
	b.X = append(b.X, ot)
	b.Y = append(b.Y, v)
}

func (b *Balance) write(fpath string) {
	if b, err := json.MarshalIndent(b, "", " "); err == nil {
		os.WriteFile(fpath, b, 0777)
	}
}

func (ch *Chart) add(ot int64, v float64, side, act string) {
	ch.X = append(ch.X, ot)
	ch.Y = append(ch.Y, v)
	ch.Side = append(ch.Side, side)
	ch.Action = append(ch.Action, act)
}
func (ch *Chart) write(fpath string) {
	if b, err := json.MarshalIndent(ch, "", " "); err == nil {
		os.WriteFile(fpath, b, 0777)
	}
}

func (p *TestPosition) has() bool {
	return p.size != 0.0
}

func (p *TestPosition) check(v float64) float64 {
	if p.side == "BUY" {
		return (v - p.price) * p.size
	} else {
		return (p.price - v) * p.size
	}
}

func (s *Summary) isProfFilled(v, profR float64) bool {
	if !s.TestPosition.has() {
		return false
	}
	if s.side == "BUY" {
		return (v-s.price)/s.price >= profR
	} else {
		return (s.price-v)/s.price >= profR
	}
}
func (s *Summary) isLossFilled(v, lossR float64) bool {
	if !s.TestPosition.has() {
		return false
	}
	if s.side == "BUY" {
		return (v-s.price)/s.price <= lossR
	} else {
		return (s.price-v)/s.price <= lossR
	}
}

func (s *Summary) open(price, size float64, otime int64, side string) {
	s.price = price
	s.size = size
	s.side = side
	s.time = otime
	s.chart.add(otime, price, side, "OPEN")
}
func (s *Summary) close(price float64, otime int64) float64 {
	var pl float64
	if s.side == "BUY" {
		pl = (price - s.price - s.spread) * s.size
		s.pl += pl
	} else {
		pl = (s.price - price - s.spread) * s.size
		s.pl += pl
	}
	//add to chart
	var side string
	if s.side == "BUY" {
		side = "SELL"
	} else {
		side = "BUY"
	}

	//print infos
	//fmt.Printf("CLOSE-%v! open:%v price:%.f close:%v price:%.f prof:%.f\n",s.side,s.time, s.price, otime, price,pl)
	s.chart.add(otime, price, side, "CLOSE")
	s.cnt++

	//reset
	s.price = 0.0
	s.size = 0.0

	s.side = ""
	s.time = 0
	return pl
}

// jsonファイルをロードしてTestCandleSticksにparse
func load(fpath string) TestCandleSticks {
	sticks := []*TestCandleStick{}
	if b, err := os.ReadFile(fpath); err == nil {
		json.Unmarshal(b, &sticks)
		return sticks
	}
	return nil
}

func breakThrough(v float64, inf *minmax.Inf) string {
	// volatility
	// vol := 1 - (inf.Minv / inf.Maxv)
	// vol := 1 - (inf.Maxv - inf.Minv)

	if v > inf.Maxv {
		// fmt.Println("-------------------------")
		// fmt.Printf("upperBread:%v\n", inf.Scaled)
		// fmt.Printf("max:%v,min:%v,maxi:%v,mini:%v,current:%v\n", inf.Maxv, inf.Minv, inf.Maxi, inf.Mini, v)
		// if vol > thresh {
		// 	return "BUY"
		// }
		// return "SELL"
		return "BUY"
	}
	if v < inf.Minv {
		// fmt.Println("-------------------------")
		// fmt.Printf("lowerBread:%v\n", inf.Scaled)
		// fmt.Printf("max:%v,min:%v,maxi:%v,mini:%v,current:%v\n", inf.Maxv, inf.Minv, inf.Maxi, inf.Mini, v)
		// if vol > thresh {
		// 	return "SELL"
		// }
		// return "BUY"
		return "SELL"
	}
	return ""
}

func logic(scaled float64) string {
	switch {
	case scaled <= -1.0:
		return "SELL"
	case scaled >= 1:
		return "BUY"
	}
	return ""
}

func toUnix(t string) int64 {
	unix, err := time.Parse(layout(), t)
	if err != nil {
		fmt.Println(err)
	}
	return unix.Unix()
}

func frame(goq *Goquest) {

	sticks := load("./testdata.json")
	if sticks == nil {
		return
	}

	span := 12 // ロジックに使うロウソク足の数
	tLossR, tProfR := -0.005, 0.005
	thresh := 0.0025 // レンジ判定の閾値
	tsize := 10000.0

	pos := &Summary{spread: 0.008}
	bal := &Balance{}

	for i, stick := range sticks[span+1:] {
		ed := span + i
		st := ed - span
		otime := sticks[ed].Unix

		highs := sticks[st:ed].Extract("H")
		lows := sticks[st:ed].Extract("L")
		// 次のopen価格を現在価格とする
		current := stick.Prices.O

		inf := minmax.NewInf(highs, lows).AddWrap(current)
		vel := 1 - (inf.Minv / inf.Maxv)

		dec := breakThrough(current, inf)
		if pos.has() {
			//逆向きポジなら強制閉じる
			if len(dec) > 0 && pos.side != dec && vel > thresh {
				pl := pos.close(current, otime)
				fmt.Printf("break:%v\n", pl)
			}
		}

		if pos.isLossFilled(current, tLossR) {
			pl := pos.close(current, otime)
			fmt.Printf("loss:%v\n", pl)
		}

		if pos.isProfFilled(current, tProfR) {
			pl := pos.close(current, otime)
			fmt.Printf("prof:%v\n", pl)
		}

		// ここはvel > threshを入れないほうが利益出てる。
		if dec != "" && !pos.has() {
			pos.open(current, tsize, otime, dec)
		}

		bal.add(otime, pos.pl)
	}
	fmt.Printf("prof:%.f trades:%v\n", pos.pl, pos.cnt)
	pos.chart.write(POS_FILE)
	bal.write(BAL_FILE)
}

// testdata.jsonにunix時間をつけるため1回だけ実行
func once() {
	sticks := load("./testdata.json")
	fmt.Println(len(sticks))
	for _, stick := range sticks {
		unixTime, _ := time.Parse(layout(), stick.Time)
		stick.Unix = unixTime.Unix()
	}
	// Unixを付けるためにファイル上書き
	writeFile("./testdata.json", sticks)
}

// 新規にテストデータ取得
func newTestData(goq *Goquest, unixStr string) {
	testFile(goq, "M5", "USD_JPY", 5000, "", unixStr)
}

type X struct {
	X []int `json:"X"`
}

func newCompareData(goq *Goquest) {
	b, _ := ioutil.ReadFile("../oanda-bot/trade.json")
	x := &X{}
	json.Unmarshal(b, x)

	// if len(x.X) > 5000 {
	// 	i := len(x.X) - 5000
	// 	x.X = x.X[i:]
	// }
	start := x.X[0]
	unixStr := strconv.Itoa(start)
	data := NewCandles(goq, 5000, "M5", "USD_JPY", unixStr, "", "").Extract()
	_ = data
	nu := nextUnix(data[len(data)-1].Time, 300)
	prevlast, _ := time.Parse(layout(), data[len(data)-1].Time)

	for {
		cd := NewCandles(goq, 5000, "M5", "USD_JPY", nu, "", "")
		sticks := cd.Extract()
		if len(sticks) == 0 {
			fmt.Println("sticks length was 0. Breaking loop")
			break
		}
		lastT := sticks[len(sticks)-1].Time
		_lastU, _ := time.Parse(layout(), lastT)
		// lastU := _lastU.Unix()
		// nuInt, _ := strconv.ParseInt(nu, 10, 0)
		if prevlast == _lastU {
			println("breaking lool:", nu)
			break
		}
		// if nuInt >= lastU {
		// 	println("breaking lool:", nu)
		// 	break
		// }
		data = append(data, sticks...)
		nu = nextUnix(sticks[len(sticks)-1].Time, 300)
	}

	writeFile("./testdata.json", data)

	// testFile(goq, "M5", "USD_JPY", 5000, unixStr, "")
}

func execBackTest(g *Goquest, isNew bool) {
	if isNew {
		newTestData(g, "")
	}
	once()
	frame(g)
}

func execCompare(g *Goquest) {
	newCompareData(g)
	once()
	frame(g)

	// python実行
	cmd := exec.Command(genPyCommand(), "./graph.py")
	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println((string(b)))

	// 取引指標を取得
	cmd = exec.Command(genPyCommand(), "../oanda-eval/main.py")
	b, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}

	// tweetメッセージ作成
	result := NewEval(string(b)).String()
	month := int(time.Now().AddDate(0, -1, 0).Month())
	msg := "💰Wavenauts:" + fmt.Sprint(month) + "月分報告とバックテスト比較💰\n"
	msg += result + "\n"
	msg += "#FX #USD/JPY #プログラミング初心者"

	// tweet
	t := NewTwitter("../oanda-bot/twitter.json")
	t.tweetImage(msg, "result.png")
}

func main() {

	// os.Args[0] -> 実行ファイルのパス。
	// os.Args[1:] -> コマンドライン引数
	// os.Argが長さ1（引数なし）なら、既存のファイルでテスト実行（[]CandleData型のjsonファイルを./testdata.jsonの名前で保存しておく必要あり）
	// os.Arg[1]が "compare"なら実取引との比較を実行。

	goq := NewGoquest("../oanda-bot/key.json", "live")
	_ = goq

	// コマンドライン引数なしなら、既存ファイルでバックテスト
	if len(os.Args) <= 1 {
		fmt.Println("backtest using existing testdata.json")
		execBackTest(goq, false)
		return
	}

	// コマンドライン引数が"test-new"なら、最新ロウソク足を取得してバックテスト
	if os.Args[1] == "test-new" {
		fmt.Println("backtest using new candle data")
		execBackTest(goq, true)
		return
	}

	// コマンドライン引数が"compare"なら、コンペア実行。事前にサーバからbalacne.json、trade.jsonが必要
	if os.Args[1] == "compare" {
		fmt.Println("compare backtest and real trade")
		execCompare(goq)
		return
	}

	fmt.Println("Your command line argument is wrong! Only no argument,'test-new',or 'compare' is accepted.")

	// ********************************************************************************************
	// 新しいデータでテストを実施する方法：
	// 1. newTestDataを呼ぶ。onceとframeはコメントアウト。APIでロウソク足をとってファイルに書き出される
	//    実取引期間と比較する際には、サーバからtrade.jsonを取って来て、代わりにnewCompareDataを呼ぶ
	// 2. onceを呼ぶ。newTestDataとframeはコメントアウト。1のファイルにUnixタイムスタンプを追加する
	// 3. frameを呼ぶ。newTestDataとonceはコメントアウト。2のファイルでバックテストが実行される。
	//
	// 既存のファイルからテストする場合は、[]CandleData型のjsonファイルを./testdata.jsonの名前で保存し、
	// 上記2.,3.を実行
	// ********************************************************************************************

	// newTestData(goq, "")
	// newCompareData(goq)
	// once()
	// frame(goq)

}

// *******************************
// * Test 済
// *******************************

/****Pricing****/
// res := NewPricing(goq, "USD_JPY,EUR_USD")
// prettyPrint(res.Prices)
// fmt.Println(res.Spread("EUR_USD"))

/****Account****/
// res := NewAccount(goq)
// prettyPrint(res)
// totalPL(res)

/****candles****/
// unix := time.Now().Unix()
// _ = unix
// ask := NewCandles(goq, 10, "H4", "USD_JPY", "", "", "A")
// bid := NewCandles(goq, 10, "H4", "USD_JPY", "", "", "B")
// ad := ask.Extract()
// bd := bid.Extract()
// lastAd := ad[len(ad)-1]
// lastBd := bd[len(bd)-1]
// fmt.Printf("%+v", lastAd.Ask)
// fmt.Println()
// fmt.Printf("%+v", lastBd.Bid)
// fmt.Println(lastAd.Ask.C - lastBd.Bid.C)

/****close****/
// res := NewMarketClose(goq, "USD_JPY", 1000, 0)
// fmt.Println(res.statusCode)
// fill := res.LongFillTransaction
// fmt.Printf("%+v\n", fill)
// fmt.Println()
// fmt.Printf("%v", res)

/****open****/
// res := NewMarketOrder(goq, "USD_JPY", 10000)
// fmt.Println(res)

/****position****/
// res := NewPosition(goq, "EUR_USD")
// fmt.Printf("%+v", res)
// prettyPrint(res)
