/*
 * Oanda API Responseの型を定める
 */
package main

type (
	base struct {
		statusCode int
	}

	// high,low,open. and close prices
	Hloc struct {
		H float64 `json:"h,string"`
		L float64 `json:"l,string"`
		C float64 `json:"c,string"`
		O float64 `json:"o,string"`
	}

	CandleData struct {
		Complete bool   `json:"complete"`
		Mid      *Hloc  `json:"mid"`
		Time     string `json:"time"`
		Volume   int32  `json:"volume"`
	}

	// Granularity: "M" "W" "D" "H4" "H1" "M15" "M5" "M1" 等
	// Instrument: "USD_JPY" "EUR_USD" 等
	Candles struct {
		base
		CandleData  []CandleData `json:"candles"`
		Granularity string       `json:"granularity"`
		Instrument  string       `json:"instrument"`
	}

	buckets struct {
		Price        float64 `json:"price,string"`
		LongPercent  float64 `json:"longCountPercent,string"`
		ShortPercent float64 `json:"shortCountPercent,string"`
	}

	Book struct {
		Price       float64   `json:"price,string"`
		BucketWidth float64   `json:"bucketWidth,string"`
		Buckets     []buckets `json:"buckets"`
		UnixTime    int64     `json:"unixTime,string"`
	}

	PositionBook struct {
		base
		Book *Book `json:"positionBook"`
	}

	OrderBook struct {
		base
		Book *Book `json:"orderBook"`
	}

	positionSide struct {
		// 保有中のポジションの枚数。過去の取引量ではない。
		Units int `json:"units,string"`
		// 保有中ポジションの平均取得価格
		Average float64 `json:"averagePrice,string"`
		// 保有中ポジションの取引id
		TradeIDs []string `json:"tradeIDs"`
		// これは何故か全期間損益。保有中のものでな無いので注意
		PL float64 `json:"pl,string"`
		// 未実現損益
		UnrealizedPL float64 `json:"unrealizedPl,string"`
	}

	position struct {
		Instrument string `json:"instrument"`
		// アカウントの累計損益なので注意
		PL float64 `json:"pl,string"`
		// 未実現損益。保有中のポジションの未実現損益
		UnrealizedPL float64 `json:"unrealizedPl,string"`
		// 謎。手数料？でも累計でも0になってる。
		Margin float64 `json:"marginUsed:string"`
		// 謎。手数料？でも累計でも0になってる。
		Commission float64 `json:"commission,string"`

		// position別の累計情報
		Long  positionSide `json:"long"`
		Short positionSide `json:"short"`
	}

	Positions struct {
		base
		Positions []position `json:"positions"`
		LastID    string     `json:"lastTransactionID"`
	}

	Position struct {
		base
		Position *position `json:"position"`
		LastID   string    `json:"lastTransactionID"`
	}

	transaction struct {
		ID        string `json:"id"`
		Time      string `json:"time"`
		BatchID   string `json:"batchID"`
		RequestID string `json:"requestID"`
	}

	fillTransaction struct {
		transaction
		Type           string  `json:"type"`
		OrderID        string  `json:"orderID"`
		Instrument     string  `json:"instrument"`
		Units          int     `json:"units,string"`
		Reason         string  `json:"reason"`
		PL             float64 `json:"pl,string"`
		Commission     float64 `json:"commission,string"`
		AccountBalance float64 `json:"accountBalance,string"`
	}

	cancelTransaction struct {
		transaction
		Type    string `json:"type"`
		OrderID string `json:"orderID"`
		Reason  string `json:"reason"`
	}

	Orders struct {
		base
		Transaction       transaction       `json:"orderCreateTransaction"`
		FillTransaction   fillTransaction   `json:"orderFillTransaction"`
		CancelTransaction cancelTransaction `json:"orderCancelTransaction"`
		LastID            string            `json:"lastTransactionID"`
	}

	CloseOrders struct {
		base
		LongFillTransaction    fillTransaction   `json:"longOrderFillTransaction"`
		ShortFillTransaction   fillTransaction   `json:"shortOrderFillTransaction"`
		LongCancelTransaction  cancelTransaction `json:"longOrderCancelTransaction "`
		ShortCancelTransaction cancelTransaction `json:"shortOrderCancelTransaction"`
		LastID                 string            `json:"lastTransactionID"`
	}
)

type Checker interface {
	Check() bool
	Status(int)
}

func (b *base) Status(code int) {
	b.statusCode = code
}

func (b *base) Check() bool {
	return b.statusCode >= 200 && b.statusCode <= 299
}

func (c *Candles) Extract() []CandleData {
	if !c.Check() {
		return nil
	}
	return c.CandleData
}
