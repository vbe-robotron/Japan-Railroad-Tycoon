package core

import (
	"hash/fnv"
	"math"
)

// Sim はゲームのシミュレーション用パラメータを集約する。
//
// Phase 1-5 ではすべて定数。Phase 7 で JSON から読み込む予定。
// Sim 自体はステートレスな純粋関数群の名前空間として予約されており、
// 実際の計算は Game のメソッドとして実装している（Game の文脈を利用するため）。
type Sim struct{}

// BaseCostPerKm は敷設コストの基本単価（円/km）。
//
// 時代によって変動する予定。Phase 1-7.6 では資産回転率を史実の鉄道に
// 近づけるため、5000 → 250000 に引き上げた（50倍）。
// 梅田〜宝塚（19.1km）の建設費は約 4,775,000 円となり、年間収入の
// 約 0.9 倍・投資回収約 4.6 年（史実の私鉄 10-20 年より短いが、
// ゲーム進行を考慮した値）になる。
const BaseCostPerKm = 250000

// MaintenancePerKm は維持費の基本単価（円/km/月）。
//
// Phase 1-7.5 で 1000 → 20000 に引き上げ、史実の私鉄の営業係数
// （コスト/収入）70-90 に近づけた。
//
// Phase 1-7.7 では需要モデルを人口のべき乗則（指数 0.6）へ変更した。
// これにより同じ人口でも旅客収入がおおむね 1/6 程度まで低下するため、
// 営業係数を維持するには維持費単価も同じ比率でスケールさせる必要がある。
// 梅田〜宝塚（人口 12 万＋3 万）で 20000 → 2800（約 1/7.1）とした。
// これにより同路線の営業係数は 70-90 を維持し、田舎路線も 150-200 を
// 維持する（需要のべき乗則化と営業係数維持の両立には不可欠な調整）。
const MaintenancePerKm = 2800

// FareElasticity は運賃弾力性（需要への影響）。
const FareElasticity = -0.7

// DemandBase は旅客需要の基準人口（人）。
//
// Phase 1-7.7 以降、需要は人口のべき乗則となり、この値を基準に
// 正規化する: (Population / DemandBase)^DemandExponent。
const DemandBase = 1000

// DemandExponent は旅客需要の人口に対するべき指数。
//
// 都市の鉄道需要は人口の 0.5〜0.7 乗に比例するという交通工学の
// 経験則（グラビティモデル）の中間値 0.6 を採用する。
// 大都市でも需要が人口に線形比例して過大にならないようにする。
const DemandExponent = 0.6

// seasonalFactors は月ごとの季節係数（0.9-1.1）。
var seasonalFactors = [12]float64{
	0.90, 0.92, 1.00, 1.05, 1.05, 1.00,
	1.05, 1.08, 1.00, 1.00, 0.95, 0.95,
}

// facilityBaseCosts は施設種別ごとの基本建設コスト（円）。
var facilityBaseCosts = map[string]int64{
	"department_store": 10_000_000,
	"housing":          5_000_000,
	"amusement":        8_000_000,
	"hotel":            7_000_000,
	"office":           12_000_000,
}

// DistanceKm は 2 都市間の距離を km で返す。
//
// Haversine 距離に地形係数をかけたものを返す。
// 都市が見つからない場合は 0 を返す。
func (g *Game) DistanceKm(fromID, toID string) float64 {
	from, ok := g.Cities[fromID]
	if !ok {
		return 0
	}
	to, ok := g.Cities[toID]
	if !ok {
		return 0
	}
	// 地形は Phase 1-5 では平地固定
	// （Phase 1-6 以降、ノード間の地形データを参照する）
	return g.DistanceKmWithTerrain(from, to, TerrainPlain)
}

// DistanceKmWithTerrain は地形を指定して距離を返す。
//
// 都市の座標は Phase 1-5 では cityLatLon の暫定実装から得る。
func (g *Game) DistanceKmWithTerrain(from, to *City, terrain TerrainKind) float64 {
	if from == nil || to == nil {
		return 0
	}
	// 都市の座標は Node から取る予定だが、Phase 1-5 では
	// cityLatLon の暫定実装（都市 ID のハッシュから擬似座標）を使う。
	// Phase 7 で nodes.json を読み込んだら Node の座標を使う。
	lat1, lon1 := g.cityLatLon(from)
	lat2, lon2 := g.cityLatLon(to)
	straight := HaversineKm(lat1, lon1, lat2, lon2)
	t := LookupTerrain(terrain)
	return straight * t.DistanceFactor
}

