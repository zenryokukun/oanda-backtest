package main

import (
	"encoding/json"
	"fmt"
)

type Eval struct {
	TotalPL        float64
	Count          int
	WinRate        float64
	ProfitPerTrade float64
	DrawDown       float64
	ProfitFactor   float64
	RecoveryFactor float64
}

// evalStrはpythonで計算した評価指標。JSON形式。
func NewEval(evalStr string) Eval {
	e := Eval{}
	json.Unmarshal([]byte(evalStr), &e)
	return e
}

func (e Eval) String() string {
	var msg string
	msg = "総利益:" + fmt.Sprint(e.TotalPL) + "\n"
	msg += "取引数:" + fmt.Sprint(e.Count) + "\n"
	msg += "勝率:" + fmt.Sprint(e.WinRate) + "\n"
	msg += "期待利益:" + fmt.Sprint(e.ProfitPerTrade) + "\n"
	msg += "ドローダウン:" + fmt.Sprint(e.DrawDown) + "\n"
	msg += "リカバリ・ファクタ:" + fmt.Sprint(e.RecoveryFactor) + "\n"
	return msg
}

// "TotalPL": totalPL(data),
// "Count": trade_count(data),
// "WinRate": win_rate(data),
// "ProfitPerTrade": profit_per_trade(data),
// "LongInfo": long_info(data),
// "ShortInfo": short_info(data),
// "DrawDown": drawdown(data),
// "ProfitFactor": profit_factor(data),
// "RecoveryFactor": recovery_factor(data)
