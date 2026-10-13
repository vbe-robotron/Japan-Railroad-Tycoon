package core

import "testing"

func TestTechnologyByID(t *testing.T) {
	tech, ok := TechnologyByID("electric_tram")
	if !ok {
		t.Fatal("electric_tram が見つからない")
	}
	if tech.Name != "電気軌道" {
		t.Errorf("Name = %q, want %q", tech.Name, "電気軌道")
	}
	if tech.UnlockYear != 1895 {
		t.Errorf("UnlockYear = %d, want 1895", tech.UnlockYear)
	}
	if tech.Kind != "traction" {
		t.Errorf("Kind = %q, want %q", tech.Kind, "traction")
	}

	if _, ok := TechnologyByID("does_not_exist"); ok {
		t.Error("未知の ID で ok = true になった")
	}
	if _, ok := TechnologyByID(""); ok {
		t.Error("空 ID で ok = true になった")
	}
}

func TestUnlockedTechnologies(t *testing.T) {
	tests := []struct {
		year int
		want int
	}{
		{1879, 0},
		{1880, 1}, // 馬車鉄道
		{1887, 2},
		{1895, 3},
		{1900, 4},
		{1905, 5},
		{1925, 6},
		{1930, 7},
		{1950, 8},
		{1964, 9},
		{2000, 9}, // 新幹線まで（リニアは未定義）
	}
	for _, tt := range tests {
		got := UnlockedTechnologies(tt.year)
		if len(got) != tt.want {
			t.Errorf("UnlockedTechnologies(%d) の件数 = %d, want %d", tt.year, len(got), tt.want)
		}
		// 解禁年が year 以下であることを確認
		for _, tech := range got {
			if tech.UnlockYear > tt.year {
				t.Errorf("UnlockedTechnologies(%d) に未解禁 %s（%d）が含まれる",
					tt.year, tech.ID, tech.UnlockYear)
			}
		}
	}

	// 負の年でも安全
	if got := UnlockedTechnologies(-100); len(got) != 0 {
		t.Errorf("UnlockedTechnologies(-100) = %d 件, want 0", len(got))
	}
}

func TestTechnologyTable_Unique(t *testing.T) {
	seen := map[string]bool{}
	for _, tech := range TechnologyTable {
		if tech.ID == "" {
			t.Error("空の ID の技術がある")
		}
		if tech.Name == "" {
			t.Errorf("%s: Name が空", tech.ID)
		}
		if tech.UnlockYear <= 0 {
			t.Errorf("%s: UnlockYear が不正: %d", tech.ID, tech.UnlockYear)
		}
		if seen[tech.ID] {
			t.Errorf("ID が重複: %s", tech.ID)
		}
		seen[tech.ID] = true
	}
	if len(seen) != len(TechnologyTable) {
		t.Errorf("件数 = %d, want %d", len(seen), len(TechnologyTable))
	}
}
