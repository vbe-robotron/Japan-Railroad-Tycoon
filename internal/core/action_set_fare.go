package core

import (
	"fmt"
	"math"
)

// このファイルは SetFareAction の本実装を提供する（Phase 1-8-1）。
//
// プレイヤーは自社路線の運賃を変更できる。これにより価格戦略
// （独占路線は値上げ、競合路線は値下げ、赤字路線は値上げ）が可能になる。
//
// Phase 1-8-1 では運賃弾力性（FareElasticity = -0.7）は Game.lineScore の
// 100/Fare を通じて市場シェアに間接的に反映される（需要そのものには
// 直接作用しない）。運賃を下げるとスコアが上がりシェアが増える。

// Validate は SetFareAction が実行可能かどうかを検証する。
//
// 検証内容:
//   - ゲームが存在する
//   - プレイヤー会社が存在する
//   - 路線が存在し、プレイヤー会社が所有している
//   - 運賃が MinFare〜MaxFare の範囲内（1〜1000円）
//   - 現在の運賃の ±MaxFareChangeRatio（±50%）以内の変更
//
// Validate は状態を変更しない。
func (a *SetFareAction) Validate(g *Game) error {
	if g == nil {
		return NewValidationError("GAME_NIL", "ゲームが初期化されていません", nil)
	}

	c := g.PlayerCompany()
	if c == nil {
		return NewValidationError("NO_PLAYER", "プレイヤー会社が存在しません", ErrNoPlayer)
	}

	// 自社路線を検索
	line := c.FindLine(a.LineID)
	if line == nil {
		// 他社の路線か確認（所有権エラーを路線なしと区別する）
		for _, other := range g.Companies {
			if other.ID == c.ID {
				continue
			}
			if other.FindLine(a.LineID) != nil {
				return NewValidationError("NOT_OWNED",
					"他社の路線の運賃は変更できません", ErrUnauthorized)
			}
		}
		return NewValidationError("LINE_NOT_FOUND",
			fmt.Sprintf("路線が見つかりません: %s", a.LineID), ErrNotFound)
	}

	// 運賃の範囲チェック
	if a.Fare < MinFare {
		return NewValidationError("INVALID_FARE",
			fmt.Sprintf("運賃は %d円 以上である必要があります", MinFare), ErrInvalidArgument)
	}
	if a.Fare > MaxFare {
		return NewValidationError("INVALID_FARE",
			fmt.Sprintf("運賃は %d円 以下である必要があります", MaxFare), ErrInvalidArgument)
	}

	// 変更幅チェック（±50%）
	if line.Fare > 0 {
		changeRatio := math.Abs(float64(a.Fare-line.Fare)) / float64(line.Fare)
		if changeRatio > MaxFareChangeRatio {
			return NewValidationError("FARE_CHANGE_TOO_LARGE",
				fmt.Sprintf("運賃の変更は現在の %.0f%% 以内にしてください（現在: %d円 → 指定: %d円）",
					MaxFareChangeRatio*100, line.Fare, a.Fare),
				ErrInvalidArgument)
		}
	}

	return nil
}

// Apply は SetFareAction を実行する。
//
// 処理:
//  1. 路線の運賃を変更
//  2. fare_changed イベントを生成（同じ運賃でも記録のため生成する）
func (a *SetFareAction) Apply(g *Game) ([]Event, error) {
	if err := a.Validate(g); err != nil {
		return nil, err
	}

	c := g.PlayerCompany()
	line := c.FindLine(a.LineID)
	oldFare := line.Fare
	line.Fare = a.Fare

	// 変化率（%）
	changePct := 0.0
	if oldFare > 0 {
		changePct = float64(a.Fare-oldFare) / float64(oldFare) * 100.0
	}

	events := []Event{
		g.Info(EventFareChanged,
			fmt.Sprintf("%sの運賃を %d円 → %d円 に変更しました（%+.1f%%）",
				line.DisplayName, oldFare, a.Fare, changePct)).
			With("line_id", line.ID).
			With("old_fare", oldFare).
			With("new_fare", a.Fare).
			With("change_pct", changePct).
			ForCompany(c.ID).
			ForLine(line.ID),
	}

	return events, nil
}
