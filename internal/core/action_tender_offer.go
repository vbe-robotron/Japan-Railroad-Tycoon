package core

import "fmt"

// このファイルは TenderOfferAction の本実装を提供する（Phase 1-8-5）。
//
// TOB（公開買付け）は対象会社の株をプレミアム付きで買い集め、目標比率まで
// 保有を引き上げる。Phase 1-8-5 では数ターンかけた成立判定は行わず、
// Apply と同時に即時成立する簡略版とする。
//
// 防衛策（DefendTakeoverAction）との対抗は Phase 8（AI）で本格実装する。
//
// プレミアム（Premium）は買付価格の上乗せ倍率で、1.05〜1.5 の範囲。
// 目標比率（TargetRatio）は発行済株に対する取得目標で、0.51〜0.67 の範囲。
// 取得後、議決権比率が 50% を超えれば子会社化する。

// TOB のパラメータ範囲。
const (
	MinTOBPremium     = 1.05
	MaxTOBPremium     = 1.5
	MinTOBTargetRatio = 0.51
	MaxTOBTargetRatio = 0.67
)

// Validate は TenderOfferAction が実行可能かどうかを検証する。
//
// 検証内容:
//   - ゲームが存在する（GAME_NIL）
//   - プレイヤー会社が存在する（NO_PLAYER）
//   - 自社でない（SELF_TARGET）
//   - 対象会社が存在する（TARGET_NOT_FOUND）
//   - 国鉄でない（TARGET_NATIONAL）
//   - プレミアムが範囲内（INVALID_PREMIUM）
//   - 目標比率が範囲内（INVALID_TARGET_RATIO）
//   - 資金が必要コスト以上（INSUFFICIENT_FUNDS）
//
// Validate は状態を変更しない。
func (a *TenderOfferAction) Validate(g *Game) error {
	if g == nil {
		return NewValidationError("GAME_NIL", "ゲームが初期化されていません", nil)
	}
	c := g.actorCompany(a.ActorCompanyID)
	if c == nil {
		return NewValidationError("NO_PLAYER", "プレイヤー会社が存在しません", ErrNoPlayer)
	}
	if a.TargetCompanyID == c.ID {
		return NewValidationError("SELF_TARGET", "自社へのTOBはできません", ErrInvalidArgument)
	}
	target, ok := g.Companies[a.TargetCompanyID]
	if !ok {
		return NewValidationError("TARGET_NOT_FOUND",
			fmt.Sprintf("対象会社が見つかりません: %s", a.TargetCompanyID), ErrNotFound)
	}
	if target.IsNational {
		return NewValidationError("TARGET_NATIONAL", "国鉄は取得できません", ErrUnauthorized)
	}
	if a.Premium < MinTOBPremium || a.Premium > MaxTOBPremium {
		return NewValidationError("INVALID_PREMIUM",
			fmt.Sprintf("プレミアムは %.2f〜%.2f の範囲です: %.2f",
				MinTOBPremium, MaxTOBPremium, a.Premium), ErrInvalidArgument)
	}
	if a.TargetRatio < MinTOBTargetRatio || a.TargetRatio > MaxTOBTargetRatio {
		return NewValidationError("INVALID_TARGET_RATIO",
			fmt.Sprintf("目標比率は %.2f〜%.2f の範囲です: %.2f",
				MinTOBTargetRatio, MaxTOBTargetRatio, a.TargetRatio), ErrInvalidArgument)
	}

	needCount := tenderOfferNeed(c, target, a.TargetRatio)
	cost := tenderOfferCost(needCount, target, a.Premium)
	if c.Cash < cost {
		return NewValidationError("INSUFFICIENT_FUNDS",
			fmt.Sprintf("資金不足（必要: %s / 所持: %s）",
				FormatYen(cost), FormatYen(c.Cash)), ErrInsufficientFunds)
	}
	return nil
}

// Apply は TenderOfferAction を実行する。
//
// 処理:
//  1. Validate と同じ検証
//  2. 取得コストを資金から差し引く
//  3. Holding[targetID] を目標比率まで加算
//  4. 議決権比率が 50% 超なら子会社化（ParentID 設定）
//  5. tob_launched / tob_success イベント（子会社化時は subsidiary_established も）を生成
func (a *TenderOfferAction) Apply(g *Game) ([]Event, error) {
	if err := a.Validate(g); err != nil {
		return nil, err
	}
	c := g.actorCompany(a.ActorCompanyID)
	target := g.Companies[a.TargetCompanyID]

	needCount := tenderOfferNeed(c, target, a.TargetRatio)
	cost := tenderOfferCost(needCount, target, a.Premium)

	c.Cash -= cost
	if c.Holding == nil {
		c.Holding = map[string]int{}
	}
	c.Holding[target.ID] += needCount

	holdingPct := float64(c.Holding[target.ID]) / float64(target.TotalShares)

	events := []Event{
		g.Info(EventTOBLaunched,
			fmt.Sprintf("%sへのTOBを開始しました（プレミアム %.2f、目標 %.1f%%）",
				target.Name, a.Premium, a.TargetRatio*100)).
			With("target_company_id", target.ID).
			With("target_company_name", target.Name).
			With("premium", a.Premium).
			With("target_ratio", a.TargetRatio).
			ForCompany(c.ID),
		g.Success(EventTOBSuccess,
			fmt.Sprintf("%sのTOBが成功しました（保有 %.1f%%）",
				target.Name, holdingPct*100)).
			With("target_company_id", target.ID).
			With("target_company_name", target.Name).
			With("acquired_ratio", holdingPct).
			With("total_cost", cost).
			ForCompany(c.ID),
	}

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

// tenderOfferNeed は目標比率を満たすために追加取得すべき株数を返す。
//
// 既に保有済みの株数を差し引いた不足分を返す。不足がなければ 0。
func tenderOfferNeed(c, target *Company, targetRatio float64) int {
	need := int(float64(target.TotalShares)*targetRatio) - c.Holding[target.ID]
	if need < 0 {
		return 0
	}
	return need
}

// tenderOfferCost は TOB の買付コストを返す。
//
// コスト = 必要株数 × 株価 × プレミアム。
func tenderOfferCost(needCount int, target *Company, premium float64) int64 {
	return int64(float64(needCount*target.StockPrice) * premium)
}
