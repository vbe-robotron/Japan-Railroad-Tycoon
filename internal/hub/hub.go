package hub

import (
	"context"
	"sync"

	"github.com/you/jrt/internal/core"
	"github.com/you/jrt/internal/data"
	"github.com/you/jrt/internal/store"
)

// Hub は複数ゲームのセッションを管理する。
//
// sessions は ID → Session のマップであり、mu で保護する。
// store は永続化層であり、セッションの生成・ロード・削除のたびに利用する。
// data は新規ゲームへ投入する都市・会社データであり、nil の場合は投入しない。
type Hub struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	store    store.Store
	data     *data.Dataset
}

// New は空の Hub を生成する。ds は nil でもよい。
func New(st store.Store, ds *data.Dataset) *Hub {
	return &Hub{
		sessions: map[string]*Session{},
		store:    st,
		data:     ds,
	}
}

// Data は Hub が保持するゲームデータを返す。未設定なら nil。
func (h *Hub) Data() *data.Dataset {
	return h.data
}

// Store は Hub が保持する永続化層を返す。未設定なら nil。
func (h *Hub) Store() store.Store {
	return h.store
}

// Session は 1 ゲームの状態・ロック・購読者を保持する。
//
// mu はゲーム状態（Game）を保護し、Lock/Unlock で外部に公開する。
// subsMu は購読者マップを保護する。ゲーム本体のロックと購読者操作を
// 分離することで、Game をロック中でも Broadcast がデッドロックしない。
type Session struct {
	mu   sync.Mutex
	Game *core.Game

	subsMu sync.RWMutex
	subs   map[*Subscriber]struct{}
}

// newSession はゲームを包む Session を生成する。
func newSession(g *core.Game) *Session {
	return &Session{
		Game: g,
		subs: map[*Subscriber]struct{}{},
	}
}

// Get は既存のセッションを返す。存在しなければ ok は false。
func (h *Hub) Get(id string) (*Session, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	s, ok := h.sessions[id]
	return s, ok
}

// GetOrLoad はセッションを返す。メモリになければ store からロードする。
//
// ロード中の競合に備え、マップへの登録は二重チェックで行い、
// 既に他ゴルーチンが登録済みならその既存セッションを返す。
func (h *Hub) GetOrLoad(ctx context.Context, id string) (*Session, error) {
	h.mu.RLock()
	s, ok := h.sessions[id]
	h.mu.RUnlock()
	if ok {
		return s, nil
	}

	g, err := h.store.LoadGame(ctx, id)
	if err != nil {
		return nil, err
	}

	s = newSession(g)
	h.mu.Lock()
	if existing, ok := h.sessions[id]; ok {
		s = existing
	} else {
		h.sessions[id] = s
	}
	h.mu.Unlock()
	return s, nil
}

// Create は新規セッションを作成し、store に保存する。
func (h *Hub) Create(ctx context.Context, seed int64, playerID string) (*Session, error) {
	g := core.NewGame(seed, playerID)
	if err := h.store.SaveGame(ctx, g); err != nil {
		return nil, err
	}

	s := newSession(g)
	h.mu.Lock()
	h.sessions[g.ID] = s
	h.mu.Unlock()
	return s, nil
}

// Delete はセッションを削除する。store からも削除し、購読者を切断する。
func (h *Hub) Delete(ctx context.Context, id string) error {
	if err := h.store.DeleteGame(ctx, id); err != nil {
		return err
	}

	h.mu.Lock()
	s, ok := h.sessions[id]
	delete(h.sessions, id)
	h.mu.Unlock()

	if ok {
		s.closeAll()
	}
	return nil
}

// List はメモリ上に存在する全セッションの ID を返す。
func (h *Hub) List() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ids := make([]string, 0, len(h.sessions))
	for id := range h.sessions {
		ids = append(ids, id)
	}
	return ids
}

// Lock は Session のゲームロックを取得する。
func (s *Session) Lock() { s.mu.Lock() }

// Unlock は Session のゲームロックを解放する。
func (s *Session) Unlock() { s.mu.Unlock() }

// ID はゲーム ID を返す。
func (s *Session) ID() string { return s.Game.ID }

// Subscribe は購読者を追加して返す。
func (s *Session) Subscribe() *Subscriber {
	sub := NewSubscriber()
	s.subsMu.Lock()
	s.subs[sub] = struct{}{}
	s.subsMu.Unlock()
	return sub
}

// Unsubscribe は購読者を削除して閉じる。
func (s *Session) Unsubscribe(sub *Subscriber) {
	s.subsMu.Lock()
	_, ok := s.subs[sub]
	delete(s.subs, sub)
	s.subsMu.Unlock()
	if ok {
		sub.Close()
	}
}

// Broadcast は全購読者にイベントを配信する。
//
// 購読者マップはスナップショットしてから配信する。Send は非ブロッキングの
// ため、遅い購読者がいても他への配信を妨げない。
func (s *Session) Broadcast(events []core.Event) {
	s.subsMu.RLock()
	subs := make([]*Subscriber, 0, len(s.subs))
	for sub := range s.subs {
		subs = append(subs, sub)
	}
	s.subsMu.RUnlock()

	for _, sub := range subs {
		sub.Send(events)
	}
}

// closeAll は全購読者を閉じる。Delete 時に呼ぶ。
func (s *Session) closeAll() {
	s.subsMu.Lock()
	subs := make([]*Subscriber, 0, len(s.subs))
	for sub := range s.subs {
		subs = append(subs, sub)
	}
	s.subs = map[*Subscriber]struct{}{}
	s.subsMu.Unlock()

	for _, sub := range subs {
		sub.Close()
	}
}
