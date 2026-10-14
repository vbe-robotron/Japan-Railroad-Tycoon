package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/you/jrt/internal/core"
	"github.com/you/jrt/internal/hub"
	"github.com/you/jrt/internal/store"
)

// defaultStartDate は新規ゲームの開始年月。
//
// core.NewGame は Date をゼロ値のまま返すため、API 側で既定値を設定する。
var defaultStartDate = core.Date{Year: 1907, Month: 10}

// session はパスパラメータ id のセッションを取得する。
//
// メモリ優先で、なければ store からロードする。見つからない場合は
// 404 を書き出して ok=false を返す。
func session(h *hub.Hub, w http.ResponseWriter, r *http.Request) (*hub.Session, bool) {
	id := r.PathValue("id")
	if s, ok := h.Get(id); ok {
		return s, true
	}
	s, err := h.GetOrLoad(r.Context(), id)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "GAME_NOT_FOUND", "ゲームが見つかりません")
			return nil, false
		}
		writeError(w, http.StatusInternalServerError, "LOAD_FAILED", err.Error())
		return nil, false
	}
	return s, true
}

// gameStateResponse はゲーム状態の要約。
type gameStateResponse struct {
	ID       string        `json:"id"`
	Date     core.Date     `json:"date"`
	Turn     int           `json:"turn"`
	PlayerID string        `json:"player_id"`
	Company  *core.Company `json:"company,omitempty"`
	Rank     int           `json:"rank"`
}

// stateOf はゲームから状態要約を生成する。呼び出し側でロックすること。
func stateOf(g *core.Game) gameStateResponse {
	return gameStateResponse{
		ID:       g.ID,
		Date:     g.Date,
		Turn:     g.Turn,
		PlayerID: g.PlayerID,
		Company:  g.PlayerCompany(),
		Rank:     rankOf(g),
	}
}

// rankOf は純資産順の順位（1 が最上位）を返す。
func rankOf(g *core.Game) int {
	player := g.PlayerCompany()
	if player == nil {
		return 0
	}
	pv := player.NetAssetsWithLines(g)
	rank := 1
	for id, c := range g.Companies {
		if id == player.ID {
			continue
		}
		if c.NetAssetsWithLines(g) > pv {
			rank++
		}
	}
	return rank
}

// createGameRequest は新規ゲーム作成のリクエスト。
type createGameRequest struct {
	CompanyID  string `json:"company_id"`
	Seed       int64  `json:"seed,omitempty"`
	Difficulty string `json:"difficulty,omitempty"`
}

// createGame は新規ゲームを作成する。
//
// seed が 0 の場合は現在時刻から決定する。作成直後に空のプレイヤー会社を
// 追加する（Phase 8 で本格的な初期会社を作るまでの暫定）。
func createGame(h *hub.Hub, st store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createGameRequest
		if r.Body != nil {
			// ボディが空の場合は EOF を無視する。
			_ = json.NewDecoder(r.Body).Decode(&req)
		}
		if req.CompanyID == "" {
			req.CompanyID = "player"
		}
		seed := req.Seed
		if seed == 0 {
			seed = time.Now().UnixNano()
		}

		sess, err := h.Create(r.Context(), seed, req.CompanyID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
			return
		}

		sess.Lock()
		// mod データが結線されていれば都市・会社を投入する。
		if ds := h.Data(); ds != nil {
			ds.ApplyToGame(sess.Game, req.CompanyID)
		}
		if _, ok := sess.Game.Companies[req.CompanyID]; !ok {
			c := core.NewCompany(req.CompanyID, req.CompanyID, req.CompanyID)
			c.IsPlayer = true
			c.Cash = 9_000_000
			c.TotalShares = 10000
			c.StockPrice = 1000
			sess.Game.AddCompany(c)
		}
		if sess.Game.Date.IsZero() {
			sess.Game.Date = defaultStartDate
		}
		// 全社に AI を割り当てる（プレイヤー会社は HistoryIgnore）。
		sess.Game.InitAIs()
		_ = st.SaveGame(r.Context(), sess.Game)
		resp := stateOf(sess.Game)
		sess.Unlock()

		writeJSON(w, http.StatusCreated, Response{OK: true, Data: resp})
	}
}

// listGames はメモリ上のゲーム ID 一覧を返す。
func listGames(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ids := h.List()
		sort.Strings(ids)
		writeOK(w, ids, nil)
	}
}

// getGame はゲーム状態を返す。
func getGame(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, ok := session(h, w, r)
		if !ok {
			return
		}
		sess.Lock()
		resp := stateOf(sess.Game)
		sess.Unlock()
		writeOK(w, resp, nil)
	}
}

// deleteGame はゲームを削除する。
func deleteGame(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if _, ok := h.Get(id); !ok {
			if _, err := h.GetOrLoad(r.Context(), id); err != nil {
				if errors.Is(err, core.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
					writeError(w, http.StatusNotFound, "GAME_NOT_FOUND", "ゲームが見つかりません")
					return
				}
			}
		}
		if err := h.Delete(r.Context(), id); err != nil {
			writeError(w, http.StatusInternalServerError, "DELETE_FAILED", err.Error())
			return
		}
		writeOK(w, map[string]string{"id": id}, nil)
	}
}

