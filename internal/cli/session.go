package cli

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/you/jrt/internal/api"
	"github.com/you/jrt/internal/core"
)

// ErrQuit は quit コマンドで戻される。呼び出し側はループを終了する。
var ErrQuit = errors.New("quit")

// ErrNotRunning はゲーム開始前にゲーム操作コマンドを実行したことを表す。
var ErrNotRunning = errors.New("ゲームが開始されていません（'start <company>' または 'load <id>'）")

// UnknownCommandError は未知のコマンドを表す。
type UnknownCommandError struct {
	Verb string
}

// Error はエラーメッセージを返す。
func (e *UnknownCommandError) Error() string {
	return "不明なコマンド: " + e.Verb
}

// promptRefresher はプロンプト再描画を行う（readline.Instance が満たす）。
type promptRefresher interface {
	Refresh()
}

// Session は 1 回の CLI 実行の状態を保持する。
//
// Client は HTTP クライアント、Game / State は現在のゲーム、
// Running はゲーム開始済みかどうかを表す。CLI はロジックを持たず、
// 全判定はサーバー側で行われる。
type Session struct {
	Client  *api.Client
	Game    *api.GameClient
	State   *api.GameState
	Running bool

	baseCtx context.Context
	timeout time.Duration

	mu     sync.Mutex
	cancel context.CancelFunc

	seenMu sync.Mutex
	seen   map[string]struct{}

	prompt promptRefresher
}

// NewSession はサーバー URL からセッションを生成する。
func NewSession(baseURL string) *Session {
	return &Session{
		Client:  api.NewClient(api.Config{BaseURL: baseURL, Timeout: 10 * time.Second}),
		baseCtx: context.Background(),
		timeout: 10 * time.Second,
		seen:    map[string]struct{}{},
	}
}

// SetRefresher はイベント受信時にプロンプトを再描画するための補完器を設定する。
func (s *Session) SetRefresher(r promptRefresher) { s.prompt = r }

// Execute は 1 コマンドを実行する。未知コマンドや未開始時のゲーム操作は
// エラーを返す。空コマンドは何もしない。
func (s *Session) Execute(cmd *Command) error {
	if cmd == nil || cmd.Verb == "" {
		return nil
	}
	verb := resolveVerb(cmd.Verb)
	h, ok := handlers[verb]
	if !ok {
		return &UnknownCommandError{Verb: cmd.Verb}
	}
	if requiresGame(verb) && !s.Running {
		return ErrNotRunning
	}
	normalized := *cmd
	normalized.Verb = verb
	return h(s, &normalized)
}

// Close はイベント購読を停止する。
func (s *Session) Close() {
	s.stopEvents()
}

// ctx はタイムアウト付きのリクエストコンテキストを返す。
func (s *Session) ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(s.baseCtx, s.timeout)
}

// startEvents は現在のゲームのイベント購読を（再）開始する。
func (s *Session) startEvents() {
	s.stopEvents()
	if s.Game == nil {
		return
	}
	ctx, cancel := context.WithCancel(s.baseCtx)
	s.mu.Lock()
	s.cancel = cancel
	s.mu.Unlock()
	s.StartEventListener(ctx)
}

// stopEvents は進行中のイベント購読を停止する。
func (s *Session) stopEvents() {
	s.mu.Lock()
	cancel := s.cancel
	s.cancel = nil
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// begin はゲーム開始時の共通処理（状態設定・イベント購読開始）を行う。
func (s *Session) begin(g *api.GameClient, st *api.GameState) {
	s.Game = g
	s.State = st
	s.Running = true
	s.seenMu.Lock()
	s.seen = map[string]struct{}{}
	s.seenMu.Unlock()
	s.startEvents()
}

// printEvents は未表示のイベントだけを表示する。
//
// アクションの HTTP 応答と WebSocket 配信の双方から同じイベントが届くため、
// イベント ID で重複を除く。ID が空のイベントは常に表示する。
func (s *Session) printEvents(events []core.Event) {
	if len(events) == 0 {
		return
	}
	fresh := make([]core.Event, 0, len(events))
	s.seenMu.Lock()
	for _, e := range events {
		if e.ID != "" {
			if _, ok := s.seen[e.ID]; ok {
				continue
			}
			s.seen[e.ID] = struct{}{}
		}
		fresh = append(fresh, e)
	}
	s.seenMu.Unlock()
	printEvents(fresh)
}
