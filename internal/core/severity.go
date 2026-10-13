package core

// イベントの重要度。フィルタや色分けに用いる。
const (
	SeverityInfo    = "info"
	SeveritySuccess = "success"
	SeverityWarning = "warning"
	SeverityDanger  = "danger"
)

// severityRank は各重要度の並び順を定義する。
var severityRank = map[string]int{
	SeverityInfo:    0,
	SeveritySuccess: 1,
	SeverityWarning: 2,
	SeverityDanger:  3,
}

// IsValidSeverity は s が既知の重要度かどうかを返す。
func IsValidSeverity(s string) bool {
	_, ok := severityRank[s]
	return ok
}

// SeverityRank は重要度の並び順を返す。
//
// info=0, success=1, warning=2, danger=3。未知の値は -1 を返す。
func SeverityRank(s string) int {
	if r, ok := severityRank[s]; ok {
		return r
	}
	return -1
}
