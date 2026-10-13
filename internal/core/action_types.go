package core

// このファイルは 25 種のアクション構造体と Type() メソッドを定義する。
//
// アクションは 建設・運用・財務・M&A・開発・政治・廃線・ターン の
// 7 カテゴリに分類される。Validate/Apply の実装は action_stubs.go に置き、
// Phase 1-4 以降で 1 つずつ本実装へ移行する。
//
// 各構造体の JSON タグはアクションログ・リプレイの往復互換を保証する。
// Type() の文字列は DecodeAction のレジストリキーと一致させること。

// ────────────────────────────────────────────
// 建設系（5種）
// ────────────────────────────────────────────

// BuildLineAction は路線を敷設する。
type BuildLineAction struct {
	From   string `json:"from"`   // CityID
	To     string `json:"to"`     // CityID
	Gauge  int    `json:"gauge"`  // 762|1067|1435
	Double bool   `json:"double"` // 複線か
}

// BuildStationAction は既存路線に中間駅を追加する。
type BuildStationAction struct {
	LineID   string  `json:"line_id"`
	CityID   string  `json:"city_id"`
	Position float64 `json:"position"` // 0.0=From, 1.0=To
	Kind     string  `json:"kind"`     // "local"|"express"|"terminal"
}

// BuildFacilityAction は開発施設を建設する。
type BuildFacilityAction struct {
	CityID string `json:"city_id"`
	Kind   string `json:"kind"`  // "department_store"|"housing"|"amusement"|"hotel"|"office"
	Scale  int    `json:"scale"` // 1-5
}

// BuildDepotAction は車庫・工場・発電所を建設する。
type BuildDepotAction struct {
	CityID string `json:"city_id"`
	Kind   string `json:"kind"` // "depot"|"factory"|"power_plant"
}

// ElectrifyAction は路線を電化する。
type ElectrifyAction struct {
	LineID string `json:"line_id"`
	System string `json:"system"` // "dc1500"|"ac20000"|"dc600"
}

// ────────────────────────────────────────────
// 運用系（3種）
// ────────────────────────────────────────────

// PurchaseRollingStockAction は車両を購入する。
type PurchaseRollingStockAction struct {
	LineID string `json:"line_id"`
	Count  int    `json:"count"`
	Kind   string `json:"kind"` // "steam"|"electric"|"diesel"|"shinkansen"
}

// SetFareAction は運賃を変更する。
type SetFareAction struct {
	LineID string `json:"line_id"`
	Fare   int    `json:"fare"` // 円
}

// SetScheduleAction はダイヤ・本数を変更する。
type SetScheduleAction struct {
	LineID       string `json:"line_id"`
	TrainsPerDay int    `json:"trains_per_day"`
}

// ────────────────────────────────────────────
// 財務系（4種）
// ────────────────────────────────────────────

// IssueBondsAction は社債を発行する。
type IssueBondsAction struct {
	Amount int64   `json:"amount"`
	Rate   float64 `json:"rate"` // 年利
	Years  int     `json:"years"`
}

// IssueSharesAction は増資する。
type IssueSharesAction struct {
	Count int `json:"count"`
}

// PayDividendAction は配当を支払う。
type PayDividendAction struct {
	PerShare int `json:"per_share"`
}

// RepayDebtAction は借金を返済する。
type RepayDebtAction struct {
	Amount int64 `json:"amount"`
}

// ────────────────────────────────────────────
// M&A系（5種）
// ────────────────────────────────────────────

// BuySharesAction は他社株を取得する。
type BuySharesAction struct {
	TargetCompanyID string `json:"target_company_id"`
	Count           int    `json:"count"`
}

// TenderOfferAction は TOB を仕掛ける。
type TenderOfferAction struct {
	TargetCompanyID string  `json:"target_company_id"`
	Premium         float64 `json:"premium"`      // 1.1 = 10%上乗せ
	TargetRatio     float64 `json:"target_ratio"` // 0.51
}

