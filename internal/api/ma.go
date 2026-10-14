package api

import (
	"context"

	"github.com/you/jrt/internal/core"
)

// このファイルは M&A 系アクションの型付きメソッドを提供する。

// BuyShares は他社株を取得する。
func (g *GameClient) BuyShares(ctx context.Context, targetCompanyID string, count int) ([]core.Event, error) {
	return g.action(ctx, "buy_shares", core.BuySharesAction{
		TargetCompanyID: targetCompanyID, Count: count,
	})
}

// TenderOffer は TOB を仕掛ける。
func (g *GameClient) TenderOffer(ctx context.Context, targetCompanyID string, premium, targetRatio float64) ([]core.Event, error) {
	return g.action(ctx, "tender_offer", core.TenderOfferAction{
		TargetCompanyID: targetCompanyID, Premium: premium, TargetRatio: targetRatio,
	})
}

// Merge は他社と合併する。
func (g *GameClient) Merge(ctx context.Context, targetCompanyID, name string, ratio float64) ([]core.Event, error) {
	return g.action(ctx, "merge", core.MergeAction{
		TargetCompanyID: targetCompanyID, Name: name, Ratio: ratio,
	})
}
