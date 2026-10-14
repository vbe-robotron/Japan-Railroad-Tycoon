package cli

import "context"

// StartEventListener は WebSocket を購読し、イベントを非同期に表示する。
//
// 受信は内部ゴルーチンで行うため呼び出しはブロックしない。イベントを
// 表示した後、設定されていればプロンプトを再描画する。ctx のキャンセルで
// 停止する。購読に失敗した場合は何もせず戻る。
func (s *Session) StartEventListener(ctx context.Context) {
	gc := s.Game
	if gc == nil {
		return
	}
	ch, err := gc.Stream(ctx)
	if err != nil {
		return
	}
	pr := s.prompt
	go func() {
		for msg := range ch {
			if msg.Type == "events" && len(msg.Events) > 0 {
				s.printEvents(msg.Events)
			}
			if pr != nil {
				pr.Refresh()
			}
		}
	}()
}
