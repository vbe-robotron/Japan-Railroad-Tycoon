package api

import (
	"net/http"
	"strings"
	"time"
)

// Config は Client の生成設定。
type Config struct {
	// BaseURL はサーバーの基底 URL（例: http://localhost:8080）。
	BaseURL string
	// Timeout は 1 リクエストのタイムアウト。HTTP が nil の場合のみ使う。
	Timeout time.Duration
	// HTTP は任意の HTTP クライアント。nil なら Timeout 付きの既定値を使う。
	HTTP *http.Client
}

// Client は HTTP クライアントの基盤。
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient は Config から Client を生成する。
func NewClient(cfg Config) *Client {
	hc := cfg.HTTP
	if hc == nil {
		hc = &http.Client{Timeout: cfg.Timeout}
	}
	return &Client{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		http:    hc,
	}
}

// ForGame はゲーム単位の GameClient を返す。
func (c *Client) ForGame(gameID string) *GameClient {
	return &GameClient{client: c, gameID: gameID}
}

// GameClient は 1 ゲームに対する操作を提供する。
type GameClient struct {
	client *Client
	gameID string
}

// ID はゲーム ID を返す。
func (g *GameClient) ID() string { return g.gameID }

// path はゲーム配下のパスを返す。
func (g *GameClient) path(suffix string) string {
	return "/api/games/" + g.gameID + suffix
}

// wsURL は HTTP の基底 URL を WebSocket 用スキームへ変換する。
func (c *Client) wsURL(path string) string {
	u := c.baseURL + path
	switch {
	case strings.HasPrefix(u, "https://"):
		return "wss://" + strings.TrimPrefix(u, "https://")
	case strings.HasPrefix(u, "http://"):
		return "ws://" + strings.TrimPrefix(u, "http://")
	default:
		return u
	}
}
