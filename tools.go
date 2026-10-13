//go:build tools

// tools.go は、まだコードから import されていない依存パッケージを
// go.mod / go.sum に固定するためのファイルである。
// ビルドタグ tools により通常のビルド・vet からは除外される。
// 各パッケージを実際に利用するようになったら、このファイルは削除してよい。
package tools

import (
	_ "github.com/chzyer/readline"
	_ "github.com/coder/websocket"
	_ "github.com/fatih/color"
	_ "github.com/matoous/go-nanoid/v2"
	_ "github.com/olekukonko/tablewriter"
	_ "modernc.org/sqlite"
)
