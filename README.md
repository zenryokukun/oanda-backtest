# Oanda-botのバックテストモジュール

## いつ使いますか？

バックテスト、もしくはリアル取引とバックテストのコンペアを実施したいときに使いたい時。

## どう使いますか？

コマンドライン引数によってモードが分かれます。

- 指定なし
- test-new
- compare

上記以外の引数が渡された場合、処理は行われません。第二移行の引数は無視します。

### バックテスト時

2つのテストモードがあります。

#### 現在から過去5000個分の5分足でテスト

**test-new**をコマンドライン引数に指定して実行します。

```bash
# compile済の場合
./oanda-backtest test-new
# compileせずに実行する場合
go run . test-new
```

#### 指定した./testdata.jsonファイルでテスト

あらかじめ準備した、./testdata.jsonでテストも可能です。本ファイルは[]CandleStick型にparse可能である必要があります。

コマンドライン引数は**指定しません**

```bash
# compile済の場合
./oanda-backtest
# compileせずに実行する場合
./oanda-backtest
```

### コンペア時

バックテストとリアル取引のコンペアを行います。バックテストは**test-new**で
実行した時と同じです。リアル取引は../oanda-botの取引データを利用します。

コマンドライン引数に**compare**を指定して実行します。

```bash
# compile済の場合
./oanda-backtest compare
# compileせずに実行する場合
./oanda-backtest compare
```

## 必要なファイル

- ../oanda-bot/key.json  
oanda-botのconfigファイル

- ../oanda-bot/trade.json  
oanda-botの取引データ

- ../oanda-bot/balance.json  
oanda-botの残高データ

- ../oanda-bot/twitter.json  
oanda-botのツイッター情報

- (optional) ./testdata.json  
テストするロウソクデータ。特定の期間でテストしたい場合のみ手動配置。
[]CandleStick型にparse可能である必要がある。

## 必要なモジュール

- ../oanda-eval
リカバリファクタ等の指標を計算するモジュール。compareモードで利用する時は必要。

## 吐き出すファイル

### いずれもバックテスト結果ファイル。graph.pyでグラフをするときに使われる。
- ./bal.json
- ./pos.json
- ./result.png （compareモードのみ）