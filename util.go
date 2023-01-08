package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// account情報から通貨別総利益を取得
func totalPL(acc *Account) {
	if !acc.Check() {
		return
	}
	if len(acc.Data.Positions) == 0 {
		return
	}

	for _, pos := range acc.Data.Positions {
		fmt.Printf("%v longPL:%v shortPL:%v\n", pos.Instrument, pos.Long.PL, pos.Short.PL)
	}
}

// time.FormatでYYYY-mm-ddTHH:MM:SS.000000000Z形式にするlayout
func layout() string {
	return "2006-01-02T15:04:05.000000000Z"
}

// json形式のデータをインデント付きでprintするヘルパー関数
func prettyPrint(i interface{}) {
	b, _ := json.MarshalIndent(i, "", "  ")
	fmt.Println(string(b))
}

// query string 生成
func queryStr(param strMap) string {
	if param == nil {
		return ""
	}
	query := "?"
	for k, v := range param {
		query += k + "=" + v + "&"
	}
	return query[:len(query)-1]
}

// test dataをファイルに出力
func testFile(goq *Goquest, granularity string, instrument string, count int, from, to string) {
	res := NewCandles(goq, count, granularity, instrument, from, to, "")
	candles := res.Extract()
	if candles == nil {
		return
	}
	writeFile("./testdata.json", candles)
	// b, err := json.MarshalIndent(candles, "", " ")
	// if err != nil {
	// 	fmt.Println(nil)
	// 	return
	// }
	// f, err := os.Create("./testdata.json")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// defer f.Close()
	// f.Write(b)
}

func writeFile(fpath string, i interface{}) {
	b, err := json.MarshalIndent(i, "", " ")
	if err != nil {
		fmt.Println(err)
		return
	}
	f, err := os.Create(fpath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	f.Write(b)
}
