package core

import "fmt"

// このファイルは MergeAction の本実装を提供する（Phase 1-8-5）。
//
// 合併（Merge）は対象会社と統合し、存続会社の社名を新社名に変更する。
// Phase 1-8-5 では即時実行とし、対象会社の路線・施設を存続会社へ引き継ぎ、
// 対象会社は解散（Dissolved）として統廃合グラフに記録する。
//
// 合併比率（Ratio）は対象会社の価値に対する買収コストの割合で 0.0〜1.0。
// 統合は新会社ノードを作る有向グラフ表現を将来導入するが、Phase 1-8-5 では
// 存続会社（プレイヤー）の ParentID と対象会社の Dissolved/ParentID で
// 簡易表現する。

// Validate は MergeAction が実行可能かどうかを検証する。
//
// 検証内容:
//   - ゲームが存在する（GAME_NIL）
//   - プレイヤー会社が存在する（NO_PLAYER）
//   - 自社でない（SELF_MERGE）
//   - 対象会社が存在する（TARGET_NOT_FOUND）
//   - 国鉄でない（TARGET_NATIONAL）
//   - 合併比率が範囲内（INVALID_RATIO）
//   - 新社名が指定されている（INVALID_NAME）
//   - 資金が合併コスト以上（INSUFFICIENT_FUNDS）
//
// Validate は状態を変更しない。
func (a *MergeAction) Validate(g *Game) error {
	if g == nil {
		return NewValidationError("GAME_NIL", "ゲームが初期化されていません", nil)
	}
	c := g.actorCompany(a.ActorCompanyID)
	if c == nil {
		return NewValidationError("NO_PLAYER", "プレイヤー会社が存在しません", ErrNoPlayer)
	}
	if a.TargetCompanyID == c.ID {
		return NewValidationError("SELF_MERGE", "自社との合併はできません", ErrInvalidArgument)
	}
	target, ok := g.Companies[a.TargetCompanyID]
	if !ok {
		return NewValidationError("TARGET_NOT_FOUND",
			fmt.Sprintf("対象会社が見つかりません: %s", a.TargetCompanyID), ErrNotFound)
	}
	if target.IsNational {
		return NewValidationError("TARGET_NATIONAL", "国鉄とは合併できません", ErrUnauthorized)
	}
	if a.Ratio < 0 || a.Ratio > 1.0 {
		return NewValidationError("INVALID_RATIO",
			fmt.Sprintf("合併比率は 0.0〜1.0 の範囲です: %.2f", a.Ratio), ErrInvalidArgument)
	}
	if a.Name == "" {
		return NewValidationError("INVALID_NAME", "新社名を指定してください", ErrInvalidArgument)
	}

	cost := mergeCost(target, a.Ratio)
	if c.Cash < cost {
		return NewValidationError("INSUFFICIENT_FUNDS",
			fmt.Sprintf("資金不足（必要: %s / 所持: %s）",
				FormatYen(cost), FormatYen(c.Cash)), ErrInsufficientFunds)
	}
	return nil
}

// Apply は MergeAction を実行する。
//
// 処理:
//  1. Validate と同じ検証
//  2. 合併コストを資金から差し引く
//  3. 対象会社の路線・施設を存続会社へ引き継ぐ
//  4. 存続会社の社名を新社名に変更
//  5. 対象会社を解散（Dissolved/ParentID 設定）
//  6. merge_completed / company_merged イベントを生成
func (a *MergeAction) Apply(g *Game) ([]Event, error) {
	if err := a.Validate(g); err != nil {
		return nil, err
	}
	c := g.actorCompany(a.ActorCompanyID)
	target := g.Companies[a.TargetCompanyID]

	cost := mergeCost(target, a.Ratio)
	c.Cash -= cost

	// 路線引継ぎ
	inheritedLines := len(target.Lines)
	for _, line := range target.Lines {
		line.CompanyID = c.ID
		c.Lines = append(c.Lines, line)
	}
	target.Lines = nil

	// 施設引継ぎ
	for _, f := range target.Facilities {
		f.CompanyID = c.ID
		c.Facilities = append(c.Facilities, f)
	}
	target.Facilities = nil

	oldName := c.Name
	c.Name = a.Name
	target.Dissolved = &g.Date
	target.ParentID = c.ID

	events := []Event{
		g.Success(EventMergeCompleted,
			fmt.Sprintf("%sと%sが合併し、%sが発足しました",
				oldName, target.Name, a.Name)).
			With("from_company_id", c.ID).
			With("from_company_name", oldName).
			With("to_company_id", target.ID).
			With("to_company_name", target.Name).
			With("new_company_name", a.Name).
			With("ratio", a.Ratio).
			With("cost", cost).
			ForCompany(c.ID),
		g.Success(EventCompanyMerged,
			fmt.Sprintf("合併により %d 路線を引継ぎました", inheritedLines)).
			With("node_id", c.ID).
			With("predecessors", []string{c.ID, target.ID}).
			With("kind", "merge").
			ForCompany(c.ID),
	}
	return events, nil
}

// mergeCost は合併コストを返す。
//
// コスト = 対象会社の発行済株 × 株価 × 合併比率。
func mergeCost(target *Company, ratio float64) int64 {
	return int64(float64(target.TotalShares*target.StockPrice) * ratio)
}
