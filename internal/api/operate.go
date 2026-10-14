package api

import (
	"context"

	"github.com/you/jrt/internal/core"
)

// このファイルは運用系アクションの型付きメソッドを提供する。

// SetFare は運賃を変更する。
func (g *GameClient) SetFare(ctx context.Context, lineID string, fare int) ([]core.Event, error) {
	return g.action(ctx, "set_fare", core.SetFareAction{LineID: lineID, Fare: fare})
}

// SetSchedule はダイヤ・本数を変更する。
func (g *GameClient) SetSchedule(ctx context.Context, lineID string, trainsPerDay int) ([]core.Event, error) {
	return g.action(ctx, "set_schedule", core.SetScheduleAction{
		LineID: lineID, TrainsPerDay: trainsPerDay,
	})
}

// PurchaseStock は車両を購入する。
func (g *GameClient) PurchaseStock(ctx context.Context, lineID string, count int, kind string) ([]core.Event, error) {
	return g.action(ctx, "purchase_stock", core.PurchaseRollingStockAction{
		LineID: lineID, Count: count, Kind: kind,
	})
}
