package hub

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/you/jrt/internal/core"
)

// fakeStore はテスト用のインメモリ Store 実装。
//
// エラー注入用のフィールドを持ち、失敗系の分岐を検証できるようにする。
type fakeStore struct {
	mu      sync.Mutex
	games   map[string]*core.Game
	events  map[string][]core.Event
	saveErr error
	loadErr error
	delErr  error
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		games:  map[string]*core.Game{},
		events: map[string][]core.Event{},
	}
}

func (f *fakeStore) SaveGame(_ context.Context, g *core.Game) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.games[g.ID] = g
	return nil
}

func (f *fakeStore) LoadGame(_ context.Context, id string) (*core.Game, error) {
	if f.loadErr != nil {
		return nil, f.loadErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	g, ok := f.games[id]
	if !ok {
		return nil, core.ErrNotFound
	}
	return g, nil
}

func (f *fakeStore) ListGames(_ context.Context) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	ids := make([]string, 0, len(f.games))
	for id := range f.games {
		ids = append(ids, id)
	}
	return ids, nil
}

func (f *fakeStore) DeleteGame(_ context.Context, id string) error {
	if f.delErr != nil {
		return f.delErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.games, id)
	return nil
}

func (f *fakeStore) AppendEvents(_ context.Context, gameID string, events []core.Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events[gameID] = append(f.events[gameID], events...)
	return nil
}

func (f *fakeStore) AppendAction(_ context.Context, _, _, _ string, _ []byte) error {
	return nil
}

func (f *fakeStore) Close() error { return nil }

func newTestHub() (*Hub, *fakeStore) {
	st := newFakeStore()
	return New(st, nil), st
}

// recvEvents は購読者からイベント列を 1 件受信する。
func recvEvents(t *testing.T, sub *Subscriber) []core.Event {
	t.Helper()
	select {
	case evs := <-sub.Events():
		return evs
	case <-time.After(time.Second):
		t.Fatal("イベントを受信できなかった")
		return nil
	}
}

func TestHub_Create(t *testing.T) {
	h, st := newTestHub()
	s, err := h.Create(context.Background(), 1907, "hankyu")
	if err != nil {
		t.Fatal(err)
	}
	if s.Game == nil {
		t.Fatal("Game が nil")
	}
	if s.ID() != s.Game.ID {
		t.Errorf("ID: %s vs %s", s.ID(), s.Game.ID)
	}
	if _, ok := st.games[s.ID()]; !ok {
		t.Error("store に保存されていない")
	}
}

func TestHub_Create_SaveError(t *testing.T) {
	h, st := newTestHub()
	st.saveErr = core.ErrNotFound
	if _, err := h.Create(context.Background(), 1, "p"); err == nil {
		t.Fatal("エラーを期待した")
	}
}

func TestHub_Get(t *testing.T) {
	h, _ := newTestHub()
	s, _ := h.Create(context.Background(), 1, "p")

	got, ok := h.Get(s.ID())
	if !ok || got != s {
		t.Fatal("既存セッションを取得できなかった")
	}
	if _, ok := h.Get("missing"); ok {
		t.Error("存在しない ID で ok が true")
	}
}

func TestHub_GetOrLoad(t *testing.T) {
	h, st := newTestHub()
	ctx := context.Background()

	g := core.NewGame(42, "takara")
	st.games[g.ID] = g

	s, err := h.GetOrLoad(ctx, g.ID)
	if err != nil {
		t.Fatal(err)
	}
	if s.Game.ID != g.ID {
		t.Errorf("ロード ID: %s vs %s", s.Game.ID, g.ID)
	}

	// 2 回目はメモリから同一セッションを返す。
	again, err := h.GetOrLoad(ctx, g.ID)
	if err != nil {
		t.Fatal(err)
	}
	if again != s {
		t.Error("同一セッションが返らない")
	}

	// Create 済みセッションもメモリから返る。
	c, _ := h.Create(ctx, 7, "p")
	got, err := h.GetOrLoad(ctx, c.ID())
	if err != nil || got != c {
		t.Error("メモリ上のセッションが返らない")
	}
}

func TestHub_GetOrLoad_LoadError(t *testing.T) {
	h, _ := newTestHub()
	if _, err := h.GetOrLoad(context.Background(), "missing"); err == nil {
		t.Fatal("エラーを期待した")
	}
}

func TestHub_Delete(t *testing.T) {
	h, st := newTestHub()
	ctx := context.Background()
	s, _ := h.Create(ctx, 1, "p")
	sub := s.Subscribe()

	if err := h.Delete(ctx, s.ID()); err != nil {
		t.Fatal(err)
	}
	if _, ok := h.Get(s.ID()); ok {
		t.Error("削除後もセッションが残っている")
	}
	if _, ok := st.games[s.ID()]; ok {
		t.Error("store から削除されていない")
	}
	select {
	case <-sub.Closed():
	case <-time.After(time.Second):
		t.Error("購読者が閉じられていない")
	}
}

func TestHub_Delete_DeleteError(t *testing.T) {
	h, st := newTestHub()
	st.delErr = core.ErrNotFound
	if err := h.Delete(context.Background(), "x"); err == nil {
		t.Fatal("エラーを期待した")
	}
}

func TestHub_List(t *testing.T) {
	h, _ := newTestHub()
	ctx := context.Background()
	if len(h.List()) != 0 {
		t.Error("初期状態で空でない")
	}
	a, _ := h.Create(ctx, 1, "p")
	b, _ := h.Create(ctx, 2, "p")

	ids := h.List()
	if len(ids) != 2 {
		t.Fatalf("件数: %d", len(ids))
	}
	set := map[string]bool{ids[0]: true, ids[1]: true}
	if !set[a.ID()] || !set[b.ID()] {
		t.Errorf("ID が揃っていない: %v", ids)
	}
}

func TestSession_Concurrent(t *testing.T) {
	h, _ := newTestHub()
	s, _ := h.Create(context.Background(), 1, "p")

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sub := s.Subscribe()
			s.Lock()
			_ = s.ID()
			s.Unlock()
			s.Broadcast([]core.Event{{Type: "x"}})
			s.Unsubscribe(sub)
		}()
	}
	wg.Wait()
}

func TestSession_Broadcast(t *testing.T) {
	h, _ := newTestHub()
	s, _ := h.Create(context.Background(), 1, "p")
	sub1 := s.Subscribe()
	sub2 := s.Subscribe()

	events := []core.Event{{ID: "e1", Type: "built"}}
	s.Broadcast(events)

	for _, sub := range []*Subscriber{sub1, sub2} {
		got := recvEvents(t, sub)
		if len(got) != 1 || got[0].ID != "e1" {
			t.Errorf("配信内容が不正: %+v", got)
		}
	}
}

func TestSession_Unsubscribe(t *testing.T) {
	h, _ := newTestHub()
	s, _ := h.Create(context.Background(), 1, "p")
	sub := s.Subscribe()
	s.Unsubscribe(sub)

	select {
	case <-sub.Closed():
	case <-time.After(time.Second):
		t.Fatal("Unsubscribe で閉じられていない")
	}

	// 購読者ゼロでも Broadcast は安全。
	s.Broadcast([]core.Event{{Type: "x"}})
}

func TestSession_Lock(t *testing.T) {
	h, _ := newTestHub()
	s, _ := h.Create(context.Background(), 1, "p")
	s.Lock()
	if s.Game == nil {
		t.Error("ロック中も Game を参照できるべき")
	}
	s.Unlock()
}
