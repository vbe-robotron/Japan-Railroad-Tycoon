package core

// TerrainKind は地形の種類を表す。
type TerrainKind string

const (
	TerrainPlain    TerrainKind = "plain"
	TerrainHill     TerrainKind = "hill"
	TerrainMountain TerrainKind = "mountain"
	TerrainRiver    TerrainKind = "river"
	TerrainUrban    TerrainKind = "urban"
	TerrainSea      TerrainKind = "sea"
)

// Terrain は地形のパラメータを表す。
//
// BuildCost は敷設コスト係数、DistanceFactor は距離係数、
// MaxSpeed はこの地形を走行できる最高速度（km/h）を表す。
type Terrain struct {
	Kind           TerrainKind
	BuildCost      float64 // 敷設コスト係数
	DistanceFactor float64 // 距離係数
	MaxSpeed       int     // km/h
}

// TerrainTable は地形ごとのパラメータ。
//
// 将来的に JSON から読み込む（Mod 対応）余地を残し、Phase 7 で差し替え可能に
// するため、定数ではなくテーブルとして保持する。
var TerrainTable = map[TerrainKind]Terrain{
	TerrainPlain:    {TerrainPlain, 1.0, 1.1, 130},
	TerrainHill:     {TerrainHill, 1.5, 1.3, 100},
	TerrainMountain: {TerrainMountain, 3.0, 1.8, 80},
	TerrainRiver:    {TerrainRiver, 2.0, 1.4, 110},
	TerrainUrban:    {TerrainUrban, 2.5, 1.15, 100},
	TerrainSea:      {TerrainSea, 999, 0.0, 0},
}

// LookupTerrain は TerrainKind から Terrain を返す。
//
// 未知の地形は平地として扱う。
func LookupTerrain(k TerrainKind) Terrain {
	if t, ok := TerrainTable[k]; ok {
		return t
	}
	return TerrainTable[TerrainPlain]
}
