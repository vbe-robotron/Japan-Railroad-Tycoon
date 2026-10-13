package core

import "fmt"

// このファイルでは主要 20 種のイベントについて Payload を構造体化する。
//
// 各構造体は ToEvent(g *Game) Event を持ち、日本語メッセージの生成と
// Payload の詰め込みを一括で行う。それ以外のイベントは呼び出し側で
// g.Info(...).With(...) のように map[string]any を直接組み立てる。

// LineConstructionStartedPayload は line_construction_started の Payload。
type LineConstructionStartedPayload struct {
	LineID        string  `json:"line_id"`
	CompanyID     string  `json:"company_id"`
	From          string  `json:"from"`
	To            string  `json:"to"`
	FromName      string  `json:"from_name"`
	ToName        string  `json:"to_name"`
	LengthKm      float64 `json:"length_km"`
	Gauge         int     `json:"gauge"`
	Double        bool    `json:"double"`
	EstimatedCost int64   `json:"estimated_cost"`
	Months        int     `json:"months"`
}

// ToEvent は Payload を Event に変換する。
func (p LineConstructionStartedPayload) ToEvent(g *Game) Event {
	return g.Info(EventLineConstructionStarted,
		fmt.Sprintf("%s〜%s間の路線敷設を開始しました（%s）",
			p.FromName, p.ToName, FormatYen(p.EstimatedCost))).
		With("line_id", p.LineID).
		With("company_id", p.CompanyID).
		With("from", p.From).
		With("to", p.To).
		With("from_name", p.FromName).
		With("to_name", p.ToName).
		With("length_km", p.LengthKm).
		With("gauge", p.Gauge).
		With("double", p.Double).
		With("estimated_cost", p.EstimatedCost).
		With("months", p.Months).
		ForCompany(p.CompanyID).
		ForLine(p.LineID)
}

// LineOpenedPayload は line_opened の Payload スキーマ。
type LineOpenedPayload struct {
	LineID   string  `json:"line_id"`
	From     string  `json:"from"`
	To       string  `json:"to"`
	FromName string  `json:"from_name"`
	ToName   string  `json:"to_name"`
	LengthKm float64 `json:"length_km"`
	Gauge    int     `json:"gauge"`
	Cost     int64   `json:"cost"`
	Months   int     `json:"months"`
}

// ToEvent は Payload を Event に変換する。
func (p LineOpenedPayload) ToEvent(g *Game) Event {
	return g.Success(EventLineOpened,
		fmt.Sprintf("%s〜%s間の路線が開業しました（%s）",
			p.FromName, p.ToName, FormatYen(p.Cost))).
		With("line_id", p.LineID).
		With("from", p.From).
		With("to", p.To).
		With("from_name", p.FromName).
		With("to_name", p.ToName).
		With("length_km", p.LengthKm).
		With("gauge", p.Gauge).
		With("cost", p.Cost).
		With("months", p.Months).
		ForCompany(g.PlayerID).
		ForLine(p.LineID)
}

// StationOpenedPayload は station_opened の Payload。
type StationOpenedPayload struct {
	LineID   string `json:"line_id"`
	CityID   string `json:"city_id"`
	CityName string `json:"city_name"`
	Kind     string `json:"kind"`
	Cost     int64  `json:"cost"`
}

// ToEvent は Payload を Event に変換する。
func (p StationOpenedPayload) ToEvent(g *Game) Event {
	return g.Success(EventStationOpened,
		fmt.Sprintf("%s駅が開業しました（%s、%s）",
			p.CityName, p.Kind, FormatYen(p.Cost))).
		With("line_id", p.LineID).
		With("city_id", p.CityID).
		With("city_name", p.CityName).
		With("kind", p.Kind).
		With("cost", p.Cost).
		ForLine(p.LineID).
		ForCity(p.CityID)
}

// FacilityBuiltPayload は facility_built の Payload。
type FacilityBuiltPayload struct {
	FacilityID string `json:"facility_id"`
	CompanyID  string `json:"company_id"`
	CityID     string `json:"city_id"`
	CityName   string `json:"city_name"`
	Kind       string `json:"kind"`
	Scale      int    `json:"scale"`
	Cost       int64  `json:"cost"`
}

