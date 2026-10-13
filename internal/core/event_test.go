package core

import (
	"encoding/json"
	"testing"
)

func TestNewEvent(t *testing.T) {
	g := NewGame(1907, "hankyu")
	g.Date = Date{1910, 4}
	g.Turn = 34

	e := g.NewEvent(EventLineOpened, SeveritySuccess, "開業しました")

	if len(e.ID) != nanoIDLength {
		t.Errorf("len(ID) = %d, want %d", len(e.ID), nanoIDLength)
	}
	if e.Type != EventLineOpened {
		t.Errorf("Type = %q, want %q", e.Type, EventLineOpened)
	}
	if e.Date != (Date{1910, 4}) {
		t.Errorf("Date = %+v, want {1910 4}", e.Date)
	}
	if e.Turn != 34 {
		t.Errorf("Turn = %d, want 34", e.Turn)
	}
	if e.Severity != SeveritySuccess {
		t.Errorf("Severity = %q, want %q", e.Severity, SeveritySuccess)
	}
	if e.Message != "開業しました" {
		t.Errorf("Message = %q, want %q", e.Message, "開業しました")
	}
	if e.Payload != nil {
		t.Errorf("Payload = %v, want nil", e.Payload)
	}
}

func TestEvent_With(t *testing.T) {
	g := NewGame(1907, "hankyu")

	e := g.Info("test", "msg").
		With("a", 1).
		With("b", "x").
		With("c", 1.5)

	if len(e.Payload) != 3 {
		t.Fatalf("Payload の長さ = %d, want 3", len(e.Payload))
	}
	if e.Payload["a"] != 1 {
		t.Errorf("Payload[a] = %v, want 1", e.Payload["a"])
	}
	if e.Payload["b"] != "x" {
		t.Errorf("Payload[b] = %v, want x", e.Payload["b"])
	}
	if e.Payload["c"] != 1.5 {
		t.Errorf("Payload[c] = %v, want 1.5", e.Payload["c"])
	}
}

func TestEvent_With_Immutable(t *testing.T) {
	g := NewGame(1907, "hankyu")

	base := g.Info("test", "msg")
	if base.Payload != nil {
		t.Fatalf("base.Payload = %v, want nil", base.Payload)
	}

	derived := base.With("a", 1)
	if base.Payload != nil {
		t.Errorf("With が元の Payload を変更した: %v", base.Payload)
	}
	if derived.Payload["a"] != 1 {
		t.Errorf("derived.Payload[a] = %v, want 1", derived.Payload["a"])
	}

	derived2 := derived.With("b", 2)
	if _, ok := derived.Payload["b"]; ok {
		t.Errorf("チェーンで元の Event が変更された: %v", derived.Payload)
	}
	if len(derived.Payload) != 1 {
		t.Errorf("derived.Payload の長さ = %d, want 1", len(derived.Payload))
	}
	if len(derived2.Payload) != 2 {
		t.Errorf("derived2.Payload の長さ = %d, want 2", len(derived2.Payload))
	}

	// 既存 Payload への上書きも元の Event に影響しない。
	over := derived2.With("a", 99)
	if derived2.Payload["a"] != 1 {
		t.Errorf("上書きが元の Event に波及した: %v", derived2.Payload["a"])
	}
	if over.Payload["a"] != 99 {
		t.Errorf("over.Payload[a] = %v, want 99", over.Payload["a"])
	}
}

func TestEvent_ForCompany_ForLine_ForCity(t *testing.T) {
	g := NewGame(1907, "hankyu")

	e := g.Info("test", "msg").
		ForCompany("c1").
		ForLine("l1").
		ForCity("city1")

	if e.CompanyID != "c1" {
		t.Errorf("CompanyID = %q, want c1", e.CompanyID)
	}
	if e.LineID != "l1" {
		t.Errorf("LineID = %q, want l1", e.LineID)
	}
	if e.CityID != "city1" {
		t.Errorf("CityID = %q, want city1", e.CityID)
	}
}

