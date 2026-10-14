package core

import "sort"

// このファイルは Phase 8-3 の M&A AI を実装する。
//
// 攻撃性（Aggression）と資金余裕のある AI は他社株を段階的に取得し、
// 保有比率が 30% を超えると TOB（公開買付け）で過半数を取り、50% 超で
// 合併して完全統合する。国鉄・自社・解散済みの会社は対象外とする。
// プレイヤー会社は高攻撃性（PlayerTakeoverAggression 以上）のときのみ
// 買収対象とする。

// M&A AI のパラメータ。
//
// Phase 9 で M&A 頻度を抑制した。史実の関西私鉄は 100 年以上独立を維持して
// おり、買収は攻撃的で資金に余裕のある会社が低確率で行うものに限定する。
const (
	// MAMinAggression は M&A を開始する最低攻撃性。
	MAMinAggression = 0.75
	// MAMinCash は M&A を開始する最低資金（2億円）。
	MAMinCash = 200_000_000
	// MATargetMinScore は買収対象とする最低スコア。
	MATargetMinScore = 80.0
	// MAMaxAcquireRatio は 1 ターンに取得する最大議決権比率（20%）。
	MAMaxAcquireRatio = 1.0 / 5.0
	// MATOBThreshold は TOB に切り替える保有比率。
	MATOBThreshold = 0.3
	// MATOBPremium は TOB のプレミアム係数。
	MATOBPremium = 1.1
	// MATOBTargetRatio は TOB の目標議決権比率。
	MATOBTargetRatio = 0.51
	// MAMergeRatio は合併コストの比率。
	MAMergeRatio = 0.5
	// PlayerTakeoverAggression はプレイヤーを買収対象にする最低攻撃性。
	PlayerTakeoverAggression = 0.9
	// MAMergeProbability は過半数保有後に合併する確率（統合準備期間）。
	MAMergeProbability = 0.2
	// MAProbabilityFactor は確率判定の係数（Aggression × 係数 が実行確率）。
	MAProbabilityFactor = 0.1
)

// aiMA は AI が他社を買収・合併する。
//
// 処理順:
//  0. 攻撃性・資金のゲートと確率判定
//  1. 既に 50% 超を保有する対象は（20% で）合併
//  2. それ以外は株取得（30% 超なら TOB）
//
// 確率判定は Aggression × MAProbabilityFactor（0.9 で 9%/ターン）。
// rng を使うため決定論は seed に依存する。
func (g *Game) aiMA(c *Company, ai *AI) []Event {
	// 攻撃性が低いと買収しない
	if ai.Personality.Aggression < MAMinAggression {
		return nil
	}
	// 資金余裕チェック
	if c.Cash < MAMinCash {
		return nil
	}
	// 確率判定（Aggression が高いほど実行しやすい）
	g.EnsureRng()
	if g.rng.Float64() > ai.Personality.Aggression*MAProbabilityFactor {
		return nil
	}

	target := g.findTakeoverTarget(c, ai)
	if target == nil {
		return nil
	}

	holdingPct := g.holdingPct(c, target)

	// 既に 50% 超保有なら合併を試みる（統合準備期間として 20% のみ）
	if holdingPct > 0.5 {
		if g.rng.Float64() < MAMergeProbability {
			return g.aiMerge(c, target)
		}
		return nil
	}

	count := g.takeoverShareCount(c, target, ai)
	if count <= 0 {
		return nil
	}

	// 30% 近くまで来たら TOB で過半数を狙う
	if holdingPct > MATOBThreshold && c.Cash > g.tobCost(c, target, ai) {
		return g.aiTenderOffer(c, target, ai)
	}

	// 通常の株取得
	return g.executeBuyShares(c, target, count)
}

