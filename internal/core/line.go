package core

// Line は 1 本の路線を表す。
//
// From/To は City.ID を指す。営業係数や連続赤字ターン数など、
// Phase 1-7 の廃線判定に必要な状態を保持する。
type Line struct {
	ID                      string  `json:"id"`
	CompanyID               string  `json:"company_id"`
	From                    string  `json:"from"`         // CityID
	To                      string  `json:"to"`           // CityID
	DisplayName             string  `json:"display_name"` // "宝塚線"
	LengthKm                float64 `json:"length_km"`
	Gauge                   int     `json:"gauge"` // 762|1067|1435
	Double                  bool    `json:"double"`
	Electrified             bool    `json:"electrified"`
	Fare                    int     `json:"fare"` // 円
	TrainsPerDay            int     `json:"trains_per_day"`
	Opened                  Date    `json:"opened"`
	Profit                  int64   `json:"profit"`          // 直近月次
	OperatingRatio          float64 `json:"operating_ratio"` // 営業係数
	Condition               float64 `json:"condition"`       // 0.0-1.0
	ConsecutiveDeficitTurns int     `json:"consecutive_deficit_turns"`
	AccumulatedLoss         int64   `json:"accumulated_loss"`
	PoliticalPressure       float64 `json:"political_pressure"` // 0.0-1.0
}

// NewLine は路線を生成する。
//
// 軌間は日本私鉄で最も一般的な狭軌（1067mm）を既定とし、状態は健全
// （Condition=1.0）、政治的圧力はなし（0.0）で初期化する。
func NewLine(id, companyID, from, to string) *Line {
	return &Line{
		ID:                id,
		CompanyID:         companyID,
		From:              from,
		To:                to,
		Gauge:             1067,
		Condition:         1.0,
		PoliticalPressure: 0.0,
	}
}

// Health は路線の健全度を表す文字列を返す。
//
// 連続赤字ターン数に基づく簡易判定。Phase 1-7 で営業係数や政治的圧力を
// 加味した判定に精緻化する予定。
func (l *Line) Health() string {
	switch {
	case l.ConsecutiveDeficitTurns <= 5:
		return "safe"
	case l.ConsecutiveDeficitTurns <= 11:
		return "caution"
	case l.ConsecutiveDeficitTurns <= 23:
		return "danger"
	default:
		return "critical"
	}
}
