package core

import (
	"fmt"
	"strings"
)

// このファイルは BuildLineAction の本実装を提供する。
//
// Phase 1-6 では「即時開業」とする。工期システム（line_construction_started
// → line_opened の 2 段階）は Phase 1-7 以降で導入する。

// Validate は BuildLineAction が実行可能かどうかを検証する。
//
// 検証内容:
//   - プレイヤー会社が存在する
//   - from と to の都市が存在する
//   - from != to
//   - 同じ区間に自社路線がない
//   - 4ゲートをすべて満たす
func (a *BuildLineAction) Validate(g *Game) error {
	if g == nil {
		return NewValidationError("GAME_NIL", "ゲームが初期化されていません", nil)
	}

	c := g.PlayerCompany()
	if c == nil {
		return NewValidationError("NO_PLAYER", "プレイヤー会社が存在しません", ErrNoPlayer)
	}

	if a.From == a.To {
		return NewValidationError("SAME_CITY", "出発地と目的地が同じです", ErrInvalidArgument)
	}

	from, ok := g.Cities[a.From]
	if !ok {
		return NewValidationError("FROM_NOT_FOUND",
			fmt.Sprintf("出発都市が見つかりません: %s", a.From), ErrCityNotFound)
	}
	to, ok := g.Cities[a.To]
	if !ok {
		return NewValidationError("TO_NOT_FOUND",
			fmt.Sprintf("目的地都市が見つかりません: %s", a.To), ErrCityNotFound)
	}

	// 既存路線チェック
	for _, l := range c.Lines {
		if (l.From == a.From && l.To == a.To) || (l.From == a.To && l.To == a.From) {
			return NewValidationError("LINE_EXISTS",
				fmt.Sprintf("%s〜%s間には既に路線があります", from.Name, to.Name), ErrLineExists)
		}
	}

	// 4ゲートチェック
	result := g.CanBuildLine(a.From, a.To, a.Gauge, a.Double)
	if !result.Can {
		for _, gate := range result.Gates {
			if gate.Status == GateFail {
				code := "GATE_" + strings.ToUpper(gate.Kind)
				return NewValidationError(code, gate.Reason, ErrRestricted)
			}
		}
	}

	return nil
}

// Apply は BuildLineAction を実行する。
//
// 処理:
//  1. 4ゲート再チェック（Validate と同じ）
//  2. 敷設コストを支払う
//  3. Line を生成し、Company.Lines に追加
//  4. 都市の需要を再計算
//  5. line_opened イベントを生成
//     （Phase 1-6 では即時開業。工期を考慮する場合、line_construction_started
//     は Phase 1-7 以降）
func (a *BuildLineAction) Apply(g *Game) ([]Event, error) {
	if err := a.Validate(g); err != nil {
		return nil, err
	}

	c := g.PlayerCompany()
	from := g.Cities[a.From]
	to := g.Cities[a.To]

	cost := g.LineCost(a.From, a.To, a.Gauge, a.Double)
	months := g.buildMonths(a.From, a.To)

	// 資金減算
	c.Cash -= cost

	// Line 生成
	lineID := g.IDGen.Next()
	line := NewLine(lineID, c.ID, a.From, a.To)
	line.DisplayName = fmt.Sprintf("%s〜%s", from.Name, to.Name)
	line.LengthKm = g.DistanceKm(a.From, a.To)
	line.Gauge = a.Gauge
	line.Double = a.Double
	line.Opened = g.Date
	line.Fare = g.DefaultFare(a.From, a.To)
	line.TrainsPerDay = 10

	// 会社に追加
	c.AddLine(line)

	// 都市の需要再計算
	g.recalcCityDemand(from)
	g.recalcCityDemand(to)

	// イベント生成
	events := []Event{
		g.Success(EventLineOpened,
			fmt.Sprintf("%s〜%s間の路線が開業しました（%s）",
				from.Name, to.Name, FormatYen(cost))).
			With("line_id", line.ID).
			With("from", from.ID).
			With("to", to.ID).
			With("from_name", from.Name).
			With("to_name", to.Name).
			With("length_km", line.LengthKm).
			With("gauge", line.Gauge).
			With("cost", cost).
			With("months", months).
			ForCompany(c.ID).
			ForLine(line.ID),
	}

	return events, nil
}
