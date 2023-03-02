package main

/*
NewTwitter -- constructor
(*twitter)tweet -- tweets text
(*twitter)tweetImage --tweets text with Image
*/

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/ChimeraCoder/anaconda"
)

type (
	keys struct {
		API_KEY       string `json:"API_KEY"`
		API_SECRET    string `json:"API_SECRET"`
		ACCESS_TOKEN  string `json:"ACCESS_TOKEN"`
		ACCESS_SECRET string `json:"ACCESS_SECRET"`
	}

	twitter struct {
		api *anaconda.TwitterApi
	}
)

func NewTwitter(confPath string) *twitter {
	t := &twitter{}
	t.setAPIkeys(confPath)
	return t
}

func (t *twitter) setAPIkeys(confPath string) {
	b, err := os.ReadFile(confPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	_keys := &keys{}
	json.Unmarshal(b, _keys)
	anaconda.SetConsumerKey(_keys.API_KEY)
	anaconda.SetConsumerSecret(_keys.API_SECRET)
	t.api = anaconda.NewTwitterApi(_keys.ACCESS_TOKEN, _keys.ACCESS_SECRET)
}

func (t *twitter) tweet(text string, v url.Values) {
	_, err := t.api.PostTweet(text, v)
	if err != nil {
		fmt.Println(err)
	}
}

func (t *twitter) tweetImage(text string, imgPath ...string) {
	ids := ""
	for _, p := range imgPath {
		bs := imgBase64(p)
		media, err := t.api.UploadMedia(bs)
		if err != nil {
			fmt.Println(err)
			return
		}
		if len(ids) == 0 {
			ids = media.MediaIDString
		} else {
			ids += "," + media.MediaIDString
		}
	}
	v := url.Values{}
	v.Add("media_ids", ids)
	t.tweet(text, v)
}

func imgBase64(imgPath string) string {
	b, err := os.ReadFile(imgPath)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	bstr := base64.StdEncoding.EncodeToString(b)
	return bstr
}
