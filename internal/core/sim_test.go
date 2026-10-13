package core

import (
	"math"
	"testing"
)

// simGame はテスト用に 2 都市（a, b）を持つゲームを返す。
func simGame() *Game {
	g := NewGame(1907, "hankyu")
	g.Date = Date{Year: 1907, Month: 6} // 季節係数 1.00
	g.AddCity(&City{ID: "a", Name: "A", Population: 100000, LandValue: 50})
	g.AddCity(&City{ID: "b", Name: "B", Population: 50000, LandValue: 50})
	return g
}

func TestDistanceKm_SameCity(t *testing.T) {
	g := simGame()
	if got := g.DistanceKm("a", "a"); got != 0 {
		t.Errorf("DistanceKm(a, a) = %v, want 0", got)
	}
}

func TestDistanceKm_UnknownCity(t *testing.T) {
	g := simGame()
	if got := g.DistanceKm("a", "missing"); got != 0 {
		t.Errorf("DistanceKm(a, missing) = %v, want 0", got)
	}
	if got := g.DistanceKm("missing", "a"); got != 0 {
		t.Errorf("DistanceKm(missing, a) = %v, want 0", got)
	}
}

func TestDistanceKm_Positive(t *testing.T) {
	g := simGame()
	got := g.DistanceKm("a", "b")
	if got <= 0 {
		t.Fatalf("DistanceKm(a, b) = %v, want > 0", got)
	}
	// 対称性
	if g.DistanceKm("b", "a") != got {
		t.Errorf("DistanceKm が対称でない")
	}
}

func TestDistanceKm_Deterministic(t *testing.T) {
	g1 := simGame()
	g2 := simGame()
	if g1.DistanceKm("a", "b") != g2.DistanceKm("a", "b") {
		t.Error("DistanceKm が決定論的でない")
	}
}

func TestDistanceKmWithTerrain(t *testing.T) {
	g := simGame()
	a, b := g.Cities["a"], g.Cities["b"]

	plain := g.DistanceKmWithTerrain(a, b, TerrainPlain)
	hill := g.DistanceKmWithTerrain(a, b, TerrainHill)
	sea := g.DistanceKmWithTerrain(a, b, TerrainSea)

	if plain <= 0 {
		t.Fatalf("平地距離 = %v, want > 0", plain)
	}
	if hill <= plain {
		t.Errorf("丘陵距離 %v が平地距離 %v 以下", hill, plain)
	}
	if sea != 0 {
		t.Errorf("海距離 = %v, want 0", sea)
	}
	// nil は 0
	if got := g.DistanceKmWithTerrain(nil, b, TerrainPlain); got != 0 {
		t.Errorf("from=nil で %v, want 0", got)
	}
	if got := g.DistanceKmWithTerrain(a, nil, TerrainPlain); got != 0 {
		t.Errorf("to=nil で %v, want 0", got)
	}
}

func TestCityLatLon_Range(t *testing.T) {
	g := simGame()
	for _, id := range []string{"a", "b", "umeda", "takarazuka"} {
		lat, lon := g.cityLatLon(&City{ID: id})
		if lat < 30.0 || lat > 46.0 {
			t.Errorf("cityLatLon(%q) lat = %v, want 30-46", id, lat)
		}
		if lon < 128.0 || lon > 146.0 {
			t.Errorf("cityLatLon(%q) lon = %v, want 128-146", id, lon)
		}
	}
}

func TestGaugeFactor(t *testing.T) {
	tests := []struct {
		gauge int
		want  float64
	}{
		{762, 0.8},
		{1067, 1.0},
		{1435, 1.2},
		{999, 1.0}, // 未知は 1.0
	}
	for _, tt := range tests {
		if got := GaugeFactor(tt.gauge); got != tt.want {
			t.Errorf("GaugeFactor(%d) = %v, want %v", tt.gauge, got, tt.want)
		}
	}
}

func TestDoubleFactor(t *testing.T) {
	if got := DoubleFactor(false); got != 1.0 {
		t.Errorf("DoubleFactor(false) = %v, want 1.0", got)
	}
	if got := DoubleFactor(true); got != 1.7 {
		t.Errorf("DoubleFactor(true) = %v, want 1.7", got)
	}
}

func TestLineCost_AllGauges(t *testing.T) {
	g := simGame()
	c762 := g.LineCost("a", "b", 762, false)
	c1067 := g.LineCost("a", "b", 1067, false)
	c1435 := g.LineCost("a", "b", 1435, false)
	cDouble := g.LineCost("a", "b", 1067, true)

	if c762 <= 0 || c1067 <= 0 || c1435 <= 0 {
		t.Fatalf("コストが非正: %d, %d, %d", c762, c1067, c1435)
	}
	if !(c762 < c1067 && c1067 < c1435) {
		t.Errorf("軌間順序が不正: 762=%d, 1067=%d, 1435=%d", c762, c1067, c1435)
	}
	if cDouble <= c1067 {
		t.Errorf("複線コスト %d が単線 %d 以下", cDouble, c1067)
	}

	// 期待値: Base × dist × build(1.0) × GaugeFactor × DoubleFactor
	dist := g.DistanceKm("a", "b")
	want := int64(math.Round(float64(BaseCostPerKm) * dist * 1.0 * 1.0 * 1.0))
	if c1067 != want {
		t.Errorf("LineCost(1067) = %d, want %d", c1067, want)
	}
}

