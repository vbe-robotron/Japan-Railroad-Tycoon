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
