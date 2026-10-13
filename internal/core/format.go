package core

import (
	"fmt"
	"strings"
)

// FormatYen は金額を日本語の単位付き文字列に整形する。
//
// 1兆円以上は "%.1f兆円"、1億円以上は "%.1f億円"、1万円以上は "%.1f万円"、
// それ未満は "%d円" となる。負の値も対応する（例: "-1.5億円"）。
func FormatYen(v int64) string {
	neg := v < 0
	if neg {
		v = -v
	}
	var s string
	switch {
	case v >= 1_0000_0000_0000:
		s = fmt.Sprintf("%.1f兆円", float64(v)/1_0000_0000_0000)
	case v >= 1_0000_0000:
		s = fmt.Sprintf("%.1f億円", float64(v)/1_0000_0000)
	case v >= 1_0000:
		s = fmt.Sprintf("%.1f万円", float64(v)/1_0000)
	default:
		s = fmt.Sprintf("%d円", v)
	}
	if neg {
		return "-" + s
	}
	return s
}

// FormatNumber は数値を3桁区切りの文字列に整形する。
//
// 例: 1234567 → "1,234,567"、-1234 → "-1,234"。
func FormatNumber(v int64) string {
	s := fmt.Sprintf("%d", v)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}

	var b strings.Builder
	n := len(s)
	for i := 0; i < n; i++ {
		if i > 0 && (n-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteByte(s[i])
	}

	if neg {
		return "-" + b.String()
	}
	return b.String()
}
