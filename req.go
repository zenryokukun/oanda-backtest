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
// priceComponent -> "M"(default):中央値？ "A":ask "B":bid
func candlesParam(
	p strMap,
	count int,
	granularity, instruments, from, to string,
	priceComponent string,
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

	if priceComponent == "A" || priceComponent == "B" {
		p["price"] = priceComponent
	} else {
		p["price"] = "M"
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

// 成行きのcloseパラメタ
func marketCloseParam(p iMap, longUnits, shortUnits int) {
	if longUnits > 0 {
		p["longUnits"] = strconv.Itoa(longUnits)
	}
	if shortUnits > 0 {
		p["shortUnits"] = strconv.Itoa(shortUnits)
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

// count -> ロウソク足何個とるか
// from,to両方指定した場合、countの指定は出来ないので、0以下の数値を渡すこと。
// from,to は　"YYYY-mm-ddTHH:MM:SS.000000000Z" もしくは unix時間を文字列にしたもの(fmt.Sprintf("%v",time.Now().Unix())とか)
// priceComponent -> "M"(default):中央値？ "A":ask "B":bid
func NewCandles(
	goq *Goquest,
	count int,
	granularity, instruments, from, to string,
	priceComponent string,
) *Candles {
	res := &Candles{}
	ep := fmt.Sprintf("/instruments/%v/candles", instruments)
	param := strMap{}
	candlesParam(param, count, granularity, instruments, from, to, priceComponent)
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
	goq.Post(ep, param, res)
	return res
}

// 成行きクローズ
// long:クローズするlongポジションunit、short:クローズするshortポジション
// 決済しないほうのポジションには 0 を指定
func NewMarketClose(goq *Goquest, instrument string, longUnits, shortUnits int) *CloseOrders {
	res := &CloseOrders{}
	param := iMap{}
	marketCloseParam(param, longUnits, shortUnits)
	ep := fmt.Sprintf("/accounts/%v/positions/%v/close", goq.Auth.Id, instrument)
	goq.Put(ep, param, res)
	return res
}

// 口座嬢王
func NewAccount(goq *Goquest) *Account {
	res := &Account{}
	ep := "/accounts/" + goq.Auth.Id
	goq.Get(ep, nil, res)
	return res
}

// instruments:"USD_JPY,EUR_USD"のように複数指定可能
func NewPricing(goq *Goquest, instruments string) *Pricing {
	res := &Pricing{}
	ep := "/accounts/" + goq.Auth.Id + "/pricing"
	p := map[string]string{
		"instruments": instruments,
	}
	goq.Get(ep, p, res)
	return res
}
