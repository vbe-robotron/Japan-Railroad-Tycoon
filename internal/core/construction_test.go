package core

import (
	"math"
	"testing"
)

// constructionGame は建設テスト用のゲームを返す。
//
// 2 都市（梅田・宝塚）とプレイヤー会社（阪急、指定資金）を持つ。
// Date は指定年 6 月（季節係数 1.00）に設定する。
func constructionGame(year int, cash int64) *Game {
	g := NewGame(1907, "hankyu")
	g.Date = Date{Year: year, Month: 6}
	g.AddCity(NewCity("umeda", "梅田", "大阪"))
	g.AddCity(NewCity("takarazuka", "宝塚", "兵庫"))
	c := NewCompany("hankyu", "hankyu", "阪急")
	c.Cash = cash
	g.AddCompany(c)
	return g
}

func TestCheckTechnologyGate(t *testing.T) {
	tests := []struct {
		name   string
		year   int
		gauge  int
		status GateStatus
	}{
		{"762 未解禁", 1879, 762, GateFail},
		{"762 解禁", 1880, 762, GatePass},
		{"1067 未解禁", 1886, 1067, GateFail},
		{"1067 解禁", 1887, 1067, GatePass},
		{"1435 未解禁", 1894, 1435, GateFail},
		{"1435 解禁", 1895, 1435, GatePass},
		{"不正な軌間", 1950, 999, GateFail},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := constructionGame(tt.year, 0)
			gate := g.checkTechnologyGate(tt.gauge)
			if gate.Kind != "technology" {
				t.Errorf("Kind = %q, want technology", gate.Kind)
			}
			if gate.Status != tt.status {
				t.Errorf("Status = %q, want %q（理由: %s）", gate.Status, tt.status, gate.Reason)
			}
			if gate.Reason == "" {
				t.Error("Reason が空")
			}
		})
	}

	// 未解禁時は Details に required / unlock_year が入る
	g := constructionGame(1886, 0)
	gate := g.checkTechnologyGate(1067)
	if gate.Details["required"] != "steam_railway" {
		t.Errorf("Details[required] = %v, want steam_railway", gate.Details["required"])
	}
	if gate.Details["unlock_year"] != 1887 {
		t.Errorf("Details[unlock_year] = %v, want 1887", gate.Details["unlock_year"])
	}
}

func TestCheckRegulationGate(t *testing.T) {
	g1886 := constructionGame(1886, 0)
	gate := g1886.checkRegulationGate()
	if gate.Kind != "regulation" {
		t.Errorf("Kind = %q, want regulation", gate.Kind)
	}
	if gate.Status != GateFail {
		t.Errorf("1886年の Status = %q, want %q", gate.Status, GateFail)
	}
	if gate.Reason == "" {
		t.Error("Reason が空")
	}

	g1887 := constructionGame(1887, 0)
	gate = g1887.checkRegulationGate()
	if gate.Status != GatePass {
		t.Errorf("1887年の Status = %q, want %q", gate.Status, GatePass)
	}
}

func TestCheckTerrainGate(t *testing.T) {
	g := constructionGame(1910, 0)
	gate := g.checkTerrainGate("umeda", "takarazuka")
	if gate.Kind != "terrain" {
		t.Errorf("Kind = %q, want terrain", gate.Kind)
	}
	// Phase 1-6 では常に conditional
	if gate.Status != GateConditional {
		t.Errorf("Status = %q, want %q", gate.Status, GateConditional)
	}
	if gate.Details["terrain"] != string(TerrainPlain) {
		t.Errorf("Details[terrain] = %v, want %q", gate.Details["terrain"], TerrainPlain)
	}
}