// actionRequest はアクション実行のリクエスト。
type actionRequest struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// doAction は任意のアクションを実行する。
func doAction(h *hub.Hub, st store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req actionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", "リクエストの JSON が不正です")
			return
		}
		if req.Type == "" {
			writeError(w, http.StatusBadRequest, "MISSING_TYPE", "アクション種別が指定されていません")
			return
		}
		action, err := core.DecodeAction(req.Type, req.Payload)
		if err != nil {
			writeError(w, http.StatusBadRequest, "UNKNOWN_ACTION", err.Error())
			return
		}

		sess, ok := session(h, w, r)
		if !ok {
			return
		}
		execute(w, r, sess, st, action, req.Payload)
	}
}

// endTurn はターンを終了する。
func endTurn(h *hub.Hub, st store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, ok := session(h, w, r)
		if !ok {
			return
		}
		execute(w, r, sess, st, &core.EndTurnAction{}, nil)
	}
}

// execute はアクションの検証・適用・永続化・配信を一貫して行う。
func execute(w http.ResponseWriter, r *http.Request, sess *hub.Session, st store.Store, action core.Action, payload []byte) {
	events, verr, serr := runAction(r.Context(), sess, st, action, payload)
	switch {
	case verr != nil:
		writeValidationError(w, verr)
	case serr != nil:
		writeError(w, http.StatusInternalServerError, "PERSIST_FAILED", serr.Error())
	default:
		writeOK(w, nil, events)
	}
}

// runAction はゲームをロックしてアクションを適用し、永続化と配信を行う。
//
// 戻り値は (イベント, 検証エラー, 永続化エラー)。検証エラーは
// ゲームルール違反、永続化エラーは store 障害を表す。
func runAction(ctx context.Context, sess *hub.Session, st store.Store, action core.Action, payload []byte) ([]core.Event, error, error) {
	sess.Lock()
	defer sess.Unlock()

	if err := action.Validate(sess.Game); err != nil {
		return nil, err, nil
	}
	events, err := action.Apply(sess.Game)
	if err != nil {
		return nil, err, nil
	}

	// 整合性のためロック保持中に保存する。
	if err := st.SaveGame(ctx, sess.Game); err != nil {
		return nil, nil, err
	}
	if err := st.AppendAction(ctx, sess.Game.ID, sess.Game.PlayerID, action.Type(), payload); err != nil {
		return nil, nil, err
	}
	if err := st.AppendEvents(ctx, sess.Game.ID, events); err != nil {
		return nil, nil, err
	}

	sess.Broadcast(events)
	return events, nil, nil
}

// listLines はプレイヤー会社の路線一覧を返す。
func listLines(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, ok := session(h, w, r)
		if !ok {
			return
		}
		sess.Lock()
		lines := []*core.Line{}
		if c := sess.Game.PlayerCompany(); c != nil {
			lines = append(lines, c.Lines...)
		}
		sess.Unlock()
		writeOK(w, lines, nil)
	}
}

// getLine は ID 指定で 1 本の路線を返す。
func getLine(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, ok := session(h, w, r)
		if !ok {
			return
		}
		lid := r.PathValue("lid")
		sess.Lock()
		var line *core.Line
		for _, c := range sess.Game.Companies {
			if l := c.FindLine(lid); l != nil {
				line = l
				break
			}
		}
		sess.Unlock()
		if line == nil {
			writeError(w, http.StatusNotFound, "LINE_NOT_FOUND", "路線が見つかりません")
			return
		}
		writeOK(w, line, nil)
	}
}

// listCompanies は会社一覧を ID 昇順で返す。
func listCompanies(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, ok := session(h, w, r)
		if !ok {
			return
		}
		sess.Lock()
		companies := make([]*core.Company, 0, len(sess.Game.Companies))
		for _, c := range sess.Game.Companies {
			companies = append(companies, c)
		}
		sess.Unlock()
		sort.Slice(companies, func(i, j int) bool { return companies[i].ID < companies[j].ID })
		writeOK(w, companies, nil)
	}
}

// listCities は都市一覧を ID 昇順で返す。
func listCities(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, ok := session(h, w, r)
		if !ok {
			return
		}
		sess.Lock()
		cities := make([]*core.City, 0, len(sess.Game.Cities))
		for _, c := range sess.Game.Cities {
			cities = append(cities, c)
		}
		sess.Unlock()
		sort.Slice(cities, func(i, j int) bool { return cities[i].ID < cities[j].ID })
		writeOK(w, cities, nil)
	}
}

// listEvents はゲーム内イベント一覧を返す。
func listEvents(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, ok := session(h, w, r)
		if !ok {
			return
		}
		sess.Lock()
		events := make([]core.Event, len(sess.Game.Events))
		copy(events, sess.Game.Events)
		sess.Unlock()
		writeOK(w, events, nil)
	}
}

// canBuild は路線敷設の可否（4 ゲート）を返す。
//
// クエリ: from, to, gauge（既定 1067）, double（既定 false）。
func canBuild(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, ok := session(h, w, r)
		if !ok {
			return
		}
		q := r.URL.Query()
		from := q.Get("from")
		to := q.Get("to")

		gauge := 1067
		if s := q.Get("gauge"); s != "" {
			v, err := strconv.Atoi(s)
			if err != nil {
				writeError(w, http.StatusBadRequest, "INVALID_GAUGE", "gauge は整数で指定してください")
				return
			}
			gauge = v
		}
		double := false
		if s := q.Get("double"); s != "" {
			v, err := strconv.ParseBool(s)
			if err != nil {
				writeError(w, http.StatusBadRequest, "INVALID_DOUBLE", "double は真偽値で指定してください")
				return
			}
			double = v
		}

		sess.Lock()
		result := sess.Game.CanBuildLine(from, to, gauge, double)
		sess.Unlock()
		writeOK(w, result, nil)
	}
}