// ToEvent は Payload を Event に変換する。
func (p FacilityBuiltPayload) ToEvent(g *Game) Event {
	return g.Success(EventFacilityBuilt,
		fmt.Sprintf("%sに%sを建設しました（規模%d、%s）",
			p.CityName, p.Kind, p.Scale, FormatYen(p.Cost))).
		With("facility_id", p.FacilityID).
		With("company_id", p.CompanyID).
		With("city_id", p.CityID).
		With("city_name", p.CityName).
		With("kind", p.Kind).
		With("scale", p.Scale).
		With("cost", p.Cost).
		ForCompany(p.CompanyID).
		ForCity(p.CityID)
}

// ElectrifiedPayload は electrified の Payload。
type ElectrifiedPayload struct {
	LineID    string `json:"line_id"`
	CompanyID string `json:"company_id"`
	System    string `json:"system"`
	Voltage   int    `json:"voltage"`
	Cost      int64  `json:"cost"`
	Months    int    `json:"months"`
}

// ToEvent は Payload を Event に変換する。
func (p ElectrifiedPayload) ToEvent(g *Game) Event {
	return g.Success(EventElectrified,
		fmt.Sprintf("路線が電化されました（%s・%dV、%s）",
			p.System, p.Voltage, FormatYen(p.Cost))).
		With("line_id", p.LineID).
		With("company_id", p.CompanyID).
		With("system", p.System).
		With("voltage", p.Voltage).
		With("cost", p.Cost).
		With("months", p.Months).
		ForCompany(p.CompanyID).
		ForLine(p.LineID)
}

// FareChangedPayload は fare_changed の Payload。
type FareChangedPayload struct {
	LineID  string `json:"line_id"`
	OldFare int    `json:"old_fare"`
	NewFare int    `json:"new_fare"`
}

// ToEvent は Payload を Event に変換する。
func (p FareChangedPayload) ToEvent(g *Game) Event {
	return g.Info(EventFareChanged,
		fmt.Sprintf("運賃を%d円から%d円に変更しました", p.OldFare, p.NewFare)).
		With("line_id", p.LineID).
		With("old_fare", p.OldFare).
		With("new_fare", p.NewFare).
		ForLine(p.LineID)
}

// LineDeficitPayload は line_deficit の Payload。
type LineDeficitPayload struct {
	LineID           string `json:"line_id"`
	MonthlyLoss      int64  `json:"monthly_loss"`
	AccumulatedLoss  int64  `json:"accumulated_loss"`
	ConsecutiveTurns int    `json:"consecutive_turns"`
	ThresholdTurns   int    `json:"threshold_turns"`
}

// ToEvent は Payload を Event に変換する。
func (p LineDeficitPayload) ToEvent(g *Game) Event {
	return g.Warn(EventLineDeficit,
		fmt.Sprintf("%dヶ月連続で赤字です（月次%s、累積%s）",
			p.ConsecutiveTurns, FormatYen(p.MonthlyLoss), FormatYen(p.AccumulatedLoss))).
		With("line_id", p.LineID).
		With("monthly_loss", p.MonthlyLoss).
		With("accumulated_loss", p.AccumulatedLoss).
		With("consecutive_turns", p.ConsecutiveTurns).
		With("threshold_turns", p.ThresholdTurns).
		ForLine(p.LineID)
}

// LineCriticalPayload は line_critical の Payload。
type LineCriticalPayload struct {
	LineID            string  `json:"line_id"`
	ConsecutiveTurns  int     `json:"consecutive_turns"`
	AccumulatedLoss   int64   `json:"accumulated_loss"`
	PoliticalPressure float64 `json:"political_pressure"`
}

