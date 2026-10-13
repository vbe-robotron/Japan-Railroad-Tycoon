package core

import "testing"

func TestLookupTerrain_AllKinds(t *testing.T) {
	kinds := []TerrainKind{
		TerrainPlain, TerrainHill, TerrainMountain,
		TerrainRiver, TerrainUrban, TerrainSea,
	}
	for _, k := range kinds {
		got := LookupTerrain(k)
		if got.Kind != k {
			t.Errorf("LookupTerrain(%q).Kind = %q, want %q", k, got.Kind, k)
		}
		if got != TerrainTable[k] {
			t.Errorf("LookupTerrain(%q) = %+v, want %+v", k, got, TerrainTable[k])
		}
	}
}

func TestLookupTerrain_Unknown(t *testing.T) {
	got := LookupTerrain(TerrainKind("volcano"))
	if got != TerrainTable[TerrainPlain] {
		t.Errorf("未知地形 = %+v, want 平地 %+v", got, TerrainTable[TerrainPlain])
	}
}

func TestTerrainTable_Values(t *testing.T) {
	tests := []struct {
		kind   TerrainKind
		build  float64
		dist   float64
		speed  int
	}{
		{TerrainPlain, 1.0, 1.1, 130},
		{TerrainHill, 1.5, 1.3, 100},
		{TerrainMountain, 3.0, 1.8, 80},
		{TerrainRiver, 2.0, 1.4, 110},
		{TerrainUrban, 2.5, 1.15, 100},
		{TerrainSea, 999, 0.0, 0},
	}
	for _, tt := range tests {
		tr, ok := TerrainTable[tt.kind]
		if !ok {
			t.Fatalf("TerrainTable に %q がない", tt.kind)
		}
		if tr.BuildCost != tt.build {
			t.Errorf("%s.BuildCost = %v, want %v", tt.kind, tr.BuildCost, tt.build)
		}
		if tr.DistanceFactor != tt.dist {
			t.Errorf("%s.DistanceFactor = %v, want %v", tt.kind, tr.DistanceFactor, tt.dist)
		}
		if tr.MaxSpeed != tt.speed {
			t.Errorf("%s.MaxSpeed = %d, want %d", tt.kind, tr.MaxSpeed, tt.speed)
		}
	}
}
