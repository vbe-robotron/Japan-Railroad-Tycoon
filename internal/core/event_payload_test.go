package core

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestPayload_ToEvent(t *testing.T) {
	g := NewGame(1907, "hankyu")
	g.Date = Date{1910, 4}
	g.Turn = 34

	tests := []struct {
		name    string
		ev      Event
		typ     string
		sev     string
		company string
		line    string
		city    string
		payload map[string]any
	}{
		{
			name: "LineConstructionStarted",
			ev: LineConstructionStartedPayload{
				LineID: "L1", CompanyID: "hankyu", From: "umeda", To: "takarazuka",
				FromName: "梅田", ToName: "宝塚", LengthKm: 24.5, Gauge: 1435,
				Double: true, EstimatedCost: 240000, Months: 8,
			}.ToEvent(g),
			typ: EventLineConstructionStarted, sev: SeverityInfo, company: "hankyu", line: "L1",
			payload: map[string]any{
				"line_id": "L1", "company_id": "hankyu", "from": "umeda", "to": "takarazuka",
				"from_name": "梅田", "to_name": "宝塚", "length_km": 24.5, "gauge": 1435,
				"double": true, "estimated_cost": int64(240000), "months": 8,
			},
		},
		{
			name: "LineOpened",
			ev: LineOpenedPayload{
				LineID: "L2", From: "umeda", To: "takarazuka",
				FromName: "梅田", ToName: "宝塚", LengthKm: 24.5, Gauge: 1435,
				Cost: 240000, Months: 8,
			}.ToEvent(g),
			typ: EventLineOpened, sev: SeveritySuccess, company: "hankyu", line: "L2",
			payload: map[string]any{
				"line_id": "L2", "from": "umeda", "to": "takarazuka",
				"from_name": "梅田", "to_name": "宝塚", "length_km": 24.5,
				"gauge": 1435, "cost": int64(240000), "months": 8,
			},
		},
		{
			name: "StationOpened",
			ev: StationOpenedPayload{
				LineID: "L3", CityID: "shonai", CityName: "庄内", Kind: "local", Cost: 30000,
			}.ToEvent(g),
			typ: EventStationOpened, sev: SeveritySuccess, line: "L3", city: "shonai",
			payload: map[string]any{
				"line_id": "L3", "city_id": "shonai", "city_name": "庄内",
				"kind": "local", "cost": int64(30000),
			},
		},
		{
			name: "FacilityBuilt",
			ev: FacilityBuiltPayload{
				FacilityID: "F1", CompanyID: "hankyu", CityID: "takarazuka",
				CityName: "宝塚", Kind: "amusement", Scale: 3, Cost: 500000,
			}.ToEvent(g),
			typ: EventFacilityBuilt, sev: SeveritySuccess, company: "hankyu", city: "takarazuka",
			payload: map[string]any{
				"facility_id": "F1", "company_id": "hankyu", "city_id": "takarazuka",
				"city_name": "宝塚", "kind": "amusement", "scale": 3, "cost": int64(500000),
			},
		},
		{
			name: "Electrified",
			ev: ElectrifiedPayload{
				LineID: "L4", CompanyID: "hankyu", System: "dc1500", Voltage: 1500,
				Cost: 100000, Months: 6,
			}.ToEvent(g),
			typ: EventElectrified, sev: SeveritySuccess, company: "hankyu", line: "L4",
			payload: map[string]any{
				"line_id": "L4", "company_id": "hankyu", "system": "dc1500",
				"voltage": 1500, "cost": int64(100000), "months": 6,
			},
		},
		{
			name: "FareChanged",
			ev: FareChangedPayload{
				LineID: "L5", OldFare: 120, NewFare: 150,
			}.ToEvent(g),
			typ: EventFareChanged, sev: SeverityInfo, line: "L5",
			payload: map[string]any{
				"line_id": "L5", "old_fare": 120, "new_fare": 150,
			},
		},
		{
			name: "LineDeficit",
			ev: LineDeficitPayload{
				LineID: "L6", MonthlyLoss: -12000, AccumulatedLoss: -72000,
				ConsecutiveTurns: 6, ThresholdTurns: 24,
			}.ToEvent(g),
			typ: EventLineDeficit, sev: SeverityWarning, line: "L6",
			payload: map[string]any{
				"line_id": "L6", "monthly_loss": int64(-12000), "accumulated_loss": int64(-72000),
				"consecutive_turns": 6, "threshold_turns": 24,
			},
		},
		{
			name: "LineCritical",
			ev: LineCriticalPayload{
				LineID: "L7", ConsecutiveTurns: 24, AccumulatedLoss: -500000, PoliticalPressure: 0.6,
			}.ToEvent(g),
			typ: EventLineCritical, sev: SeverityDanger, line: "L7",
			payload: map[string]any{
				"line_id": "L7", "consecutive_turns": 24, "accumulated_loss": int64(-500000),
				"political_pressure": 0.6,
			},
		},
		{
			name: "BondsIssued",
			ev: BondsIssuedPayload{
				CompanyID: "hankyu", Amount: 5000000, Rate: 0.05, Years: 10, Purpose: "設備投資",
			}.ToEvent(g),
			typ: EventBondsIssued, sev: SeveritySuccess, company: "hankyu",
			payload: map[string]any{
				"company_id": "hankyu", "amount": int64(5000000), "rate": 0.05,
				"years": 10, "purpose": "設備投資",
			},
		},
		{
			name: "SharesAcquired",
			ev: SharesAcquiredPayload{
				CompanyID: "hankyu", TargetCompanyID: "hanshin", TargetCompanyName: "阪神電鉄",
				Count: 1000, Price: 500, TotalCost: 500000, HoldingPct: 0.15,
			}.ToEvent(g),
			typ: EventSharesAcquired, sev: SeveritySuccess, company: "hankyu",
			payload: map[string]any{
				"company_id": "hankyu", "target_company_id": "hanshin",
				"target_company_name": "阪神電鉄", "count": 1000, "price": 500,
				"total_cost": int64(500000), "holding_pct": 0.15,
			},
		},
		{
			name: "SubsidiaryEstablished",
			ev: SubsidiaryEstablishedPayload{
				ParentCompanyID: "hankyu", SubsidiaryCompanyID: "hanshin",
				SubsidiaryName: "阪神電鉄", HoldingPct: 0.52,
			}.ToEvent(g),
			typ: EventSubsidiaryEstablished, sev: SeveritySuccess, company: "hankyu",
			payload: map[string]any{
				"parent_company_id": "hankyu", "subsidiary_company_id": "hanshin",
				"subsidiary_name": "阪神電鉄", "holding_pct": 0.52,
			},
		},
		{
			name: "TOBLaunched",
			ev: TOBLaunchedPayload{
				TargetCompanyID: "hanshin", TargetCompanyName: "阪神電鉄",
				Premium: 1.1, TargetRatio: 0.51, TotalCost: 5000000, DeadlineTurn: 40,
			}.ToEvent(g),
			typ: EventTOBLaunched, sev: SeverityInfo, company: "hanshin",
			payload: map[string]any{
				"target_company_id": "hanshin", "target_company_name": "阪神電鉄",
				"premium": 1.1, "target_ratio": 0.51, "total_cost": int64(5000000),
				"deadline_turn": 40,
			},
		},
		{
			name: "TOBSuccess",
			ev: TOBSuccessPayload{
				AcquirerCompanyID: "hankyu", TargetCompanyID: "hanshin",
				TargetCompanyName: "阪神電鉄", AcquiredRatio: 0.52,
				TotalCost: 5000000, NewHoldingPct: 0.52,
			}.ToEvent(g),
			typ: EventTOBSuccess, sev: SeveritySuccess, company: "hankyu",
			payload: map[string]any{
				"acquirer_company_id": "hankyu", "target_company_id": "hanshin",
				"target_company_name": "阪神電鉄", "acquired_ratio": 0.52,
				"total_cost": int64(5000000), "new_holding_pct": 0.52,
			},
		},
		{
			name: "MergeCompleted",
			ev: MergeCompletedPayload{
				NewCompanyID: "hankyu_hanshin_hd", NewCompanyName: "阪急阪神HD",
				FromCompanyID: "hankyu", FromName: "阪急電鉄",
				ToCompanyID: "hanshin", ToName: "阪神電鉄", Ratio: 0.5,
			}.ToEvent(g),
			typ: EventMergeCompleted, sev: SeveritySuccess, company: "hankyu_hanshin_hd",
			payload: map[string]any{
				"new_company_id": "hankyu_hanshin_hd", "new_company_name": "阪急阪神HD",
				"from_company_id": "hankyu", "from_name": "阪急電鉄",
				"to_company_id": "hanshin", "to_name": "阪神電鉄", "ratio": 0.5,
			},
		},
		{
			name: "LineAbandoned",
			ev: LineAbandonedPayload{
				LineID: "L8", FromName: "梅田", ToName: "宝塚", Reason: "economic", SalvageValue: 120000,
			}.ToEvent(g),
			typ: EventLineAbandoned, sev: SeverityWarning, line: "L8",
			payload: map[string]any{
				"line_id": "L8", "from_name": "梅田", "to_name": "宝塚",
				"reason": "economic", "salvage_value": int64(120000),
			},
		},
		{
			name: "LineDamaged",
			ev: LineDamagedPayload{
				LineID: "L9", Kind: "earthquake", DamageRate: 0.35,
				ConditionAfter: 0.65, Cause: "関東大震災",
			}.ToEvent(g),
			typ: EventLineDamaged, sev: SeverityDanger, line: "L9",
			payload: map[string]any{
				"line_id": "L9", "kind": "earthquake", "damage_rate": 0.35,
				"condition_after": 0.65, "cause": "関東大震災",
			},
		},
		{
			name: "LandAcquired",
			ev: LandAcquiredPayload{
				CompanyID: "hankyu", CityID: "takarazuka", CityName: "宝塚", Area: 1200.5, Cost: 800000,
			}.ToEvent(g),
			typ: EventLandAcquired, sev: SeveritySuccess, company: "hankyu", city: "takarazuka",
			payload: map[string]any{
				"company_id": "hankyu", "city_id": "takarazuka", "city_name": "宝塚",
				"area": 1200.5, "cost": int64(800000),
			},
		},
		{
			name: "CityGrown",
			ev: CityGrownPayload{
				CityID: "takarazuka", CityName: "宝塚", OldPopulation: 50000,
				NewPopulation: 80000, Developed: 0.4, Cause: "沿線開発",
			}.ToEvent(g),
			typ: EventCityGrown, sev: SeveritySuccess, city: "takarazuka",
			payload: map[string]any{
				"city_id": "takarazuka", "city_name": "宝塚", "old_population": 50000,
				"new_population": 80000, "developed": 0.4, "cause": "沿線開発",
			},
		},
		{
			name: "SubsidyReceived",
			ev: SubsidyReceivedPayload{
				CompanyID: "hankyu", CityID: "takarazuka", Amount: 1000000,
				Conditions: []string{"keep_local_lines"},
			}.ToEvent(g),
			typ: EventSubsidyReceived, sev: SeveritySuccess, company: "hankyu", city: "takarazuka",
			payload: map[string]any{
				"company_id": "hankyu", "city_id": "takarazuka",
				"amount": int64(1000000), "conditions": []string{"keep_local_lines"},
			},
		},
		{
			name: "Disaster",
			ev: DisasterPayload{
				Kind: "earthquake", Region: "kanto", AffectedLines: []string{"L1", "L2"},
				DamageRate: 0.35, Casualties: 105000,
			}.ToEvent(g),
			typ: EventDisaster, sev: SeverityDanger,
			payload: map[string]any{
				"kind": "earthquake", "region": "kanto", "affected_lines": []string{"L1", "L2"},
				"damage_rate": 0.35, "casualties": 105000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ev.Type != tt.typ {
				t.Errorf("Type = %q, want %q", tt.ev.Type, tt.typ)
			}
			if tt.ev.Severity != tt.sev {
				t.Errorf("Severity = %q, want %q", tt.ev.Severity, tt.sev)
			}
			if tt.ev.ID == "" {
				t.Error("ID が空です")
			}
			if len(tt.ev.ID) != nanoIDLength {
				t.Errorf("len(ID) = %d, want %d", len(tt.ev.ID), nanoIDLength)
			}
			if tt.ev.Message == "" {
				t.Error("Message が空です")
			}
			if tt.ev.Date != g.Date {
				t.Errorf("Date = %+v, want %+v", tt.ev.Date, g.Date)
			}
			if tt.ev.Turn != g.Turn {
				t.Errorf("Turn = %d, want %d", tt.ev.Turn, g.Turn)
			}
			if tt.company != "" && tt.ev.CompanyID != tt.company {
				t.Errorf("CompanyID = %q, want %q", tt.ev.CompanyID, tt.company)
			}
			if tt.line != "" && tt.ev.LineID != tt.line {
				t.Errorf("LineID = %q, want %q", tt.ev.LineID, tt.line)
			}
			if tt.city != "" && tt.ev.CityID != tt.city {
				t.Errorf("CityID = %q, want %q", tt.ev.CityID, tt.city)
			}
			for k, want := range tt.payload {
				got, ok := tt.ev.Payload[k]
				if !ok {
					t.Errorf("Payload に %q がありません", k)
					continue
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("Payload[%q] = %#v, want %#v", k, got, want)
				}
			}
			if len(tt.ev.Payload) != len(tt.payload) {
				t.Errorf("Payload の数 = %d, want %d", len(tt.ev.Payload), len(tt.payload))
			}
		})
	}
}

