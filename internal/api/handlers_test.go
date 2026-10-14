package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/you/jrt/internal/core"
	"github.com/you/jrt/internal/data"
	"github.com/you/jrt/internal/hub"
	"github.com/you/jrt/internal/store"
)

// ────────────────────────────────────────────
// テストヘルパー
// ────────────────────────────────────────────

func newSQLiteStore(t *testing.T) *store.SQLiteStore {
	t.Helper()
	st, err := store.NewSQLite(":memory:")
	if err != nil {
		t.Fatalf("NewSQLite: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func newTestServer(t *testing.T, st store.Store) (*httptest.Server, *hub.Hub) {
	t.Helper()
	h := hub.New(st, nil)
	srv := httptest.NewServer(NewRouter(h, st))
	t.Cleanup(srv.Close)
	return srv, h
}

// gameStateResponse はサーバーが返す状態要約。GameState の別名。
type gameStateResponse = GameState

// envelope は Response を汎用的に読み取るための型。
type envelope struct {
	OK     bool            `json:"ok"`
	Data   json.RawMessage `json:"data"`
	Events []core.Event    `json:"events"`
	Errors []APIError      `json:"errors"`
}

func doRequest(t *testing.T, method, url, body string) (*http.Response, envelope) {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, r)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	var env envelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	resp.Body.Close()
	return resp, env
}

// rawRequest は JSON を期待しないリクエストに使う。
func rawRequest(t *testing.T, method, url, body string) *http.Response {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, r)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	return resp
}

func createTestGame(t *testing.T, srv *httptest.Server, body string) (string, envelope) {
	t.Helper()
	resp, env := doRequest(t, http.MethodPost, srv.URL+"/api/games", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, want 201", resp.StatusCode)
	}
	var data gameStateResponse
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatalf("unmarshal create data: %v", err)
	}
	return data.ID, env
}

func addCities(t *testing.T, h *hub.Hub, id string, cities ...*core.City) {
	t.Helper()
	sess, ok := h.Get(id)
	if !ok {
		t.Fatalf("session not found: %s", id)
	}
	sess.Lock()
	for _, c := range cities {
		sess.Game.AddCity(c)
	}
	sess.Unlock()
}

func testCities() (*core.City, *core.City) {
	umeda := &core.City{ID: "umeda", Name: "梅田", Prefecture: "大阪府",
		Lat: 34.7055, Lon: 135.4983, Population: 120000, LandValue: 60}
	takarazuka := &core.City{ID: "takarazuka", Name: "宝塚", Prefecture: "兵庫県",
		Lat: 34.7995, Lon: 135.3468, Population: 30000, LandValue: 20}
	return umeda, takarazuka
}

// ────────────────────────────────────────────
// createGame / listGames / getGame / deleteGame
// ────────────────────────────────────────────

