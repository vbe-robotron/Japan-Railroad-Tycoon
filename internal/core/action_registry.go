package core

import (
	"encoding/json"
	"fmt"
	"sort"
)

// actionFactory はアクションのゼロ値を生成する関数。
type actionFactory func() Action

// actionRegistry は type 文字列からアクションを復元するためのレジストリ。
//
// Mod で新アクションを追加する場合はここに登録する。
var actionRegistry = map[string]actionFactory{
	"build_line":      func() Action { return &BuildLineAction{} },
	"build_station":   func() Action { return &BuildStationAction{} },
	"build_facility":  func() Action { return &BuildFacilityAction{} },
	"build_depot":     func() Action { return &BuildDepotAction{} },
	"electrify":       func() Action { return &ElectrifyAction{} },
	"purchase_stock":  func() Action { return &PurchaseRollingStockAction{} },
	"set_fare":        func() Action { return &SetFareAction{} },
	"set_schedule":    func() Action { return &SetScheduleAction{} },
	"issue_bonds":     func() Action { return &IssueBondsAction{} },
	"issue_shares":    func() Action { return &IssueSharesAction{} },
	"pay_dividend":    func() Action { return &PayDividendAction{} },
	"repay_debt":      func() Action { return &RepayDebtAction{} },
	"buy_shares":      func() Action { return &BuySharesAction{} },
	"tender_offer":    func() Action { return &TenderOfferAction{} },
	"merge":           func() Action { return &MergeAction{} },
	"alliance":        func() Action { return &AllianceAction{} },
	"defend_takeover": func() Action { return &DefendTakeoverAction{} },
	"acquire_land":    func() Action { return &AcquireLandAction{} },
	"lobby":           func() Action { return &LobbyAction{} },
	"nationalize":     func() Action { return &NationalizeAction{} },
	"split_jnr":       func() Action { return &SplitJnrAction{} },
	"accept_subsidy":  func() Action { return &AcceptSubsidyAction{} },
	"abandon_line":    func() Action { return &AbandonLineAction{} },
	"restore_line":    func() Action { return &RestoreAbandonedLineAction{} },
	"end_turn":        func() Action { return &EndTurnAction{} },
}

// DecodeAction はアクションタイプと JSON ペイロードから Action を復元する。
//
// payload が空の場合はゼロ値のアクションを返す。未知の type や不正な
// JSON の場合はエラーを返す。
func DecodeAction(actionType string, payload json.RawMessage) (Action, error) {
	factory, ok := actionRegistry[actionType]
	if !ok {
		return nil, fmt.Errorf("unknown action type: %s", actionType)
	}
	a := factory()
	if len(payload) > 0 {
		if err := json.Unmarshal(payload, a); err != nil {
			return nil, fmt.Errorf("invalid payload for %s: %w", actionType, err)
		}
	}
	return a, nil
}

// AllActionTypes は全アクションタイプをソート済みで返す。
func AllActionTypes() []string {
	types := make([]string, 0, len(actionRegistry))
	for t := range actionRegistry {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

// IsValidActionType は s が既知のアクションタイプかどうかを返す。
func IsValidActionType(s string) bool {
	_, ok := actionRegistry[s]
	return ok
}
