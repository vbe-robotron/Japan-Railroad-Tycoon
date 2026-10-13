package core

// City は市区町村の代表駅を持つ都市を表す。
//
// RepresentativeNode は後の Phase で導入する駅ノード（Node.ID）を指す。
// 地価や人口、開発度など不動産・流通収益の基礎となる。
type City struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Prefecture         string   `json:"prefecture"`
	RepresentativeNode string   `json:"representative_node"` // NodeID
	Population         int      `json:"population"`
	LandValue          int      `json:"land_value"` // 地価係数 0-100
	Demand             int      `json:"demand"`
	Developed          float64  `json:"developed"` // 0.0-1.0
	OwnerID            string   `json:"owner_id,omitempty"`
	PoliticalPressure  float64  `json:"political_pressure"`
	ResidentsAffected  int      `json:"residents_affected"`
	Tags               []string `json:"tags"`
}

// NewCity は都市を生成する。
//
// 開発度と政治的圧力は 0.0 で、Tags は空スライスで初期化する。
func NewCity(id, name, prefecture string) *City {
	return &City{
		ID:                id,
		Name:              name,
		Prefecture:        prefecture,
		Developed:         0.0,
		PoliticalPressure: 0.0,
		Tags:              []string{},
	}
}
