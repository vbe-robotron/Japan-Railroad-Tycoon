package core

import "testing"

// setupAIAbandonGame は廃線 AI テスト用のゲームを返す。
func setupAIAbandonGame(t *testing.T) (*Game, *Company, *Line) {
	t.Helper()
	g := NewGame(1907, "player")
	g.Date = Date{Year: 1910, Month: 1}
	g.AddCity(&City{ID: "umeda", Name: "梅田", Population: 120000,
		Lat: 34.7055, Lon: 135.4983, LandValue: 60})
	g.AddCity(&City{ID: "takarazuka", Name: "宝塚", Population: 30000,
		Lat: 34.7995, Lon: 135.3468, LandValue: 20})

	c := NewCompany("hankyu", "hankyu", "阪急")
	c.Cash = 9_000_000
	g.AddCompany(c)

	line := NewLine("l1", c.ID, "umeda", "takarazuka")
	line.LengthKm = g.DistanceKm("umeda", "takarazuka")
	c.AddLine(line)
	return g, c, line
}

func TestAIAbandon_NoDeficit(t *testing.T) {
	g, c, line := setupAIAbandonGame(t)
	line.ConsecutiveDeficitTurns = 0
	line.AccumulatedLoss = 0
	ai := &AI{CompanyID: "hankyu", Personality: DefaultPersonality("hankyu")}

	if evs := g.aiAbandon(c, ai); evs != nil {
		t.Errorf("黒字路線でイベントが生成された: %d", len(evs))
	}
	if len(c.Lines) != 1 {
		t.Error("黒字路線が廃止された")
	}
}

func TestAIAbandon_Deficit(t *testing.T) {
	g, c, line := setupAIAbandonGame(t)
	line.ConsecutiveDeficitTurns = AbandonDeficitTurns
	ai := &AI{CompanyID: "hankyu", Personality: DefaultPersonality("hankyu")}

	evs := g.aiAbandon(c, ai)

	if !hasEventType(evs, EventLineAbandoned) {
		t.Error("line_abandoned が生成されなかった")
	}
	if len(c.Lines) != 0 {
		t.Errorf("Lines = %d, want 0", len(c.Lines))
	}
}

func TestAIAbandon_AccumulatedLoss(t *testing.T) {
	g, c, line := setupAIAbandonGame(t)
	line.ConsecutiveDeficitTurns = 0
	line.AccumulatedLoss = -(lineValue(line)*AbandonLossRatio/100 + 1)
	ai := &AI{CompanyID: "hankyu", Personality: DefaultPersonality("hankyu")}

	if evs := g.aiAbandon(c, ai); !hasEventType(evs, EventLineAbandoned) {
		t.Errorf("累積損失ゲートで廃線されなかった: %d", len(evs))
	}
}

func TestAIAbandon_HighPoliticalPressure(t *testing.T) {
	g, c, line := setupAIAbandonGame(t)
	line.ConsecutiveDeficitTurns = AbandonDeficitTurns
	g.Cities["umeda"].PoliticalPressure = 0.9
	ai := &AI{CompanyID: "hankyu", Personality: DefaultPersonality("hankyu")}

	if evs := g.aiAbandon(c, ai); evs != nil {
		t.Errorf("政治圧力高で廃線された: %d", len(evs))
	}
	if len(c.Lines) != 1 {
		t.Error("政治圧力高で路線が廃止された")
	}
}

func TestAIAbandon_HighConservatism(t *testing.T) {
	g, c, line := setupAIAbandonGame(t)
	line.ConsecutiveDeficitTurns = AbandonDeficitTurns
	ai := &AI{CompanyID: "hankyu", Personality: Personality{Aggression: 0.5, Conservatism: 0.8}}

	if evs := g.aiAbandon(c, ai); evs != nil {
		t.Errorf("保守的会社で廃線された: %d", len(evs))
	}
}

func TestAIAbandon_NoLines(t *testing.T) {
	g, c, _ := setupAIAbandonGame(t)
	c.Lines = nil
	ai := &AI{CompanyID: "hankyu", Personality: DefaultPersonality("hankyu")}
	if evs := g.aiAbandon(c, ai); evs != nil {
		t.Errorf("路線なしでイベントが生成された: %d", len(evs))
	}
}

func TestShouldAbandon_NilCity(t *testing.T) {
	g, c, line := setupAIAbandonGame(t)
	// 都市が存在しなくても政治ゲートは通過する（圧力なし扱い）
	line.From = "nowhere"
	line.ConsecutiveDeficitTurns = AbandonDeficitTurns
	ai := &AI{CompanyID: "hankyu", Personality: DefaultPersonality("hankyu")}
	if !g.shouldAbandon(c, line, ai) {
		t.Error("都市が存在しない場合に廃線判定が false になった")
	}
}

func TestCityPressureExceeds(t *testing.T) {
	g := NewGame(1907, "p")
	g.AddCity(&City{ID: "a", PoliticalPressure: 0.8})
	g.AddCity(&City{ID: "b", PoliticalPressure: 0.5})
	if !g.cityPressureExceeds("a") {
		t.Error("高圧力都市が検出されなかった")
	}
	if g.cityPressureExceeds("b") {
		t.Error("低圧力都市が高圧力と判定された")
	}
	if g.cityPressureExceeds("missing") {
		t.Error("存在しない都市が高圧力と判定された")
	}
}
