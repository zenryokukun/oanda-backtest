/*
 * Oanda API Responseの型を定める
 */
package main

const (
	// 想定外の計算値
	CalcError = -1
	// StructのFieldがnil
	MissingError = -2
	// 配列が空
	EmptyError = -3
	// 一致無し
	NoMatch = -4
)

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

	// CandleDataの MId or Ask or BidとTimeをマージしたもの。
	CandleStick struct {
		Complete bool
		Time     string
		Prices   *Hloc
	}

	CandleSticks []CandleStick

	CandleData struct {
		Complete bool `json:"complete"`
		// ******************************
		// request パラメタで
		// price:"B" -> Bidが埋まる
		// price:"A" -> Askが埋まる
		// price:"M" -> Midが埋まる(default)
		Mid *Hloc `json:"mid"`
		Ask *Hloc `json:"ask"`
		Bid *Hloc `json:"bid"`
		// ******************************
		Time   string `json:"time"`
		Volume int32  `json:"volume"`
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

	PositionDataSide struct {
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

	PositionData struct {
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
		Long  PositionDataSide `json:"long"`
		Short PositionDataSide `json:"short"`
	}

	Positions struct {
		base
		Positions []PositionData `json:"positions"`
		LastID    string         `json:"lastTransactionID"`
	}

	Position struct {
		base
		Position *PositionData `json:"position"`
		LastID   string        `json:"lastTransactionID"`
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

	Order struct {
		Id          string `json:"id"`
		CreatedTime string `json:"createdTime"`
		// PENDING,FILLED,TRIGGERED,CANCELLED
		State string `json:"state"`
	}

	AccountData struct {
		// 証拠金維持率と思われる
		MarginRate        float64 `json:"marginRate,string"`
		MarginUsed        float64 `json:"marginUsed,string"`
		OpenTradeCount    int     `json:"openTradeCount"`
		OpenPositionCount int     `json:"openPositionCount"`
		PendingOrderCount int     `json:"pendingOrderCount"`
		// 総未実現利益
		UnrealizedPL float64 `json:"unrealizedPl,string"`
		// 総利益
		PL float64 `json:"pl,string"`
		// 残高
		Balance    float64        `json:"balance,string"`
		Commission float64        `json:"commission,string"`
		Positions  []PositionData `json:"positions"`
		Orders     []Order        `json:"orders"`
	}

	Account struct {
		base
		Data   AccountData `json:"account"`
		LastID string      `json:"lastTransactionID"`
	}

	Ticker struct {
		Price     float64 `json:"price,string"`
		Liquidity int64   `json:"liquidity"`
	}

	Price struct {
		Time       string   `json:"time"`
		Instrument string   `json:"instrument"`
		Bids       []Ticker `json:"bids"`
		Asks       []Ticker `json:"asks"`
	}

	Pricing struct {
		base
		Time string `json:"time"`
		// "USD_JPY,EUR_USD"のように複数通貨指定できる。
		// 通貨単位でPriceが埋まる。
		Prices []Price `json:"prices"`
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

func (c *Candles) ExtractMid() CandleSticks {
	data := c.Extract()
	if data == nil || len(data) == 0 {
		return nil
	}
	sticks := []CandleStick{}
	for _, d := range data {
		if d.Mid == nil {
			continue
		}
		stick := CandleStick{
			Complete: d.Complete,
			Prices:   d.Mid,
			Time:     d.Time,
		}
		sticks = append(sticks, stick)
	}
	return sticks
}

func (s CandleSticks) Extract(hloc string) []float64 {
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

// Spread計算。ask-bid。片道の取引は(ask-bid)/2になるはずだが、考慮しない。
func (p *Price) Spread() float64 {
	if len(p.Asks) == 0 || len(p.Bids) == 0 {
		return EmptyError
	}
	lastAsk := p.Asks[len(p.Asks)-1]
	lastBid := p.Bids[len(p.Bids)-1]
	return lastAsk.Price - lastBid.Price
}

// Pricingは複数の通貨の指定が出来きる。instrumentでfilterする。
// マッチしない場合はnilを返す
func (p *Pricing) filter(instrument string) *Price {
	for _, p := range p.Prices {
		if p.Instrument == instrument {
			return &p
		}
	}
	return nil
}

// Pricingからspreadを計算して返す
// 計算できないときはMissingErrorかEmptyErrorを返す
func (p *Pricing) Spread(instrument string) float64 {
	if !p.Check() {
		return MissingError
	}
	if len(p.Prices) == 0 {
		return EmptyError
	}
	price := p.filter(instrument)
	if price == nil {
		return NoMatch
	}
	return price.Spread()
}
