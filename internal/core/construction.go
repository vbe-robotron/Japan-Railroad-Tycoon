package core

import (
	"fmt"
	"math"
)

// このファイルは路線敷設の 4 ゲート判定（技術・法規・地形・経済）を実装する。
//
// 各ゲートは GatePass / GateFail / GateConditional の 3 状態を返す。
// CanBuildLine はすべてのゲートを評価し、1 つでも GateFail があれば
// 敷設不可（Can=false）とする。GateConditional は条件付きで実行可能と
// みなし、敷設を妨げない。

// GateStatus はゲートの判定結果を表す。
type GateStatus string

const (
	// GatePass は条件を満たすことを表す。
	GatePass GateStatus = "pass"
	// GateFail は条件を満たさないことを表す（理由付き）。
	GateFail GateStatus = "fail"
	// GateConditional は条件付きで可能であることを表す。
	GateConditional GateStatus = "conditional"
)

// ConstructionGate は 1 つのゲートの判定結果。
type ConstructionGate struct {
	Kind    string         // "technology"|"regulation"|"terrain"|"economy"
	Status  GateStatus     // pass|fail|conditional
	Reason  string         // 人間可読なメッセージ
	Details map[string]any // 追加情報（任意）
}

// CanBuildResult は 4 ゲートの総合判定。
type CanBuildResult struct {
	Can    bool               // 敷設可能か
	Gates  []ConstructionGate // 各ゲートの判定結果
	Cost   int64              // 敷設コスト（円）
	Months int                // 工期（ヶ月）
}

// CanBuildLine は路線敷設の可否を 4 ゲートで判定する。
//
// 全ゲートが pass または conditional なら Can = true。
// 1 つでも fail があれば Can = false。
func (g *Game) CanBuildLine(fromID, toID string, gauge int, double bool) CanBuildResult {
	gates := []ConstructionGate{
		g.checkTechnologyGate(gauge),
		g.checkRegulationGate(),
		g.checkTerrainGate(fromID, toID),
		g.checkEconomyGate(fromID, toID, gauge, double),
	}

	can := true
	for _, gate := range gates {
		if gate.Status == GateFail {
			can = false
			break
		}
	}

	return CanBuildResult{
		Can:    can,
		Gates:  gates,
		Cost:   g.LineCost(fromID, toID, gauge, double),
		Months: g.buildMonths(fromID, toID),
	}
}

// checkTechnologyGate は技術ゲートを判定する。
//
// 軌間に応じて必要な技術を決める:
//   - 762mm: 馬車鉄道（1880）
//   - 1067mm: 蒸気鉄道（1887）
//   - 1435mm: 電気軌道（1895）
func (g *Game) checkTechnologyGate(gauge int) ConstructionGate {
	var requiredID string
	switch gauge {
	case 762:
		requiredID = "horse_tram" // 1880
	case 1067:
		requiredID = "steam_railway" // 1887
	case 1435:
		requiredID = "electric_tram" // 1895
	default:
		return ConstructionGate{
			Kind:   "technology",
			Status: GateFail,
			Reason: fmt.Sprintf("不正な軌間です: %d", gauge),
		}
	}

	tech, ok := TechnologyByID(requiredID)
	if !ok {
		return ConstructionGate{
			Kind:   "technology",
			Status: GateFail,
			Reason: "技術データが見つかりません",
		}
	}

	if g.Date.Year < tech.UnlockYear {
		return ConstructionGate{
			Kind:   "technology",
			Status: GateFail,
			Reason: fmt.Sprintf("%sが未解禁です（%d年解禁）", tech.Name, tech.UnlockYear),
			Details: map[string]any{
				"required":    tech.ID,
				"unlock_year": tech.UnlockYear,
			},
		}
	}

	return ConstructionGate{
		Kind:   "technology",
		Status: GatePass,
		Reason: fmt.Sprintf("%s（OK）", tech.Name),
		Details: map[string]any{
			"required":    tech.ID,
			"unlock_year": tech.UnlockYear,
		},
	}
}

// checkRegulationGate は法規ゲートを判定する。
//
// 私設鉄道条例（1887）が制定されていないと敷設不可。
func (g *Game) checkRegulationGate() ConstructionGate {
	if !IsRegulationActive("private_railway_ordinance", g.Date.Year) {
		return ConstructionGate{
			Kind:   "regulation",
			Status: GateFail,
			Reason: "私設鉄道条例が未制定です（1887年制定）",
		}
	}
	return ConstructionGate{
		Kind:   "regulation",
		Status: GatePass,
		Reason: "私設鉄道条例（OK）",
		Details: map[string]any{
			"regulation": "private_railway_ordinance",
		},
	}
}

// checkTerrainGate は地形ゲートを判定する。
//
// Phase 1-6 では地形は平地固定（conditional）。
// Phase 7 で nodes.json から地形を取得する予定。
func (g *Game) checkTerrainGate(fromID, toID string) ConstructionGate {
	// Phase 1-6 では地形は平地固定
	return ConstructionGate{
		Kind:   "terrain",
		Status: GateConditional,
		Reason: "平地（OK）",
		Details: map[string]any{
			"terrain": string(TerrainPlain),
		},
	}
}

// checkEconomyGate は経済ゲートを判定する。
//
// 資金 ≥ 敷設コスト を満たすか。
func (g *Game) checkEconomyGate(fromID, toID string, gauge int, double bool) ConstructionGate {
	c := g.PlayerCompany()
	if c == nil {
		return ConstructionGate{
			Kind:   "economy",
			Status: GateFail,
			Reason: "プレイヤー会社なし",
		}
	}
	cost := g.LineCost(fromID, toID, gauge, double)
	if c.Cash < cost {
		return ConstructionGate{
			Kind:   "economy",
			Status: GateFail,
			Reason: fmt.Sprintf("資金不足（必要: %s / 所持: %s）",
				FormatYen(cost), FormatYen(c.Cash)),
			Details: map[string]any{
				"required":  cost,
				"available": c.Cash,
			},
		}
	}
	return ConstructionGate{
		Kind:   "economy",
		Status: GatePass,
		Reason: fmt.Sprintf("資金OK（必要: %s / 所持: %s）",
			FormatYen(cost), FormatYen(c.Cash)),
		Details: map[string]any{
			"required":  cost,
			"available": c.Cash,
		},
	}
}

// buildMonths は工期（ヶ月）を返す。
//
// 距離に比例: months = max(1, ceil(DistanceKm / 5))
func (g *Game) buildMonths(fromID, toID string) int {
	dist := g.DistanceKm(fromID, toID)
	months := int(math.Ceil(dist / 5.0))
	if months < 1 {
		months = 1
	}
	return months
}
