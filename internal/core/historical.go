package core

// このファイルは core 内の史実路線型を定義する。
//
// data パッケージが historical_lines.json を読み込み、この型に変換して
// Game へ注入する（core は data に依存しないため）。

// HistoricalLine は史実の路線を表す。
//
// data パッケージから注入される参照データであり、AI の史実追従に使う。
type HistoricalLine struct {
	CompanyID   string `json:"company_id"`
	From        string `json:"from"`
	To          string `json:"to"`
	Gauge       int    `json:"gauge"`
	Double      bool   `json:"double"`
	OpenedYear  int    `json:"opened_year"`
	OpenedMonth int    `json:"opened_month"`
	DisplayName string `json:"display_name"`
}

// SetHistoricalLines は史実路線データを設定する。
func (g *Game) SetHistoricalLines(lines []HistoricalLine) {
	g.HistoricalLines = lines
}

// AutoBuildHistoricalLine は史実の既存路線をコストなしで敷設する。
//
// ゲーム開始時点で史実の鉄道網を再現するために使う。
// イベントは生成しない（初期状態）。
func (g *Game) AutoBuildHistoricalLine(c *Company, h HistoricalLine) {
	if c == nil {
		return
	}
	from := g.Cities[h.From]
	to := g.Cities[h.To]
	if from == nil || to == nil {
		return
	}
	if g.lineExists(h.From, h.To, c.ID) {
		return
	}

	lineID := g.IDGen.Next()
	line := NewLine(lineID, c.ID, h.From, h.To)
	line.DisplayName = h.DisplayName
	line.LengthKm = g.DistanceKm(h.From, h.To)
	line.Gauge = h.Gauge
	line.Double = h.Double
	line.Opened = Date{Year: h.OpenedYear, Month: h.OpenedMonth}
	line.Fare = g.DefaultFare(h.From, h.To)
	line.TrainsPerDay = 10

	c.AddLine(line)
	g.recalcCityDemand(from)
	g.recalcCityDemand(to)
}
