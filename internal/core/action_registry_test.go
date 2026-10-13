package core

import (
	"encoding/json"
	"reflect"
	"testing"
)

// filledActions は全 25 種を有意な値で埋めたアクションのテーブルを返す。
//
// レジストリのデコード・JSON 往復テストで、フィールドが正しく
// バインドされることを確認するために使用する。
func filledActions() []Action {
	return []Action{
		&BuildLineAction{From: "osaka", To: "kobe", Gauge: 1435, Double: true},
		&BuildStationAction{LineID: "L1", CityID: "kyoto", Position: 0.5, Kind: "express"},
		&BuildFacilityAction{CityID: "osaka", Kind: "department_store", Scale: 3},
		&BuildDepotAction{CityID: "kobe", Kind: "factory"},
		&ElectrifyAction{LineID: "L1", System: "dc1500"},
		&PurchaseRollingStockAction{LineID: "L1", Count: 4, Kind: "electric"},
		&SetFareAction{LineID: "L1", Fare: 320},
		&SetScheduleAction{LineID: "L1", TrainsPerDay: 12},
		&IssueBondsAction{Amount: 1000000, Rate: 0.035, Years: 10},
		&IssueSharesAction{Count: 500},
		&PayDividendAction{PerShare: 5},
		&RepayDebtAction{Amount: 250000},
		&BuySharesAction{TargetCompanyID: "nankai", Count: 100},
		&TenderOfferAction{TargetCompanyID: "nankai", Premium: 1.1, TargetRatio: 0.51},
		&MergeAction{TargetCompanyID: "nankai", Ratio: 0.5, Name: "南海電鉄"},
		&AllianceAction{TargetCompanyID: "hankyu", Kind: "fare_integration"},
		&DefendTakeoverAction{Kind: "poison_pill"},
		&AcquireLandAction{CityID: "osaka", Amount: 2.5},
		&LobbyAction{CityID: "osaka", Kind: "subsidy"},
		&NationalizeAction{TargetCompanyID: "jnr", Compensation: 50000000},
		&SplitJnrAction{Companies: []string{"jr_east", "jr_west"}},
		&AcceptSubsidyAction{Amount: 300000, Conditions: []string{"keep_local"}},
		&AbandonLineAction{LineID: "L1", Reason: "economic"},
		&RestoreAbandonedLineAction{LineID: "L1"},
		&EndTurnAction{},
	}
}

// TestDecodeAction_AllTypes は全 25 種が JSON から復元できることを検証する。
func TestDecodeAction_AllTypes(t *testing.T) {
	for _, want := range filledActions() {
		t.Run(want.Type(), func(t *testing.T) {
			payload, err := json.Marshal(want)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			got, err := DecodeAction(want.Type(), payload)
			if err != nil {
				t.Fatalf("DecodeAction: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("DecodeAction = %#v, want %#v", got, want)
			}
		})
	}
}

// TestDecodeAction_EmptyPayload は payload が空のときゼロ値を返すことを検証する。
func TestDecodeAction_EmptyPayload(t *testing.T) {
	for _, want := range allActions() {
		t.Run(want.Type(), func(t *testing.T) {
			got, err := DecodeAction(want.Type(), nil)
			if err != nil {
				t.Fatalf("DecodeAction: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("DecodeAction = %#v, want %#v", got, want)
			}
		})
	}
}

// TestDecodeAction_Unknown は未知の type でエラーになることを検証する。
func TestDecodeAction_Unknown(t *testing.T) {
	_, err := DecodeAction("does_not_exist", json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("未知の type でエラーにならなかった")
	}
}

// TestDecodeAction_InvalidPayload は不正な JSON でエラーになることを検証する。
func TestDecodeAction_InvalidPayload(t *testing.T) {
	tests := []struct {
		name       string
		actionType string
		payload    string
	}{
		{"壊れたJSON", "build_line", `{invalid`},
		{"型不一致", "build_line", `{"gauge": "not-an-int"}`},
		{"配列を渡す", "end_turn", `[1,2,3]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeAction(tt.actionType, json.RawMessage(tt.payload))
			if err == nil {
				t.Fatalf("不正な payload でエラーにならなかった: %s", tt.payload)
			}
		})
	}
}

// TestDecodeAction_RoundTrip はエンコード→デコードで同じ値になることを検証する。
func TestDecodeAction_RoundTrip(t *testing.T) {
	for _, want := range filledActions() {
		t.Run(want.Type(), func(t *testing.T) {
			data, err := json.Marshal(want)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			got, err := DecodeAction(want.Type(), data)
			if err != nil {
				t.Fatalf("DecodeAction: %v", err)
			}
			gotJSON, err := json.Marshal(got)
			if err != nil {
				t.Fatalf("Marshal(got): %v", err)
			}
			if string(data) != string(gotJSON) {
				t.Errorf("round trip: %s != %s", data, gotJSON)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("round trip の値が一致しない: %#v != %#v", got, want)
			}
		})
	}
}

// TestAction_JSONRoundTrip は全 25 種がエンコード→デコード可能なことを検証する。
func TestAction_JSONRoundTrip(t *testing.T) {
	for _, want := range allActions() {
		t.Run(want.Type(), func(t *testing.T) {
			data, err := json.Marshal(want)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			got, err := DecodeAction(want.Type(), data)
			if err != nil {
				t.Fatalf("DecodeAction: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("round trip: %#v != %#v", got, want)
			}
		})
	}
}

// TestAllActionTypes は 25 種がソート済みで返ることを検証する。
func TestAllActionTypes(t *testing.T) {
	got := AllActionTypes()
	if len(got) != 25 {
		t.Fatalf("len(AllActionTypes()) = %d, want 25", len(got))
	}
	for i := 1; i < len(got); i++ {
		if got[i-1] > got[i] {
			t.Errorf("ソートされていない: %q > %q", got[i-1], got[i])
		}
	}
	// レジストリと一致することを確認する。
	set := map[string]bool{}
	for _, a := range allActions() {
		set[a.Type()] = true
	}
	for _, typ := range got {
		if !set[typ] {
			t.Errorf("未登録の type が含まれる: %s", typ)
		}
	}
}

// TestIsValidActionType は既知/未知の判定を検証する。
func TestIsValidActionType(t *testing.T) {
	for _, a := range allActions() {
		if !IsValidActionType(a.Type()) {
			t.Errorf("IsValidActionType(%q) = false, want true", a.Type())
		}
	}
	if IsValidActionType("does_not_exist") {
		t.Error("未知の type が valid と判定された")
	}
	if IsValidActionType("") {
		t.Error("空文字が valid と判定された")
	}
}
