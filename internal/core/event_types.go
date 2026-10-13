package core

// イベントタイプの定数。Event.Type には必ずこれらのいずれかを設定する。
//
// イベントは Type で分類し、Payload で詳細を伝える。60 種以上のタイプを
// 定義しており、AllEventTypes/IsValidEventType で一覧・検証できる。

// 建設系。
const (
	EventLineConstructionStarted = "line_construction_started"
	EventLineOpened              = "line_opened"
	EventStationOpened           = "station_opened"
	EventFacilityBuilt           = "facility_built"
	EventDepotBuilt              = "depot_built"
	EventElectrified             = "electrified"
	EventConstructionCompleted   = "construction_completed"
)

// 運用系。
const (
	EventFareChanged     = "fare_changed"
	EventScheduleChanged = "schedule_changed"
	EventStockPurchased  = "stock_purchased"
	EventLineProfitable  = "line_profitable"
	EventLineDeficit     = "line_deficit"
	EventLineCritical    = "line_critical"
)

// 財務系。
const (
	EventBondsIssued    = "bonds_issued"
	EventSharesIssued   = "shares_issued"
	EventDividendPaid   = "dividend_paid"
	EventDebtRepaid     = "debt_repaid"
	EventBankruptcyRisk = "bankruptcy_risk"
	EventBankruptcy     = "bankruptcy"
	EventStockPriceUp   = "stock_price_up"
	EventStockPriceDown = "stock_price_down"
)

// M&A系。
const (
	EventSharesAcquired        = "shares_acquired"
	EventSubsidiaryEstablished = "subsidiary_established"
	EventTOBLaunched           = "tob_launched"
	EventTOBSuccess            = "tob_success"
	EventTOBFailed             = "tob_failed"
	EventMergeCompleted        = "merge_completed"
	EventCompanyMerged         = "company_merged"
	EventTakeoverDefended      = "takeover_defended"
	EventTakeoverDefeated      = "takeover_defeated"
	EventAllianceFormed        = "alliance_formed"
	EventHoldingEstablished    = "holding_established"
)

// 廃線系。
const (
	EventLineAbandoned = "line_abandoned"
	EventLineRestored  = "line_restored"
	EventLineDamaged   = "line_damaged"
	EventLineDestroyed = "line_destroyed"
)

// 開発系。
const (
	EventLandAcquired        = "land_acquired"
	EventCityGrown           = "city_grown"
	EventPopulationIncreased = "population_increased"
	EventLandValueUp         = "land_value_up"
	EventFacilityEffect      = "facility_effect"
)

// 政治系。
const (
	EventLobbySuccess          = "lobby_success"
	EventLobbyFailed           = "lobby_failed"
	EventSubsidyReceived       = "subsidy_received"
	EventRegulationChanged     = "regulation_changed"
	EventPoliticalPressureUp   = "political_pressure_up"
	EventPoliticalPressureDown = "political_pressure_down"
	EventNationalized          = "nationalized"
	EventPrivatized            = "privatized"
)

// 歴史系。
const (
	EventHistoricalEvent = "historical_event"
	EventDisaster        = "disaster"
	EventWarEffect       = "war_effect"
	EventEconomicBoom    = "economic_boom"
	EventEconomicCrisis  = "economic_crisis"
)

// ターン系。
const (
	EventTurnEnded   = "turn_ended"
	EventYearChanged = "year_changed"
	EventRankChanged = "rank_changed"
)

// システム系。
const (
	EventGameStarted   = "game_started"
	EventGameEnded     = "game_ended"
	EventVictory       = "victory"
	EventDefeat        = "defeat"
	EventSaveCompleted = "save_completed"
)

// allEventTypes は定義済みの全イベントタイプを、カテゴリ順に列挙したもの。
//
// このスライスが唯一の情報源であり、allEventTypeSet と AllEventTypes は
// ここから導出する。追加漏れを避けるため、定数を増やしたら必ず追記する。
var allEventTypes = []string{
	// 建設系
	EventLineConstructionStarted,
	EventLineOpened,
	EventStationOpened,
	EventFacilityBuilt,
	EventDepotBuilt,
	EventElectrified,
	EventConstructionCompleted,
	// 運用系
	EventFareChanged,
	EventScheduleChanged,
	EventStockPurchased,
	EventLineProfitable,
	EventLineDeficit,
	EventLineCritical,
	// 財務系
	EventBondsIssued,
	EventSharesIssued,
	EventDividendPaid,
	EventDebtRepaid,
	EventBankruptcyRisk,
	EventBankruptcy,
	EventStockPriceUp,
	EventStockPriceDown,
	// M&A系
	EventSharesAcquired,
	EventSubsidiaryEstablished,
	EventTOBLaunched,
	EventTOBSuccess,
	EventTOBFailed,
	EventMergeCompleted,
	EventCompanyMerged,
	EventTakeoverDefended,
	EventTakeoverDefeated,
	EventAllianceFormed,
	EventHoldingEstablished,
	// 廃線系
	EventLineAbandoned,
	EventLineRestored,
	EventLineDamaged,
	EventLineDestroyed,
	// 開発系
	EventLandAcquired,
	EventCityGrown,
	EventPopulationIncreased,
	EventLandValueUp,
	EventFacilityEffect,
	// 政治系
	EventLobbySuccess,
	EventLobbyFailed,
	EventSubsidyReceived,
	EventRegulationChanged,
	EventPoliticalPressureUp,
	EventPoliticalPressureDown,
	EventNationalized,
	EventPrivatized,
	// 歴史系
	EventHistoricalEvent,
	EventDisaster,
	EventWarEffect,
	EventEconomicBoom,
	EventEconomicCrisis,
	// ターン系
	EventTurnEnded,
	EventYearChanged,
	EventRankChanged,
	// システム系
	EventGameStarted,
	EventGameEnded,
	EventVictory,
	EventDefeat,
	EventSaveCompleted,
}

// allEventTypeSet は IsValidEventType を O(1) で判定するための集合。
var allEventTypeSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(allEventTypes))
	for _, t := range allEventTypes {
		m[t] = struct{}{}
	}
	return m
}()

// AllEventTypes は全イベントタイプのスライスのコピーを返す。
//
// 呼び出し側が返り値を変更しても内部状態に影響しないよう、毎回コピーする。
func AllEventTypes() []string {
	out := make([]string, len(allEventTypes))
	copy(out, allEventTypes)
	return out
}

// IsValidEventType は s が既知のイベントタイプかどうかを返す。
func IsValidEventType(s string) bool {
	_, ok := allEventTypeSet[s]
	return ok
}
