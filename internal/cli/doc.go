// Package cli は対話型の REPL クライアントを提供する。
//
// 責務:
//   - コマンドのパース（parser.go）
//   - HTTP 経由のサーバー操作（handlers.go）
//   - 色付き出力（printer.go）
//   - Tab 補完（completer.go）
//   - WebSocket 購読（events.go）
//
// CLI はロジックを持たない。全判定はサーバー側。
package cli