func TestCreateGame(t *testing.T) {
	st := newSQLiteStore(t)
	srv, _ := newTestServer(t, st)

	resp, env := doRequest(t, http.MethodPost, srv.URL+"/api/games",
		`{"company_id":"hankyu","seed":42}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	if !env.OK {
		t.Fatalf("ok = false: %+v", env.Errors)
	}
	var data gameStateResponse
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if data.PlayerID != "hankyu" {
		t.Errorf("player_id = %q, want hankyu", data.PlayerID)
	}
	if data.Company == nil {
		t.Fatalf("company is nil")
	}
	if data.Company.Name == "" {
		t.Error("company.Name is empty")
	}
	if data.Company.Cash != 9_000_000 {
		t.Errorf("cash = %d, want 9000000", data.Company.Cash)
	}
	if data.Company.Reputation != 50 {
		t.Errorf("reputation = %d, want 50", data.Company.Reputation)
	}
	if data.Date.IsZero() {
		t.Error("date should not be zero")
	}
	if data.Rank != 1 {
		t.Errorf("rank = %d, want 1", data.Rank)
	}
}

func TestCreateGame_DefaultPlayerID(t *testing.T) {
	st := newSQLiteStore(t)
	srv, _ := newTestServer(t, st)

	_, env := createTestGame(t, srv, "")
	var data gameStateResponse
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if data.PlayerID != "player" {
		t.Errorf("player_id = %q, want player", data.PlayerID)
	}
}

func TestListGames(t *testing.T) {
	st := newSQLiteStore(t)
	srv, _ := newTestServer(t, st)

	id1, _ := createTestGame(t, srv, `{"company_id":"a"}`)
	id2, _ := createTestGame(t, srv, `{"company_id":"b"}`)

	_, env := doRequest(t, http.MethodGet, srv.URL+"/api/games", "")
	var ids []string
	if err := json.Unmarshal(env.Data, &ids); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("len(ids) = %d, want 2", len(ids))
	}
	if !contains(ids, id1) || !contains(ids, id2) {
		t.Errorf("ids = %v, want contains %s and %s", ids, id1, id2)
	}
}

func TestGetGame(t *testing.T) {
	st := newSQLiteStore(t)
	srv, _ := newTestServer(t, st)
	id, _ := createTestGame(t, srv, `{"company_id":"hankyu"}`)

	resp, env := doRequest(t, http.MethodGet, srv.URL+"/api/games/"+id, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var data gameStateResponse
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if data.ID != id || data.PlayerID != "hankyu" {
		t.Errorf("data = %+v, want id=%s player=hankyu", data, id)
	}
}

func TestDeleteGame(t *testing.T) {
	st := newSQLiteStore(t)
	srv, _ := newTestServer(t, st)
	id, _ := createTestGame(t, srv, `{"company_id":"hankyu"}`)

	resp, env := doRequest(t, http.MethodDelete, srv.URL+"/api/games/"+id, "")
	if resp.StatusCode != http.StatusOK || !env.OK {
		t.Fatalf("delete status = %d ok=%v", resp.StatusCode, env.OK)
	}

	resp2 := rawRequest(t, http.MethodGet, srv.URL+"/api/games/"+id, "")
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusNotFound {
		t.Errorf("get after delete status = %d, want 404", resp2.StatusCode)
	}
}

// ────────────────────────────────────────────
// アクション
// ────────────────────────────────────────────

func TestDoAction_BuildLine(t *testing.T) {
	st := newSQLiteStore(t)
	srv, h := newTestServer(t, st)
	id, _ := createTestGame(t, srv, `{"company_id":"hankyu","seed":7}`)

	umeda, takarazuka := testCities()
	addCities(t, h, id, umeda, takarazuka)

	resp, env := doRequest(t, http.MethodPost, srv.URL+"/api/games/"+id+"/actions",
		`{"type":"build_line","payload":{"from":"umeda","to":"takarazuka","gauge":1067,"double":false}}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200: %+v", resp.StatusCode, env.Errors)
	}
	if !env.OK {
		t.Fatalf("ok = false: %+v", env.Errors)
	}
	if !hasEventType(env.Events, core.EventLineOpened) {
		t.Errorf("events = %+v, want line_opened", env.Events)
	}

	// 路線一覧に 1 件
	_, le := doRequest(t, http.MethodGet, srv.URL+"/api/games/"+id+"/lines", "")
	var lines []*core.Line
	if err := json.Unmarshal(le.Data, &lines); err != nil {
		t.Fatalf("unmarshal lines: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("len(lines) = %d, want 1", len(lines))
	}

	// 個別取得
	resp3, _ := doRequest(t, http.MethodGet, srv.URL+"/api/games/"+id+"/lines/"+lines[0].ID, "")
	if resp3.StatusCode != http.StatusOK {
		t.Errorf("get line status = %d, want 200", resp3.StatusCode)
	}
}

func TestEndTurn(t *testing.T) {
	st := newSQLiteStore(t)
	srv, _ := newTestServer(t, st)
	id, _ := createTestGame(t, srv, `{"company_id":"hankyu"}`)

	resp, env := doRequest(t, http.MethodPost, srv.URL+"/api/games/"+id+"/turn/end", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200: %+v", resp.StatusCode, env.Errors)
	}
	if !hasEventType(env.Events, core.EventTurnEnded) {
		t.Errorf("events = %+v, want turn_ended", env.Events)
	}

	_, ge := doRequest(t, http.MethodGet, srv.URL+"/api/games/"+id, "")
	var data gameStateResponse
	if err := json.Unmarshal(ge.Data, &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if data.Turn != 1 {
		t.Errorf("turn = %d, want 1", data.Turn)
	}
}

func TestCanBuild(t *testing.T) {
	st := newSQLiteStore(t)
	srv, h := newTestServer(t, st)
	id, _ := createTestGame(t, srv, `{"company_id":"hankyu"}`)

	umeda, takarazuka := testCities()
	addCities(t, h, id, umeda, takarazuka)

	resp, env := doRequest(t, http.MethodGet,
		srv.URL+"/api/games/"+id+"/can-build?from=umeda&to=takarazuka&gauge=1067", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var result core.CanBuildResult
	if err := json.Unmarshal(env.Data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !result.Can {
		t.Errorf("can = false, gates = %+v", result.Gates)
	}
	if result.Cost <= 0 {
		t.Errorf("cost = %d, want > 0", result.Cost)
	}
}

func TestCanBuild_BadParams(t *testing.T) {
	st := newSQLiteStore(t)
	srv, _ := newTestServer(t, st)
	id, _ := createTestGame(t, srv, `{"company_id":"hankyu"}`)

	resp, env := doRequest(t, http.MethodGet,
		srv.URL+"/api/games/"+id+"/can-build?from=a&to=b&gauge=abc", "")
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("bad gauge status = %d, want 400", resp.StatusCode)
	}
	if len(env.Errors) == 0 || env.Errors[0].Code != "INVALID_GAUGE" {
		t.Errorf("errors = %+v, want INVALID_GAUGE", env.Errors)
	}

	resp2, env2 := doRequest(t, http.MethodGet,
		srv.URL+"/api/games/"+id+"/can-build?from=a&to=b&double=xyz", "")
	if resp2.StatusCode != http.StatusBadRequest {
		t.Errorf("bad double status = %d, want 400", resp2.StatusCode)
	}
	if len(env2.Errors) == 0 || env2.Errors[0].Code != "INVALID_DOUBLE" {
		t.Errorf("errors = %+v, want INVALID_DOUBLE", env2.Errors)
	}
}

func TestLists(t *testing.T) {
	st := newSQLiteStore(t)
	srv, h := newTestServer(t, st)
	id, _ := createTestGame(t, srv, `{"company_id":"hankyu"}`)

	umeda, takarazuka := testCities()
	addCities(t, h, id, umeda, takarazuka)
	doRequest(t, http.MethodPost, srv.URL+"/api/games/"+id+"/turn/end", "")

	// companies
	_, ce := doRequest(t, http.MethodGet, srv.URL+"/api/games/"+id+"/companies", "")
	var companies []*core.Company
	if err := json.Unmarshal(ce.Data, &companies); err != nil {
		t.Fatalf("companies: %v", err)
	}
	if len(companies) != 1 {
		t.Errorf("len(companies) = %d, want 1", len(companies))
	}

	// cities
	_, ci := doRequest(t, http.MethodGet, srv.URL+"/api/games/"+id+"/cities", "")
	var cities []*core.City
	if err := json.Unmarshal(ci.Data, &cities); err != nil {
		t.Fatalf("cities: %v", err)
	}
	if len(cities) != 2 {
		t.Errorf("len(cities) = %d, want 2", len(cities))
	}

	// events
	_, ev := doRequest(t, http.MethodGet, srv.URL+"/api/games/"+id+"/events", "")
	var events []core.Event
	if err := json.Unmarshal(ev.Data, &events); err != nil {
		t.Fatalf("events: %v", err)
	}
	if !hasEventType(events, core.EventTurnEnded) {
		t.Errorf("events = %+v, want turn_ended", events)
	}
}

// ────────────────────────────────────────────
// エラー系
// ────────────────────────────────────────────

func TestErrorResponses(t *testing.T) {
	st := newSQLiteStore(t)
	srv, _ := newTestServer(t, st)

	// 404: 存在しないゲーム
	resp := rawRequest(t, http.MethodGet, srv.URL+"/api/games/nope", "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("missing game status = %d, want 404", resp.StatusCode)
	}

	// 404: 存在しないゲームへのアクション
	resp2 := rawRequest(t, http.MethodPost, srv.URL+"/api/games/nope/actions",
		`{"type":"end_turn"}`)
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusNotFound {
		t.Errorf("action missing game status = %d, want 404", resp2.StatusCode)
	}

	// 400: 不正 JSON
	resp3 := rawRequest(t, http.MethodPost, srv.URL+"/api/games/nope/actions", `{`)
	resp3.Body.Close()
	if resp3.StatusCode != http.StatusBadRequest {
		t.Errorf("invalid json status = %d, want 400", resp3.StatusCode)
	}

	id, _ := createTestGame(t, srv, `{"company_id":"hankyu"}`)

	// 400: type 未指定
	resp4, env4 := doRequest(t, http.MethodPost, srv.URL+"/api/games/"+id+"/actions", `{}`)
	if resp4.StatusCode != http.StatusBadRequest || env4.Errors[0].Code != "MISSING_TYPE" {
		t.Errorf("missing type: status=%d errors=%+v", resp4.StatusCode, env4.Errors)
	}

	// 400: 未知のアクション
	resp5, env5 := doRequest(t, http.MethodPost, srv.URL+"/api/games/"+id+"/actions",
		`{"type":"does_not_exist"}`)
	if resp5.StatusCode != http.StatusBadRequest || env5.Errors[0].Code != "UNKNOWN_ACTION" {
		t.Errorf("unknown action: status=%d errors=%+v", resp5.StatusCode, env5.Errors)
	}

	// 400: 検証エラー（SAME_CITY）
	resp6, env6 := doRequest(t, http.MethodPost, srv.URL+"/api/games/"+id+"/actions",
		`{"type":"build_line","payload":{"from":"x","to":"x","gauge":1067}}`)
	if resp6.StatusCode != http.StatusBadRequest || env6.Errors[0].Code != "SAME_CITY" {
		t.Errorf("same city: status=%d errors=%+v", resp6.StatusCode, env6.Errors)
	}

	// 404: 検証エラー（FROM_NOT_FOUND）
	resp7, env7 := doRequest(t, http.MethodPost, srv.URL+"/api/games/"+id+"/actions",
		`{"type":"build_line","payload":{"from":"zzz","to":"yyy","gauge":1067}}`)
	if resp7.StatusCode != http.StatusNotFound || env7.Errors[0].Code != "FROM_NOT_FOUND" {
		t.Errorf("from not found: status=%d errors=%+v", resp7.StatusCode, env7.Errors)
	}
}

// failStore は SaveGame を強制的に失敗させる Store ラッパー。
type failStore struct {
	store.Store
}

func (f *failStore) SaveGame(_ context.Context, _ *core.Game) error {
	return errors.New("save failed")
}

func TestErrorResponse_InternalServerError(t *testing.T) {
	base := newSQLiteStore(t)
	fs := &failStore{Store: base}
	h := hub.New(fs, nil)
	srv := httptest.NewServer(NewRouter(h, fs))
	defer srv.Close()

	resp, env := doRequest(t, http.MethodPost, srv.URL+"/api/games", `{"company_id":"x"}`)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	if env.OK || len(env.Errors) == 0 || env.Errors[0].Code != "CREATE_FAILED" {
		t.Errorf("errors = %+v, want CREATE_FAILED", env.Errors)
	}
}

// newDataServer は vanilla データを結線したサーバーを用意する。
func newDataServer(t *testing.T) (*httptest.Server, *hub.Hub) {
	t.Helper()
	st := newSQLiteStore(t)
	ds, err := data.Load(filepath.Join("..", "..", "data"))
	if err != nil {
		t.Fatalf("data.Load: %v", err)
	}
	h := hub.New(st, ds)
	srv := httptest.NewServer(NewRouter(h, st))
	t.Cleanup(srv.Close)
	return srv, h
}

// ────────────────────────────────────────────
// データ結線
// ────────────────────────────────────────────

func TestCreateGame_WithData(t *testing.T) {
	srv, _ := newDataServer(t)
	id, _ := createTestGame(t, srv, `{"company_id":"hankyu","seed":42}`)

	_, ce := doRequest(t, http.MethodGet, srv.URL+"/api/games/"+id+"/companies", "")
	var companies []*core.Company
	if err := json.Unmarshal(ce.Data, &companies); err != nil {
		t.Fatalf("companies: %v", err)
	}
	if len(companies) != 20 {
		t.Fatalf("len(companies) = %d, want 20", len(companies))
	}

	_, ci := doRequest(t, http.MethodGet, srv.URL+"/api/games/"+id+"/cities", "")
	var cities []*core.City
	if err := json.Unmarshal(ci.Data, &cities); err != nil {
		t.Fatalf("cities: %v", err)
	}
	if len(cities) != 60 {
		t.Fatalf("len(cities) = %d, want 60", len(cities))
	}

	var player *core.Company
	for _, c := range companies {
		if c.ID == "hankyu" {
			player = c
		}
	}
	if player == nil || !player.IsPlayer {
		t.Errorf("hankyu がプレイヤー会社でない: %+v", player)
	}
}

func TestBuildLine_AfterCreate(t *testing.T) {
	srv, _ := newDataServer(t)
	id, _ := createTestGame(t, srv, `{"company_id":"hankyu","seed":7}`)

	// 都市投入はデータ結線済み。直接 build できる。
	resp, env := doRequest(t, http.MethodPost, srv.URL+"/api/games/"+id+"/actions",
		`{"type":"build_line","payload":{"from":"umeda","to":"takarazuka","gauge":1067,"double":false}}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200: %+v", resp.StatusCode, env.Errors)
	}
	if !env.OK {
		t.Fatalf("ok = false: %+v", env.Errors)
	}
	if !hasEventType(env.Events, core.EventLineOpened) {
		t.Errorf("events = %+v, want line_opened", env.Events)
	}

	_, le := doRequest(t, http.MethodGet, srv.URL+"/api/games/"+id+"/lines", "")
	var lines []*core.Line
	if err := json.Unmarshal(le.Data, &lines); err != nil {
		t.Fatalf("unmarshal lines: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("len(lines) = %d, want 1", len(lines))
	}
}

// ────────────────────────────────────────────
// ユーティリティ
// ────────────────────────────────────────────
func hasEventType(events []core.Event, typ string) bool {
	for _, e := range events {
		if e.Type == typ {
			return true
		}
	}
	return false
}

func contains(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}
