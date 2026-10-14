package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/coder/websocket"

	"github.com/you/jrt/internal/hub"
)

// wsWriteTimeout は 1 フレームの書き込みに許容する時間。
const wsWriteTimeout = 5 * time.Second

// streamWS は WebSocket でターン通知とイベントを配信する。
//
// 接続直後にスナップショットを 1 通送信し、以降はセッションへ
// Broadcast されたイベントを非ブロッキングに転送する。
func streamWS(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, ok := session(h, w, r)
		if !ok {
			return
		}

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: []string{"*"},
		})
		if err != nil {
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")

		sub := sess.Subscribe()
		defer sess.Unsubscribe(sub)

		// クライアント切断を検知するため、読み取り監視付きの context を使う。
		ctx := conn.CloseRead(r.Context())

		// 初回スナップショット
		sess.Lock()
		snapshot := map[string]any{
			"type": "snapshot",
			"date": sess.Game.Date,
			"turn": sess.Game.Turn,
			"game": buildGameState(sess.Game),
		}
		sess.Unlock()

		if err := writeJSONWS(ctx, conn, snapshot); err != nil {
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case events := <-sub.Events():
				msg := map[string]any{"type": "events", "events": events}
				wctx, cancel := context.WithTimeout(ctx, wsWriteTimeout)
				err := writeJSONWS(wctx, conn, msg)
				cancel()
				if err != nil {
					return
				}
			case <-sub.Closed():
				return
			}
		}
	}
}

// writeJSONWS は値を JSON 化して 1 テキストフレームとして送信する。
func writeJSONWS(ctx context.Context, conn *websocket.Conn, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return conn.Write(ctx, websocket.MessageText, data)
}
