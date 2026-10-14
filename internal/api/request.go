package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/you/jrt/internal/core"
)

// request は内部共通の HTTP リクエスト表現。
type request struct {
	Method string
	Path   string
	Query  map[string]string
	Body   any
}

// actionBody はアクション実行のリクエストボディ。
type actionBody struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// serverResponse はサーバーの統一レスポンス形式。
type serverResponse struct {
	OK     bool            `json:"ok"`
	Data   json.RawMessage `json:"data"`
	Events []core.Event    `json:"events"`
	Errors []Error         `json:"errors"`
}

// do はリクエストを送信し、統一レスポンスを返す。
//
// 接続失敗は ErrConnection、タイムアウトは ErrTimeout、JSON 不正は
// ErrInvalid を返す。HTTP 4xx/5xx または ok=false の場合は、サーバーが
// 返した最初のエラーを *Error として返す。
func (c *Client) do(ctx context.Context, req request) (*serverResponse, error) {
	u := c.baseURL + req.Path
	if len(req.Query) > 0 {
		q := url.Values{}
		for k, v := range req.Query {
			q.Set(k, v)
		}
		u += "?" + q.Encode()
	}

	var body io.Reader
	if req.Body != nil {
		data, err := json.Marshal(req.Body)
		if err != nil {
			return nil, ErrInvalid
		}
		body = bytes.NewReader(data)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, u, body)
	if err != nil {
		return nil, &Error{Code: "REQUEST_FAILED", Message: err.Error()}
	}
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		if isTimeout(err) {
			return nil, ErrTimeout
		}
		return nil, ErrConnection
	}
	defer resp.Body.Close()

	var sr serverResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, ErrInvalid
	}
	if resp.StatusCode >= 400 || !sr.OK {
		return nil, errorFromResponse(resp.StatusCode, sr.Errors)
	}
	return &sr, nil
}

// decode はリクエストを送信し、Data を out へデコードして Events を返す。
//
// out が nil の場合は Data をデコードしない（アクション系の用途）。
func (c *Client) decode(ctx context.Context, req request, out any) ([]core.Event, error) {
	sr, err := c.do(ctx, req)
	if err != nil {
		return nil, err
	}
	if out != nil && len(sr.Data) > 0 {
		if err := json.Unmarshal(sr.Data, out); err != nil {
			return nil, ErrInvalid
		}
	}
	return sr.Events, nil
}

// action は型付きアクションを実行し、発生イベントを返す。
func (g *GameClient) action(ctx context.Context, actionType string, payload any) ([]core.Event, error) {
	body := actionBody{Type: actionType}
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, ErrInvalid
		}
		body.Payload = raw
	}
	return g.client.decode(ctx, request{
		Method: http.MethodPost,
		Path:   g.path("/actions"),
		Body:   body,
	}, nil)
}

// isTimeout は err がタイムアウトかどうかを判定する。
func isTimeout(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}