func TestCheckEconomyGate(t *testing.T) {
	// プレイヤー会社なし
	gNoPlayer := constructionGame(1910, 0)
	delete(gNoPlayer.Companies, "hankyu")
	gate := gNoPlayer.checkEconomyGate("umeda", "takarazuka", 1067, false)
	if gate.Kind != "economy" {
		t.Errorf("Kind = %q, want economy", gate.Kind)
	}
	if gate.Status != GateFail {
		t.Errorf("プレイヤーなしの Status = %q, want %q", gate.Status, GateFail)
	}

	// 資金不足
	gPoor := constructionGame(1910, 0)
	cost := gPoor.LineCost("umeda", "takarazuka", 1067, false)
	gPoor.PlayerCompany().Cash = cost - 1
	gate = gPoor.checkEconomyGate("umeda", "takarazuka", 1067, false)
	if gate.Status != GateFail {
		t.Errorf("資金不足の Status = %q, want %q", gate.Status, GateFail)
	}
	if gate.Details["required"] != cost {
		t.Errorf("Details[required] = %v, want %d", gate.Details["required"], cost)
	}
	if gate.Details["available"] != cost-1 {
		t.Errorf("Details[available] = %v, want %d", gate.Details["available"], cost-1)
	}

	// 資金十分
	gRich := constructionGame(1910, cost)
	gate = gRich.checkEconomyGate("umeda", "takarazuka", 1067, false)
	if gate.Status != GatePass {
		t.Errorf("資金十分の Status = %q, want %q", gate.Status, GatePass)
	}
}

func TestCanBuildLine_AllGates(t *testing.T) {
	g := constructionGame(1910, 100_000_000)
	result := g.CanBuildLine("umeda", "takarazuka", 1067, false)

	if !result.Can {
		t.Fatalf("Can = false, want true（ゲート: %+v）", result.Gates)
	}
	if len(result.Gates) != 4 {
		t.Fatalf("ゲート数 = %d, want 4", len(result.Gates))
	}

	kinds := map[string]bool{}
	for _, gate := range result.Gates {
		kinds[gate.Kind] = true
		if gate.Status == GateFail {
			t.Errorf("ゲート %s が fail: %s", gate.Kind, gate.Reason)
		}
	}
	for _, want := range []string{"technology", "regulation", "terrain", "economy"} {
		if !kinds[want] {
			t.Errorf("ゲート %s が含まれない", want)
		}
	}

	if result.Cost <= 0 {
		t.Errorf("Cost = %d, want > 0", result.Cost)
	}
	if result.Months < 1 {
		t.Errorf("Months = %d, want >= 1", result.Months)
	}
	if result.Cost != g.LineCost("umeda", "takarazuka", 1067, false) {
		t.Errorf("Cost が LineCost と一致しない")
	}
}

func TestCanBuildLine_TechnologyFail(t *testing.T) {
	g := constructionGame(1880, 100_000_000) // 蒸気鉄道は 1887 解禁
	result := g.CanBuildLine("umeda", "takarazuka", 1067, false)
	if result.Can {
		t.Fatal("Can = true, want false")
	}
	found := false
	for _, gate := range result.Gates {
		if gate.Kind == "technology" && gate.Status == GateFail {
			found = true
		}
	}
	if !found {
		t.Error("技術ゲートが fail でない")
	}
}

func TestCanBuildLine_RegulationFail(t *testing.T) {
	g := constructionGame(1886, 100_000_000) // 私設鉄道条例は 1887
	result := g.CanBuildLine("umeda", "takarazuka", 762, false)
	if result.Can {
		t.Fatal("Can = true, want false")
	}
	found := false
	for _, gate := range result.Gates {
		if gate.Kind == "regulation" && gate.Status == GateFail {
			found = true
		}
	}
	if !found {
		t.Error("法規ゲートが fail でない")
	}
}

func TestCanBuildLine_EconomyFail(t *testing.T) {
	g := constructionGame(1910, 0)
	result := g.CanBuildLine("umeda", "takarazuka", 1067, false)
	if result.Can {
		t.Fatal("Can = true, want false")
	}
	found := false
	for _, gate := range result.Gates {
		if gate.Kind == "economy" && gate.Status == GateFail {
			found = true
		}
	}
	if !found {
		t.Error("経済ゲートが fail でない")
	}
}

func TestBuildMonths(t *testing.T) {
	g := constructionGame(1910, 0)
	dist := g.DistanceKm("umeda", "takarazuka")
	want := int(math.Ceil(dist / 5.0))
	if want < 1 {
		want = 1
	}
	if got := g.buildMonths("umeda", "takarazuka"); got != want {
		t.Errorf("buildMonths = %d, want %d", got, want)
	}

	// 未知都市は距離 0 → 最低 1 ヶ月
	if got := g.buildMonths("umeda", "missing"); got != 1 {
		t.Errorf("buildMonths(未知) = %d, want 1", got)
	}
}
