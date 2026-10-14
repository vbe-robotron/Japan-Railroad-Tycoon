package api

import (
	"context"
	"errors"
	"testing"

	"github.com/you/jrt/internal/core"
)

func TestBuildLine(t *testing.T) {
	c, _, h, _ := newClientServer(t)
	ctx := context.Background()
	g, _, err := c.CreateGame(ctx, CreateGameRequest{CompanyID: "hankyu", Seed: 7})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	umeda, takarazuka := testCities()
	addCities(t, h, g.ID(), umeda, takarazuka)

	events, err := g.BuildLine(ctx, "umeda", "takarazuka", 1067, false)
	if err != nil {
		t.Fatalf("BuildLine: %v", err)
	}
	if !hasEventType(events, core.EventLineOpened) {
		t.Errorf("events = %+v, want line_opened", events)
	}
}

func TestBuildLine_Error(t *testing.T) {
	c, _, _, _ := newClientServer(t)
	ctx := context.Background()
	g, _, err := c.CreateGame(ctx, CreateGameRequest{CompanyID: "hankyu"})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}

	_, err = g.BuildLine(ctx, "zzz", "yyy", 1067, false)
	if err == nil {
		t.Fatal("BuildLine err = nil, want not found")
	}
	if !IsNotFound(err) {
		t.Errorf("err = %v, want not found", err)
	}
	var e *Error
	if !errors.As(err, &e) {
		t.Fatalf("err type = %T, want *api.Error", err)
	}
	if e.Code != "FROM_NOT_FOUND" {
		t.Errorf("code = %q, want FROM_NOT_FOUND", e.Code)
	}
}

func TestClient_EndTurn(t *testing.T) {
	c, _, _, _ := newClientServer(t)
	ctx := context.Background()
	g, _, err := c.CreateGame(ctx, CreateGameRequest{CompanyID: "hankyu"})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}

	events, err := g.EndTurn(ctx)
	if err != nil {
		t.Fatalf("EndTurn: %v", err)
	}
	if !hasEventType(events, core.EventTurnEnded) {
		t.Errorf("events = %+v, want turn_ended", events)
	}

	st, err := g.Status(ctx)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if st.Turn != 1 {
		t.Errorf("turn = %d, want 1", st.Turn)
	}
}

func TestAction_Generic(t *testing.T) {
	c, _, _, _ := newClientServer(t)
	ctx := context.Background()
	g, _, err := c.CreateGame(ctx, CreateGameRequest{CompanyID: "hankyu"})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}

	events, err := g.Action(ctx, &core.EndTurnAction{})
	if err != nil {
		t.Fatalf("Action: %v", err)
	}
	if !hasEventType(events, core.EventTurnEnded) {
		t.Errorf("events = %+v, want turn_ended", events)
	}
}

func TestAction_Nil(t *testing.T) {
	c, _, _, _ := newClientServer(t)
	ctx := context.Background()
	g, _, err := c.CreateGame(ctx, CreateGameRequest{CompanyID: "hankyu"})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	_, err = g.Action(ctx, nil)
	if err == nil {
		t.Fatal("Action(nil) err = nil, want error")
	}
	var e *Error
	if !errors.As(err, &e) || e.Code != "INVALID_ACTION" {
		t.Errorf("err = %v, want INVALID_ACTION", err)
	}
}

// TestActionMethods_All は全アクションラッパーが往復することを確認する。
//
// 成功・失敗のどちらでもよいが、返るエラーは *api.Error でなければならない。
func TestActionMethods_All(t *testing.T) {
	c, _, h, _ := newClientServer(t)
	ctx := context.Background()
	g, _, err := c.CreateGame(ctx, CreateGameRequest{CompanyID: "hankyu", Seed: 7})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	umeda, takarazuka := testCities()
	addCities(t, h, g.ID(), umeda, takarazuka)
	events, err := g.BuildLine(ctx, "umeda", "takarazuka", 1067, false)
	if err != nil {
		t.Fatalf("BuildLine: %v", err)
	}
	lineID := events[0].LineID

	calls := []struct {
		name string
		call func() ([]core.Event, error)
	}{
		{"BuildStation", func() ([]core.Event, error) { return g.BuildStation(ctx, lineID, umeda.ID, "local", 0.5) }},
		{"BuildFacility", func() ([]core.Event, error) { return g.BuildFacility(ctx, umeda.ID, "housing", 1) }},
		{"BuildDepot", func() ([]core.Event, error) { return g.BuildDepot(ctx, umeda.ID, "depot") }},
		{"Electrify", func() ([]core.Event, error) { return g.Electrify(ctx, lineID, "dc1500") }},
		{"SetFare", func() ([]core.Event, error) { return g.SetFare(ctx, lineID, 250) }},
		{"SetSchedule", func() ([]core.Event, error) { return g.SetSchedule(ctx, lineID, 10) }},
		{"PurchaseStock", func() ([]core.Event, error) { return g.PurchaseStock(ctx, lineID, 2, "steam") }},
		{"IssueBonds", func() ([]core.Event, error) { return g.IssueBonds(ctx, 1_000_000, 0.05, 10) }},
		{"IssueShares", func() ([]core.Event, error) { return g.IssueShares(ctx, 100) }},
		{"PayDividend", func() ([]core.Event, error) { return g.PayDividend(ctx, 10) }},
		{"RepayDebt", func() ([]core.Event, error) { return g.RepayDebt(ctx, 100_000) }},
		{"BuyShares", func() ([]core.Event, error) { return g.BuyShares(ctx, "other", 10) }},
		{"TenderOffer", func() ([]core.Event, error) { return g.TenderOffer(ctx, "other", 1.1, 0.51) }},
		{"Merge", func() ([]core.Event, error) { return g.Merge(ctx, "other", "新社", 1.0) }},
		{"AbandonLine", func() ([]core.Event, error) { return g.AbandonLine(ctx, lineID, "economic") }},
	}
	for _, tc := range calls {
		_, err := tc.call()
		if err != nil {
			var e *Error
			if !errors.As(err, &e) {
				t.Errorf("%s: err type = %T, want *api.Error", tc.name, err)
			}
		}
	}
}
