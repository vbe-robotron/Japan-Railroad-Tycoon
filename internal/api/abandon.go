package api

import (
	"context"

	"github.com/you/jrt/internal/core"
)

// AbandonLine は路線を廃止する。
func (g *GameClient) AbandonLine(ctx context.Context, lineID, reason string) ([]core.Event, error) {
	return g.action(ctx, "abandon_line", core.AbandonLineAction{LineID: lineID, Reason: reason})
}