// ToEvent は Payload を Event に変換する。
func (p LineCriticalPayload) ToEvent(g *Game) Event {
	return g.Danger(EventLineCritical,
		fmt.Sprintf("路線が廃線の危機にあります（%dヶ月連続赤字、累積%s）",
			p.ConsecutiveTurns, FormatYen(p.AccumulatedLoss))).
		With("line_id", p.LineID).
		With("consecutive_turns", p.ConsecutiveTurns).
		With("accumulated_loss", p.AccumulatedLoss).
		With("political_pressure", p.PoliticalPressure).
		ForLine(p.LineID)
}

// BondsIssuedPayload は bonds_issued の Payload。
type BondsIssuedPayload struct {
	CompanyID string  `json:"company_id"`
	Amount    int64   `json:"amount"`
	Rate      float64 `json:"rate"`
	Years     int     `json:"years"`
	Purpose   string  `json:"purpose"`
}

// ToEvent は Payload を Event に変換する。
func (p BondsIssuedPayload) ToEvent(g *Game) Event {
	return g.Success(EventBondsIssued,
		fmt.Sprintf("社債%sを発行しました（年利%.1f%%、%d年、%s）",
			FormatYen(p.Amount), p.Rate*100, p.Years, p.Purpose)).
		With("company_id", p.CompanyID).
		With("amount", p.Amount).
		With("rate", p.Rate).
		With("years", p.Years).
		With("purpose", p.Purpose).
		ForCompany(p.CompanyID)
}

// SharesAcquiredPayload は shares_acquired の Payload。
type SharesAcquiredPayload struct {
	CompanyID         string  `json:"company_id"`
	TargetCompanyID   string  `json:"target_company_id"`
	TargetCompanyName string  `json:"target_company_name"`
	Count             int     `json:"count"`
	Price             int     `json:"price"`
	TotalCost         int64   `json:"total_cost"`
	HoldingPct        float64 `json:"holding_pct"`
}

// ToEvent は Payload を Event に変換する。
func (p SharesAcquiredPayload) ToEvent(g *Game) Event {
	return g.Success(EventSharesAcquired,
		fmt.Sprintf("%sの株式%d株を取得しました（%s、保有%.1f%%）",
			p.TargetCompanyName, p.Count, FormatYen(p.TotalCost), p.HoldingPct*100)).
		With("company_id", p.CompanyID).
		With("target_company_id", p.TargetCompanyID).
		With("target_company_name", p.TargetCompanyName).
		With("count", p.Count).
		With("price", p.Price).
		With("total_cost", p.TotalCost).
		With("holding_pct", p.HoldingPct).
		ForCompany(p.CompanyID)
}

// SubsidiaryEstablishedPayload は subsidiary_established の Payload。
type SubsidiaryEstablishedPayload struct {
	ParentCompanyID     string  `json:"parent_company_id"`
	SubsidiaryCompanyID string  `json:"subsidiary_company_id"`
	SubsidiaryName      string  `json:"subsidiary_name"`
	HoldingPct          float64 `json:"holding_pct"`
}

// ToEvent は Payload を Event に変換する。
func (p SubsidiaryEstablishedPayload) ToEvent(g *Game) Event {
	return g.Success(EventSubsidiaryEstablished,
		fmt.Sprintf("%sを子会社化しました（保有%.1f%%）",
			p.SubsidiaryName, p.HoldingPct*100)).
		With("parent_company_id", p.ParentCompanyID).
		With("subsidiary_company_id", p.SubsidiaryCompanyID).
		With("subsidiary_name", p.SubsidiaryName).
		With("holding_pct", p.HoldingPct).
		ForCompany(p.ParentCompanyID)
}

// TOBLaunchedPayload は tob_launched の Payload。
type TOBLaunchedPayload struct {
	TargetCompanyID   string  `json:"target_company_id"`
	TargetCompanyName string  `json:"target_company_name"`
	Premium           float64 `json:"premium"`
	TargetRatio       float64 `json:"target_ratio"`
	TotalCost         int64   `json:"total_cost"`
	DeadlineTurn      int     `json:"deadline_turn"`
}