// MergeAction は他社と合併する。
type MergeAction struct {
	TargetCompanyID string  `json:"target_company_id"`
	Ratio           float64 `json:"ratio"`
	Name            string  `json:"name"` // 新社名
}

// AllianceAction は他社と提携する。
type AllianceAction struct {
	TargetCompanyID string `json:"target_company_id"`
	Kind            string `json:"kind"` // "through_service"|"fare_integration"|"mutual_holding"
}

// DefendTakeoverAction は買収防衛策を講じる。
type DefendTakeoverAction struct {
	Kind string `json:"kind"` // "poison_pill"|"white_knight"|"buyback"
}

// ────────────────────────────────────────────
// 開発系（2種）
// ────────────────────────────────────────────

// AcquireLandAction は用地を取得する。
type AcquireLandAction struct {
	CityID string  `json:"city_id"`
	Amount float64 `json:"amount"` // km²
}

// LobbyAction は政治家に工作する。
type LobbyAction struct {
	CityID string `json:"city_id"`
	Kind   string `json:"kind"` // "route_approval"|"subsidy"|"fare_hike"
}

// ────────────────────────────────────────────
// 政治系（3種）
// ────────────────────────────────────────────

// NationalizeAction は対象会社を国有化する。
type NationalizeAction struct {
	TargetCompanyID string `json:"target_company_id"`
	Compensation    int64  `json:"compensation"`
}

// SplitJnrAction は国鉄を分割民営化する。
type SplitJnrAction struct {
	Companies []string `json:"companies"`
}

// AcceptSubsidyAction は補助金を受諾する。
type AcceptSubsidyAction struct {
	Amount     int64    `json:"amount"`
	Conditions []string `json:"conditions"`
}

// ────────────────────────────────────────────
// 廃線系（2種）
// ────────────────────────────────────────────

// AbandonLineAction は路線を廃止する。
type AbandonLineAction struct {
	LineID string `json:"line_id"`
	Reason string `json:"reason"` // "economic"|"disaster"|"political"|"wartime"
}

// RestoreAbandonedLineAction は廃線を復活させる。
type RestoreAbandonedLineAction struct {
	LineID string `json:"line_id"`
}

// ────────────────────────────────────────────
// ターン系（1種）
// ────────────────────────────────────────────

// EndTurnAction はターンを終了する。
type EndTurnAction struct{}

// ────────────────────────────────────────────
// Type() メソッド
// ────────────────────────────────────────────

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *BuildLineAction) Type() string { return "build_line" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *BuildStationAction) Type() string { return "build_station" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *BuildFacilityAction) Type() string { return "build_facility" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *BuildDepotAction) Type() string { return "build_depot" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *ElectrifyAction) Type() string { return "electrify" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *PurchaseRollingStockAction) Type() string { return "purchase_stock" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *SetFareAction) Type() string { return "set_fare" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *SetScheduleAction) Type() string { return "set_schedule" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *IssueBondsAction) Type() string { return "issue_bonds" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *IssueSharesAction) Type() string { return "issue_shares" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *PayDividendAction) Type() string { return "pay_dividend" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *RepayDebtAction) Type() string { return "repay_debt" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *BuySharesAction) Type() string { return "buy_shares" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *TenderOfferAction) Type() string { return "tender_offer" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *MergeAction) Type() string { return "merge" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *AllianceAction) Type() string { return "alliance" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *DefendTakeoverAction) Type() string { return "defend_takeover" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *AcquireLandAction) Type() string { return "acquire_land" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *LobbyAction) Type() string { return "lobby" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *NationalizeAction) Type() string { return "nationalize" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *SplitJnrAction) Type() string { return "split_jnr" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *AcceptSubsidyAction) Type() string { return "accept_subsidy" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *AbandonLineAction) Type() string { return "abandon_line" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *RestoreAbandonedLineAction) Type() string { return "restore_line" }

// Type はアクションタイプ文字列（レジストリキー）を返す。
func (a *EndTurnAction) Type() string { return "end_turn" }
