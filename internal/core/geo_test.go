package core

import (
	"math"
	"testing"
)

func TestHaversineKm_TokyoOsaka(t *testing.T) {
	// 東京駅: 北緯 35.6812、東経 139.7671
	// 大阪駅: 北緯 34.7024、東経 135.4959
	got := HaversineKm(35.6812, 139.7671, 34.7024, 135.4959)
	t.Logf("東京〜大阪の Haversine 距離 = %.1f km", got)

	const want = 400.0
	lo, hi := want*0.95, want*1.05
	if got < lo || got > hi {
		t.Errorf("HaversineKm(東京, 大阪) = %.1f km, want %.1f-%.1f km", got, lo, hi)
	}
}

func TestHaversineKm_SamePoint(t *testing.T) {
	if got := HaversineKm(35.0, 139.0, 35.0, 139.0); got != 0 {
		t.Errorf("同一地点 = %v, want 0", got)
	}
}

func TestHaversineKm_Equator(t *testing.T) {
	// 赤道で経度 1 度 ≈ 111.19 km
	got := HaversineKm(0, 0, 0, 1)
	if math.Abs(got-111.19) > 0.5 {
		t.Errorf("赤道 1 度 = %.2f km, want ≈111.19 km", got)
	}
}

func TestToRad(t *testing.T) {
	if got := toRad(180); math.Abs(got-math.Pi) > 1e-12 {
		t.Errorf("toRad(180) = %v, want π", got)
	}
	if got := toRad(0); got != 0 {
		t.Errorf("toRad(0) = %v, want 0", got)
	}
}
