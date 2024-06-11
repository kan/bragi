# Bragi

goによるSKKサーバーの実装です。最終的に以下の機能を備えることを目標にしています。

- OpenAI APIを使った変換候補の作成
- 各種SKK辞書の読み込み
- URLで指定したSKK辞書の自動取り込み&定期更新
- 上記の変換をモジュール化してON/OFF可能に

## 使い方

現状、dockerを用いて動作させることを想定しています。docker-compose.ymlがあるので、以下の手順で起動できます。

```
docker-compose up -d
docker-compose exec golang bash

# 以下、コンテナ内
$ export BRAGI_OPENAI_API_KEY=xxxxx # OpenAI API Keyの指定
$ go run main.go # 1234ポートでSKKサーバーが起動
```

