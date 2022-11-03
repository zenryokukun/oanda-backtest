/*
 * Oanda API Requestのパラメタの型を定める
 */
package main

import (
	"fmt"
	"strconv"
)

type (
	TakeProfitParam struct {
		Price float64 `json:"price,string"`
		Tif   string  `json:"timeInForce"`
		Gtd   string  `json:"gtdTime"`
	}

	StopLossParam struct {
		Price    float64 `json:"price,string"`
		Distance float64 `json:"distance,string"`
		Tif      string  `json:"timeInForce"`
		Gtd      string  `json:"gtdTime"`
	}

	TrailingStopLossParam struct {
		Distance float64 `json:"distance,string"`
		Tif      string  `json:"timeInForce"`
		Gtd      string  `json:"gtdTime"`
	}

	OrderParam struct {
		Type       string `json:"type"`
		Instrument string `json:"instrument"`
		Units      int    `json:"units,string"`
		Tif        string `json:"timeInForce"`
		// PriceBound float64                `json:"priceBound,string"`
		TakeProfit TakeProfitParam       `json:"takeProfitOnFill"`
		StopLoss   StopLossParam         `json:"stopLossOnFill"`
		Trailing   TrailingStopLossParam `json:"trailingStopLossOnFill"`
	}
)

// from,to両方指定した場合、countの指定は出来ないので、0以下の数値を渡すこと。
// from,to は　"YYYY-mm-ddTHH:MM:SS.000000000Z" もしくは unix時間を文字列にしたもの(fmt.Sprintf("%v",time.Now().Unix())とか)
func candlesParam(
	p strMap,
	count int,
	granularity, instruments, from, to string,
) {
	if count > 0 {
		cntStr := strconv.Itoa(count)
		p["count"] = cntStr
	}
	p["granularity"] = granularity
	p["instruments"] = instruments
	if from != "" {
		p["from"] = from
	}
	if to != "" {
		p["to"] = to
	}
}

// 成行注文のbaseパラメタ
func marketOrderParam(p iMap, instrument string, units int, tif string) {
	if tif == "" {
		tif = "FOK" // DEFAULT Fill or Kill.
	}
	p["order"] = iMap{
		"type":        "MARKET",
		"instrument":  instrument,
		"units":       units,
		"timeInForce": tif,
	}
}

// dtime は　"YYYY-mm-ddTHH:MM:SS.000000000Z" もしくは unix時間を文字列にしたもの(fmt.Sprintf("%v",time.Now().Unix())とか)
// "YYYY～"で指定するときは、UTCで指定すること。time.Now().Add(- 9 * time.Hour)　とかして、日本時間から９時間引く必要ある。
// 面倒だったらunix時間でやること
func timeParam(p strMap, dtime string) {
	if dtime != "" {
		p["time"] = dtime
	}
}

func instrumentParam(p strMap, instrument string) {
	if instrument != "" {
		p["instrument"] = instrument
	}
}

func NewCandles(
	goq *Goquest,
	count int,
	granularity, instruments, from, to string,
) *Candles {
	res := &Candles{}
	ep := fmt.Sprintf("/instruments/%v/candles", instruments)
	param := strMap{}
	candlesParam(param, count, granularity, instruments, from, to)
	goq.Get(ep, param, res)
	return res
}

func populateBook(goq *Goquest, ep string, dtime string, i Checker) {
	param := strMap{}
	timeParam(param, dtime)
	goq.Get(ep, param, i)
}

// dtimeはうまく効かない。どういうデータが返ってきているかよく分からない
func NewPositionBook(goq *Goquest, instruments string, dtime string) *PositionBook {
	res := &PositionBook{}
	ep := fmt.Sprintf("/instruments/%v/positionBook", instruments)
	populateBook(goq, ep, dtime, res)
	return res
}

// dtimeはうまく効かない。どういうデータが返ってきているかよく分からない
func NewOrderBook(goq *Goquest, instruments string, dtime string) *OrderBook {
	res := &OrderBook{}
	ep := fmt.Sprintf("/instruments/%v/orderBook", instruments)
	populateBook(goq, ep, dtime, res)
	return res
}

// 通貨単位のPositionを取得
func NewPosition(goq *Goquest, instrument string) *Position {
	res := &Position{}
	ep := fmt.Sprintf("/accounts/%v/positions/%v", goq.Auth.Id, instrument)
	goq.Get(ep, nil, res)
	return res
}

// 取引したことのあるポジション情報を取得
// Responseは全期間利益とか癖のあるデータがあるのでstructのコメント見ておくこと
func NewPositions(goq *Goquest) *Positions {
	res := &Positions{}
	ep := fmt.Sprintf("/accounts/%v/positions", goq.Auth.Id)
	goq.Get(ep, nil, res)
	return res
}

// ポジションを持っている通貨の情報を取得。
// Responseのデータは癖があるのでstructのコメント見ておくこと
func NewOpenPositions(goq *Goquest) *Positions {
	res := &Positions{}
	ep := fmt.Sprintf("/accounts/%v/openPositions", goq.Auth.Id)
	goq.Get(ep, nil, res)
	return res
}

// 成行き新規
func NewMarketOrder(goq *Goquest, instrument string, units int) *Orders {
	res := &Orders{}
	ep := fmt.Sprintf("/accounts/%v/orders", goq.Auth.Id)
	param := iMap{}
	marketOrderParam(param, instrument, units, "")
	goq.Post("POST", ep, param, res)
	return res
}

// 成行きクローズ
func NewMarketClose(goq *Goquest, instrument string) {
	ep := fmt.Sprintf("/accounts/%v/positions/%v/close", goq.Auth.Id, instrument)
	goq.Post("PUT", ep, nil, nil)
}