// cityLatLon は都市の緯度経度を返す。
//
// City.Lat/Lon が設定されていれば実座標を返す。未設定の場合は
// Phase 1-5 の暫定実装として City ID のハッシュから決定論的に
// 日本の緯度経度範囲（北緯 30-46、東経 128-146）にマッピングする。
// Phase 7 で nodes.json を読み込んだら、Node の座標を使うように差し替える。
func (g *Game) cityLatLon(c *City) (float64, float64) {
	if c.Lat != 0 && c.Lon != 0 {
		return c.Lat, c.Lon
	}
	h := fnv.New64a()
	h.Write([]byte(c.ID))
	v := h.Sum64()
	lat := 30.0 + float64(v%1600)/100.0        // 30.00-46.00
	lon := 128.0 + float64((v>>16)%1800)/100.0 // 128.00-146.00
	return lat, lon
}

// LineCost は路線の敷設コストを円で返す。
//
// Cost = BaseCostPerKm × DistanceKm × TerrainBuildCost × GaugeFactor × DoubleFactor
func (g *Game) LineCost(fromID, toID string, gauge int, double bool) int64 {
	return g.LineCostWithTerrain(fromID, toID, TerrainPlain, gauge, double)
}

// LineCostWithTerrain は地形を指定してコストを返す。
func (g *Game) LineCostWithTerrain(fromID, toID string, terrain TerrainKind, gauge int, double bool) int64 {
	from, ok := g.Cities[fromID]
	if !ok {
		return 0
	}
	to, ok := g.Cities[toID]
	if !ok {
		return 0
	}
	dist := g.DistanceKmWithTerrain(from, to, terrain)
	t := LookupTerrain(terrain)
	cost := float64(BaseCostPerKm) * dist * t.BuildCost *
		GaugeFactor(gauge) * DoubleFactor(double)
	return int64(math.Round(cost))
}

// GaugeFactor は軌間のコスト係数を返す。
//
// 1067=1.0, 1435=1.2, 762=0.8, それ以外は 1.0。
func GaugeFactor(gauge int) float64 {
	switch gauge {
	case 1067:
		return 1.0
	case 1435:
		return 1.2
	case 762:
		return 0.8
	default:
		return 1.0
	}
}

// DoubleFactor は複線化のコスト係数を返す。
//
// 単線=1.0, 複線=1.7。
func DoubleFactor(double bool) float64 {
	if double {
		return 1.7
	}
	return 1.0
}

// DefaultFare は 2 都市間のデフォルト運賃を円で返す。
//
// 距離に応じた線形モデル: Fare = 50 + 8 × DistanceKm
// 10 円単位に丸める。
func (g *Game) DefaultFare(fromID, toID string) int {
	dist := g.DistanceKm(fromID, toID)
	fare := 50.0 + 8.0*dist
	// 10 円単位に丸める
	return int(math.Round(fare/10.0) * 10)
}

// Demand は都市の旅客需要を返す。
//
// Phase 1-7.7 で人口に線形比例するモデルからべき乗則へ変更した。
//
//	Demand = (Population / DemandBase)^0.6 × (1 + Developed) × SeasonalFactor
//
// 指数 0.6 は交通工学の経験則（鉄道需要は人口の 0.5〜0.7 乗）に基づく。
// これにより大都市（例: 神戸 40 万）の需要が人口に線形比例して
// 過大になるのを防ぎ、大都市間路線の営業係数を現実的な範囲に収める。
func (g *Game) Demand(city *City) int {
	if city == nil {
		return 0
	}
	// べき乗則（基準人口で正規化）
	base := math.Pow(float64(city.Population), DemandExponent) /
		math.Pow(DemandBase, DemandExponent)
	developed := 1.0 + city.Developed
	seasonal := g.SeasonalFactor()
	return int(base * developed * seasonal)
}

// ProjectDemand は将来の需要を予測する（簡易版）。
//
// years 年後を想定した開発度の増加を線形で仮定する。
// Phase 1-5 では Develop の増加のみ考慮する。
func (g *Game) ProjectDemand(city *City, years int) int {
	if city == nil || years <= 0 {
		return g.Demand(city)
	}
	// 開発度が年 0.02 ずつ上がると仮定
	future := *city
	d := future.Developed + 0.02*float64(years)
	if d > 1.0 {
		d = 1.0
	}
	future.Developed = d
	return g.Demand(&future)
}

// FacilityCost は施設の建設コストを円で返す。
//
// 種類・規模・地価に依存する。scale は 1-5 にクランプされる。
func (g *Game) FacilityCost(city *City, kind string, scale int) int64 {
	if city == nil {
		return 0
	}
	base, ok := facilityBaseCosts[kind]
	if !ok {
		base = 5_000_000
	}
	if scale < 1 {
		scale = 1
	}
	if scale > 5 {
		scale = 5
	}
	// 地価補正: 地価が高いほどコスト増
	landFactor := 0.5 + float64(city.LandValue)/100.0
	cost := float64(base) * float64(scale) * landFactor
	return int64(math.Round(cost))
}

// SeasonalFactor は季節係数を返す。
//
// 月ごとに 0.9-1.1 の範囲で変動する。月が不正な場合は 1.0。
func (g *Game) SeasonalFactor() float64 {
	m := g.Date.Month
	if m < 1 || m > 12 {
		return 1.0
	}
	return seasonalFactors[m-1]
}
