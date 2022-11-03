package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	LIVE_URL = "https://api-fxtrade.oanda.com/v3"
	DEMO_URL = "https://api-fxpractice.oanda.com/v3"
)

type (
	Goquest struct {
		Auth   *apiKey
		Client *http.Client
		url    string
	}
)

type strMap map[string]string
type iMap map[string]interface{}

func NewGoquest(fpath string, mode string) *Goquest {
	url := ""
	if mode == "live" {
		url = LIVE_URL
	} else if mode == "demo" {
		url = DEMO_URL
	}
	return &Goquest{
		Auth:   newApiKey(fpath, mode),
		Client: &http.Client{},
		url:    url,
	}
}

// urlパラメタ生成
func (g *Goquest) genUrl(ep string, param strMap) string {
	return g.url + ep + queryStr(param)
}

// 認証authをヘッダにセット
func (g *Goquest) auth(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+g.Auth.Token)
}

// Content-Typeをヘッダにセット
func (g *Goquest) contenType(req *http.Request, mime string) {
	req.Header.Set("Content-Type", mime)
}

// GET 実行。Responseはjsonを想定。responseのbodyをinterfaceにpopulateする。
func (goq *Goquest) Get(ep string, param strMap, i Checker) {
	uri := goq.genUrl(ep, param)
	req, err := http.NewRequest("GET", uri, nil)
	fmt.Println(uri)
	if err != nil {
		fmt.Println(err)
	}

	goq.auth(req)

	res, err := goq.Client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(b, i)
	if err != nil {
		fmt.Println(err)
	}
	i.Status(res.StatusCode)

}

func (goq *Goquest) Post(method string, ep string, param iMap, i Checker) {
	uri := goq.genUrl(ep, nil)
	body, err := json.Marshal(param)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(uri)
	req, err := http.NewRequest(method, uri, strings.NewReader(string(body)))
	if err != nil {
		fmt.Println(err)
	}

	goq.auth(req)
	goq.contenType(req, "application/json")

	// add content-type
	res, err := goq.Client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer res.Body.Close()

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(resBody, i)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res.Status)
	i.Status(res.StatusCode)
}

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

// time.FormatでYYYY-mm-ddTHH:MM:SS.000000000Z形式にするlayout
func layout() string {
	return "2006-01-02T15:04:05.000000000Z"
}

func prettyPrint(i interface{}) {
	b, _ := json.MarshalIndent(i, "", "  ")
	fmt.Println(string(b))
}

func main() {
	goq := NewGoquest("./key.json", "live")
	_ = goq
	res := NewMarketOrder(goq, "USD_JPY", 10000)
	fmt.Println(res)
	// res := NewPosition(goq, "EUR_USD")
	// fmt.Printf("%+v", res)
	// prettyPrint(res)
	// now := time.Now().Add(-time.Hour * 9)

	// delta := time.Second * 60 * 5
	// diff := now.Add(-delta)
	// fmtDiff := diff.Format(layout())
	// fmtNow := now.Format(layout())
	// _, _ = fmtDiff, fmtNow
	// fmt.Println(fmtDiff, fmtNow)
	// to := time.Now().Unix() - 300
	// from := to - 300
	// _ = from
	// res := NewCandles(goq, 0, "M1", "USD_JPY", fmt.Sprintf("%v", from), fmt.Sprintf("%v", to))
	// fmt.Printf("%v", res)
	// fmt.Println(res.CandleData[len(res.CandleData)-1].Mid)
	// t := fmt.Sprintf("%v", time.Now().Unix()-300)
	// fmt.Println(t)
	// res := NewPositionBook(goq, "USD_JPY", "")
	// fmt.Printf("%#v\n", res.Book.UnixTime)
}
