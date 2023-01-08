/*
 * Oanda APIのidとtokenをファイルから読み取る
 */

package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type (
	apiKey struct {
		Id    string `json:"id"`
		Token string `json:"token"`
	}

	apiKeys struct {
		Live *apiKey `json:"live"`
		Demo *apiKey `json:"demo"`
	}
)

func readConf(fpath string) (*apiKeys, error) {
	b, err := os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	keys := &apiKeys{}
	json.Unmarshal(b, keys)
	return keys, nil
}

func newApiKey(fpath string, mode string) *apiKey {
	conf, err := readConf(fpath)
	if err != nil {
		fmt.Println(err)
	}
	if mode == "live" {
		return conf.Live
	}
	if mode == "demo" {
		return conf.Demo
	}
	return nil
}
