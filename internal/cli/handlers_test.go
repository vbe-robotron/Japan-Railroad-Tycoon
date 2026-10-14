package cli

import (
	"bytes"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/you/jrt/internal/api"
	"github.com/you/jrt/internal/core"
	"github.com/you/jrt/internal/hub"
	"github.com/you/jrt/internal/store"
)

// captureOutput は Stdout をバッファへ差し替え、テスト終了時に戻す。
func captureOutput(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	old := Stdout
	Stdout = buf
	t.Cleanup(func() { Stdout = old })
	return buf
}

// newTestSession は httptest サーバーに接続した Session を用意する。
func newTestSession(t *testing.T) (*Session, *hub.Hub) {
	t.Helper()
	st, err := store.NewSQLite(":memory:")
	if err != nil {
		t.Fatalf("NewSQLite: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	h := hub.New(st)
	srv := httptest.NewServer(api.NewRouter(h, st))
	t.Cleanup(srv.Close)

	s := NewSession(srv.URL)
	t.Cleanup(s.Close)
	return s, h
}

// startGame はテスト用にゲームを開始する。
func startGame(t *testing.T, s *Session) {
	t.Helper()
	if err := s.Execute(Parse("start hankyu --seed 42")); err != nil {
		t.Fatalf("start: %v", err)
	}
}

// addCities は実行中ゲームへ都市と競合会社を追加する。
func addCities(t *testing.T, h *hub.Hub, gameID string) {
	t.Helper()
	sess, ok := h.Get(gameID)
	if !ok {
		t.Fatalf("session %s not found", gameID)
	}
	sess.Lock()
	defer sess.Unlock()
	sess.Game.AddCity(&core.City{
		ID: "umeda", Name: "梅田", Population: 120000,
		Lat: 34.7055, Lon: 135.4983,
	})
	sess.Game.AddCity(&core.City{
		ID: "takarazuka", Name: "宝塚", Population: 30000,
		Lat: 34.7995, Lon: 135.3468,
	})
	rival := core.NewCompany("rival", "rival", "競合電鉄")
	rival.Cash = 5_000_000
	rival.TotalShares = 1000
	rival.StockPrice = 100
	sess.Game.AddCompany(rival)
}

func TestHandleStart(t *testing.T) {
	captureOutput(t)
	s, _ := newTestSession(t)
	startGame(t, s)

	if !s.Running {
		t.Error("Running = false, want true")
	}
	if s.Game == nil {
		t.Fatal("Game is nil")
	}
	if s.State == nil || s.State.PlayerID != "hankyu" {
		t.Fatalf("State = %+v", s.State)
	}
	if s.State.Company == nil || !s.State.Company.IsPlayer {
		t.Errorf("company = %+v, want player", s.State.Company)
	}
}

func TestHandleStatus(t *testing.T) {
	captureOutput(t)
	s, _ := newTestSession(t)
	startGame(t, s)

	s.State = nil
	if err := s.Execute(Parse("status")); err != nil {
		t.Fatalf("status: %v", err)
	}
	if s.State == nil || s.State.PlayerID != "hankyu" {
		t.Errorf("State = %+v, want hankyu", s.State)
	}
}

func TestHandleBuild(t *testing.T) {
	captureOutput(t)
	s, h := newTestSession(t)
	startGame(t, s)
	addCities(t, h, s.Game.ID())

	if err := s.Execute(Parse("build umeda takarazuka")); err != nil {
		t.Fatalf("build: %v", err)
	}
	ctx, cancel := s.ctx()
	defer cancel()
	lines, err := s.Game.ListLines(ctx)
	if err != nil {
		t.Fatalf("ListLines: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("len(lines) = %d, want 1", len(lines))
	}
}

func TestHandleBuild_MissingArgs(t *testing.T) {
	captureOutput(t)
	s, _ := newTestSession(t)
	startGame(t, s)
	if err := s.Execute(Parse("build umeda")); err == nil {
		t.Error("build with one arg: want error")
	}
}

func TestHandleEndTurn(t *testing.T) {
	captureOutput(t)
	s, _ := newTestSession(t)
	startGame(t, s)

	if err := s.Execute(Parse("et")); err != nil {
		t.Fatalf("et: %v", err)
	}
	if s.State.Turn != 1 {
		t.Errorf("Turn = %d, want 1", s.State.Turn)
	}
}

func TestHandleTurnN(t *testing.T) {
	captureOutput(t)
	s, _ := newTestSession(t)
	startGame(t, s)

	if err := s.Execute(Parse("turn 3")); err != nil {
		t.Fatalf("turn 3: %v", err)
	}
	if s.State.Turn != 3 {
		t.Errorf("Turn = %d, want 3", s.State.Turn)
	}
	if err := s.Execute(Parse("turn x")); err == nil {
		t.Error("turn x: want error")
	}
}

func TestHandleQuit(t *testing.T) {
	captureOutput(t)
	s, _ := newTestSession(t)
	startGame(t, s)

	err := s.Execute(Parse("quit"))
	if !errors.Is(err, ErrQuit) {
		t.Fatalf("quit err = %v, want ErrQuit", err)
	}
	if s.Running {
		t.Error("Running = true after quit")
	}
	if err := s.Execute(Parse("status")); !errors.Is(err, ErrNotRunning) {
		t.Errorf("post-quit status err = %v, want ErrNotRunning", err)
	}
}

func TestHandleList(t *testing.T) {
	buf := captureOutput(t)
	s, _ := newTestSession(t)
	startGame(t, s)

	if err := s.Execute(Parse("list")); err != nil {
		t.Fatalf("list: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("list produced no output")
	}
}

func TestHandleViewsAndActions(t *testing.T) {
	captureOutput(t)
	s, h := newTestSession(t)
	startGame(t, s)
	addCities(t, h, s.Game.ID())

	must := func(line string) {
		t.Helper()
		if err := s.Execute(Parse(line)); err != nil {
			t.Fatalf("%s: %v", line, err)
		}
	}
	must("build umeda takarazuka")

	ctx, cancel := s.ctx()
	lines, err := s.Game.ListLines(ctx)
	cancel()
	if err != nil || len(lines) == 0 {
		t.Fatalf("no lines: %v", err)
	}
	lineID := lines[0].ID

	must("lines")
	must("lines --sort profit")
	must("line " + lineID)
	must("companies")
	must("cities")
	must("can-build umeda takarazuka")
	must("can-build umeda takarazuka --gauge 1435")
	must("set-fare " + lineID + " 250")
	must("build-facility takarazuka housing")
	must("buy-shares rival 10")
	must("help")
	must("help build")
}

func TestHandleAbandonAndMA(t *testing.T) {
	captureOutput(t)
	s, h := newTestSession(t)
	startGame(t, s)
	addCities(t, h, s.Game.ID())

	if err := s.Execute(Parse("build umeda takarazuka")); err != nil {
		t.Fatalf("build: %v", err)
	}
	ctx, cancel := s.ctx()
	lines, err := s.Game.ListLines(ctx)
	cancel()
	if err != nil || len(lines) == 0 {
		t.Fatalf("no lines: %v", err)
	}

	// 成功は前提とせず、ハンドラが最後まで実行されることを確認する。
	_ = s.Execute(Parse("abandon-line " + lines[0].ID + " --reason economic"))
	_ = s.Execute(Parse("tob rival --premium 1.1 --ratio 0.51"))
	_ = s.Execute(Parse("merge rival 新会社 --ratio 0.5"))
}

func TestHandleBuildFacility_MissingArgs(t *testing.T) {
	captureOutput(t)
	s, _ := newTestSession(t)
	startGame(t, s)
	if err := s.Execute(Parse("build-facility takarazuka")); err == nil {
		t.Error("want error for missing kind")
	}
}

func TestHandleSetFare_MissingArgs(t *testing.T) {
	captureOutput(t)
	s, _ := newTestSession(t)
	startGame(t, s)
	if err := s.Execute(Parse("set-fare x")); err == nil {
		t.Error("want error for missing fare")
	}
	if err := s.Execute(Parse("set-fare x abc")); err == nil {
		t.Error("want error for non-numeric fare")
	}
}

func TestHandleLoad_InvalidID(t *testing.T) {
	captureOutput(t)
	s, _ := newTestSession(t)
	if err := s.Execute(Parse("load does-not-exist")); err == nil {
		t.Error("load invalid id: want error")
	}
}
