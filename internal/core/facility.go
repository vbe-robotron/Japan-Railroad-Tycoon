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
//
// Phase 1-8-2.5 では開発が強すぎた（施設収入で資金が余り、投資回収が
// 短すぎた）ため、基本コストを旧比 0.6 倍（40% 削減）、収入率を旧比
// 0.625 倍に調整した。これにより開発効果と施設収入が抑えられ、旧実測
// （1931 年: 資金 62 倍・年率 19%）から目標（24 年 5〜10 倍）に近づく。
var FacilityTable = map[string]FacilityDef{
	FacilityDepartmentStore: {FacilityDepartmentStore, "百貨店", 6_000_000, 0.15, 0.005},
	FacilityHousing:         {FacilityHousing, "住宅地", 3_000_000, 0.20, 0.003},
	FacilityAmusement:       {FacilityAmusement, "遊園地", 4_800_000, 0.10, 0.004},
	FacilityHotel:           {FacilityHotel, "ホテル", 4_200_000, 0.08, 0.0035},
	FacilityOffice:          {FacilityOffice, "オフィス", 7_200_000, 0.12, 0.0055},
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
