package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/you/jrt/internal/core"
)

// Status はゲームの現在状態を返す。
func (g *GameClient) Status(ctx context.Context) (*GameState, error) {
	var st GameState
	if _, err := g.client.decode(ctx, request{Method: http.MethodGet, Path: g.path("")}, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

// ListLines はプレイヤー会社の路線一覧を返す。
func (g *GameClient) ListLines(ctx context.Context) ([]core.Line, error) {
	var lines []core.Line
	if _, err := g.client.decode(ctx, request{Method: http.MethodGet, Path: g.path("/lines")}, &lines); err != nil {
		return nil, err
	}
	return lines, nil
}

// GetLine は ID 指定で 1 本の路線を返す。
func (g *GameClient) GetLine(ctx context.Context, lineID string) (*core.Line, error) {
	var line core.Line
	if _, err := g.client.decode(ctx, request{Method: http.MethodGet, Path: g.path("/lines/" + lineID)}, &line); err != nil {
		return nil, err
	}
	return &line, nil
}

// ListCompanies は会社一覧を返す。
func (g *GameClient) ListCompanies(ctx context.Context) ([]core.Company, error) {
	var companies []core.Company
	if _, err := g.client.decode(ctx, request{Method: http.MethodGet, Path: g.path("/companies")}, &companies); err != nil {
		return nil, err
	}
	return companies, nil
}

// ListCities は都市一覧を返す。
func (g *GameClient) ListCities(ctx context.Context) ([]core.City, error) {
	var cities []core.City
	if _, err := g.client.decode(ctx, request{Method: http.MethodGet, Path: g.path("/cities")}, &cities); err != nil {
		return nil, err
	}
	return cities, nil
}

// CanBuild は路線敷設の可否（4 ゲート）を返す。
func (g *GameClient) CanBuild(ctx context.Context, from, to string, gauge int) (*CanBuildResult, error) {
	query := map[string]string{
		"from":  from,
		"to":    to,
		"gauge": strconv.Itoa(gauge),
	}
	var result CanBuildResult
	if _, err := g.client.decode(ctx, request{
		Method: http.MethodGet,
		Path:   g.path("/can-build"),
		Query:  query,
	}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Events はゲーム内イベントを返す。last > 0 なら直近 last 件に絞る。
func (g *GameClient) Events(ctx context.Context, last int) ([]core.Event, error) {
	var events []core.Event
	if _, err := g.client.decode(ctx, request{Method: http.MethodGet, Path: g.path("/events")}, &events); err != nil {
		return nil, err
	}
	if last > 0 && len(events) > last {
		events = events[len(events)-last:]
	}
	return events, nil
}
