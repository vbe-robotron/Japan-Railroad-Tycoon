package api

import (
	"context"
	"testing"
	"time"

	"github.com/you/jrt/internal/core"
)

func TestStream_Snapshot(t *testing.T) {
	c, _, _, _ := newClientServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	g, _, err := c.CreateGame(ctx, CreateGameRequest{CompanyID: "hankyu", Seed: 1})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	ch, err := g.Stream(ctx)
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}

	msg := recvMessage(t, ch)
	if msg.Type != "snapshot" {
		t.Errorf("type = %q, want snapshot", msg.Type)
	}
	if msg.Date.IsZero() {
		t.Error("snapshot date should not be zero")
	}
}

func TestStream_Events(t *testing.T) {
	c, _, _, _ := newClientServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	g, _, err := c.CreateGame(ctx, CreateGameRequest{CompanyID: "hankyu", Seed: 2})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	ch, err := g.Stream(ctx)
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}

	// スナップショットを読み切る。
	if msg := recvMessage(t, ch); msg.Type != "snapshot" {
		t.Fatalf("first message type = %q, want snapshot", msg.Type)
	}

	if _, err := g.EndTurn(ctx); err != nil {
		t.Fatalf("EndTurn: %v", err)
	}

	for {
		msg := recvMessage(t, ch)
		if msg.Type != "events" {
			continue
		}
		if !hasEventType(msg.Events, core.EventTurnEnded) {
			t.Errorf("events = %+v, want turn_ended", msg.Events)
		}
		return
	}
}

func TestStream_ConnectionError(t *testing.T) {
	c := NewClient(Config{BaseURL: "http://127.0.0.1:0"})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := c.ForGame("x").Stream(ctx); err == nil {
		t.Fatal("Stream err = nil, want connection error")
	}
}

// recvMessage はチャネルから 1 通受信する。タイムアウトで失敗させる。
func recvMessage(t *testing.T, ch <-chan StreamMessage) StreamMessage {
	t.Helper()
	select {
	case msg, ok := <-ch:
		if !ok {
			t.Fatal("stream channel closed")
		}
		return msg
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for stream message")
		return StreamMessage{}
	}
}
