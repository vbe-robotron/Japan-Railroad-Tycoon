package core

import (
	"errors"
	"testing"
)

func TestAbandonLineAction_Type(t *testing.T) {
	a := &AbandonLineAction{}
	if got := a.Type(); got != "abandon_line" {
		t.Errorf("Type() = %q, want %q", got, "abandon_line")
	}
}

func TestAbandonLineAction_Validate_NilGame(t *testing.T) {
	a := &AbandonLineAction{LineID: "L1", Reason: "economic"}
	err := a.Validate(nil)
	if err == nil {
		t.Fatal("nil ゲームでエラーにならなかった")
	}
	ve, ok := err.(*ValidationError)
	if !ok || ve.Code != "GAME_NIL" {
		t.Errorf("GAME_NIL でない: %v", err)
	}
}

func TestAbandonLineAction_Validate_NoPlayer(t *testing.T) {
	g := NewGame(1907, "hankyu")
	a := &AbandonLineAction{LineID: "L1", Reason: "economic"}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("プレイヤーなしでエラーにならなかった")
	}
	ve, _ := err.(*ValidationError)
	if ve.Code != "NO_PLAYER" {
		t.Errorf("NO_PLAYER でない: %q", ve.Code)
	}
}

func TestAbandonLineAction_Validate_LineNotFound(t *testing.T) {
	g := setupAbandonTestGame()
	a := &AbandonLineAction{LineID: "not_exist", Reason: "economic"}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("路線なしでエラーにならなかった")
	}
	ve, _ := err.(*ValidationError)
	if ve.Code != "LINE_NOT_FOUND" {
		t.Errorf("LINE_NOT_FOUND でない: %q", ve.Code)
	}
	if !errors.Is(ve, ErrNotFound) {
		t.Error("ErrNotFound をラップしていない")
	}
}

func TestAbandonLineAction_Validate_NotOwned(t *testing.T) {
	g := setupAbandonTestGame()
	other := NewCompany("other", "other", "他社")
	g.AddCompany(other)
	line := NewLine("L_other", "other", "umeda", "takarazuka")
	other.AddLine(line)

	a := &AbandonLineAction{LineID: "L_other", Reason: "economic"}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("他社路線でエラーにならなかった")
	}
	ve, _ := err.(*ValidationError)
	if ve.Code != "NOT_OWNED" {
		t.Errorf("NOT_OWNED でない: %q", ve.Code)
	}
}

func TestAbandonLineAction_Validate_InvalidReason(t *testing.T) {
	g := setupAbandonTestGame()
	line := g.PlayerCompany().Lines[0]
	a := &AbandonLineAction{LineID: line.ID, Reason: "invalid"}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("不正な理由でエラーにならなかった")
	}
	ve, _ := err.(*ValidationError)
	if ve.Code != "INVALID_REASON" {
		t.Errorf("INVALID_REASON でない: %q", ve.Code)
	}
}

func TestAbandonLineAction_Validate_EconomicGate_NotMet(t *testing.T) {
	g := setupAbandonTestGame()
	line := g.PlayerCompany().Lines[0]
	// 黒字路線（連続赤字 0）
	line.ConsecutiveDeficitTurns = 0
	line.AccumulatedLoss = 0

	a := &AbandonLineAction{LineID: line.ID, Reason: "economic"}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("赤字でない路線でエラーにならなかった")
	}
	ve, _ := err.(*ValidationError)
	if ve.Code != "GATE_ECONOMIC" {
		t.Errorf("GATE_ECONOMIC でない: %q", ve.Code)
	}
}

func TestAbandonLineAction_Validate_PoliticalGate_NotMet(t *testing.T) {
	g := setupAbandonTestGame()
	line := g.PlayerCompany().Lines[0]
	line.ConsecutiveDeficitTurns = 24
	line.AccumulatedLoss = -1_000_000

	g.Cities["umeda"].PoliticalPressure = 0.9 // 閾値超え

	a := &AbandonLineAction{LineID: line.ID, Reason: "economic"}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("政治圧力高でエラーにならなかった")
	}
	ve, _ := err.(*ValidationError)
	if ve.Code != "GATE_POLITICAL" {
		t.Errorf("GATE_POLITICAL でない: %q", ve.Code)
	}
}

