package core

// Event はゲーム内で発生した出来事を表す。
//
// Phase 1-2 で本実装する。ここでは Game.Events の型を確定させるための
// 最小限の定義のみを置く。
type Event struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Date      Date           `json:"date"`
	Turn      int            `json:"turn"`
	Severity  string         `json:"severity"`
	Message   string         `json:"message"`
	Payload   map[string]any `json:"payload,omitempty"`
	CompanyID string         `json:"company_id,omitempty"`
	CityID    string         `json:"city_id,omitempty"`
	LineID    string         `json:"line_id,omitempty"`
}
