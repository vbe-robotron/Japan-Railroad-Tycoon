package core

import (
	"fmt"
	"math"
	"sort"
)

// このファイルは Phase 1-7 の経済シミュレーションを実装する。
//
// 収益モデル:
//
//	収益 = 旅客収入 + 貨物収入 + 不動産収入 + 流通収入 - 維持費 - 減価償却
//
// Phase 1-7 では旅客収入・維持費・減価償却のみを実装する。
// 貨物・不動産・流通収入は Phase 1-8 以降で追加する。

const (
	// DaysPerMonth は 1 ヶ月の日数（旅客需要の月次換算に用いる）。
	DaysPerMonth = 30

	// PassengersPerTrain は 1 本あたりの輸送力（人/日）。
	PassengersPerTrain = 500

	// LineDepreciationMonths は路線の減価償却期間（50 年 = 600 ヶ月）。
	LineDepreciationMonths = 600
)

// simulateEconomy は全社の月次経済を計算し、イベントを返す。
//
// EndTurn() から呼ばれる。時間進行前の Date を用いて計算する。
//
// 処理:
//  1. 各社の旅客収入・維持費・減価償却を計算
//  2. Company.Cash を更新
//  3. Line.Profit, Line.OperatingRatio を更新
//  4. 赤字継続ターンを更新し、line_deficit / line_critical を生成
//  5. 黒字転換時に line_profitable を生成
//
// 決定論のため、会社は ID の昇順で処理する
// （g.Companies は map であり、順序が不定なため）。
func (g *Game) simulateEconomy() []Event {
	var events []Event

	ids := make([]string, 0, len(g.Companies))
	for id := range g.Companies {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		events = append(events, g.simulateCompany(g.Companies[id])...)
	}
	return events
}

// simulateCompany は 1 社の経済を計算する。
func (g *Game) simulateCompany(c *Company) []Event {
	var events []Event
	var totalRevenue, totalCost int64

	for _, line := range c.Lines {
		revenue, cost := g.simulateLine(line, c)
		totalRevenue += revenue
		totalCost += cost

		profit := revenue - cost
		line.Profit = profit

		// 営業係数（コスト / 収入 × 100）
		if revenue > 0 {
			line.OperatingRatio = float64(cost) / float64(revenue) * 100.0
		} else {
			line.OperatingRatio = 999.9
		}

		// 赤字継続の更新
		events = append(events, g.updateLineHealth(c, line, profit)...)
	}

	// 減価償却（路線: 50 年定額）
	depreciation := g.calculateDepreciation(c)
	totalCost += depreciation

	// 資金更新
	c.Cash += totalRevenue - totalCost

	return events
}

// simulateLine は 1 路線の月次収入とコストを返す。
//
// 収入:
//   - 旅客収入 = 旅客数 × 運賃
//
// コスト:
//   - 維持費 = LengthKm × MaintenancePerKm × 補正
func (g *Game) simulateLine(line *Line, c *Company) (revenue int64, cost int64) {
	// 旅客収入
	passengers := g.linePassengers(line)
	revenue = int64(passengers) * int64(line.Fare)

	// 維持費
	maint := float64(line.LengthKm) * MaintenancePerKm
	if line.Double {
		maint *= 1.5
	}
	if line.Electrified {
		maint *= 1.3
	}
	cost = int64(maint)

	return revenue, cost
}

// linePassengers は路線の月次旅客数を返す。
//
// passengers = 需要 × シェア × 輸送力制約
//
// 需要（Game.Demand）は日次需要として扱い、月次旅客数へ DaysPerMonth を
// 掛けて換算する。輸送力は TrainsPerDay × PassengersPerTrain × DaysPerMonth。
func (g *Game) linePassengers(line *Line) int {
	from := g.Cities[line.From]
	to := g.Cities[line.To]
	if from == nil || to == nil {
		return 0
	}

	// 需要: 両端の都市の平均（日次）
	demand := (g.Demand(from) + g.Demand(to)) / 2

	// シェア
	share := g.lineShare(line)

	// 日次需要を月次旅客数へ換算（丸めで浮動小数の切り捨て誤差を防ぐ）
	passengers := int(math.Round(float64(demand) * share * DaysPerMonth))

	// 輸送力制約（月次）
	capacity := line.TrainsPerDay * PassengersPerTrain * DaysPerMonth
	if passengers > capacity {
		passengers = capacity
	}

	return passengers
}

