package core

import "testing"

// allActions はテストで使用する全 25 種のアクションのゼロ値インスタンスを返す。
func allActions() []Action {
	return []Action{
		&BuildLineAction{},
		&BuildStationAction{},
		&BuildFacilityAction{},
		&BuildDepotAction{},
		&ElectrifyAction{},
		&PurchaseRollingStockAction{},
		&SetFareAction{},
		&SetScheduleAction{},
		&IssueBondsAction{},
		&IssueSharesAction{},
		&PayDividendAction{},
		&RepayDebtAction{},
		&BuySharesAction{},
		&TenderOfferAction{},
		&MergeAction{},
		&AllianceAction{},
		&DefendTakeoverAction{},
		&AcquireLandAction{},
		&LobbyAction{},
		&NationalizeAction{},
		&SplitJnrAction{},
		&AcceptSubsidyAction{},
		&AbandonLineAction{},
		&RestoreAbandonedLineAction{},
		&EndTurnAction{},
	}
}

// TestActionTypes_AllTypes は全 25 種が空でない Type() を返し、重複がないことを検証する。
func TestActionTypes_AllTypes(t *testing.T) {
	seen := map[string]bool{}
	for _, a := range allActions() {
		typ := a.Type()
		if typ == "" {
			t.Errorf("%T: empty Type()", a)
		}
		if seen[typ] {
			t.Errorf("duplicate Type(): %s", typ)
		}
		seen[typ] = true
	}
	if len(seen) != 25 {
		t.Errorf("expected 25 action types, got %d", len(seen))
	}
}

// TestActionTypes_Unique は Type() 文字列の重複がないことを検証する。
func TestActionTypes_Unique(t *testing.T) {
	seen := map[string]Action{}
	for _, a := range allActions() {
		typ := a.Type()
		if prev, ok := seen[typ]; ok {
			t.Errorf("duplicate Type() %q: %T と %T", typ, prev, a)
			continue
		}
		seen[typ] = a
	}
}
