package api

import (
	"net/http"

	"github.com/you/jrt/internal/hub"
	"github.com/you/jrt/internal/store"
)

// NewRouter は HTTP API のルータを生成する。
//
// Go 1.22+ のメソッド付きパターン（"METHOD /path"）と、
// r.PathValue によるパスパラメータ取得を用いる。
func NewRouter(h *hub.Hub, st store.Store) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", health)

	mux.HandleFunc("POST /api/games", createGame(h))
	mux.HandleFunc("GET /api/games", listGames(h))
	mux.HandleFunc("GET /api/games/{id}", getGame(h))
	mux.HandleFunc("DELETE /api/games/{id}", deleteGame(h))
	mux.HandleFunc("POST /api/games/{id}/actions", doAction(h, st))
	mux.HandleFunc("POST /api/games/{id}/turn/end", endTurn(h, st))
	mux.HandleFunc("GET /api/games/{id}/lines", listLines(h))
	mux.HandleFunc("GET /api/games/{id}/lines/{lid}", getLine(h))
	mux.HandleFunc("GET /api/games/{id}/companies", listCompanies(h))
	mux.HandleFunc("GET /api/games/{id}/cities", listCities(h))
	mux.HandleFunc("GET /api/games/{id}/events", listEvents(h))
	mux.HandleFunc("GET /api/games/{id}/can-build", canBuild(h))
	mux.HandleFunc("GET /api/games/{id}/stream", streamWS(h))

	return withRecover(withCORS(mux))
}

// health は死活監視用のエンドポイント。
func health(w http.ResponseWriter, r *http.Request) {
	writeOK(w, map[string]string{"status": "ok"}, nil)
}
