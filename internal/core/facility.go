package core

// 施設の種類。
const (
	FacilityDepartmentStore = "department_store"
	FacilityHousing         = "housing"
	FacilityAmusement       = "amusement"
	FacilityHotel           = "hotel"
	FacilityOffice          = "office"
)

// FacilityDef は施設の種類ごとの基本パラメータを表す。
//
// DevelopGain は規模 1 のときの開発度上昇量、IncomeRate は建設コストに
// 対する月次収入率である。
type FacilityDef struct {
	Kind        string
	Name        string  // 日本語名
	BaseCost    int64   // 基本コスト（円）
	DevelopGain float64 // 開発度上昇（scale=1）
	IncomeRate  float64 // 月次収入率（コストに対する割合）
}

// FacilityTable は施設の定義テーブル。
//
// 種類（Kind）から FacilityDef を引く。基本コストは Game.FacilityCost が
// 地価・規模補正を掛ける前の値である。
var FacilityTable = map[string]FacilityDef{
	FacilityDepartmentStore: {FacilityDepartmentStore, "百貨店", 10_000_000, 0.15, 0.008},
	FacilityHousing:         {FacilityHousing, "住宅地", 5_000_000, 0.20, 0.005},
	FacilityAmusement:       {FacilityAmusement, "遊園地", 8_000_000, 0.10, 0.007},
	FacilityHotel:           {FacilityHotel, "ホテル", 7_000_000, 0.08, 0.006},
	FacilityOffice:          {FacilityOffice, "オフィス", 12_000_000, 0.12, 0.009},
}

// FacilityByKind は種類から定義を返す。
//
// 未知の種類の場合は ok=false を返す。
func FacilityByKind(kind string) (FacilityDef, bool) {
	def, ok := FacilityTable[kind]
	return def, ok
}

// ScaleDevelopFactor は規模による開発度上昇の倍率を返す。
//
// scale=1: 1.0, scale=2: 1.3, ..., scale=5: 2.2。
// 範囲外の規模は 1-5 にクランプする。
func ScaleDevelopFactor(scale int) float64 {
	if scale < 1 {
		scale = 1
	}
	if scale > 5 {
		scale = 5
	}
	return 1.0 + float64(scale-1)*0.3
}

// Facility は会社が都市に建設した開発施設を表す。
//
// 種類（Kind）は "department_store"|"housing"|"amusement"|"hotel"|"office"。
// 規模（Scale）は 1-5 の範囲にクランプされる。
type Facility struct {
	ID        string `json:"id"`
	CompanyID string `json:"company_id"`
	CityID    string `json:"city_id"`
	Kind      string `json:"kind"`  // "department_store"|"housing"|"amusement"|"hotel"|"office"
	Scale     int    `json:"scale"` // 1-5
	BuiltAt   Date   `json:"built_at"`
	Income    int64  `json:"income"` // 月次収入
}

// NewFacility は施設を生成する。
//
// Scale は 1-5 にクランプする。範囲外の値でも panic せず丸める。
func NewFacility(id, companyID, cityID, kind string, scale int) *Facility {
	if scale < 1 {
		scale = 1
	} else if scale > 5 {
		scale = 5
	}
	return &Facility{
		ID:        id,
		CompanyID: companyID,
		CityID:    cityID,
		Kind:      kind,
		Scale:     scale,
		Income:    0,
	}
}
