package main

import (
	"encoding/json"
	"fmt"
	"os"
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

func main() {
	goq := NewGoquest("./key.json", "live")
	_ = goq

	// ********************************************************************************************
	// 新しいデータでテストを実施する方法：
	// 1. newTestDataを呼ぶ。onceとframeはコメントアウト。APIでロウソク足をとってファイルに書き出される
	// 2. onceを呼ぶ。newTestDataとframeはコメントアウト。1のファイルにUnixタイムスタンプを追加する
	// 3. frameを呼ぶ。newTestDataとonceはコメントアウト。2のファイルでバックテストが実行される。
	// ********************************************************************************************

	// newTestData(goq, "")
	// once()
	frame(goq)
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
