package core

import "fmt"

// Date はゲーム内の年月を表す。
//
// time.Time は使わず、年と月のみを保持する。年0や負の年も許容し、
// 史実の前史を扱えるようにしている。
type Date struct {
	Year  int `json:"year"`
	Month int `json:"month"` // 1-12
}

// floorDiv は a を b で割った床（floor）を返す。b > 0 を前提とする。
//
// Go の除算は 0 方向へ切り捨てるため、負の被除数では floor と一致しない。
// 例: -7/12 は Go では 0 だが、floor では -1。
func floorDiv(a, b int) int {
	q := a / b
	r := a % b
	if r != 0 && (r < 0) != (b < 0) {
		q--
	}
	return q
}

// floorMod は a を b で割った剰余（常に 0 <= r < b）を返す。b > 0 を前提とする。
//
// Go の剰余は被除数と同じ符号になるため、負の被除数では負の値を返す。
// 例: -7%12 は Go では -7 だが、floor では 5。
func floorMod(a, b int) int {
	r := a % b
	if r < 0 {
		r += b
	}
	return r
}

// NewDate は Date を生成する。
//
// 月が 1-12 の範囲外の場合は panic せず、年を繰り上げ・繰り下げして
// 正規化する。例: NewDate(1910, 13) は 1911年1月、NewDate(1910, 0) は
// 1909年12月になる。負の年も floor 除算で正しく正規化する。
func NewDate(year, month int) Date {
	total := year*12 + (month - 1)
	return Date{
		Year:  floorDiv(total, 12),
		Month: floorMod(total, 12) + 1,
	}
}

// String は "1910年3月" 形式の文字列を返す。
func (d Date) String() string {
	return fmt.Sprintf("%d年%d月", d.Year, d.Month)
}

// AddMonths は n ヶ月後の Date を返す。n は負でもよい。
//
// 負の年でも正しく動作するよう floor 除算を用いる。
func (d Date) AddMonths(n int) Date {
	total := d.Year*12 + (d.Month - 1) + n
	return Date{
		Year:  floorDiv(total, 12),
		Month: floorMod(total, 12) + 1,
	}
}

// Before は d が o より厳密に前かどうかを返す。
func (d Date) Before(o Date) bool {
	if d.Year != o.Year {
		return d.Year < o.Year
	}
	return d.Month < o.Month
}

// After は d が o より厳密に後かどうかを返す。
func (d Date) After(o Date) bool {
	if d.Year != o.Year {
		return d.Year > o.Year
	}
	return d.Month > o.Month
}

// Equal は d と o が同じ年月かどうかを返す。
func (d Date) Equal(o Date) bool {
	return d.Year == o.Year && d.Month == o.Month
}

// IsZero は d がゼロ値（Year==0 && Month==0）かどうかを返す。
func (d Date) IsZero() bool {
	return d.Year == 0 && d.Month == 0
}
