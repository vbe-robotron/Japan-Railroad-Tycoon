package core

import (
	"math/rand"
	"testing"
)

func TestDate_String(t *testing.T) {
	tests := []struct {
		name string
		date Date
		want string
	}{
		{"通常", Date{Year: 1910, Month: 3}, "1910年3月"},
		{"1桁月", Date{Year: 1872, Month: 10}, "1872年10月"},
		{"年0", Date{Year: 0, Month: 1}, "0年1月"},
		{"負の年", Date{Year: -100, Month: 12}, "-100年12月"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.date.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDate_AddMonths(t *testing.T) {
	tests := []struct {
		name string
		date Date
		n    int
		want Date
	}{
		{"0ヶ月", Date{1910, 3}, 0, Date{1910, 3}},
		{"月内", Date{1910, 3}, 5, Date{1910, 8}},
		{"12ヶ月で1年", Date{1910, 3}, 12, Date{1911, 3}},
		{"年末繰り上がり", Date{1910, 11}, 3, Date{1911, 2}},
		{"24ヶ月で2年", Date{1910, 1}, 24, Date{1912, 1}},
		{"負の月", Date{1910, 3}, -5, Date{1909, 10}},
		{"負で年をまたぐ", Date{1910, 1}, -1, Date{1909, 12}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.date.AddMonths(tt.n); got != tt.want {
				t.Errorf("AddMonths(%d) = %+v, want %+v", tt.n, got, tt.want)
			}
		})
	}
}

func TestDate_Before(t *testing.T) {
	tests := []struct {
		name string
		d, o Date
		want bool
	}{
		{"年が前", Date{1910, 12}, Date{1911, 1}, true},
		{"月が前", Date{1910, 1}, Date{1910, 2}, true},
		{"同一", Date{1910, 3}, Date{1910, 3}, false},
		{"後", Date{1911, 1}, Date{1910, 12}, false},
		{"年が後で月が前", Date{1911, 1}, Date{1910, 12}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.Before(tt.o); got != tt.want {
				t.Errorf("%+v.Before(%+v) = %v, want %v", tt.d, tt.o, got, tt.want)
			}
		})
	}
}

func TestDate_After(t *testing.T) {
	tests := []struct {
		name string
		d, o Date
		want bool
	}{
		{"年が後", Date{1911, 1}, Date{1910, 12}, true},
		{"月が後", Date{1910, 2}, Date{1910, 1}, true},
		{"同一", Date{1910, 3}, Date{1910, 3}, false},
		{"前", Date{1910, 12}, Date{1911, 1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.After(tt.o); got != tt.want {
				t.Errorf("%+v.After(%+v) = %v, want %v", tt.d, tt.o, got, tt.want)
			}
		})
	}
}

func TestDate_Equal(t *testing.T) {
	tests := []struct {
		name string
		d, o Date
		want bool
	}{
		{"同じ", Date{1910, 3}, Date{1910, 3}, true},
		{"月違い", Date{1910, 3}, Date{1910, 4}, false},
		{"年違い", Date{1910, 3}, Date{1911, 3}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.Equal(tt.o); got != tt.want {
				t.Errorf("%+v.Equal(%+v) = %v, want %v", tt.d, tt.o, got, tt.want)
			}
		})
	}
}

func TestDate_IsZero(t *testing.T) {
	tests := []struct {
		name string
		date Date
		want bool
	}{
		{"ゼロ値", Date{}, true},
		{"年のみ", Date{Year: 1910}, false},
		{"月のみ", Date{Month: 3}, false},
		{"通常", Date{1910, 3}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.date.IsZero(); got != tt.want {
				t.Errorf("%+v.IsZero() = %v, want %v", tt.date, got, tt.want)
			}
		})
	}
}

func TestNewDate_Normalize(t *testing.T) {
	tests := []struct {
		name  string
		year  int
		month int
		want  Date
	}{
		{"通常", 1910, 3, Date{1910, 3}},
		{"月13は翌年1月", 1910, 13, Date{1911, 1}},
		{"月0は前年12月", 1910, 0, Date{1909, 12}},
		{"月25は2年後の1月", 1910, 25, Date{1912, 1}},
		{"月-1は前年11月", 1910, -1, Date{1909, 11}},
		{"12月はそのまま", 1910, 12, Date{1910, 12}},
		{"負の年", -1, 6, Date{-1, 6}},
		{"負の年月0", -1, 0, Date{-2, 12}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDate(tt.year, tt.month); got != tt.want {
				t.Errorf("NewDate(%d, %d) = %+v, want %+v", tt.year, tt.month, got, tt.want)
			}
		})
	}
}

func TestFloorDiv(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"正の割り切れる", 24, 12, 2},
		{"正の端数", 25, 12, 2},
		{"零", 0, 12, 0},
		{"負の割り切れる", -24, 12, -2},
		{"負の端数は切り下げ", -7, 12, -1},
		{"負の端数2", -13, 12, -2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := floorDiv(tt.a, tt.b); got != tt.want {
				t.Errorf("floorDiv(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestFloorMod(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"正", 25, 12, 1},
		{"零", 24, 12, 0},
		{"負の剰余は正へ", -7, 12, 5},
		{"負の剰余2", -1, 12, 11},
		{"負の割り切れる", -24, 12, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := floorMod(tt.a, tt.b); got != tt.want {
				t.Errorf("floorMod(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestDate_AddMonths_Negative(t *testing.T) {
	tests := []struct {
		name string
		date Date
		n    int
		want Date
	}{
		{"負の年で0ヶ月", Date{-1, 6}, 0, Date{-1, 6}},
		{"0年1月の1ヶ月前", Date{0, 1}, -1, Date{-1, 12}},
		{"1910年1月の1ヶ月前", Date{1910, 1}, -1, Date{1909, 12}},
		{"負の年の前進", Date{-1, 6}, 12, Date{0, 6}},
		{"負の年から負へ", Date{-1, 6}, -13, Date{-2, 5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.date.AddMonths(tt.n); got != tt.want {
				t.Errorf("%+v.AddMonths(%d) = %+v, want %+v", tt.date, tt.n, got, tt.want)
			}
		})
	}
}

func TestDate_AddMonths_FloorDiv(t *testing.T) {
	// floor 除算が使われていなければ月が 1-12 の範囲外になるケース。
	d := Date{-1, 6}
	got := d.AddMonths(0)
	if got.Month < 1 || got.Month > 12 {
		t.Fatalf("AddMonths(0) の月が範囲外: %+v", got)
	}
	if got != (Date{-1, 6}) {
		t.Fatalf("AddMonths(0) = %+v, want {-1 6}", got)
	}
}

func TestDate_AddMonths_Random(t *testing.T) {
	r := rand.New(rand.NewSource(1907))
	for i := 0; i < 1000; i++ {
		d := NewDate(r.Intn(4000)-2000, r.Intn(24)-12)
		n := r.Intn(2401) - 1200
		back := d.AddMonths(n).AddMonths(-n)
		if back != d {
			t.Fatalf("case %d: %+v.AddMonths(%d).AddMonths(%d) = %+v, want %+v",
				i, d, n, -n, back, d)
		}
		if back.Month < 1 || back.Month > 12 {
			t.Fatalf("case %d: 月が範囲外: %+v", i, back)
		}
	}
}
