package core

import "fmt"

// このファイルは AbandonLineAction の本実装を提供する（Phase 1-8-3）。
//
// プレイヤーは収益の悪化した路線を廃止できる。廃線には次の効果がある:
//
//   - 路線を削除する
//   - 用地売却益を得る（地価 × 距離 × 係数）
//   - 沿線都市の政治圧力が +0.2 上昇する（両端都市）
//   - 会社の評判が -5 低下する
//   - 沿線需要を再計算する
//
// 廃線は 4 ゲートのうち経済（Gate 1）と政治（Gate 3）を満たす必要がある。
// 代替路線（Gate 2）と物理状態（Gate 4）は Phase 1-8-3 では簡略化する。

// validAbandonReasons は廃線理由として許可する文字列の集合。
var validAbandonReasons = map[string]bool{
	"economic":  true,
	"disaster":  true,
	"political": true,
	"wartime":   true,
}

// Validate は AbandonLineAction が実行可能かどうかを検証する。
//
// 検証内容:
//   - ゲームが存在する
//   - プレイヤー会社が存在する
//   - 理由が既知
//   - 路線が存在する（自社所有）
//   - Gate 1（経済）: 連続赤字 ≥ 24ヶ月、または累積損失 ≥ 路線価値の30%
//   - Gate 3（政治）: 両端都市の政治圧力 ≤ 0.7
//
// Validate は状態を変更しない。
func (a *AbandonLineAction) Validate(g *Game) error {
	if g == nil {
		return NewValidationError("GAME_NIL", "ゲームが初期化されていません", nil)
	}
	c := g.actorCompany(a.ActorCompanyID)
	if c == nil {
		return NewValidationError("NO_PLAYER", "プレイヤー会社が存在しません", ErrNoPlayer)
	}
	if !validAbandonReasons[a.Reason] {
		return NewValidationError("INVALID_REASON",
			fmt.Sprintf("不正な理由です: %s", a.Reason), ErrInvalidArgument)
	}

	line := c.FindLine(a.LineID)
	if line == nil {
		for _, other := range g.Companies {
			if other.ID != c.ID && other.FindLine(a.LineID) != nil {
				return NewValidationError("NOT_OWNED", "他社の路線は廃止できません", ErrUnauthorized)
			}
		}
		return NewValidationError("LINE_NOT_FOUND",
			fmt.Sprintf("路線が見つかりません: %s", a.LineID), ErrNotFound)
	}

	// Gate 1: 経済
	if line.ConsecutiveDeficitTurns < 24 && line.AccumulatedLoss > -(lineValue(line)*30/100) {
		return NewValidationError("GATE_ECONOMIC",
			fmt.Sprintf("赤字が不十分です（連続%dヶ月 / 累積損失 %s）",
				line.ConsecutiveDeficitTurns, FormatYen(line.AccumulatedLoss)),
			ErrRestricted)
	}

	// Gate 3: 政治圧力
	from := g.Cities[line.From]
	to := g.Cities[line.To]
	if from != nil && from.PoliticalPressure > 0.7 {
		return NewValidationError("GATE_POLITICAL",
			fmt.Sprintf("%sの政治圧力が高すぎます（%.2f）", from.Name, from.PoliticalPressure),
			ErrRestricted)
	}
	if to != nil && to.PoliticalPressure > 0.7 {
		return NewValidationError("GATE_POLITICAL",
			fmt.Sprintf("%sの政治圧力が高すぎます（%.2f）", to.Name, to.PoliticalPressure),
			ErrRestricted)
	}

	return nil
}

// Apply は AbandonLineAction を実行する。
//
// 処理:
//  1. Validate と同じ検証
//  2. 用地売却益を得る（地価 × 距離 × 係数）
//  3. 路線を削除
//  4. 評判を -5
//  5. 両端都市の政治圧力を +0.2（上限 1.0）
//  6. 沿線需要を再計算
//  7. line_abandoned イベントを生成
func (a *AbandonLineAction) Apply(g *Game) ([]Event, error) {
	if err := a.Validate(g); err != nil {
		return nil, err
	}

	c := g.actorCompany(a.ActorCompanyID)
	line := c.FindLine(a.LineID)
	from := g.Cities[line.From]
	to := g.Cities[line.To]

	// 用地売却益（地価 × 距離 × 係数）
	landValue := int64(0)
	if from != nil {
		landValue += int64(from.LandValue) * 10000
	}
	if to != nil {
		landValue += int64(to.LandValue) * 10000
	}
	proceeds := int64(line.LengthKm) * landValue / 100000

	c.Cash += proceeds
	c.RemoveLine(line.ID)
	c.Reputation -= 5

	if from != nil {
		from.PoliticalPressure += 0.2
		if from.PoliticalPressure > 1.0 {
			from.PoliticalPressure = 1.0
		}
		g.RecalcCityDemand(from)
	}
	if to != nil {
		to.PoliticalPressure += 0.2
		if to.PoliticalPressure > 1.0 {
			to.PoliticalPressure = 1.0
		}
		g.RecalcCityDemand(to)
	}

	events := []Event{
		g.Warn(EventLineAbandoned,
			fmt.Sprintf("%sを廃止しました（理由: %s、売却益: %s）",
				line.DisplayName, a.Reason, FormatYen(proceeds))).
			With("line_id", line.ID).
			With("reason", a.Reason).
			With("proceeds", proceeds).
			With("accumulated_loss", line.AccumulatedLoss).
			ForCompany(c.ID).
			ForLine(line.ID),
	}
	return events, nil
}

// lineValue は路線の簿価を返す。
//
// 路線価値は敷設コスト相当（距離 × BaseCostPerKm）とみなす簡易評価。
func lineValue(line *Line) int64 {
	return int64(line.LengthKm * float64(BaseCostPerKm))
}
