package core

import "testing"

func TestIsValidSeverity(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"info", SeverityInfo, true},
		{"success", SeveritySuccess, true},
		{"warning", SeverityWarning, true},
		{"danger", SeverityDanger, true},
		{"未知", "critical", false},
		{"空", "", false},
		{"大文字", "INFO", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidSeverity(tt.s); got != tt.want {
				t.Errorf("IsValidSeverity(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestSeverityRank(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{"info", SeverityInfo, 0},
		{"success", SeveritySuccess, 1},
		{"warning", SeverityWarning, 2},
		{"danger", SeverityDanger, 3},
		{"未知", "unknown", -1},
		{"空", "", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SeverityRank(tt.s); got != tt.want {
				t.Errorf("SeverityRank(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}