func TestLineCost_WithTerrain(t *testing.T) {
	g := simGame()
	plain := g.LineCostWithTerrain("a", "b", TerrainPlain, 1067, false)
	hill := g.LineCostWithTerrain("a", "b", TerrainHill, 1067, false)
	mountain := g.LineCostWithTerrain("a", "b", TerrainMountain, 1067, false)

	if !(plain < hill && hill < mountain) {
		t.Errorf("地形によるコスト差が不正: plain=%d, hill=%d, mountain=%d", plain, hill, mountain)
	}

	// 未知都市は 0
	if got := g.LineCostWithTerrain("missing", "b", TerrainPlain, 1067, false); got != 0 {
		t.Errorf("from 未知で %d, want 0", got)
	}
	if got := g.LineCostWithTerrain("a", "missing", TerrainPlain, 1067, false); got != 0 {
		t.Errorf("to 未知で %d, want 0", got)
	}
}

func TestDefaultFare(t *testing.T) {
	g := simGame()
	dist := g.DistanceKm("a", "b")
	want := int(math.Round((50.0+8.0*dist)/10.0) * 10)
	got := g.DefaultFare("a", "b")
	if got != want {
		t.Errorf("DefaultFare = %d, want %d", got, want)
	}
	if got%10 != 0 {
		t.Errorf("DefaultFare = %d, 10 円単位でない", got)
	}

	// 未知都市は距離 0 → 50 円
	if got := g.DefaultFare("a", "missing"); got != 50 {
		t.Errorf("未知都市の DefaultFare = %d, want 50", got)
	}
}

func TestDemand(t *testing.T) {
	g := simGame()
	g.Date = Date{Year: 1907, Month: 6} // 季節係数 1.00

	// Phase 1-8-2.5: 需要はべき乗則 (Population/DemandBase)^0.6 に
	// 開発度倍率 (1 + Developed×1.0) を掛けたもの（Developed=0 は 1.0 倍）。
	base := &City{ID: "x", Population: 100000, Developed: 0.0}
	if got := g.Demand(base); got != 15 {
		t.Errorf("Demand(base) = %d, want 15", got)
	}

	// 開発度 0.5 → 需要倍率 1.5 倍。base=15.84（100^0.6）、15.84×1.5=23.7 → 23
	dev := &City{ID: "x", Population: 100000, Developed: 0.5}
	if got := g.Demand(dev); got != 23 {
		t.Errorf("Demand(developed=0.5) = %d, want 23", got)
	}

	small := &City{ID: "x", Population: 50000, Developed: 0.0}
	if got := g.Demand(small); got != 10 {
		t.Errorf("Demand(pop=50000) = %d, want 10", got)
	}

	if got := g.Demand(nil); got != 0 {
		t.Errorf("Demand(nil) = %d, want 0", got)
	}
}

func TestDemand_SeasonalFactor(t *testing.T) {
	g := simGame()
	base := &City{ID: "x", Population: 100000, Developed: 0.0}

	prev := -1
	for m := 1; m <= 12; m++ {
		g.Date = Date{Year: 1907, Month: m}
		f := g.SeasonalFactor()
		if f < 0.9 || f > 1.1 {
			t.Errorf("月 %d の季節係数 = %v, want 0.9-1.1", m, f)
		}
		d := g.Demand(base)
		if d <= 0 {
			t.Errorf("月 %d の需要 = %d, want > 0", m, d)
		}
		prev = d
	}
	_ = prev

	// 不正な月は 1.0
	g.Date = Date{Year: 1907, Month: 0}
	if got := g.SeasonalFactor(); got != 1.0 {
		t.Errorf("月 0 の季節係数 = %v, want 1.0", got)
	}
	g.Date = Date{Year: 1907, Month: 13}
	if got := g.SeasonalFactor(); got != 1.0 {
		t.Errorf("月 13 の季節係数 = %v, want 1.0", got)
	}
}

