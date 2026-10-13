package core

import "fmt"

// このファイルは BuildFacilityAction の本実装を提供する（Phase 1-8-2）。
//
// プレイヤーは都市に開発施設（百貨店・住宅地・遊園地・ホテル・オフィス）を
// 建設できる。施設建設は次の効果を持つ:
//
//   - 建設コストを支払う（種類・規模・地価に依存）
//   - 都市の開発度（Developed）を上昇させる（阪急モデルの中核）
//   - 都市の所有者（OwnerID）を自社にする
//   - 施設収入（Facility.Income）を毎月得る
//
// 開発度は Game.Demand の需要倍率（1.0 + Developed×1.0）を通じて旅客需要を
// 押し上げるため、沿線開発が路線の収益を改善する（Phase 1-8-2.5 で 4.0 → 1.0）。

// Validate は BuildFacilityAction が実行可能かどうかを検証する。
//
// 検証内容:
//   - ゲームが存在する
//   - プレイヤー会社が存在する
//   - 都市が存在する
//   - 種類が既知
//   - 規模が 1〜5
//   - 他社が所有していない
//   - 資金 ≥ 建設コスト
//
// Validate は状態を変更しない。
func (a *BuildFacilityAction) Validate(g *Game) error {
	if g == nil {
		return NewValidationError("GAME_NIL", "ゲームが初期化されていません", nil)
	}

	c := g.PlayerCompany()
	if c == nil {
		return NewValidationError("NO_PLAYER", "プレイヤー会社が存在しません", ErrNoPlayer)
	}

	city, ok := g.Cities[a.CityID]
	if !ok {
		return NewValidationError("CITY_NOT_FOUND",
			fmt.Sprintf("都市が見つかりません: %s", a.CityID), ErrCityNotFound)
	}

	if _, ok := FacilityByKind(a.Kind); !ok {
		return NewValidationError("UNKNOWN_FACILITY_KIND",
			fmt.Sprintf("未知の施設種類です: %s", a.Kind), ErrInvalidArgument)
	}

	if a.Scale < 1 || a.Scale > 5 {
		return NewValidationError("INVALID_SCALE",
			fmt.Sprintf("規模は 1〜5 の範囲で指定してください: %d", a.Scale), ErrInvalidArgument)
	}

	if city.OwnerID != "" && city.OwnerID != c.ID {
		return NewValidationError("OWNED_BY_OTHER",
			fmt.Sprintf("%sは他社が開発済みです", city.Name), ErrUnauthorized)
	}

	cost := g.FacilityCost(city, a.Kind, a.Scale)
	if c.Cash < cost {
		return NewValidationError("INSUFFICIENT_FUNDS",
			fmt.Sprintf("資金不足（必要: %s / 所持: %s）",
				FormatYen(cost), FormatYen(c.Cash)), ErrInsufficientFunds)
	}

	return nil
}

// Apply は BuildFacilityAction を実行する。
//
// 処理:
//  1. 建設コストを支払う
//  2. Facility を生成し、Company.Facilities に追加
//  3. 都市の開発度を上昇（種類別の上昇量 × 規模補正、上限 1.0）
//  4. 都市の需要を再計算
//  5. facility_built / facility_effect イベントを生成
func (a *BuildFacilityAction) Apply(g *Game) ([]Event, error) {
	if err := a.Validate(g); err != nil {
		return nil, err
	}

	c := g.PlayerCompany()
	city := g.Cities[a.CityID]
	def := FacilityTable[a.Kind]

	cost := g.FacilityCost(city, a.Kind, a.Scale)
	c.Cash -= cost

	facilityID := g.IDGen.Next()
	facility := NewFacility(facilityID, c.ID, a.CityID, a.Kind, a.Scale)
	facility.BuiltAt = g.Date
	facility.Income = int64(float64(cost) * def.IncomeRate)

	c.Facilities = append(c.Facilities, facility)

	// 開発度上昇（種類別の上昇量 × 規模補正）
	gain := def.DevelopGain * ScaleDevelopFactor(a.Scale)
	oldDeveloped := city.Developed
	city.Developed += gain
	if city.Developed > 1.0 {
		city.Developed = 1.0
	}
	city.OwnerID = c.ID

	// 需要再計算（開発度上昇を需要へ反映）
	g.RecalcCityDemand(city)

	events := []Event{
		g.Success(EventFacilityBuilt,
			fmt.Sprintf("%sに%sを建設しました（%s）",
				city.Name, def.Name, FormatYen(cost))).
			With("facility_id", facility.ID).
			With("city_id", city.ID).
			With("city_name", city.Name).
			With("kind", a.Kind).
			With("scale", a.Scale).
			With("cost", cost).
			With("income", facility.Income).
			ForCompany(c.ID).
			ForCity(city.ID),
		g.Success(EventFacilityEffect,
			fmt.Sprintf("%sの開発度が %.2f → %.2f に上昇しました",
				city.Name, oldDeveloped, city.Developed)).
			With("facility_id", facility.ID).
			With("city_id", city.ID).
			With("old_developed", oldDeveloped).
			With("new_developed", city.Developed).
			ForCompany(c.ID).
			ForCity(city.ID),
	}

	return events, nil
}
