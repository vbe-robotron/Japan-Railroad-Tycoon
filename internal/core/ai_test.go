package core

import (
	"encoding/json"
	"testing"
)

// newAITestGame は AI テスト用のゲームを返す。
//
// 2 都市（a, b）と会社を 1 社追加し、必要なら史実路線を 1 本設定する。
func newAITestGame(t *testing.T, companyID string, cash int64, line *HistoricalLine) (*Game, *Company) {
	t.Helper()
	g := NewGame(1907, "player")
	g.Date = Date{Year: 1910, Month: 3}
	g.AddCity(&City{ID: "a", Name: "A", Population: 100000, Lat: 35.0, Lon: 135.0})
	g.AddCity(&City{ID: "b", Name: "B", Population: 100000, Lat: 35.0, Lon: 135.2})

	c := NewCompany(companyID, companyID, companyID)
	c.Cash = cash
	g.AddCompany(c)

	if line != nil {
		g.SetHistoricalLines([]HistoricalLine{*line})
	}
	return g, c
}

func TestInitAIs(t *testing.T) {
	g := NewGame(1, "player")
	p := NewCompany("player", "player", "プレイヤー")
	p.IsPlayer = true
	g.AddCompany(p)
	g.AddCompany(NewCompany("hankyu", "hankyu", "阪急"))

	g.InitAIs()

	if len(g.AIs) != 2 {
		t.Fatalf("len(AIs) = %d, want 2", len(g.AIs))
	}
	if g.AIs["player"] == nil || g.AIs["hankyu"] == nil {
		t.Fatal("AI が割り当てられていない会社がある")
	}
	if g.AIs["player"].HistoryMode != HistoryIgnore {
		t.Errorf("player.HistoryMode = %v, want HistoryIgnore", g.AIs["player"].HistoryMode)
	}
	if g.AIs["hankyu"].HistoryMode != HistoryStrict {
		t.Errorf("hankyu.HistoryMode = %v, want HistoryStrict", g.AIs["hankyu"].HistoryMode)
	}
}

func TestInitAIs_Idempotent(t *testing.T) {
	g := NewGame(1, "player")
	g.AddCompany(NewCompany("hankyu", "hankyu", "阪急"))

	g.InitAIs()
	first := g.AIs["hankyu"]
	g.InitAIs()

	if g.AIs["hankyu"] != first {
		t.Error("InitAIs が既存の AI を上書きした")
	}
}

func TestInitAIs_PlayerFlag(t *testing.T) {
	g := NewGame(1, "hankyu")
	c := NewCompany("hankyu", "hankyu", "阪急")
	c.IsPlayer = true
	g.AddCompany(c)

	g.InitAIs()

	if got := g.AIs["hankyu"].HistoryMode; got != HistoryIgnore {
		t.Errorf("IsPlayer の HistoryMode = %v, want HistoryIgnore", got)
	}
}

func TestDefaultPersonality(t *testing.T) {
	h := DefaultPersonality("hankyu")
	if h.RealEstate < 0.9 {
		t.Errorf("hankyu.RealEstate = %v, want >= 0.9", h.RealEstate)
	}

	j := DefaultPersonality("jnr")
	if j.RealEstate > 0.1 {
		t.Errorf("jnr.RealEstate = %v, want <= 0.1", j.RealEstate)
	}
	if j.Politics < 0.9 {
		t.Errorf("jnr.Politics = %v, want >= 0.9", j.Politics)
	}

	def := DefaultPersonality("unknown")
	want := Personality{Aggression: 0.5, Conservatism: 0.5, Innovation: 0.5, RealEstate: 0.5, Politics: 0.5}
	if def != want {
		t.Errorf("不明会社の性格 = %+v, want %+v", def, want)
	}
}

