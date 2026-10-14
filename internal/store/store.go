package store

import (
	"context"

	"github.com/you/jrt/internal/core"
)

// Store はゲーム状態の永続化を抽象化するインターフェース。
//
// 実装は SQLiteStore のほか、テスト用のインメモリ実装や将来の
// クラウド保存などを想定している。core 側はこのインターフェースに
// 依存しない（純粋ロジックは core に閉じる）。
type Store interface {
	// SaveGame はゲーム状態を保存する。既存の同一 ID は上書きする。
	SaveGame(ctx context.Context, g *core.Game) error

	// LoadGame は指定した ID のゲームを復元する。
	LoadGame(ctx context.Context, id string) (*core.Game, error)

	// ListGames は保存済みゲームの ID を更新日時の新しい順に返す。
	ListGames(ctx context.Context) ([]string, error)

	// DeleteGame は指定した ID のゲームと関連データを削除する。
	DeleteGame(ctx context.Context, id string) error

	// AppendEvents はイベント列を追記する。
	AppendEvents(ctx context.Context, gameID string, events []core.Event) error

	// AppendAction はアクションログを追記する。
	AppendAction(ctx context.Context, gameID, companyID, actionType string, payload []byte) error

	// Close は内部リソースを解放する。
	Close() error
}
