package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/you/jrt/internal/hub"
	"github.com/you/jrt/internal/store"
)

// newClientServer は httptest サーバーと接続済み Client を用意する。
func newClientServer(t *testing.T) (*Client, *httptest.Server, *hub.Hub, store.Store) {
	t.Helper()
	st := newSQLiteStore(t)
	srv, h := newTestServer(t, st)
	return NewClient(Config{BaseURL: srv.URL}), srv, h, st
}

func TestClient_ForGameID(t *testing.T) {
	c := NewClient(Config{BaseURL: "http://example.test/"})
	g := c.ForGame("abc")
	if g.ID() != "abc" {
		t.Errorf("ID() = %q, want abc", g.ID())
	}
	if got := g.path("/lines"); got != "/api/games/abc/lines" {
		t.Errorf("path = %q", got)
	}
}

func TestClient_WSURL(t *testing.T) {
	cases := map[string]string{
		"http://x.test":  "ws://x.test",
		"https://x.test": "wss://x.test",
		"ws://x.test":    "ws://x.test",
	}
	for in, want := range cases {
		if got := NewClient(Config{BaseURL: in}).wsURL("/p"); got != want+"/p" {
			t.Errorf("wsURL(%q) = %q, want %q", in, got, want+"/p")
		}
	}
}

func TestClient_CreateGame(t *testing.T) {
	c, _, _, _ := newClientServer(t)
	ctx := context.Background()

	g, state, err := c.CreateGame(ctx, CreateGameRequest{CompanyID: "hankyu", Seed: 42})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if g.ID() != state.ID {
		t.Errorf("client id = %q, state id = %q", g.ID(), state.ID)
	}
	if state.PlayerID != "hankyu" {
		t.Errorf("player_id = %q, want hankyu", state.PlayerID)
	}
	if state.Company == nil || state.Company.Cash != 9_000_000 {
		t.Fatalf("company = %+v, want player company with 9000000 cash", state.Company)
	}
	if state.Rank != 1 {
		t.Errorf("rank = %d, want 1", state.Rank)
	}
}

func TestClient_ListGames(t *testing.T) {
	c, _, _, _ := newClientServer(t)
	ctx := context.Background()

	g1, _, err := c.CreateGame(ctx, CreateGameRequest{CompanyID: "a"})
	if err != nil {
		t.Fatalf("CreateGame a: %v", err)
	}
	if _, _, err := c.CreateGame(ctx, CreateGameRequest{CompanyID: "b"}); err != nil {
		t.Fatalf("CreateGame b: %v", err)
	}

	games, err := c.ListGames(ctx)
	if err != nil {
		t.Fatalf("ListGames: %v", err)
	}
	if len(games) != 2 {
		t.Fatalf("len(games) = %d, want 2", len(games))
	}
	found := false
	for _, s := range games {
		if s.ID == g1.ID() {
			found = true
			if s.CompanyID != "a" {
				t.Errorf("company_id = %q, want a", s.CompanyID)
			}
			if s.Date.IsZero() {
				t.Error("date should not be zero")
			}
		}
	}
	if !found {
		t.Errorf("created game %s not in list", g1.ID())
	}
}

func TestLoadGame(t *testing.T) {
	c, _, _, _ := newClientServer(t)
	ctx := context.Background()

	created, _, err := c.CreateGame(ctx, CreateGameRequest{CompanyID: "hankyu"})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}

	g, state, err := c.LoadGame(ctx, created.ID())
	if err != nil {
		t.Fatalf("LoadGame: %v", err)
	}
	if g.ID() != created.ID() {
		t.Errorf("id = %q, want %q", g.ID(), created.ID())
	}
	if state.ID != created.ID() {
		t.Errorf("state id = %q, want %q", state.ID, created.ID())
	}
}

func TestLoadGame_NotFound(t *testing.T) {
	c, _, _, _ := newClientServer(t)

	_, _, err := c.LoadGame(context.Background(), "does-not-exist")
	if err == nil {
		t.Fatal("LoadGame err = nil, want not found")
	}
	if !IsNotFound(err) {
		t.Errorf("err = %v, want not found", err)
	}
}

func TestDelete(t *testing.T) {
	c, _, _, _ := newClientServer(t)
	ctx := context.Background()

	g, _, err := c.CreateGame(ctx, CreateGameRequest{CompanyID: "hankyu"})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if err := g.Delete(ctx); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, _, err := c.LoadGame(ctx, g.ID()); err == nil {
		t.Error("LoadGame after delete err = nil, want not found")
	}
}

func TestDelete_NotFound(t *testing.T) {
	c, _, _, _ := newClientServer(t)
	if err := c.ForGame("nope").Delete(context.Background()); err == nil {
		t.Error("Delete err = nil, want not found")
	}
}

func TestCreateGame_ConnectionError(t *testing.T) {
	c := NewClient(Config{BaseURL: "http://127.0.0.1:0"})
	_, _, err := c.CreateGame(context.Background(), CreateGameRequest{})
	if err == nil {
		t.Fatal("err = nil, want connection error")
	}
	if !errors.Is(err, ErrConnection) {
		t.Errorf("err = %v, want ErrConnection", err)
	}
}

func TestCreateGame_InvalidResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	c := NewClient(Config{BaseURL: srv.URL})
	_, _, err := c.CreateGame(context.Background(), CreateGameRequest{})
	if err == nil {
		t.Fatal("err = nil, want invalid response")
	}
	if !errors.Is(err, ErrInvalid) {
		t.Errorf("err = %v, want ErrInvalid", err)
	}
}
