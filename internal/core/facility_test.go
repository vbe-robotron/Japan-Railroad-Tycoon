package core

import (
	"encoding/json"
	"testing"
)

func TestNewFacility(t *testing.T) {
	tests := []struct {
		name  string
		scale int
		want  int
	}{
		{"最小1", 1, 1},
		{"中央3", 3, 3},
		{"最大5", 5, 5},
		{"0は1にクランプ", 0, 1},
		{"負は1にクランプ", -10, 1},
		{"6は5にクランプ", 6, 5},
		{"100は5にクランプ", 100, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFacility("f1", "hankyu", "osaka", "department_store", tt.scale)
			if f.Scale != tt.want {
				t.Errorf("Scale = %d, want %d", f.Scale, tt.want)
			}
			if f.Income != 0 {
				t.Errorf("Income = %d, want 0", f.Income)
			}
		})
	}
}

func TestNewFacility_Fields(t *testing.T) {
	f := NewFacility("f1", "hankyu", "osaka", "hotel", 4)
	if f.ID != "f1" || f.CompanyID != "hankyu" || f.CityID != "osaka" || f.Kind != "hotel" {
		t.Errorf("フィールドが不正: %+v", f)
	}
}

func TestFacility_JSONRoundTrip(t *testing.T) {
	f := NewFacility("f1", "hankyu", "osaka", "amusement", 5)
	f.BuiltAt = Date{1920, 4}
	f.Income = 50000

	data, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var restored Facility
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if restored != *f {
		t.Errorf("復元後が異なる:\n got %+v\nwant %+v", restored, *f)
	}
	if restored.BuiltAt != (Date{1920, 4}) {
		t.Errorf("BuiltAt = %+v", restored.BuiltAt)
	}
}
