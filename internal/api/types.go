package api

import "github.com/you/jrt/internal/core"

// このファイルは API クライアントが受け取る DTO 型を定義する。

// GameState はゲームの現在状態の要約。
type GameState struct {
	ID             string        `json:"id"`
	Date           core.Date     `json:"date"`
	Turn           int           `json:"turn"`
	PlayerID       string        `json:"player_id"`
	Company        *core.Company `json:"company,omitempty"`
	Rank           int           `json:"rank"`
	TotalCompanies int           `json:"total_companies"`
}

// LineView は路線とその派生情報（健全度・距離）を表す。
type LineView struct {
	Line     *core.Line `json:"line"`
	Health   string     `json:"health"`
	Distance float64    `json:"distance_km"`
}

// CompanySummary は会社一覧表示用の要約。
type CompanySummary struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Cash       int64  `json:"cash"`
	Debt       int64  `json:"debt"`
	StockPrice int    `json:"stock_price"`
	IsPlayer   bool   `json:"is_player"`
	IsNational bool   `json:"is_national"`
	LineCount  int    `json:"line_count"`
}

// BuildGate は敷設可否 1 ゲートの判定結果。
//
// core.ConstructionGate は JSON タグを持たないため、フィールド名がそのまま
// キーになる点に合わせる。
type BuildGate struct {
	Kind    string         `json:"Kind"`
	Status  string         `json:"Status"`
	Reason  string         `json:"Reason"`
	Details map[string]any `json:"Details"`
}

// CanBuildResult は敷設可否の総合判定。
type CanBuildResult struct {
	Can    bool        `json:"Can"`
	Gates  []BuildGate `json:"Gates"`
	Cost   int64       `json:"Cost"`
	Months int         `json:"Months"`
}
