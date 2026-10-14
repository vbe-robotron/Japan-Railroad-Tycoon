package core

import "fmt"

// このファイルは Phase 8-2 の運用 AI を実装する。
//
// 運賃調整（営業係数に応じた値上げ・値下げ）と車両増備（輸送力増強）を
// 自社路線ごとに行う。

// aiOperate は運賃調整と車両購入を行う。
func (g *Game) aiOperate(c *Company, ai *AI) []Event {
	var events []Event
	for _, line := range c.Lines {
		events = append(events, g.aiAdjustFare(c, line)...)
		events = append(events, g.aiPurchaseStock(c, line)...)
	}
	return events
}

// aiAdjustFare は運賃を調整する。
//
//   - 営業係数 > 100: 値上げ（最大 +10%、上限 MaxFare）
//   - 営業係数 < 70: 値下げ（最大 -5%、下限 MinFare+10）
//   - それ以外: 変更なし
func (g *Game) aiAdjustFare(c *Company, line *Line) []Event {
	if line.Fare <= 0 {
		return nil
	}

	var newFare int
	switch {
	case line.OperatingRatio > 100 && line.Fare < MaxFare:
		newFare = line.Fare + line.Fare/10 // +10%
		if newFare > MaxFare {
			newFare = MaxFare
		}
	case line.OperatingRatio < 70 && line.Fare > MinFare+10:
		newFare = line.Fare - line.Fare/20 // -5%
		if newFare < MinFare+10 {
			newFare = MinFare + 10
		}
	default:
		return nil
	}

	if newFare == line.Fare {
		return nil
	}

	// ±MaxFareChangeRatio 制約チェック
	changeRatio := float64(newFare-line.Fare) / float64(line.Fare)
	if changeRatio > MaxFareChangeRatio || changeRatio < -MaxFareChangeRatio {
		return nil
	}

	oldFare := line.Fare
	line.Fare = newFare

	return []Event{
		g.Info(EventFareChanged,
			fmt.Sprintf("%sが%sの運賃を %d円 → %d円 に変更しました",
				c.Name, line.DisplayName, oldFare, newFare)).
			With("line_id", line.ID).
			With("old_fare", oldFare).
			With("new_fare", newFare).
			With("ai", true).
			ForCompany(c.ID).
			ForLine(line.ID),
	}
}

// aiPurchaseStock は車両を購入する。
//
// 簡易実装: 予算（50 万円）以上あれば車両を +5 本する。上限 20 本。
func (g *Game) aiPurchaseStock(c *Company, line *Line) []Event {
	// 簡易: 予算 50万円以上あれば車両 +5
	if c.Cash < 500_000 {
		return nil
	}
	if line.TrainsPerDay >= 20 {
		return nil // 上限
	}
	oldTrains := line.TrainsPerDay
	line.TrainsPerDay += 5
	cost := int64(100_000) // 10万円
	c.Cash -= cost

	return []Event{
		g.Info(EventStockPurchased,
			fmt.Sprintf("%sが%sの車両を増備しました（%d本 → %d本）",
				c.Name, line.DisplayName, oldTrains, line.TrainsPerDay)).
			With("line_id", line.ID).
			With("old_trains", oldTrains).
			With("new_trains", line.TrainsPerDay).
			With("cost", cost).
			With("ai", true).
			ForCompany(c.ID).
			ForLine(line.ID),
	}
}
