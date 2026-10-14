package api

import (
	"context"

	"github.com/you/jrt/internal/core"
)

// このファイルは財務系アクションの型付きメソッドを提供する。

// IssueBonds は社債を発行する。
func (g *GameClient) IssueBonds(ctx context.Context, amount int64, rate float64, years int) ([]core.Event, error) {
	return g.action(ctx, "issue_bonds", core.IssueBondsAction{
		Amount: amount, Rate: rate, Years: years,
	})
}

// IssueShares は増資する。
func (g *GameClient) IssueShares(ctx context.Context, count int) ([]core.Event, error) {
	return g.action(ctx, "issue_shares", core.IssueSharesAction{Count: count})
}

// PayDividend は配当を支払う。
func (g *GameClient) PayDividend(ctx context.Context, perShare int) ([]core.Event, error) {
	return g.action(ctx, "pay_dividend", core.PayDividendAction{PerShare: perShare})
}

// RepayDebt は借金を返済する。
func (g *GameClient) RepayDebt(ctx context.Context, amount int64) ([]core.Event, error) {
	return g.action(ctx, "repay_debt", core.RepayDebtAction{Amount: amount})
}
