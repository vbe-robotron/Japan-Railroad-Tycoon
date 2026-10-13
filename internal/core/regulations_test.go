package core

import "testing"

func TestRegulationByID(t *testing.T) {
	r, ok := RegulationByID("private_railway_ordinance")
	if !ok {
		t.Fatal("private_railway_ordinance が見つからない")
	}
	if r.Name != "私設鉄道条例" {
		t.Errorf("Name = %q, want %q", r.Name, "私設鉄道条例")
	}
	if r.EnactedYear != 1887 {
		t.Errorf("EnactedYear = %d, want 1887", r.EnactedYear)
	}
	if len(r.Effects) == 0 || r.Effects[0] != "private_railway_allowed" {
		t.Errorf("Effects = %v, want 先頭が private_railway_allowed", r.Effects)
	}

	if _, ok := RegulationByID("does_not_exist"); ok {
		t.Error("未知の ID で ok = true になった")
	}
}

func TestEffectiveRegulations(t *testing.T) {
	tests := []struct {
		year int
		want int
	}{
		{1886, 0},
		{1887, 1},
		{1892, 2},
		{1900, 3},
		{1906, 4},
		{1919, 5},
		{1943, 6},
		{1949, 7},
		{1987, 8},
		{2000, 8},
	}
	for _, tt := range tests {
		got := EffectiveRegulations(tt.year)
		if len(got) != tt.want {
			t.Errorf("EffectiveRegulations(%d) の件数 = %d, want %d", tt.year, len(got), tt.want)
		}
		for _, r := range got {
			if r.EnactedYear > tt.year {
				t.Errorf("EffectiveRegulations(%d) に未制定 %s が含まれる", tt.year, r.ID)
			}
		}
	}
}

func TestIsRegulationActive(t *testing.T) {
	if IsRegulationActive("private_railway_ordinance", 1886) {
		t.Error("1886年で私設鉄道条例が有効と判定された")
	}
	if !IsRegulationActive("private_railway_ordinance", 1887) {
		t.Error("1887年で私設鉄道条例が無効と判定された")
	}
	if !IsRegulationActive("private_railway_ordinance", 2000) {
		t.Error("2000年で私設鉄道条例が無効と判定された")
	}
	if IsRegulationActive("does_not_exist", 2000) {
		t.Error("未知の法規が有効と判定された")
	}
}

func TestRegulationTable_Unique(t *testing.T) {
	seen := map[string]bool{}
	for _, r := range RegulationTable {
		if r.ID == "" {
			t.Error("空の ID の法規がある")
		}
		if r.Name == "" {
			t.Errorf("%s: Name が空", r.ID)
		}
		if r.EnactedYear <= 0 {
			t.Errorf("%s: EnactedYear が不正: %d", r.ID, r.EnactedYear)
		}
		if seen[r.ID] {
			t.Errorf("ID が重複: %s", r.ID)
		}
		seen[r.ID] = true
	}
}