func TestPayload_JSONRoundTrip(t *testing.T) {
	payloads := []struct {
		name string
		v    any
	}{
		{"LineConstructionStarted", LineConstructionStartedPayload{
			LineID: "L1", CompanyID: "hankyu", From: "umeda", To: "takarazuka",
			FromName: "梅田", ToName: "宝塚", LengthKm: 24.5, Gauge: 1435,
			Double: true, EstimatedCost: 240000, Months: 8,
		}},
		{"LineOpened", LineOpenedPayload{
			LineID: "L2", From: "umeda", To: "takarazuka", FromName: "梅田",
			ToName: "宝塚", LengthKm: 24.5, Gauge: 1435, Cost: 240000, Months: 8,
		}},
		{"StationOpened", StationOpenedPayload{
			LineID: "L3", CityID: "shonai", CityName: "庄内", Kind: "local", Cost: 30000,
		}},
		{"FacilityBuilt", FacilityBuiltPayload{
			FacilityID: "F1", CompanyID: "hankyu", CityID: "takarazuka",
			CityName: "宝塚", Kind: "amusement", Scale: 3, Cost: 500000,
		}},
		{"Electrified", ElectrifiedPayload{
			LineID: "L4", CompanyID: "hankyu", System: "dc1500", Voltage: 1500,
			Cost: 100000, Months: 6,
		}},
		{"FareChanged", FareChangedPayload{LineID: "L5", OldFare: 120, NewFare: 150}},
		{"LineDeficit", LineDeficitPayload{
			LineID: "L6", MonthlyLoss: -12000, AccumulatedLoss: -72000,
			ConsecutiveTurns: 6, ThresholdTurns: 24,
		}},
		{"LineCritical", LineCriticalPayload{
			LineID: "L7", ConsecutiveTurns: 24, AccumulatedLoss: -500000, PoliticalPressure: 0.6,
		}},
		{"BondsIssued", BondsIssuedPayload{
			CompanyID: "hankyu", Amount: 5000000, Rate: 0.05, Years: 10, Purpose: "設備投資",
		}},
		{"SharesAcquired", SharesAcquiredPayload{
			CompanyID: "hankyu", TargetCompanyID: "hanshin", TargetCompanyName: "阪神電鉄",
			Count: 1000, Price: 500, TotalCost: 500000, HoldingPct: 0.15,
		}},
		{"SubsidiaryEstablished", SubsidiaryEstablishedPayload{
			ParentCompanyID: "hankyu", SubsidiaryCompanyID: "hanshin",
			SubsidiaryName: "阪神電鉄", HoldingPct: 0.52,
		}},
		{"TOBLaunched", TOBLaunchedPayload{
			TargetCompanyID: "hanshin", TargetCompanyName: "阪神電鉄",
			Premium: 1.1, TargetRatio: 0.51, TotalCost: 5000000, DeadlineTurn: 40,
		}},
		{"TOBSuccess", TOBSuccessPayload{
			AcquirerCompanyID: "hankyu", TargetCompanyID: "hanshin",
			TargetCompanyName: "阪神電鉄", AcquiredRatio: 0.52,
			TotalCost: 5000000, NewHoldingPct: 0.52,
		}},
		{"MergeCompleted", MergeCompletedPayload{
			NewCompanyID: "hankyu_hanshin_hd", NewCompanyName: "阪急阪神HD",
			FromCompanyID: "hankyu", FromName: "阪急電鉄",
			ToCompanyID: "hanshin", ToName: "阪神電鉄", Ratio: 0.5,
		}},
		{"LineAbandoned", LineAbandonedPayload{
			LineID: "L8", FromName: "梅田", ToName: "宝塚", Reason: "economic", SalvageValue: 120000,
		}},
		{"LineDamaged", LineDamagedPayload{
			LineID: "L9", Kind: "earthquake", DamageRate: 0.35,
			ConditionAfter: 0.65, Cause: "関東大震災",
		}},
		{"LandAcquired", LandAcquiredPayload{
			CompanyID: "hankyu", CityID: "takarazuka", CityName: "宝塚", Area: 1200.5, Cost: 800000,
		}},
		{"CityGrown", CityGrownPayload{
			CityID: "takarazuka", CityName: "宝塚", OldPopulation: 50000,
			NewPopulation: 80000, Developed: 0.4, Cause: "沿線開発",
		}},
		{"SubsidyReceived", SubsidyReceivedPayload{
			CompanyID: "hankyu", CityID: "takarazuka", Amount: 1000000,
			Conditions: []string{"keep_local_lines"},
		}},
		{"Disaster", DisasterPayload{
			Kind: "earthquake", Region: "kanto", AffectedLines: []string{"L1", "L2"},
			DamageRate: 0.35, Casualties: 105000,
		}},
	}

	for _, tt := range payloads {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.v)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}

			// 同じ型のゼロ値に復元して比較する。
			out := reflect.New(reflect.TypeOf(tt.v)).Interface()
			if err := json.Unmarshal(data, out); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if !reflect.DeepEqual(reflect.ValueOf(out).Elem().Interface(), tt.v) {
				t.Errorf("JSON 往復で一致しません:\n got %#v\nwant %#v",
					reflect.ValueOf(out).Elem().Interface(), tt.v)
			}
		})
	}
}
