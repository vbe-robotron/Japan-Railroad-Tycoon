package api

import (
	"context"
	"net/http"

	"github.com/you/jrt/internal/core"
)

// CreateGameRequest は新規ゲーム作成のリクエスト。
type CreateGameRequest struct {
	CompanyID  string `json:"company_id"`
	Seed       int64  `json:"seed,omitempty"`
	Difficulty string `json:"difficulty,omitempty"`
}

// GameSummary はゲーム一覧の 1 件分の要約。
type GameSummary struct {
	ID          string    `json:"id"`
	CompanyID   string    `json:"company_id"`
	CompanyName string    `json:"company_name"`
	Date        core.Date `json:"date"`
	Turn        int       `json:"turn"`
}

// CreateGame はゲームを作成し、GameClient と初期状態を返す。
func (c *Client) CreateGame(ctx context.Context, req CreateGameRequest) (*GameClient, *GameState, error) {
	var st GameState
	if _, err := c.decode(ctx, request{Method: http.MethodPost, Path: "/api/games", Body: req}, &st); err != nil {
		return nil, nil, err
	}
	return c.ForGame(st.ID), &st, nil
}

// ListGames は既存ゲームの要約一覧を返す。
//
// サーバーは ID の一覧を返すため、各ゲームの状態を取得して要約を組み立てる。
func (c *Client) ListGames(ctx context.Context) ([]GameSummary, error) {
	var ids []string
	if _, err := c.decode(ctx, request{Method: http.MethodGet, Path: "/api/games"}, &ids); err != nil {
		return nil, err
	}
	summaries := make([]GameSummary, 0, len(ids))
	for _, id := range ids {
		_, st, err := c.LoadGame(ctx, id)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, summaryOf(st))
	}
	return summaries, nil
}

// LoadGame はゲーム ID から GameClient と状態を返す。
func (c *Client) LoadGame(ctx context.Context, gameID string) (*GameClient, *GameState, error) {
	g := c.ForGame(gameID)
	var st GameState
	if _, err := c.decode(ctx, request{Method: http.MethodGet, Path: g.path("")}, &st); err != nil {
		return nil, nil, err
	}
	return g, &st, nil
}

// Delete はゲームを削除する。
func (g *GameClient) Delete(ctx context.Context) error {
	_, err := g.client.decode(ctx, request{Method: http.MethodDelete, Path: g.path("")}, nil)
	return err
}

// summaryOf は GameState から GameSummary を生成する。
func summaryOf(st *GameState) GameSummary {
	s := GameSummary{
		ID:        st.ID,
		CompanyID: st.PlayerID,
		Date:      st.Date,
		Turn:      st.Turn,
	}
	if st.Company != nil {
		s.CompanyName = st.Company.Name
	}
	return s
}
