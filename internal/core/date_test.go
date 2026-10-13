package core

import "testing"

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDate(tt.year, tt.month); got != tt.want {
				t.Errorf("NewDate(%d, %d) = %+v, want %+v", tt.year, tt.month, got, tt.want)
			}
		})
	}
}
