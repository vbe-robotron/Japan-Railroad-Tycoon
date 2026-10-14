package core

import "fmt"

// このファイルは BuySharesAction の本実装を提供する（Phase 1-8-4）。
//
// プレイヤーは他社の株式を取得できる。株取得には 5% のプレミアムが付く。
// 議決権比率（Holding / TotalShares）が 50% を超えると子会社化し、
// 対象会社の ParentID を自社に設定する。国鉄（IsNational）と自社株は
// 取得できない。
//
// 株価（StockPrice）は Phase 1-8-4 では Company のフィールドをそのまま使う。
// 動的な株価計算は後続フェーズで導入する。

// BuySharesPremium は株取得時に上乗せされるプレミアム係数（5%）。
const BuySharesPremium = 1.05 // 5% プレミアム

// Validate は BuySharesAction が実行可能かどうかを検証する。
//
// 検証内容:
//   - ゲームが存在する（GAME_NIL）
//   - プレイヤー会社が存在する（NO_PLAYER）
//   - 株数が正の整数（INVALID_COUNT）
//   - 自社株でない（SELF_PURCHASE）
//   - 対象会社が存在する（TARGET_NOT_FOUND）
//   - 国鉄でない（TARGET_NATIONAL）
//   - 資金が取得コスト以上（INSUFFICIENT_FUNDS）
//
// Validate は状態を変更しない。
func (a *BuySharesAction) Validate(g *Game) error {
	if g == nil {
		return NewValidationError("GAME_NIL", "ゲームが初期化されていません", nil)
	}
	c := g.PlayerCompany()
	if c == nil {
		return NewValidationError("NO_PLAYER", "プレイヤー会社が存在しません", ErrNoPlayer)
	}
	if a.Count <= 0 {
		return NewValidationError("INVALID_COUNT",
			fmt.Sprintf("株数は正の整数である必要があります: %d", a.Count), ErrInvalidArgument)
	}
	if a.TargetCompanyID == c.ID {
		return NewValidationError("SELF_PURCHASE", "自社株は取得できません", ErrInvalidArgument)
	}
	target, ok := g.Companies[a.TargetCompanyID]
	if !ok {
		return NewValidationError("TARGET_NOT_FOUND",
			fmt.Sprintf("対象会社が見つかりません: %s", a.TargetCompanyID), ErrNotFound)
	}
	if target.IsNational {
		return NewValidationError("TARGET_NATIONAL", "国鉄は取得できません", ErrUnauthorized)
	}

	cost := g.buySharesCost(target, a.Count)
	if c.Cash < cost {
		return NewValidationError("INSUFFICIENT_FUNDS",
			fmt.Sprintf("資金不足（必要: %s / 所持: %s）",
				FormatYen(cost), FormatYen(c.Cash)), ErrInsufficientFunds)
	}
	return nil
}

// Apply は BuySharesAction を実行する。
//
// 処理:
//  1. Validate と同じ検証
//  2. 取得コストを資金から差し引く
//  3. Holding[targetID] に取得株数を加算
//  4. 議決権比率が 50% 超なら子会社化（ParentID 設定）
//  5. shares_acquired イベント（子会社化時は subsidiary_established も）を生成
func (a *BuySharesAction) Apply(g *Game) ([]Event, error) {
	if err := a.Validate(g); err != nil {
		return nil, err
	}
	c := g.PlayerCompany()
	target := g.Companies[a.TargetCompanyID]
	cost := g.buySharesCost(target, a.Count)

	c.Cash -= cost
	if c.Holding == nil {
		c.Holding = map[string]int{}
	}
	c.Holding[target.ID] += a.Count

	total := c.Holding[target.ID]
	holdingPct := float64(total) / float64(target.TotalShares)

	events := []Event{
		g.Info(EventSharesAcquired,
			fmt.Sprintf("%sの株を %d株 取得しました（%s、保有 %d株 / %.1f%%）",
				target.Name, a.Count, FormatYen(cost), total, holdingPct*100)).
			With("target_company_id", target.ID).
			With("target_company_name", target.Name).
			With("count", a.Count).
			With("price_per_share", target.StockPrice).
			With("cost", cost).
			With("total_holding", total).
			With("holding_pct", holdingPct).
			ForCompany(c.ID),
	}

	// 50%超で子会社化
	if holdingPct > 0.5 && target.ParentID != c.ID {
		target.ParentID = c.ID
		events = append(events, g.Success(EventSubsidiaryEstablished,
			fmt.Sprintf("%sを子会社化しました（保有 %.1f%%）", target.Name, holdingPct*100)).
			With("parent_company_id", c.ID).
			With("child_company_id", target.ID).
			With("child_company_name", target.Name).
			With("holding_pct", holdingPct).
			ForCompany(c.ID))
	}

	return events, nil
}

// buySharesCost は株取得コストを返す。
//
// コスト = 株数 × 株価 × BuySharesPremium（5% プレミアム）。
func (g *Game) buySharesCost(target *Company, count int) int64 {
	return int64(float64(count*target.StockPrice) * BuySharesPremium)
}
