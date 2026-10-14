package core

import "fmt"

// このファイルは Phase 8-2 の開発 AI を実装する。
//
// 沿線都市の開発度が低く、不動産志向（Personality.RealEstate）があり、
// 資金に余裕があれば住宅地を建設する（阪急モデルの沿線開発）。

// aiDevelop は施設を建設する。
//
// 1 ターンにつき最大 1 施設を建設する。対象がなければ nil を返す。
func (g *Game) aiDevelop(c *Company, ai *AI) []Event {
	if ai.Personality.RealEstate < 0.3 {
		return nil
	}
	if c.Cash < 15_000_000 {
		return nil
	}

	// 開発対象候補
	for _, line := range c.Lines {
		for _, cityID := range []string{line.From, line.To} {
			city := g.Cities[cityID]
			if city == nil {
				continue
			}
			if city.OwnerID != "" && city.OwnerID != c.ID {
				continue // 他社所有
			}
			if city.Developed > 0.7 {
				continue // 十分開発済み
			}
			cost := g.FacilityCost(city, FacilityHousing, 2)
			if c.Cash < cost*3 {
				continue
			}

			// 住宅地を建設
			return g.executeBuildFacility(c, city, FacilityHousing, 2)
		}
	}
	return nil
}

// executeBuildFacility は AI 用の施設建設を実行する。
//
// BuildFacilityAction の Validate を経由せず内部処理を行う
// （AI はプレイヤー会社を前提としないため）。
func (g *Game) executeBuildFacility(c *Company, city *City, kind string, scale int) []Event {
	def, ok := FacilityByKind(kind)
	if !ok {
		return nil
	}
	cost := g.FacilityCost(city, kind, scale)
	if c.Cash < cost {
		return nil
	}
	c.Cash -= cost

	facilityID := g.IDGen.Next()
	f := NewFacility(facilityID, c.ID, city.ID, kind, scale)
	f.BuiltAt = g.Date
	f.Income = int64(float64(cost) * def.IncomeRate)
	c.Facilities = append(c.Facilities, f)

	oldDeveloped := city.Developed
	city.Developed += def.DevelopGain * ScaleDevelopFactor(scale)
	if city.Developed > 1.0 {
		city.Developed = 1.0
	}
	city.OwnerID = c.ID
	g.recalcCityDemand(city)

	return []Event{
		g.Success(EventFacilityBuilt,
			fmt.Sprintf("%sが%sに%sを建設しました（%s）",
				c.Name, city.Name, def.Name, FormatYen(cost))).
			With("facility_id", f.ID).
			With("city_id", city.ID).
			With("kind", kind).
			With("scale", scale).
			With("cost", cost).
			With("old_developed", oldDeveloped).
			With("new_developed", city.Developed).
			With("ai", true).
			ForCompany(c.ID).
			ForCity(city.ID),
	}
}
