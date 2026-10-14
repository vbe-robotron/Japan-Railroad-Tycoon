package core

import "fmt"

// このファイルは AI の史実ルート追従を実装する。
//
// 史実路線データ（Game.HistoricalLines）から「この年・この月」に開業した
// 路線を探し、未敷設なら敷設する。資金不足の場合は借金で補う。

// aiFollowHistory は史実ルートを追従する。
//
// この年・この月に史実で敷設された路線を探し、敷設する。
// 資金不足なら借金で補う。該当路線がなければ nil を返す。
// 1 ターンに敷設するのは最大 1 本とする。
func (g *Game) aiFollowHistory(c *Company, ai *AI) []Event {
	for _, h := range g.HistoricalLines {
		if h.CompanyID != c.ID {
			continue
		}
		if h.OpenedYear != g.Date.Year || h.OpenedMonth != g.Date.Month {
			continue
		}
		// 既に路線があるか
		if g.lineExists(h.From, h.To, c.ID) {
			continue
		}
		// 資金チェック（不足分は借金で補う）
		cost := g.LineCost(h.From, h.To, h.Gauge, h.Double)
		if c.Cash < cost {
			needed := cost - c.Cash
			c.Cash += needed
			c.Debt += needed
		}
		// 敷設
		return g.executeBuildLine(c, h.From, h.To, h.Gauge, h.Double, h.DisplayName)
	}
	return nil
}

// lineExists は指定区間に自社路線があるかを返す。
//
// From/To の向きは問わない。
func (g *Game) lineExists(from, to, companyID string) bool {
	c := g.Companies[companyID]
	if c == nil {
		return false
	}
	for _, l := range c.Lines {
		if (l.From == from && l.To == to) || (l.From == to && l.To == from) {
			return true
		}
	}
	return false
}

// executeBuildLine は AI 用の路線敷設を実行する。
//
// BuildLineAction を直接呼ばず、内部処理を行って強制的に敷設する
// （AI は 4 ゲートの Validate をスキップする。史実路線は史実に存在した
// 路線であり、ゲームのゲート事情と無関係に敷設されるため）。
func (g *Game) executeBuildLine(c *Company, from, to string, gauge int, double bool, displayName string) []Event {
	fromCity := g.Cities[from]
	toCity := g.Cities[to]
	if fromCity == nil || toCity == nil {
		return nil
	}
	cost := g.LineCost(from, to, gauge, double)
	c.Cash -= cost

	lineID := g.IDGen.Next()
	line := NewLine(lineID, c.ID, from, to)
	if displayName != "" {
		line.DisplayName = displayName
	} else {
		line.DisplayName = fmt.Sprintf("%s〜%s", fromCity.Name, toCity.Name)
	}
	line.LengthKm = g.DistanceKm(from, to)
	line.Gauge = gauge
	line.Double = double
	line.Opened = g.Date
	line.Fare = g.DefaultFare(from, to)
	line.TrainsPerDay = 10

	c.AddLine(line)

	g.recalcCityDemand(fromCity)
	g.recalcCityDemand(toCity)

	return []Event{
		g.Success(EventLineOpened,
			fmt.Sprintf("%sが%s〜%s間の路線を開業しました（%s）",
				c.Name, fromCity.Name, toCity.Name, FormatYen(cost))).
			With("line_id", line.ID).
			With("from", from).
			With("to", to).
			With("from_name", fromCity.Name).
			With("to_name", toCity.Name).
			With("length_km", line.LengthKm).
			With("gauge", gauge).
			With("cost", cost).
			With("ai", true).
			ForCompany(c.ID).
			ForLine(line.ID),
	}
}
