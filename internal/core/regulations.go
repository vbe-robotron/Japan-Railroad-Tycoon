package core

// このファイルは法規・制度（Gate 2: 法規ゲート）のテーブルを定義する。
//
// Phase 1-6 ではコード上のテーブルとして保持する。Phase 7 で JSON から
// 読み込む予定。Effects は法規の効果を表すタグで、他システムが参照する。

// Regulation は法規・制度を表す。
type Regulation struct {
	ID          string   // "private_railway_ordinance"
	Name        string   // "私設鉄道条例"
	EnactedYear int      // 1887
	Effects     []string // 効果のタグ
}

// RegulationTable は全法規のテーブル。
//
// 制定年の昇順で並べている。ID は不変の識別子であり、JSON の Mod でも
// このキーで参照する。
var RegulationTable = []Regulation{
	{ID: "private_railway_ordinance", Name: "私設鉄道条例", EnactedYear: 1887, Effects: []string{"private_railway_allowed"}},
	{ID: "railway_construction_law", Name: "鉄道敷設法", EnactedYear: 1892, Effects: []string{"trunk_lines_state_led"}},
	{ID: "private_railway_law", Name: "私設鉄道法", EnactedYear: 1900, Effects: []string{"strict_licensing"}},
	{ID: "railway_nationalization_law", Name: "鉄道国有法", EnactedYear: 1906, Effects: []string{"major_private_nationalized"}},
	{ID: "local_railway_law", Name: "地方鉄道法", EnactedYear: 1919, Effects: []string{"suburban_railway_boom"}},
	{ID: "wartime_acquisition", Name: "戦時買収", EnactedYear: 1943, Effects: []string{"wartime_nationalization"}},
	{ID: "jnr_law", Name: "日本国有鉄道法", EnactedYear: 1949, Effects: []string{"jnr_founded"}},
	{ID: "jnr_privatization", Name: "国鉄分割民営化", EnactedYear: 1987, Effects: []string{"jr_group_founded"}},
}

// RegulationByID は ID から Regulation を返す。
//
// 見つからない場合はゼロ値と false を返す。
func RegulationByID(id string) (Regulation, bool) {
	for _, r := range RegulationTable {
		if r.ID == id {
			return r, true
		}
	}
	return Regulation{}, false
}

// EffectiveRegulations は指定年で有効な法規を返す。
//
// 制定年が year 以下（EnactedYear <= year）の法規を RegulationTable の
// 順序で返す。
func EffectiveRegulations(year int) []Regulation {
	out := []Regulation{}
	for _, r := range RegulationTable {
		if r.EnactedYear <= year {
			out = append(out, r)
		}
	}
	return out
}

// IsRegulationActive は指定年で指定法規が有効かを返す。
//
// 未知の ID や未制定の法規は false を返す。
func IsRegulationActive(id string, year int) bool {
	r, ok := RegulationByID(id)
	if !ok {
		return false
	}
	return r.EnactedYear <= year
}
