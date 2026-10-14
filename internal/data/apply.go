package data

import "github.com/you/jrt/internal/core"

// defaultStartDate は史実初期状態を再現する基準日。
//
// Game.Date が未設定（ゼロ値）の場合にこの年月以前の史実路線を自動敷設する。
var defaultStartDate = core.Date{Year: 1907, Month: 10}

// ApplyToGame は Dataset を Game に投入する。
//
//   - Cities: Game.AddCity で追加
//   - Companies: Game.AddCompany で追加し、ID が playerID と一致する会社を
//     プレイヤー、ID が "jnr" の会社を国鉄として設定する
//   - HistoricalLines: core.HistoricalLine に変換して Game へ注入する
//   - 開始年月以前の史実路線を AI 会社へ自動敷設する（プレイヤーは除く）
//
// d または g が nil の場合は何もしない（呼び出し側の結線を簡略化するため）。
func (d *Dataset) ApplyToGame(g *core.Game, playerID string) {
	if d == nil || g == nil {
		return
	}

	for _, cd := range d.Cities {
		g.AddCity(&core.City{
			ID:                 cd.ID,
			Name:               cd.Name,
			Prefecture:         cd.Prefecture,
			RepresentativeNode: cd.RepresentativeNode,
			Lat:                cd.Lat,
			Lon:                cd.Lon,
			Population:         cd.Population,
			LandValue:          cd.LandValue,
			Tags:               cd.Tags,
		})
	}

	for _, cd := range d.Companies {
		c := core.NewCompany(cd.ID, cd.GroupID, cd.Name)
		c.LegalName = cd.LegalName
		c.FoundedYear = cd.FoundedYear
		c.Cash = cd.InitialCash
		c.TotalShares = cd.TotalShares
		c.StockPrice = cd.StockPrice
		c.Region = cd.Region
		c.Tags = cd.Tags
		c.Reputation = cd.Reputation
		c.IsPlayer = cd.ID == playerID
		c.IsNational = cd.ID == "jnr"
		g.AddCompany(c)
	}

	// 史実路線（core.HistoricalLine へ変換して注入）
	lines := make([]core.HistoricalLine, 0, len(d.HistoricalLines))
	for _, h := range d.HistoricalLines {
		lines = append(lines, core.HistoricalLine{
			CompanyID:   h.CompanyID,
			From:        h.From,
			To:          h.To,
			Gauge:       h.Gauge,
			Double:      h.Double,
			OpenedYear:  h.OpenedYear,
			OpenedMonth: h.OpenedMonth,
			DisplayName: h.DisplayName,
		})
	}
	g.SetHistoricalLines(lines)

	// 開始年以前の史実路線を自動敷設する。
	//
	// プレイヤー会社は自分で敷設するため対象外。開業年が開始年月より後の
	// 路線は AI（aiFollowHistory）に任せる。
	start := g.Date
	if start.Year == 0 {
		start = defaultStartDate
	}
	for _, h := range lines {
		if h.OpenedYear > start.Year ||
			(h.OpenedYear == start.Year && h.OpenedMonth > start.Month) {
			continue
		}
		c := g.Companies[h.CompanyID]
		if c == nil || c.IsPlayer {
			continue
		}
		g.AutoBuildHistoricalLine(c, h)
	}
}
