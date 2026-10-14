package data

import "github.com/you/jrt/internal/core"

// ApplyToGame は Dataset を Game に投入する。
//
//   - Cities: Game.AddCity で追加
//   - Companies: Game.AddCompany で追加し、ID が playerID と一致する会社を
//     プレイヤー、ID が "jnr" の会社を国鉄として設定する
//   - HistoricalLines: core.HistoricalLine に変換して Game へ注入する
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
}
