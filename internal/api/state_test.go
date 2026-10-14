package api

import (
	"context"
	"testing"

	"github.com/you/jrt/internal/core"
)

func TestStatus(t *testing.T) {
	c, _, _, _ := newClientServer(t)
	ctx := context.Background()

	g, _, err := c.CreateGame(ctx, CreateGameRequest{CompanyID: "hankyu"})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	st, err := g.Status(ctx)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if st.ID != g.ID() || st.PlayerID != "hankyu" {
		t.Errorf("state = %+v, want id=%s player=hankyu", st, g.ID())
	}
}

func TestListLines(t *testing.T) {
	c, _, h, _ := newClientServer(t)
	ctx := context.Background()

	g, _, err := c.CreateGame(ctx, CreateGameRequest{CompanyID: "hankyu", Seed: 7})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	umeda, takarazuka := testCities()
	addCities(t, h, g.ID(), umeda, takarazuka)
	if _, err := g.BuildLine(ctx, "umeda", "takarazuka", 1067, false); err != nil {
		t.Fatalf("BuildLine: %v", err)
	}

	lines, err := g.ListLines(ctx)
	if err != nil {
		t.Fatalf("ListLines: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("len(lines) = %d, want 1", len(lines))
	}

	line, err := g.GetLine(ctx, lines[0].ID)
	if err != nil {
		t.Fatalf("GetLine: %v", err)
	}
	if line.ID != lines[0].ID {
		t.Errorf("line id = %q, want %q", line.ID, lines[0].ID)
	}
}

func TestGetLine_NotFound(t *testing.T) {
	c, _, _, _ := newClientServer(t)
	ctx := context.Background()
	g, _, err := c.CreateGame(ctx, CreateGameRequest{})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if _, err := g.GetLine(ctx, "missing"); !IsNotFound(err) {
		t.Errorf("err = %v, want not found", err)
	}
}

func TestListCompanies(t *testing.T) {
	c, _, _, _ := newClientServer(t)
	ctx := context.Background()
	g, _, err := c.CreateGame(ctx, CreateGameRequest{CompanyID: "hankyu"})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	companies, err := g.ListCompanies(ctx)
	if err != nil {
		t.Fatalf("ListCompanies: %v", err)
	}
	if len(companies) != 1 {
		t.Fatalf("len(companies) = %d, want 1", len(companies))
	}
	if companies[0].ID != "hankyu" {
		t.Errorf("company id = %q, want hankyu", companies[0].ID)
	}
}

func TestListCities(t *testing.T) {
	c, _, h, _ := newClientServer(t)
	ctx := context.Background()
	g, _, err := c.CreateGame(ctx, CreateGameRequest{CompanyID: "hankyu"})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	umeda, takarazuka := testCities()
	addCities(t, h, g.ID(), umeda, takarazuka)

	cities, err := g.ListCities(ctx)
	if err != nil {
		t.Fatalf("ListCities: %v", err)
	}
	if len(cities) != 2 {
		t.Fatalf("len(cities) = %d, want 2", len(cities))
	}
}

func TestClient_CanBuild(t *testing.T) {
	c, _, h, _ := newClientServer(t)
	ctx := context.Background()
	g, _, err := c.CreateGame(ctx, CreateGameRequest{CompanyID: "hankyu"})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	umeda, takarazuka := testCities()
	addCities(t, h, g.ID(), umeda, takarazuka)

	result, err := g.CanBuild(ctx, "umeda", "takarazuka", 1067)
	if err != nil {
		t.Fatalf("CanBuild: %v", err)
	}
	if !result.Can {
		t.Errorf("can = false, gates = %+v", result.Gates)
	}
	if result.Cost <= 0 {
		t.Errorf("cost = %d, want > 0", result.Cost)
	}
	if len(result.Gates) != 4 {
		t.Errorf("len(gates) = %d, want 4", len(result.Gates))
	}
}

func TestEvents(t *testing.T) {
	c, _, _, _ := newClientServer(t)
	ctx := context.Background()
	g, _, err := c.CreateGame(ctx, CreateGameRequest{CompanyID: "hankyu"})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if _, err := g.EndTurn(ctx); err != nil {
		t.Fatalf("EndTurn: %v", err)
	}

	events, err := g.Events(ctx, 0)
	if err != nil {
		t.Fatalf("Events: %v", err)
	}
	if !hasEventType(events, core.EventTurnEnded) {
		t.Errorf("events = %+v, want turn_ended", events)
	}

	last, err := g.Events(ctx, 1)
	if err != nil {
		t.Fatalf("Events(last=1): %v", err)
	}
	if len(last) != 1 {
		t.Errorf("len(last) = %d, want 1", len(last))
	}
}
