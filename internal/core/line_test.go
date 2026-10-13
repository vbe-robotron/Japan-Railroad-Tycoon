package core

import (
	"encoding/json"
	"testing"
)

func TestNewLine(t *testing.T) {
	l := NewLine("l1", "hankyu", "osaka", "kobe")

	if l.ID != "l1" || l.CompanyID != "hankyu" {
		t.Errorf("ID/CompanyID = %q/%q", l.ID, l.CompanyID)
	}
	if l.From != "osaka" || l.To != "kobe" {
		t.Errorf("From/To = %q/%q", l.From, l.To)
	}
	if l.Gauge != 1067 {
		t.Errorf("Gauge = %d, want 1067", l.Gauge)
	}
	if l.Condition != 1.0 {
		t.Errorf("Condition = %v, want 1.0", l.Condition)
	}
	if l.PoliticalPressure != 0.0 {
		t.Errorf("PoliticalPressure = %v, want 0.0", l.PoliticalPressure)
	}
}

func TestLine_Health(t *testing.T) {
	tests := []struct {
		name  string
		turns int
		want  string
	}{
		{"0はsafe", 0, "safe"},
		{"5はsafe", 5, "safe"},
		{"6はcaution", 6, "caution"},
		{"11はcaution", 11, "caution"},
		{"12はdanger", 12, "danger"},
		{"23はdanger", 23, "danger"},
		{"24はcritical", 24, "critical"},
		{"100はcritical", 100, "critical"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLine("l", "c", "a", "b")
			l.ConsecutiveDeficitTurns = tt.turns
			if got := l.Health(); got != tt.want {
				t.Errorf("Health() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLine_JSONRoundTrip(t *testing.T) {
	l := NewLine("l1", "hankyu", "osaka", "kobe")
	l.DisplayName = "神戸線"
	l.LengthKm = 32.5
	l.Gauge = 1435
	l.Double = true
	l.Electrified = true
	l.Fare = 300
	l.TrainsPerDay = 120
	l.Opened = Date{1907, 8}
	l.Profit = 123456
	l.OperatingRatio = 85.5
	l.Condition = 0.8
	l.ConsecutiveDeficitTurns = 13
	l.AccumulatedLoss = 9999
	l.PoliticalPressure = 0.4

	data, err := json.Marshal(l)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var restored Line
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if restored != *l {
		t.Errorf("復元後が異なる:\n got %+v\nwant %+v", restored, *l)
	}
	if restored.Opened != (Date{1907, 8}) {
		t.Errorf("Opened = %+v", restored.Opened)
	}
}
