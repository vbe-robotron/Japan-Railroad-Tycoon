package core

import (
	"encoding/json"
	"testing"
)

func TestNewGame(t *testing.T) {
	g := NewGame(1907, "hankyu")

	if len(g.ID) != nanoIDLength {
		t.Errorf("ID の長さ = %d, want %d", len(g.ID), nanoIDLength)
	}
	if g.Seed != 1907 {
		t.Errorf("Seed = %d, want 1907", g.Seed)
	}
	if g.IDGen == nil {
		t.Fatal("IDGen が nil")
	}
	if g.IDGen.Seed != 1907 {
		t.Errorf("IDGen.Seed = %d, want 1907", g.IDGen.Seed)
	}
	if g.IDGen.Counter != 1 {
		t.Errorf("IDGen.Counter = %d, want 1", g.IDGen.Counter)
	}
	if g.Turn != 0 {
		t.Errorf("Turn = %d, want 0", g.Turn)
	}
	if g.PlayerID != "hankyu" {
		t.Errorf("PlayerID = %q, want hankyu", g.PlayerID)
	}
	if !g.Date.IsZero() {
		t.Errorf("Date = %+v, want ゼロ値", g.Date)
	}
	if g.Companies == nil {
		t.Error("Companies が nil")
	}
	if g.Cities == nil {
		t.Error("Cities が nil")
	}
	if g.Events == nil {
		t.Error("Events が nil")
	}
	if g.rng == nil {
		t.Error("rng が nil")
	}
}

func TestNewGame_Deterministic(t *testing.T) {
	a := NewGame(1907, "hankyu")
	b := NewGame(1907, "hankyu")
	if a.ID != b.ID {
		t.Errorf("同じ seed で異なる ID: %q != %q", a.ID, b.ID)
	}
}

func TestGame_PlayerCompany(t *testing.T) {
	tests := []struct {
		name     string
		playerID string
		add      []*Company
		wantNil  bool
		wantName string
	}{
		{
			name:     "存在する",
			playerID: "hankyu",
			add:      []*Company{NewCompany("hankyu", "hankyu-group", "阪急")},
			wantName: "阪急",
		},
		{
			name:     "存在しない",
			playerID: "missing",
			add:      []*Company{NewCompany("hankyu", "", "阪急")},
			wantNil:  true,
		},
		{
			name:     "会社なし",
			playerID: "any",
			wantNil:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGame(1, tt.playerID)
			for _, c := range tt.add {
				g.AddCompany(c)
			}
			got := g.PlayerCompany()
			if tt.wantNil {
				if got != nil {
					t.Fatalf("PlayerCompany() = %+v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("PlayerCompany() = nil, want 会社")
			}
			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
		})
	}
}

func TestGame_AddCompany_AddCity(t *testing.T) {
	g := NewGame(1, "player")

	c1 := NewCompany("a", "", "A社")
	c2 := NewCompany("b", "", "B社")
	g.AddCompany(c1)
	g.AddCompany(c2)
	if len(g.Companies) != 2 {
		t.Fatalf("Companies の数 = %d, want 2", len(g.Companies))
	}
	if g.Companies["a"] != c1 || g.Companies["b"] != c2 {
		t.Error("Companies に正しく登録されていない")
	}

	city := NewCity("tokyo", "東京", "東京府")
	g.AddCity(city)
	if len(g.Cities) != 1 {
		t.Fatalf("Cities の数 = %d, want 1", len(g.Cities))
	}
	if g.Cities["tokyo"] != city {
		t.Error("Cities に正しく登録されていない")
	}
}

func TestGame_AppendEvents(t *testing.T) {
	g := NewGame(1, "player")
	if len(g.Events) != 0 {
		t.Fatalf("初期 Events の数 = %d, want 0", len(g.Events))
	}

	e1 := Event{ID: "e1", Type: "test", Message: "1"}
	e2 := Event{ID: "e2", Type: "test", Message: "2"}
	g.AppendEvents(e1)
	g.AppendEvents(e2, Event{ID: "e3", Type: "test"})

	if len(g.Events) != 3 {
		t.Fatalf("Events の数 = %d, want 3", len(g.Events))
	}
	if g.Events[0].ID != "e1" || g.Events[1].ID != "e2" || g.Events[2].ID != "e3" {
		t.Errorf("Events の中身が不正: %+v", g.Events)
	}
}

func TestGame_EnsureRng(t *testing.T) {
	g := NewGame(42, "player")
	g.rng = nil

	g.EnsureRng()
	if g.rng == nil {
		t.Fatal("EnsureRng 後も rng が nil")
	}

	// 既に存在する場合は上書きしない
	r := g.Rng()
	g.EnsureRng()
	if g.rng != r {
		t.Error("EnsureRng が既存の rng を上書きした")
	}
}

func TestGame_Rng_Deterministic(t *testing.T) {
	a := NewGame(1907, "p")
	b := NewGame(1907, "p")
	for i := 0; i < 10; i++ {
		if a.Rng().Int63() != b.Rng().Int63() {
			t.Fatalf("iteration %d で乱数列が一致しない", i)
		}
	}
}

func TestGame_JSONRoundTrip(t *testing.T) {
	g := NewGame(1907, "hankyu")
	g.Date = Date{1880, 1}
	g.Turn = 12
	g.AddCompany(NewCompany("hankyu", "hankyu-group", "阪急"))
	g.AddCity(NewCity("osaka", "大阪", "大阪府"))
	g.AppendEvents(Event{ID: "e1", Type: "founded", Severity: SeverityInfo})

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var restored Game
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	restored.EnsureRng()

	if restored.ID != g.ID {
		t.Errorf("ID = %q, want %q", restored.ID, g.ID)
	}
	if restored.Seed != g.Seed {
		t.Errorf("Seed = %d, want %d", restored.Seed, g.Seed)
	}
	if restored.Date != g.Date {
		t.Errorf("Date = %+v, want %+v", restored.Date, g.Date)
	}
	if restored.Turn != g.Turn {
		t.Errorf("Turn = %d, want %d", restored.Turn, g.Turn)
	}
	if restored.PlayerID != g.PlayerID {
		t.Errorf("PlayerID = %q, want %q", restored.PlayerID, g.PlayerID)
	}
	if len(restored.Companies) != 1 {
		t.Errorf("Companies の数 = %d, want 1", len(restored.Companies))
	}
	if len(restored.Cities) != 1 {
		t.Errorf("Cities の数 = %d, want 1", len(restored.Cities))
	}
	if len(restored.Events) != 1 {
		t.Errorf("Events の数 = %d, want 1", len(restored.Events))
	}
	if restored.rng == nil {
		t.Error("EnsureRng 後も rng が nil")
	}

	// IDGen の Counter が保たれ、同じ ID 列が続く
	if restored.IDGen == nil {
		t.Fatal("IDGen が nil")
	}
	if restored.IDGen.Counter != g.IDGen.Counter {
		t.Errorf("IDGen.Counter = %d, want %d", restored.IDGen.Counter, g.IDGen.Counter)
	}
	if g.IDGen.Next() != restored.IDGen.Next() {
		t.Error("復元後に同じ ID が続かない")
	}
}

func TestGame_JSONExcludesRng(t *testing.T) {
	g := NewGame(1, "p")
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if _, ok := m["rng"]; ok {
		t.Error("JSON に rng が含まれている")
	}
	if _, ok := m["id_gen"]; !ok {
		t.Error("JSON に id_gen が含まれていない")
	}
}