// lineShare は路線の市場シェアを返す（0.0-1.0）。
//
// 同じ区間（from-to または to-from）を走る全路線の
// スコアに対する自社路線の比率を返す。
func (g *Game) lineShare(line *Line) float64 {
	myScore := g.lineScore(line)
	totalScore := 0.0

	for _, c := range g.Companies {
		for _, l := range c.Lines {
			if sameOD(l, line) {
				totalScore += g.lineScore(l)
			}
		}
	}

	if totalScore == 0 {
		return 1.0
	}
	return myScore / totalScore
}

// sameOD は 2 路線が同じ区間かどうかを返す。
//
// from-to が一致するほか、逆方向（to-from）も同一区間とみなす。
func sameOD(a, b *Line) bool {
	return (a.From == b.From && a.To == b.To) ||
		(a.From == b.To && a.To == b.From)
}

// lineScore は路線の競争力スコアを返す。
//
// score = (100 / Fare) × (TrainsPerDay / 10) × 電化補正 × 老朽化補正
//
// 運賃が安いほど、本数が多いほど、電化されているほど、
// 状態が良いほど競争力が高い。
func (g *Game) lineScore(line *Line) float64 {
	score := 1.0

	// 運賃（安いほど高い）。Fare=0 は無料扱いで運賃補正なし。
	if line.Fare > 0 {
		score *= 100.0 / float64(line.Fare)
	}

	// 本数
	score *= float64(line.TrainsPerDay) / 10.0

	// 電化補正
	if line.Electrified {
		score *= 1.3
	}

	// 老朽化補正
	score *= line.Condition

	return score
}

// updateLineHealth は路線の健全性を更新し、イベントを返す。
//
// 赤字継続が 6 ターンで line_deficit（警告）、24 ターンで line_critical
// （危機）を生成する。黒字に転換した場合は line_profitable を生成し、
// 赤字継続カウンタと累積損失をリセットする。
func (g *Game) updateLineHealth(c *Company, line *Line, profit int64) []Event {
	var events []Event

	if profit < 0 {
		line.ConsecutiveDeficitTurns++
		line.AccumulatedLoss += profit

		switch line.ConsecutiveDeficitTurns {
		case 24:
			// 24 ターン（= 2 年）連続赤字で廃線寸前
			events = append(events, g.Danger(EventLineCritical,
				fmt.Sprintf("%sが廃線寸前です（24ヶ月連続赤字）", line.DisplayName)).
				With("line_id", line.ID).
				With("consecutive_turns", line.ConsecutiveDeficitTurns).
				With("accumulated_loss", line.AccumulatedLoss).
				ForCompany(c.ID).
				ForLine(line.ID))
		case 6:
			// 6 ターン連続赤字で警告
			events = append(events, g.Warn(EventLineDeficit,
				fmt.Sprintf("%sが赤字です（月次 %s）",
					line.DisplayName, FormatYen(profit))).
				With("line_id", line.ID).
				With("monthly_loss", profit).
				With("accumulated_loss", line.AccumulatedLoss).
				With("consecutive_turns", line.ConsecutiveDeficitTurns).
				ForCompany(c.ID).
				ForLine(line.ID))
		}
	} else {
		// 黒字転換
		if line.ConsecutiveDeficitTurns > 0 {
			events = append(events, g.Success(EventLineProfitable,
				fmt.Sprintf("%sが黒字転換しました（月次 %s）",
					line.DisplayName, FormatYen(profit))).
				With("line_id", line.ID).
				With("monthly_profit", profit).
				ForCompany(c.ID).
				ForLine(line.ID))
		}
		line.ConsecutiveDeficitTurns = 0
		line.AccumulatedLoss = 0
	}

	return events
}

// calculateDepreciation は会社の月次減価償却費を返す。
//
// 路線: 50 年定額（LineDepreciationMonths = 600 ヶ月）。
// 路線価値は敷設コスト相当（距離 × BaseCostPerKm）とみなす。
// 車両・施設の減価償却は Phase 1-8 以降で追加する。
func (g *Game) calculateDepreciation(c *Company) int64 {
	var total int64
	for _, line := range c.Lines {
		value := int64(line.LengthKm * float64(BaseCostPerKm))
		total += value / LineDepreciationMonths
	}
	return total
}
