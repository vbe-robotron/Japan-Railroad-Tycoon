package api

import (
	"context"

	"github.com/you/jrt/internal/core"
)

// EndTurn はターンを終了する。
func (g *GameClient) EndTurn(ctx context.Context) ([]core.Event, error) {
	return g.action(ctx, "end_turn", nil)
}

// Action は任意の core.Action を実行する。
func (g *GameClient) Action(ctx context.Context, a core.Action) ([]core.Event, error) {
	if a == nil {
		return nil, &Error{Code: "INVALID_ACTION", Message: "アクションが nil です"}
	}
	return g.action(ctx, a.Type(), a)
}
