package core

// このファイルは Phase 8-3 の廃線 AI を実装する。
//
// 赤字が続く路線を経済ゲート（連続赤字ターン / 累積損失）と政治ゲート
// （沿線都市の政治圧力）で判定し、条件を満たすものを 1 ターンに 1 本だけ
// 廃止する。保守的な性格（Conservatism > AbandonConservatismLimit）の
// 会社は廃線しない。

// 廃線 AI のゲート閾値（Phase 9 で抑制）。
const (
	// AbandonDeficitTurns は廃線の経済ゲート（連続赤字ターン数、3年）。
	AbandonDeficitTurns = 36
	// AbandonLossRatio は廃線の経済ゲート（簿価に対する累積損失の割合 %）。
	AbandonLossRatio = 50
	// AbandonPoliticalLimit は廃線の政治ゲート（政治圧力の上限）。
	AbandonPoliticalLimit = 0.7
	// AbandonConservatismLimit は廃線を拒否する保守性の閾値。
	AbandonConservatismLimit = 0.5
)

// aiAbandon は AI が赤字路線を廃止する。
//
// 1 ターンに最大 1 路線を廃止する。対象がなければ nil を返す。
func (g *Game) aiAbandon(c *Company, ai *AI) []Event {
	for _, line := range c.Lines {
		if g.shouldAbandon(c, line, ai) {
			return g.executeAbandonLine(c, line)
		}
	}
	return nil
}

// shouldAbandon は廃線すべきかを判定する。
//
// 3 条件をすべて満たす場合のみ true:
//   - 経済: 連続赤字 ≥ 36 ターン、または累積損失 ≥ 簿価の 50%
//   - 政治: 両端都市の政治圧力 ≤ 0.7
//   - 性格: 保守的すぎない（Conservatism ≤ 0.5）
func (g *Game) shouldAbandon(c *Company, line *Line, ai *AI) bool {
	// 保守的な会社は廃線しにくい
	if ai.Personality.Conservatism > AbandonConservatismLimit {
		return false
	}

	// 経済ゲート
	value := lineValue(line)
	ecoOK := line.ConsecutiveDeficitTurns >= AbandonDeficitTurns ||
		line.AccumulatedLoss <= -(value*AbandonLossRatio/100)
	if !ecoOK {
		return false
	}

	// 政治ゲート（沿線都市の政治圧力）
	if g.cityPressureExceeds(line.From) || g.cityPressureExceeds(line.To) {
		return false
	}

	return true
}

// cityPressureExceeds は指定都市の政治圧力が閾値を超えているかを返す。
func (g *Game) cityPressureExceeds(cityID string) bool {
	city := g.Cities[cityID]
	return city != nil && city.PoliticalPressure > AbandonPoliticalLimit
}

// executeAbandonLine は AI 用の廃線。
func (g *Game) executeAbandonLine(c *Company, line *Line) []Event {
	action := &AbandonLineAction{
		ActorCompanyID: c.ID,
		LineID:         line.ID,
		Reason:         "economic",
	}
	events, err := g.applyAIAction(action)
	if err != nil {
		return nil
	}
	return events
}