func TestProjectDemand(t *testing.T) {
	g := simGame()
	g.Date = Date{Year: 1907, Month: 6}
	city := &City{ID: "x", Population: 100000, Developed: 0.0}

	// years <= 0 は現在の需要
	if got := g.ProjectDemand(city, 0); got != g.Demand(city) {
		t.Errorf("ProjectDemand(years=0) = %d, want %d", got, g.Demand(city))
	}
	if got := g.ProjectDemand(city, -1); got != g.Demand(city) {
		t.Errorf("ProjectDemand(years=-1) = %d, want %d", got, g.Demand(city))
	}

	// 将来は増加
	future := g.ProjectDemand(city, 10)
	if future <= g.Demand(city) {
		t.Errorf("ProjectDemand(years=10) = %d, 現在 %d 以下", future, g.Demand(city))
	}

	// 開発度は 1.0 でクランプ
	maxed := g.ProjectDemand(city, 1000)
	capped := &City{ID: "x", Population: 100000, Developed: 1.0}
	if maxed != g.Demand(capped) {
		t.Errorf("ProjectDemand のクランプ = %d, want %d", maxed, g.Demand(capped))
	}

	if got := g.ProjectDemand(nil, 10); got != 0 {
		t.Errorf("ProjectDemand(nil) = %d, want 0", got)
	}
}

func TestFacilityCost(t *testing.T) {
	g := simGame()
	city := &City{ID: "x", LandValue: 50} // landFactor = 1.0

	// 種類既知（Phase 1-8-2.5 で基本コストを旧比 0.6 倍に削減）
	if got := g.FacilityCost(city, "housing", 1); got != 3_000_000 {
		t.Errorf("FacilityCost(housing, 1) = %d, want 3000000", got)
	}
	if got := g.FacilityCost(city, "office", 2); got != 14_400_000 {
		t.Errorf("FacilityCost(office, 2) = %d, want 14400000", got)
	}
	// 種類未知は既定値 5,000,000
	if got := g.FacilityCost(city, "unknown", 1); got != 5_000_000 {
		t.Errorf("FacilityCost(unknown, 1) = %d, want 5000000", got)
	}
	// 地価補正
	low := &City{ID: "x", LandValue: 0}
	high := &City{ID: "x", LandValue: 100}
	cl := g.FacilityCost(low, "housing", 1)
	ch := g.FacilityCost(high, "housing", 1)
	if !(cl < ch) {
		t.Errorf("地価補正が不正: land=0 → %d, land=100 → %d", cl, ch)
	}
	if cl != 1_500_000 {
		t.Errorf("FacilityCost(land=0) = %d, want 1500000", cl)
	}
	if ch != 4_500_000 {
		t.Errorf("FacilityCost(land=100) = %d, want 4500000", ch)
	}
	if got := g.FacilityCost(nil, "housing", 1); got != 0 {
		t.Errorf("FacilityCost(nil) = %d, want 0", got)
	}
}

func TestFacilityCost_ScaleClamp(t *testing.T) {
	g := simGame()
	city := &City{ID: "x", LandValue: 50}

	if got, want := g.FacilityCost(city, "housing", 0), g.FacilityCost(city, "housing", 1); got != want {
		t.Errorf("scale=0 → %d, want scale=1 と同じ %d", got, want)
	}
	if got, want := g.FacilityCost(city, "housing", -3), g.FacilityCost(city, "housing", 1); got != want {
		t.Errorf("scale=-3 → %d, want scale=1 と同じ %d", got, want)
	}
	if got, want := g.FacilityCost(city, "housing", 6), g.FacilityCost(city, "housing", 5); got != want {
		t.Errorf("scale=6 → %d, want scale=5 と同じ %d", got, want)
	}
}

func TestSeasonalFactor_Values(t *testing.T) {
	g := simGame()
	want := [12]float64{
		0.90, 0.92, 1.00, 1.05, 1.05, 1.00,
		1.05, 1.08, 1.00, 1.00, 0.95, 0.95,
	}
	for m := 1; m <= 12; m++ {
		g.Date = Date{Year: 1907, Month: m}
		if got := g.SeasonalFactor(); got != want[m-1] {
			t.Errorf("月 %d の季節係数 = %v, want %v", m, got, want[m-1])
		}
	}
}

func TestSim_Deterministic(t *testing.T) {
	g1 := NewGame(1907, "hankyu")
	g2 := NewGame(1907, "hankyu")
	g1.AddCity(&City{ID: "umeda", Name: "梅田", Population: 100000})
	g2.AddCity(&City{ID: "umeda", Name: "梅田", Population: 100000})
	g1.AddCity(&City{ID: "takarazuka", Name: "宝塚", Population: 50000})
	g2.AddCity(&City{ID: "takarazuka", Name: "宝塚", Population: 50000})

	if g1.DistanceKm("umeda", "takarazuka") != g2.DistanceKm("umeda", "takarazuka") {
		t.Error("DistanceKm が決定論的でない")
	}
	if g1.LineCost("umeda", "takarazuka", 1067, false) !=
		g2.LineCost("umeda", "takarazuka", 1067, false) {
		t.Error("LineCost が決定論的でない")
	}
	if g1.DefaultFare("umeda", "takarazuka") != g2.DefaultFare("umeda", "takarazuka") {
		t.Error("DefaultFare が決定論的でない")
	}
}
