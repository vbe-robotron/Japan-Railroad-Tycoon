package core

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
