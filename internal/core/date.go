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

// NewDate は Date を生成する。
//
// 月が 1-12 の範囲外の場合は panic せず、年を繰り上げ・繰り下げして
// 正規化する。例: NewDate(1910, 13) は 1911年1月、NewDate(1910, 0) は
// 1909年12月になる。
func NewDate(year, month int) Date {
	total := year*12 + (month - 1)
	return Date{Year: total / 12, Month: total%12 + 1}
}

// String は "1910年3月" 形式の文字列を返す。
func (d Date) String() string {
	return fmt.Sprintf("%d年%d月", d.Year, d.Month)
}

// AddMonths は n ヶ月後の Date を返す。n は負でもよい。
func (d Date) AddMonths(n int) Date {
	total := d.Year*12 + (d.Month - 1) + n
	return Date{Year: total / 12, Month: total%12 + 1}
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
