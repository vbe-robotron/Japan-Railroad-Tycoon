package core

import "sort"

// このファイルは Phase 8-2 の建設 AI を実装する。
//
// 収益ポテンシャル（需要・地価・競合・開発度）で候補路線を評価し、
// 最高スコアの候補を 1 本だけ敷設する。史実ルートに該当がないターンに
// 適応ルートとして呼ばれる。

// 建設 AI の閾値。
const (
	// MinAIBuildCash は建設 AI が行動する最低資金（円）。
	MinAIBuildCash = 2_000_000

	// MinBuildScore は敷設を実行する最低スコア。
	//
	// 需要は人口のべき乗則（指数 0.6）で、史実データの最大都市（50 万）でも
	// 両端平均の需要は 40 程度にしかならない。初期案の 30.0 では史実デー
	// タで一度も敷設が発生しないため、実データの主要都市ペア（東京〜横浜
	// など、スコア 25〜30）で発火する 15.0 に調整した。
	MinBuildScore = 15.0
)

// BuildCandidate は敷設候補を表す。
type BuildCandidate struct {
	From   string
	To     string
	Gauge  int
	Double bool
	Cost   int64
	Score  float64
}

// aiBuild は AI が路線を敷設する。
//
// 収益ポテンシャルを評価し、最高スコアの候補を敷設する。
// 資金不足・スコア閾値未満なら何もしない。
func (g *Game) aiBuild(c *Company, ai *AI) []Event {
	// 最低資金チェック
	if c.Cash < MinAIBuildCash {
		return nil
	}

	candidates := g.buildCandidates(c, ai)
	if len(candidates) == 0 {
		return nil
	}

	best := candidates[0]
	for _, cand := range candidates[1:] {
		if cand.Score > best.Score {
			best = cand
		}
	}

	if best.Score < MinBuildScore {
		return nil
	}

	return g.executeBuildLine(c, best.From, best.To, best.Gauge, best.Double, "")
}

// buildCandidates は敷設候補を列挙する。
//
// 自社路線の隣接都市、または拠点都市から候補を生成する。
// map 走査の非決定性を避けるため、候補は From→To の昇順でソートして返す。
func (g *Game) buildCandidates(c *Company, ai *AI) []BuildCandidate {
	var candidates []BuildCandidate
	seen := map[string]bool{} // from:to

	// 1. 自社路線の端点から隣接都市へ延伸
	for _, line := range c.Lines {
		for _, endpoint := range []string{line.From, line.To} {
			for _, target := range g.Cities {
				if target.ID == endpoint {
					continue
				}
				key := endpoint + ":" + target.ID
				if seen[key] {
					continue
				}
				seen[key] = true

				if g.lineExists(endpoint, target.ID, c.ID) {
					continue
				}

				cost := g.LineCost(endpoint, target.ID, 1067, false)
				if c.Cash < cost {
					continue
				}
				score := g.scoreBuildCandidate(c, endpoint, target, ai)
				candidates = append(candidates, BuildCandidate{
					From: endpoint, To: target.ID,
					Gauge: 1067, Double: false,
					Cost: cost, Score: score,
				})
			}
		}
	}

	// 2. 会社の拠点都市から候補（初回敷設用）
	if len(c.Lines) == 0 {
		home := g.companyHomeCity(c)
		if home != "" {
			for _, target := range g.Cities {
				if target.ID == home {
					continue
				}
				key := home + ":" + target.ID
				if seen[key] {
					continue
				}
				seen[key] = true

				cost := g.LineCost(home, target.ID, 1067, false)
				if c.Cash < cost {
					continue
				}
				score := g.scoreBuildCandidate(c, home, target, ai)
				candidates = append(candidates, BuildCandidate{
					From: home, To: target.ID,
					Gauge: 1067, Double: false,
					Cost: cost, Score: score,
				})
			}
		}
	}

	// 決定論のため From→To の昇順にソートする。
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].From != candidates[j].From {
			return candidates[i].From < candidates[j].From
		}
		return candidates[i].To < candidates[j].To
	})

	return candidates
}

// companyHomeCity は会社の拠点都市を返す。
//
// Region から推定する簡易版。該当がなければ空文字を返す。
func (g *Game) companyHomeCity(c *Company) string {
	switch c.Region {
	case "kansai":
		return "umeda"
	case "kanto":
		return "tokyo"
	case "chubu":
		return "nagoya"
	case "shikoku":
		return "matsuyama"
	default:
		return ""
	}
}

// scoreBuildCandidate は敷設候補のスコアを返す。
//
// スコア = 需要ポテンシャル + 地価の安さ + 競合の少なさ + 開発度の将来性。
// 性格（Aggression）と距離補正を掛ける。
func (g *Game) scoreBuildCandidate(c *Company, fromID string, to *City, ai *AI) float64 {
	from := g.Cities[fromID]
	if from == nil || to == nil {
		return 0
	}

	score := 0.0

	// 需要ポテンシャル（両端の需要の平均）
	demand := (g.Demand(from) + g.Demand(to)) / 2
	score += float64(demand) * 0.5

	// 地価が安いほど加点
	score += float64(100-to.LandValue) * 0.1

	// 競合の少なさ
	competitors := g.countCompetitorsAt(to.ID, c.ID)
	score -= float64(competitors) * 5.0

	// 開発度による将来性
	score += to.Developed * 20.0

	// 性格補正
	score *= (1.0 + ai.Personality.Aggression*0.3)

	// 距離補正（遠すぎると減点）
	dist := g.DistanceKm(fromID, to.ID)
	if dist > 50 {
		score *= 0.7
	}

	return score
}

// countCompetitorsAt は指定都市に接続する他社路線数を返す。
//
// excludeCompanyID は自社（除外する会社）を指定する。
func (g *Game) countCompetitorsAt(cityID, excludeCompanyID string) int {
	n := 0
	for id, c := range g.Companies {
		if id == excludeCompanyID {
			continue
		}
		for _, l := range c.Lines {
			if l.From == cityID || l.To == cityID {
				n++
			}
		}
	}
	return n
}
