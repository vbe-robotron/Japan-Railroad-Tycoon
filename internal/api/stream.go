package api

import (
	"context"
	"encoding/json"

	"github.com/coder/websocket"

	"github.com/you/jrt/internal/core"
)

// StreamMessage は WebSocket で受信する 1 メッセージ。
//
// Type は "snapshot" または "events"。snapshot では Date/Turn が入り、
// events では Events が入る。
type StreamMessage struct {
	Type   string       `json:"type"`
	Date   core.Date    `json:"date,omitempty"`
	Turn   int          `json:"turn,omitempty"`
	Events []core.Event `json:"events,omitempty"`
}

// Stream は WebSocket へ接続し、メッセージの受信チャネルを返す。
//
// 受信処理は内部ゴルーチンで行うため、呼び出しはブロックしない。
// ctx がキャンセルされるか接続が切れるとチャネルは閉じられる。
func (g *GameClient) Stream(ctx context.Context) (<-chan StreamMessage, error) {
	url := g.client.wsURL(g.path("/stream"))
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return nil, ErrConnection
	}

	ch := make(chan StreamMessage, 16)
	go func() {
		defer close(ch)
		defer conn.Close(websocket.StatusNormalClosure, "")
		for {
			_, data, err := conn.Read(ctx)
			if err != nil {
				return
			}
			var msg StreamMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				continue
			}
			select {
			case ch <- msg:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}
