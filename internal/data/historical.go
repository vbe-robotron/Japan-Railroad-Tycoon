package data

// HistoricalLinesByCompany は史実路線を会社IDごとに集計する。
//
// 返り値のスライスは各会社の定義順（HistoricalLines の出現順）を保つ。
func (d *Dataset) HistoricalLinesByCompany() map[string][]HistoricalLineDef {
	out := make(map[string][]HistoricalLineDef)
	if d == nil {
		return out
	}
	for _, l := range d.HistoricalLines {
		out[l.CompanyID] = append(out[l.CompanyID], l)
	}
	return out
}

// HistoricalLinesByYear は史実路線を開業年ごとに集計する。
func (d *Dataset) HistoricalLinesByYear() map[int][]HistoricalLineDef {
	out := make(map[int][]HistoricalLineDef)
	if d == nil {
		return out
	}
	for _, l := range d.HistoricalLines {
		out[l.OpenedYear] = append(out[l.OpenedYear], l)
	}
	return out
}
