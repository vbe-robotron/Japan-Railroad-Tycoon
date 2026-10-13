package core

import "testing"

func TestFormatYen(t *testing.T) {
	tests := []struct {
		name string
		v    int64
		want string
	}{
		{"0円", 0, "0円"},
		{"1円", 1, "1円"},
		{"9999円", 9999, "9999円"},
		{"1万円", 1_0000, "1.0万円"},
		{"15000円", 1_5000, "1.5万円"},
		{"9999万円", 9999_0000, "9999.0万円"},
		{"1億円", 1_0000_0000, "1.0億円"},
		{"1億5000万円", 1_5000_0000, "1.5億円"},
		{"1兆円", 1_0000_0000_0000, "1.0兆円"},
		{"2.5兆円", 2_5000_0000_0000, "2.5兆円"},
		{"負の億", -1_5000_0000, "-1.5億円"},
		{"負の万円", -1_5000, "-1.5万円"},
		{"負の円", -100, "-100円"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatYen(tt.v); got != tt.want {
				t.Errorf("FormatYen(%d) = %q, want %q", tt.v, got, tt.want)
			}
		})
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name string
		v    int64
		want string
	}{
		{"0", 0, "0"},
		{"1桁", 7, "7"},
		{"3桁", 123, "123"},
		{"4桁", 1234, "1,234"},
		{"7桁", 1234567, "1,234,567"},
		{"10桁", 1234567890, "1,234,567,890"},
		{"負", -1234, "-1,234"},
		{"負1桁", -7, "-7"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatNumber(tt.v); got != tt.want {
				t.Errorf("FormatNumber(%d) = %q, want %q", tt.v, got, tt.want)
			}
		})
	}
}