func TestGame_Info_Success_Warn_Danger(t *testing.T) {
	tests := []struct {
		name string
		fn   func(g *Game, t, msg string) Event
		want string
	}{
		{"Info", (*Game).Info, SeverityInfo},
		{"Success", (*Game).Success, SeveritySuccess},
		{"Warn", (*Game).Warn, SeverityWarning},
		{"Danger", (*Game).Danger, SeverityDanger},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGame(1907, "hankyu")
			e := tt.fn(g, "test", "メッセージ")
			if e.Severity != tt.want {
				t.Errorf("Severity = %q, want %q", e.Severity, tt.want)
			}
			if e.Type != "test" {
				t.Errorf("Type = %q, want test", e.Type)
			}
			if e.Message != "メッセージ" {
				t.Errorf("Message = %q, want メッセージ", e.Message)
			}
			if !IsValidSeverity(e.Severity) {
				t.Errorf("Severity %q が不正", e.Severity)
			}
		})
	}
}

func TestGame_Evt(t *testing.T) {
	g := NewGame(1907, "hankyu")

	e := g.Evt(EventFareChanged, SeverityWarning, "運賃を%d円から%d円に変更", 120, 150)

	if e.Type != EventFareChanged {
		t.Errorf("Type = %q, want %q", e.Type, EventFareChanged)
	}
	if e.Severity != SeverityWarning {
		t.Errorf("Severity = %q, want %q", e.Severity, SeverityWarning)
	}
	if e.Message != "運賃を120円から150円に変更" {
		t.Errorf("Message = %q, want %q", e.Message, "運賃を120円から150円に変更")
	}

	// 引数なしでも動作する。
	e2 := g.Evt("t", SeverityInfo, "そのまま")
	if e2.Message != "そのまま" {
		t.Errorf("Message = %q, want そのまま", e2.Message)
	}
}

func TestEventID_Deterministic(t *testing.T) {
	g1 := NewGame(1907, "hankyu")
	g2 := NewGame(1907, "hankyu")

	for i := 0; i < 100; i++ {
		e1 := g1.Info("test", "msg")
		e2 := g2.Info("test", "msg")
		if e1.ID != e2.ID {
			t.Fatalf("iteration %d: %s != %s", i, e1.ID, e2.ID)
		}
	}
}

func TestEventID_ContinuesAfterJSONRoundTrip(t *testing.T) {
	g := NewGame(1907, "hankyu")
	g.Date = Date{1910, 4}
	g.Turn = 34

	// 往復前に 1 件生成して Counter を進める。
	first := g.NewEvent(EventGameStarted, SeverityInfo, "開始")

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var restored Game
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	restored.EnsureRng()

	if restored.IDGen == nil {
		t.Fatal("IDGen が nil")
	}

	// 復元後も同じイベント ID 列が続く。
	next1 := g.NewEvent(EventTurnEnded, SeverityInfo, "次")
	next2 := restored.NewEvent(EventTurnEnded, SeverityInfo, "次")
	if next1.ID != next2.ID {
		t.Errorf("復元後のイベント ID が一致しない: %s != %s", next1.ID, next2.ID)
	}
	if next1.ID == first.ID {
		t.Error("イベント ID が重複した")
	}

	// Payload と CompanyID が JSON 往復しても保持される。
	rich := g.Success(EventLineOpened, "開業").
		With("line_id", "L1").
		ForCompany("hankyu").
		ForLine("L1")
	data2, err := json.Marshal(rich)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var restoredEvent Event
	if err := json.Unmarshal(data2, &restoredEvent); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if restoredEvent.Payload["line_id"] != "L1" {
		t.Errorf("Payload[line_id] = %v, want L1", restoredEvent.Payload["line_id"])
	}
	if restoredEvent.CompanyID != "hankyu" {
		t.Errorf("CompanyID = %q, want hankyu", restoredEvent.CompanyID)
	}
	if restoredEvent.LineID != "L1" {
		t.Errorf("LineID = %q, want L1", restoredEvent.LineID)
	}
}