func TestAbandonLineAction_Validate_OK(t *testing.T) {
	g := setupAbandonTestGame()
	line := g.PlayerCompany().Lines[0]
	line.ConsecutiveDeficitTurns = 24
	line.AccumulatedLoss = -1_000_000

	a := &AbandonLineAction{LineID: line.ID, Reason: "economic"}
	if err := a.Validate(g); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}

func TestAbandonLineAction_Apply(t *testing.T) {
	g := setupAbandonTestGame()
	c := g.PlayerCompany()
	line := c.Lines[0]
	line.ConsecutiveDeficitTurns = 24
	line.AccumulatedLoss = -1_000_000
	beforeCash := c.Cash
	beforeReputation := c.Reputation
	beforePressure := g.Cities["umeda"].PoliticalPressure

	a := &AbandonLineAction{LineID: line.ID, Reason: "economic"}
	events, err := a.Apply(g)
	if err != nil {
		t.Fatal(err)
	}

	// 路線削除
	if len(c.Lines) != 0 {
		t.Errorf("Lines 数 = %d, want 0", len(c.Lines))
	}
	// 用地売却益
	if c.Cash <= beforeCash {
		t.Errorf("資金が増えていない: %d → %d", beforeCash, c.Cash)
	}
	// 評判低下
	if c.Reputation != beforeReputation-5 {
		t.Errorf("Reputation = %d, want %d", c.Reputation, beforeReputation-5)
	}
	// 政治圧力上昇
	if g.Cities["umeda"].PoliticalPressure <= beforePressure {
		t.Error("政治圧力が上昇していない")
	}
	// イベント
	if len(events) == 0 {
		t.Fatal("events が空")
	}
	found := false
	for _, e := range events {
		if e.Type == EventLineAbandoned {
			found = true
		}
	}
	if !found {
		t.Error("line_abandoned イベントがない")
	}
}

func TestAbandonLineAction_Apply_ViaRegistry(t *testing.T) {
	g := setupAbandonTestGame()
	line := g.PlayerCompany().Lines[0]
	line.ConsecutiveDeficitTurns = 24
	line.AccumulatedLoss = -1_000_000

	payload := []byte(`{"line_id":"` + line.ID + `","reason":"economic"}`)
	a, err := DecodeAction("abandon_line", payload)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Apply(g); err != nil {
		t.Fatal(err)
	}
	if len(g.PlayerCompany().Lines) != 0 {
		t.Errorf("Lines 数 = %d, want 0", len(g.PlayerCompany().Lines))
	}
}

func TestAbandonLineAction_Deterministic(t *testing.T) {
	g1 := setupAbandonTestGame()
	g2 := setupAbandonTestGame()
	line1 := g1.PlayerCompany().Lines[0]
	line2 := g2.PlayerCompany().Lines[0]
	line1.ConsecutiveDeficitTurns = 24
	line2.ConsecutiveDeficitTurns = 24
	line1.AccumulatedLoss = -1_000_000
	line2.AccumulatedLoss = -1_000_000

	a1 := &AbandonLineAction{LineID: line1.ID, Reason: "economic"}
	a2 := &AbandonLineAction{LineID: line2.ID, Reason: "economic"}
	e1, _ := a1.Apply(g1)
	e2, _ := a2.Apply(g2)

	if e1[0].Message != e2[0].Message {
		t.Errorf("Message が異なる: %q vs %q", e1[0].Message, e2[0].Message)
	}
}

func setupAbandonTestGame() *Game {
	g := NewGame(1907, "hankyu")
	g.Date = Date{Year: 1910, Month: 1}
	g.AddCity(&City{ID: "umeda", Name: "梅田", Population: 120000,
		Lat: 34.7055, Lon: 135.4983, LandValue: 60})
	g.AddCity(&City{ID: "takarazuka", Name: "宝塚", Population: 30000,
		Lat: 34.7995, Lon: 135.3468, LandValue: 20})

	hankyu := NewCompany("hankyu", "hankyu", "阪急")
	hankyu.Cash = 9_000_000
	g.AddCompany(hankyu)

	a := &BuildLineAction{From: "umeda", To: "takarazuka", Gauge: 1067}
	if _, err := a.Apply(g); err != nil {
		panic(err)
	}
	return g
}