func TestAIActionOrder(t *testing.T) {
	g := NewGame(1, "p")
	g.AddCompany(NewCompany("zeta", "zeta", "Z"))
	g.AddCompany(NewCompany("alpha", "alpha", "A"))
	g.AddCompany(NewCompany("mike", "mike", "M"))

	order := g.aiActionOrder()
	want := []string{"alpha", "mike", "zeta"}
	if len(order) != len(want) {
		t.Fatalf("len(order) = %d, want %d", len(order), len(want))
	}
	for i := range want {
		if order[i] != want[i] {
			t.Errorf("order[%d] = %q, want %q", i, order[i], want[i])
		}
	}
}

func TestRunAI_SkipsPlayer(t *testing.T) {
	g, c := newAITestGame(t, "player", 100_000_000, &HistoricalLine{
		CompanyID: "player", From: "a", To: "b", Gauge: 1067,
		OpenedYear: 1910, OpenedMonth: 3, DisplayName: "テスト線",
	})
	c.IsPlayer = true
	g.InitAIs()

	events := g.RunAI()

	if len(events) != 0 {
		t.Errorf("プレイヤーが行動した: %d events", len(events))
	}
	if len(c.Lines) != 0 {
		t.Errorf("プレイヤーの路線 = %d, want 0", len(c.Lines))
	}
}

func TestRunAI_DissolvedSkip(t *testing.T) {
	g, c := newAITestGame(t, "hankyu", 100_000_000, &HistoricalLine{
		CompanyID: "hankyu", From: "a", To: "b", Gauge: 1067,
		OpenedYear: 1910, OpenedMonth: 3, DisplayName: "テスト線",
	})
	dissolved := Date{Year: 1909, Month: 1}
	c.Dissolved = &dissolved
	g.InitAIs()

	events := g.RunAI()

	if len(events) != 0 {
		t.Errorf("解散会社が行動した: %d events", len(events))
	}
	if len(c.Lines) != 0 {
		t.Errorf("解散会社の路線 = %d, want 0", len(c.Lines))
	}
}

func TestGame_JSONRoundTrip_AIAndHistorical(t *testing.T) {
	g := NewGame(1, "player")
	p := NewCompany("player", "player", "プレイヤー")
	p.IsPlayer = true
	g.AddCompany(p)
	g.AddCompany(NewCompany("hankyu", "hankyu", "阪急"))
	g.InitAIs()
	g.SetHistoricalLines([]HistoricalLine{{
		CompanyID: "hankyu", From: "umeda", To: "takarazuka",
		Gauge: 1435, OpenedYear: 1910, OpenedMonth: 3, DisplayName: "宝塚線",
	}})

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var restored Game
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if len(restored.AIs) != 2 {
		t.Fatalf("len(AIs) = %d, want 2", len(restored.AIs))
	}
	if restored.AIs["hankyu"].Personality.RealEstate != 0.9 {
		t.Errorf("hankyu.RealEstate = %v, want 0.9",
			restored.AIs["hankyu"].Personality.RealEstate)
	}
	if restored.AIs["hankyu"].HistoryMode != HistoryStrict {
		t.Errorf("hankyu.HistoryMode = %v, want HistoryStrict",
			restored.AIs["hankyu"].HistoryMode)
	}
	if len(restored.HistoricalLines) != 1 {
		t.Fatalf("len(HistoricalLines) = %d, want 1", len(restored.HistoricalLines))
	}
	if restored.HistoricalLines[0].DisplayName != "宝塚線" {
		t.Errorf("DisplayName = %q", restored.HistoricalLines[0].DisplayName)
	}
}

func TestRunAI_RunsNonPlayer(t *testing.T) {
	g, c := newAITestGame(t, "hankyu", 100_000_000, &HistoricalLine{
		CompanyID: "hankyu", From: "a", To: "b", Gauge: 1067,
		OpenedYear: 1910, OpenedMonth: 3, DisplayName: "テスト線",
	})
	g.InitAIs()

	events := g.RunAI()

	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}
	if events[0].Type != EventLineOpened {
		t.Errorf("events[0].Type = %q, want %q", events[0].Type, EventLineOpened)
	}
	if len(c.Lines) != 1 {
		t.Errorf("hankyu の路線 = %d, want 1", len(c.Lines))
	}
}