// ToEvent は Payload を Event に変換する。
func (p TOBLaunchedPayload) ToEvent(g *Game) Event {
	return g.Info(EventTOBLaunched,
		fmt.Sprintf("%sへのTOBを開始しました（対象%.1f%%、%s）",
			p.TargetCompanyName, p.TargetRatio*100, FormatYen(p.TotalCost))).
		With("target_company_id", p.TargetCompanyID).
		With("target_company_name", p.TargetCompanyName).
		With("premium", p.Premium).
		With("target_ratio", p.TargetRatio).
		With("total_cost", p.TotalCost).
		With("deadline_turn", p.DeadlineTurn).
		ForCompany(p.TargetCompanyID)
}

// TOBSuccessPayload は tob_success の Payload。
type TOBSuccessPayload struct {
	AcquirerCompanyID string  `json:"acquirer_company_id"`
	TargetCompanyID   string  `json:"target_company_id"`
	TargetCompanyName string  `json:"target_company_name"`
	AcquiredRatio     float64 `json:"acquired_ratio"`
	TotalCost         int64   `json:"total_cost"`
	NewHoldingPct     float64 `json:"new_holding_pct"`
}

// ToEvent は Payload を Event に変換する。
func (p TOBSuccessPayload) ToEvent(g *Game) Event {
	return g.Success(EventTOBSuccess,
		fmt.Sprintf("%sへのTOBが成立しました（取得%.1f%%、%s）",
			p.TargetCompanyName, p.AcquiredRatio*100, FormatYen(p.TotalCost))).
		With("acquirer_company_id", p.AcquirerCompanyID).
		With("target_company_id", p.TargetCompanyID).
		With("target_company_name", p.TargetCompanyName).
		With("acquired_ratio", p.AcquiredRatio).
		With("total_cost", p.TotalCost).
		With("new_holding_pct", p.NewHoldingPct).
		ForCompany(p.AcquirerCompanyID)
}

// MergeCompletedPayload は merge_completed の Payload。
type MergeCompletedPayload struct {
	NewCompanyID   string  `json:"new_company_id"`
	NewCompanyName string  `json:"new_company_name"`
	FromCompanyID  string  `json:"from_company_id"`
	FromName       string  `json:"from_name"`
	ToCompanyID    string  `json:"to_company_id"`
	ToName         string  `json:"to_name"`
	Ratio          float64 `json:"ratio"`
}

// ToEvent は Payload を Event に変換する。
func (p MergeCompletedPayload) ToEvent(g *Game) Event {
	return g.Success(EventMergeCompleted,
		fmt.Sprintf("%sと%sが合併し、%sが発足しました",
			p.FromName, p.ToName, p.NewCompanyName)).
		With("new_company_id", p.NewCompanyID).
		With("new_company_name", p.NewCompanyName).
		With("from_company_id", p.FromCompanyID).
		With("from_name", p.FromName).
		With("to_company_id", p.ToCompanyID).
		With("to_name", p.ToName).
		With("ratio", p.Ratio).
		ForCompany(p.NewCompanyID)
}

// LineAbandonedPayload は line_abandoned の Payload。
type LineAbandonedPayload struct {
	LineID       string `json:"line_id"`
	FromName     string `json:"from_name"`
	ToName       string `json:"to_name"`
	Reason       string `json:"reason"`
	SalvageValue int64  `json:"salvage_value"`
}

// ToEvent は Payload を Event に変換する。
func (p LineAbandonedPayload) ToEvent(g *Game) Event {
	return g.Warn(EventLineAbandoned,
		fmt.Sprintf("%s〜%s間の路線を廃止しました（理由: %s、売却益%s）",
			p.FromName, p.ToName, p.Reason, FormatYen(p.SalvageValue))).
		With("line_id", p.LineID).
		With("from_name", p.FromName).
		With("to_name", p.ToName).
		With("reason", p.Reason).
		With("salvage_value", p.SalvageValue).
		ForLine(p.LineID)
}

// LineDamagedPayload は line_damaged の Payload。
type LineDamagedPayload struct {
	LineID         string  `json:"line_id"`
	Kind           string  `json:"kind"`
	DamageRate     float64 `json:"damage_rate"`
	ConditionAfter float64 `json:"condition_after"`
	Cause          string  `json:"cause"`
}

