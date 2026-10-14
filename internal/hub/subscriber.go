package hub

import (
	"sync"

	"github.com/you/jrt/internal/core"
)

// Subscriber は WebSocket 購読者を表す。
//
// ch はイベント列を受信するためのバッファ付きチャネルである。
// バッファが満杯の場合は Send が切断扱いとして Close を呼ぶ。
// closed は Close 済みかを通知するチャネルで、once により 1 度だけ閉じる。
type Subscriber struct {
	ch     chan []core.Event
	closed chan struct{}
	once   sync.Once
}

// NewSubscriber は新しい購読者を生成する。
func NewSubscriber() *Subscriber {
	return &Subscriber{
		ch:     make(chan []core.Event, 16),
		closed: make(chan struct{}),
	}
}

// Events はイベント受信用チャネルを返す。
func (s *Subscriber) Events() <-chan []core.Event {
	return s.ch
}

// Send はイベント列を送信する。バッファが満杯なら切断扱いで閉じる。
//
// 送信は非ブロッキングであり、受信側の速度に依存しない。
func (s *Subscriber) Send(events []core.Event) {
	select {
	case s.ch <- events:
	case <-s.closed:
	default:
		// バッファ溢れ → 切断
		s.Close()
	}
}

// Close は購読を終了する。複数回呼ばれても内部的には 1 度だけ実行する。
func (s *Subscriber) Close() {
	s.once.Do(func() {
		close(s.closed)
	})
}

// Closed は閉じられたかを通知するチャネルを返す。
func (s *Subscriber) Closed() <-chan struct{} {
	return s.closed
}
