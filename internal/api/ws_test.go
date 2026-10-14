package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/you/jrt/internal/core"
)

// snapshotMsg は WebSocket のスナップショットメッセージ。
type snapshotMsg struct {
	Type string            `json:"type"`
	Date core.Date         `json:"date"`
	Turn int               `json:"turn"`
	Game gameStateResponse `json:"game"`
}

// eventsMsg は WebSocket のイベントメッセージ。
type eventsMsg struct {
	Type   string       `json:"type"`
	Events []core.Event `json:"events"`
}

func wsURL(srvURL, id string) string {
	return "ws" + strings.TrimPrefix(srvURL, "http") + "/api/games/" + id + "/stream"
}

func dialStream(t *testing.T, srv *httptest.Server, id string) (*websocket.Conn, context.Context, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	conn, _, err := websocket.Dial(ctx, wsURL(srv.URL, id), nil)
	if err != nil {
		cancel()
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() {
		conn.Close(websocket.StatusNormalClosure, "")
		cancel()
	})
	return conn, ctx, cancel
}

func TestStreamWS_Snapshot(t *testing.T) {
	st := newSQLiteStore(t)
	srv, _ := newTestServer(t, st)
	id, _ := createTestGame(t, srv, `{"company_id":"hankyu","seed":1}`)

	conn, ctx, _ := dialStream(t, srv, id)

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var msg snapshotMsg
	if err := json.Unmarshal(data, &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if msg.Type != "snapshot" {
		t.Errorf("type = %q, want snapshot", msg.Type)
	}
	if msg.Game.ID != id {
		t.Errorf("game.id = %q, want %q", msg.Game.ID, id)
	}
}

func TestStreamWS_Events(t *testing.T) {
	st := newSQLiteStore(t)
	srv, _ := newTestServer(t, st)
	id, _ := createTestGame(t, srv, `{"company_id":"hankyu","seed":2}`)

	conn, ctx, _ := dialStream(t, srv, id)

	// スナップショットを読み切ってから購読済みを確認する。
	if _, _, err := conn.Read(ctx); err != nil {
		t.Fatalf("read snapshot: %v", err)
	}

	// ターン終了でイベントを発生させる。
	resp, _ := doRequest(t, http.MethodPost, srv.URL+"/api/games/"+id+"/turn/end", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("end turn status = %d", resp.StatusCode)
	}

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			t.Fatalf("read events: %v", err)
		}
		var msg eventsMsg
		if err := json.Unmarshal(data, &msg); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if msg.Type != "events" {
			continue
		}
		if !hasEventType(msg.Events, core.EventTurnEnded) {
			t.Errorf("events = %+v, want turn_ended", msg.Events)
		}
		return
	}
}