// ToEvent は Payload を Event に変換する。
func (p LineDamagedPayload) ToEvent(g *Game) Event {
	return g.Danger(EventLineDamaged,
		fmt.Sprintf("路線が%sにより被害を受けました（被害率%.0f%%、健全度%.2f）",
			p.Kind, p.DamageRate*100, p.ConditionAfter)).
		With("line_id", p.LineID).
		With("kind", p.Kind).
		With("damage_rate", p.DamageRate).
		With("condition_after", p.ConditionAfter).
		With("cause", p.Cause).
		ForLine(p.LineID)
}

// LandAcquiredPayload は land_acquired の Payload。
type LandAcquiredPayload struct {
	CompanyID string  `json:"company_id"`
	CityID    string  `json:"city_id"`
	CityName  string  `json:"city_name"`
	Area      float64 `json:"area"`
	Cost      int64   `json:"cost"`
}

// ToEvent は Payload を Event に変換する。
func (p LandAcquiredPayload) ToEvent(g *Game) Event {
	return g.Success(EventLandAcquired,
		fmt.Sprintf("%sの用地を取得しました（面積%.1f、%s）",
			p.CityName, p.Area, FormatYen(p.Cost))).
		With("company_id", p.CompanyID).
		With("city_id", p.CityID).
		With("city_name", p.CityName).
		With("area", p.Area).
		With("cost", p.Cost).
		ForCompany(p.CompanyID).
		ForCity(p.CityID)
}

// CityGrownPayload は city_grown の Payload。
type CityGrownPayload struct {
	CityID        string  `json:"city_id"`
	CityName      string  `json:"city_name"`
	OldPopulation int     `json:"old_population"`
	NewPopulation int     `json:"new_population"`
	Developed     float64 `json:"developed"`
	Cause         string  `json:"cause"`
}

// ToEvent は Payload を Event に変換する。
func (p CityGrownPayload) ToEvent(g *Game) Event {
	return g.Success(EventCityGrown,
		fmt.Sprintf("%sの人口が%d人から%d人に増加しました（開発度%.2f）",
			p.CityName, p.OldPopulation, p.NewPopulation, p.Developed)).
		With("city_id", p.CityID).
		With("city_name", p.CityName).
		With("old_population", p.OldPopulation).
		With("new_population", p.NewPopulation).
		With("developed", p.Developed).
		With("cause", p.Cause).
		ForCity(p.CityID)
}

// SubsidyReceivedPayload は subsidy_received の Payload。
type SubsidyReceivedPayload struct {
	CompanyID  string   `json:"company_id"`
	CityID     string   `json:"city_id"`
	Amount     int64    `json:"amount"`
	Conditions []string `json:"conditions"`
}

// ToEvent は Payload を Event に変換する。
func (p SubsidyReceivedPayload) ToEvent(g *Game) Event {
	return g.Success(EventSubsidyReceived,
		fmt.Sprintf("自治体から補助金%sを受領しました", FormatYen(p.Amount))).
		With("company_id", p.CompanyID).
		With("city_id", p.CityID).
		With("amount", p.Amount).
		With("conditions", p.Conditions).
		ForCompany(p.CompanyID).
		ForCity(p.CityID)
}

// DisasterPayload は disaster の Payload。
type DisasterPayload struct {
	Kind          string   `json:"kind"`
	Region        string   `json:"region"`
	AffectedLines []string `json:"affected_lines"`
	DamageRate    float64  `json:"damage_rate"`
	Casualties    int      `json:"casualties"`
}

// ToEvent は Payload を Event に変換する。
func (p DisasterPayload) ToEvent(g *Game) Event {
	return g.Danger(EventDisaster,
		fmt.Sprintf("%sで%sが発生しました（被害率%.0f%%、犠牲者%d人）",
			p.Region, p.Kind, p.DamageRate*100, p.Casualties)).
		With("kind", p.Kind).
		With("region", p.Region).
		With("affected_lines", p.AffectedLines).
		With("damage_rate", p.DamageRate).
		With("casualties", p.Casualties)
}
