package hub

import (
	"testing"
	"time"

	"github.com/you/jrt/internal/core"
)

func TestSubscriber_Send(t *testing.T) {
	sub := NewSubscriber()
	events := []core.Event{{ID: "e1", Type: "built"}}
	sub.Send(events)

	got := recvEvents(t, sub)
	if len(got) != 1 || got[0].ID != "e1" {
		t.Errorf("受信内容が不正: %+v", got)
	}

	select {
	case <-sub.Closed():
		t.Error("送信直後に閉じられてはならない")
	default:
	}
}

func TestSubscriber_Close(t *testing.T) {
	sub := NewSubscriber()
	select {
	case <-sub.Closed():
		t.Fatal("初期状態で閉じている")
	default:
	}

	sub.Close()
	sub.Close() // 二重 Close でも panic しない。

	select {
	case <-sub.Closed():
	case <-time.After(time.Second):
		t.Fatal("Close 後に閉じられていない")
	}
}

func TestSubscriber_Overflow(t *testing.T) {
	sub := NewSubscriber()
	// バッファ（16）を埋める。空きがあるため送信は成功する。
	for i := 0; i < cap(sub.ch); i++ {
		sub.Send([]core.Event{{Type: "x"}})
	}
	select {
	case <-sub.Closed():
		t.Fatal("溢れる前に閉じられた")
	default:
	}

	// 溢れる送信で切断される。
	sub.Send([]core.Event{{Type: "overflow"}})
	select {
	case <-sub.Closed():
	case <-time.After(time.Second):
		t.Fatal("バッファ溢れで閉じられていない")
	}

	// 閉じた後は channel バッファが埋まっていても安全。
	sub.Send([]core.Event{{Type: "after"}})
}

func TestSubscriber_EventsChannel(t *testing.T) {
	sub := NewSubscriber()
	if sub.Events() == nil {
		t.Fatal("Events チャネルが nil")
	}
	if sub.Closed() == nil {
		t.Fatal("Closed チャネルが nil")
	}
}

func TestNewSubscriber(t *testing.T) {
	sub := NewSubscriber()
	if sub == nil || cap(sub.ch) != 16 {
		t.Fatalf("初期化が不正: %+v", sub)
	}
}
