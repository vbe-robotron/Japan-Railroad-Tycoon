package core

// このファイルは鉄道技術（Gate 1: 技術ゲート）のテーブルを定義する。
//
// Phase 1-6 ではコード上のテーブルとして保持する。Phase 7 で JSON から
// 読み込む（Mod 対応）予定のため、定数ではなくテーブルと検索関数を提供する。

// Technology は鉄道技術を表す。
type Technology struct {
	ID          string // "electric_tram"
	Name        string // "電気軌道"
	UnlockYear  int    // 1895
	Kind        string // "track"|"traction"|"structure"
	Description string // 説明（任意）
}

// TechnologyTable は全技術のテーブル。
//
// 解禁年の昇順で並べている。ID は不変の識別子であり、JSON の Mod でも
// このキーで参照する。
var TechnologyTable = []Technology{
	{ID: "horse_tram", Name: "馬車鉄道", UnlockYear: 1880, Kind: "traction"},
	{ID: "steam_railway", Name: "蒸気鉄道", UnlockYear: 1887, Kind: "traction"},
	{ID: "electric_tram", Name: "電気軌道", UnlockYear: 1895, Kind: "traction"},
	{ID: "street_car", Name: "路面電車", UnlockYear: 1900, Kind: "traction"},
	{ID: "suburban_railway", Name: "郊外電鉄", UnlockYear: 1905, Kind: "traction"},
	{ID: "subway", Name: "地下鉄", UnlockYear: 1925, Kind: "structure"},
	{ID: "electrification", Name: "電化（長距離）", UnlockYear: 1930, Kind: "traction"},
	{ID: "diesel", Name: "ディーゼル", UnlockYear: 1950, Kind: "traction"},
	{ID: "shinkansen", Name: "新幹線", UnlockYear: 1964, Kind: "traction"},
}

// TechnologyByID は ID から Technology を返す。
//
// 見つからない場合はゼロ値と false を返す。
func TechnologyByID(id string) (Technology, bool) {
	for _, t := range TechnologyTable {
		if t.ID == id {
			return t, true
		}
	}
	return Technology{}, false
}

// UnlockedTechnologies は指定年で解禁済みの技術を返す。
//
// 解禁年が year 以下（UnlockYear <= year）の技術を TechnologyTable の
// 順序で返す。year が負の場合でも安全に空スライスを返す。
func UnlockedTechnologies(year int) []Technology {
	out := []Technology{}
	for _, t := range TechnologyTable {
		if t.UnlockYear <= year {
			out = append(out, t)
		}
	}
	return out
}
