package api

import (
	"context"

	"github.com/you/jrt/internal/core"
)

// このファイルは建設系アクションの型付きメソッドを提供する。

// BuildLine は路線を敷設する。
func (g *GameClient) BuildLine(ctx context.Context, from, to string, gauge int, double bool) ([]core.Event, error) {
	return g.action(ctx, "build_line", core.BuildLineAction{
		From: from, To: to, Gauge: gauge, Double: double,
	})
}

// BuildStation は既存路線に中間駅を追加する。
func (g *GameClient) BuildStation(ctx context.Context, lineID, cityID, kind string, pos float64) ([]core.Event, error) {
	return g.action(ctx, "build_station", core.BuildStationAction{
		LineID: lineID, CityID: cityID, Kind: kind, Position: pos,
	})
}

// BuildFacility は開発施設を建設する。
func (g *GameClient) BuildFacility(ctx context.Context, cityID, kind string, scale int) ([]core.Event, error) {
	return g.action(ctx, "build_facility", core.BuildFacilityAction{
		CityID: cityID, Kind: kind, Scale: scale,
	})
}

// BuildDepot は車庫・工場・発電所を建設する。
func (g *GameClient) BuildDepot(ctx context.Context, cityID, kind string) ([]core.Event, error) {
	return g.action(ctx, "build_depot", core.BuildDepotAction{CityID: cityID, Kind: kind})
}

// Electrify は路線を電化する。
func (g *GameClient) Electrify(ctx context.Context, lineID, system string) ([]core.Event, error) {
	return g.action(ctx, "electrify", core.ElectrifyAction{LineID: lineID, System: system})
}
