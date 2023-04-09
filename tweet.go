package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/zenryokukun/gotweet"
)

func tweet(msg string, paths ...string) {
	// 実行ファイルと同じフォルダを取得
	exe, err := os.Executable()
	if err != nil {
		fmt.Println(err)
	}
	root := filepath.Dir(exe)
	// さらに一つ上のフォルダを取得
	parent := filepath.Dir(root)
	// twitter confファイル
	conf := filepath.Join(parent, "oanda-bot", "twitter.json")

	// twitter初期化
	twitter := gotweet.NewTwitter(conf)
	twitter.Tweet(msg, paths...)
}