// findTakeoverTarget は買収対象を探す。
//
// 対象の選定は決定論のため会社IDの昇順で走査し、最高スコアの 1 社を返す
// （同点は ID が小さい方を優先）。対象がなければ nil を返す。
func (g *Game) findTakeoverTarget(c *Company, ai *AI) *Company {
	ids := make([]string, 0, len(g.Companies))
	for id := range g.Companies {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var best *Company
	bestScore := MATargetMinScore
	for _, id := range ids {
		t := g.Companies[id]
		if t.ID == c.ID {
			continue
		}
		if t.Dissolved != nil {
			continue
		}
		if t.IsNational {
			continue // 国鉄は買収不可
		}
		if t.IsPlayer && ai.Personality.Aggression < PlayerTakeoverAggression {
			continue // 高攻撃性のみプレイヤーを狙う
		}
		score := g.takeoverScore(c, t, ai)
		if score > bestScore {
			best = t
			bestScore = score
		}
	}
	return best
}

// takeoverScore は買収対象のスコアを返す。
func (g *Game) takeoverScore(c, t *Company, ai *AI) float64 {
	score := 0.0

	// 時価総額が小さいほど狙いやすい
	marketCap := int64(t.TotalShares * t.StockPrice)
	if marketCap > 0 {
		score += 100.0 - float64(marketCap)/1_000_000.0
	}

	// 既存路線を持つ会社は買収が困難（路線数が多いほど統合コストが高い）
	score -= float64(len(t.Lines)) * 3.0

	// 既に子会社を持つ会社は対象外に近い
	if g.hasSubsidiaries(t) {
		score -= 30.0
	}

	// 路線が隣接
	if g.adjacentToCompany(c, t) {
		score += 30
	}

	// 同地域
	if c.Region == t.Region {
		score += 20
	}

	// 攻撃性補正
	score *= 1.0 + ai.Personality.Aggression*0.5

	// 資金力補正（買えない規模なら優先度を下げる）
	if marketCap > 0 && c.Cash < marketCap {
		score *= 0.3
	}

	return score
}

// hasSubsidiaries は会社 t が子会社を持つか（他社が t を親としているか）を返す。
//
// 買収対象の保護に用いる。子会社を抱える会社は統合が難しいため減点する。
func (g *Game) hasSubsidiaries(t *Company) bool {
	for _, c := range g.Companies {
		if c != t && c.ParentID == t.ID {
			return true
		}
	}
	return false
}

// adjacentToCompany は路線が隣接しているかを返す。
func (g *Game) adjacentToCompany(c, t *Company) bool {
	for _, l := range c.Lines {
		for _, ol := range t.Lines {
			if l.From == ol.From || l.From == ol.To ||
				l.To == ol.From || l.To == ol.To {
				return true
			}
		}
	}
	return false
}

// holdingPct は c が t を保有する議決権比率を返す。
func (g *Game) holdingPct(c, t *Company) float64 {
	if t.TotalShares <= 0 {
		return 0
	}
	return float64(c.Holding[t.ID]) / float64(t.TotalShares)
}

// takeoverShareCount は取得株数を返す。
//
// 発行済株の 20% と「資金の半分で買える株数」の小さい方を返す。
// 100 株未満なら 0（見送り）。
func (g *Game) takeoverShareCount(c, t *Company, ai *AI) int {
	maxCount := int(float64(t.TotalShares) * MAMaxAcquireRatio)
	affordable := 0
	if t.StockPrice > 0 {
		affordable = int(c.Cash / int64(t.StockPrice) / 2) // 資金の半分まで
	}
	if affordable < maxCount {
		maxCount = affordable
	}
	if maxCount < 100 {
		return 0
	}
	return maxCount
}

// tobCost は TOB のコストを概算する。
func (g *Game) tobCost(c, t *Company, ai *AI) int64 {
	needCount := tenderOfferNeed(c, t, MATOBTargetRatio)
	return int64(float64(needCount*t.StockPrice) * MATOBPremium)
}

// aiTenderOffer は TOB を実行する。
func (g *Game) aiTenderOffer(c, t *Company, ai *AI) []Event {
	action := &TenderOfferAction{
		ActorCompanyID:  c.ID,
		TargetCompanyID: t.ID,
		Premium:         MATOBPremium,
		TargetRatio:     MATOBTargetRatio,
	}
	events, err := g.applyAIAction(action)
	if err != nil {
		return nil
	}
	return events
}

// executeBuyShares は AI 用の株取得。
func (g *Game) executeBuyShares(c, t *Company, count int) []Event {
	action := &BuySharesAction{
		ActorCompanyID:  c.ID,
		TargetCompanyID: t.ID,
		Count:           count,
	}
	events, err := g.applyAIAction(action)
	if err != nil {
		return nil
	}
	return events
}

// aiMerge は合併を実行する。
func (g *Game) aiMerge(c, t *Company) []Event {
	action := &MergeAction{
		ActorCompanyID:  c.ID,
		TargetCompanyID: t.ID,
		Ratio:           MAMergeRatio,
		Name:            c.Name, // 社名は維持（簡易）
	}
	events, err := g.applyAIAction(action)
	if err != nil {
		return nil
	}
	return events
}

// applyAIAction は AI のアクションを検証して適用し、イベントに ai フラグを付ける。
func (g *Game) applyAIAction(a Action) ([]Event, error) {
	if err := a.Validate(g); err != nil {
		return nil, err
	}
	events, err := a.Apply(g)
	if err != nil {
		return nil, err
	}
	for i := range events {
		events[i] = events[i].With("ai", true)
	}
	return events, nil
}
