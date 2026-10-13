# Japan Railroad Tycoon

1880〜2020年の日本鉄道史を舞台にした、私鉄版 HOI4 ともいえる鉄道会社経営シミュレーションゲーム。
プレイヤーは鉄道会社を選び、路線を敷設し、街を開発しながら、2020年に日本一の営業利益を目指す。

## 技術スタック

- Go 1.22+
- net/http（標準ライブラリ）
- database/sql + modernc.org/sqlite（cgo 不要）
- github.com/coder/websocket
- github.com/matoous/go-nanoid/v2
- github.com/chzyer/readline
- github.com/fatih/color
- github.com/olekukonko/tablewriter

## セットアップ

```sh
git clone <this repo>
cd Japan-Railroad-Tycoon
go mod download
go build ./...
```

## 実行

```sh
# サーバー起動（既定ポート 8080）
make run-server
curl localhost:8080/health   # => {"ok":true}

# CLI 起動
make run-cli
```

## テスト / Lint

```sh
make test
make lint
```

## ディレクトリ構造

```
jrt/
├── go.mod
├── go.sum
├── Makefile
├── .gitignore
├── README.md
├── PROGRESS.md
├── cmd/
│   ├── server/main.go     # サーバー
│   └── cli/main.go        # CLI クライアント
├── internal/
│   ├── core/doc.go        # 純粋ロジック（HTTP/SQL 非依存）
│   ├── store/doc.go       # SQL 永続化
│   ├── hub/doc.go         # ゲーム管理 + WebSocket
│   ├── api/doc.go         # HTTP ハンドラ + api.Client
│   └── data/doc.go        # データローダー
└── data/
    ├── active.json
    └── mods/
        └── vanilla/
            └── manifest.json
```

## アーキテクチャ

```
cmd → api → hub → core
         ↘ store ↗
```

**core は誰にも依存しない。** ゲームロジックは `internal/core` に閉じ込め、
乱数は決定論のため `g.rng` 経由のみで使用する。

## ゲーム仕様

詳細は [docs/GAME_SPEC.md](docs/GAME_SPEC.md) および [docs/TECHNICAL/ARCHITECTURE.md](docs/TECHNICAL/ARCHITECTURE.md) を参照。

## コミット規約

Conventional Commits（`feat:`, `fix:`, `test:`, `docs:`, `refactor:`）。
