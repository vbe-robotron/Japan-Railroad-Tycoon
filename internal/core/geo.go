package core

import "math"

// earthRadiusKm は地球の平均半径（km）。
const earthRadiusKm = 6371.0

// HaversineKm は 2 点の緯度経度から大円距離を km で返す。
//
// lat1, lon1, lat2, lon2 は度数法（degree）。
func HaversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

// toRad は度数法をラジアンに変換する。
func toRad(deg float64) float64 {
	return deg * math.Pi / 180
}
