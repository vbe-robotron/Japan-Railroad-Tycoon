package core

import "fmt"

// このファイルは Phase 8-2 の財務 AI を実装する。
//
// 資金不足なら借入（CreditLine の範囲内）、資金余剰なら返済を行う。
// AI が長期的に路線網を維持・拡大できるよう資金を管理する。

// aiFinance は借金と返済を行う。
func (g *Game) aiFinance(c *Company, ai *AI) []Event {
	var events []Event

	// 資金不足 → 借金
	if c.Cash < MinAIBuildCash && c.Debt < c.CreditLine {
		amount := int64(5_000_000)
		if c.Cash < 0 {
			amount = -c.Cash + 3_000_000
		}
		if c.Debt+amount > c.CreditLine {
			amount = c.CreditLine - c.Debt
		}
		if amount > 0 {
			c.Cash += amount
			c.Debt += amount
			events = append(events, g.Info(EventBondsIssued,
				fmt.Sprintf("%sが資金調達しました（%s）", c.Name, FormatYen(amount))).
				With("amount", amount).
				With("ai", true).
				ForCompany(c.ID))
		}
	}

	// 資金余剰 → 返済
	if c.Cash > 30_000_000 && c.Debt > 0 {
		repay := c.Debt / 2
		if repay > c.Cash-10_000_000 {
			repay = c.Cash - 10_000_000
		}
		if repay > 0 {
			c.Cash -= repay
			c.Debt -= repay
			events = append(events, g.Info(EventDebtRepaid,
				fmt.Sprintf("%sが借金を返済しました（%s）", c.Name, FormatYen(repay))).
				With("amount", repay).
				With("ai", true).
				ForCompany(c.ID))
		}
	}

	return events
}
