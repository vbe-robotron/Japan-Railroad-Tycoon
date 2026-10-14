package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/you/jrt/internal/core"
)

func TestRouter_Methods(t *testing.T) {
	st := newSQLiteStore(t)
	srv, h := newTestServer(t, st)
	id, _ := createTestGame(t, srv, `{"company_id":"hankyu","seed":3}`)

	umeda, takarazuka := testCities()
	addCities(t, h, id, umeda, takarazuka)

	// 路線を 1 本敷設しておく。
	resp, _ := doRequest(t, http.MethodPost, srv.URL+"/api/games/"+id+"/actions",
		`{"type":"build_line","payload":{"from":"umeda","to":"takarazuka","gauge":1067,"double":false}}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("build line status = %d", resp.StatusCode)
	}

	_, le := doRequest(t, http.MethodGet, srv.URL+"/api/games/"+id+"/lines", "")
	var lines []*core.Line
	if err := json.Unmarshal(le.Data, &lines); err != nil || len(lines) != 1 {
		t.Fatalf("lines = %v err=%v", le.Data, err)
	}
	lid := lines[0].ID

	gets := []string{
		"/health",
		"/api/games",
		"/api/games/" + id,
		"/api/games/" + id + "/lines",
		"/api/games/" + id + "/lines/" + lid,
		"/api/games/" + id + "/companies",
		"/api/games/" + id + "/cities",
		"/api/games/" + id + "/events",
		"/api/games/" + id + "/can-build?from=umeda&to=takarazuka&gauge=1067",
	}
	for _, path := range gets {
		resp := rawRequest(t, http.MethodGet, srv.URL+path, "")
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s status = %d, want 200", path, resp.StatusCode)
		}
	}

	// 欠番の路線は 404。
	respMissing := rawRequest(t, http.MethodGet, srv.URL+"/api/games/"+id+"/lines/nope", "")
	respMissing.Body.Close()
	if respMissing.StatusCode != http.StatusNotFound {
		t.Errorf("missing line status = %d, want 404", respMissing.StatusCode)
	}

	// POST turn/end
	respTurn := rawRequest(t, http.MethodPost, srv.URL+"/api/games/"+id+"/turn/end", "")
	respTurn.Body.Close()
	if respTurn.StatusCode != http.StatusOK {
		t.Errorf("turn end status = %d, want 200", respTurn.StatusCode)
	}

	// DELETE
	respDel := rawRequest(t, http.MethodDelete, srv.URL+"/api/games/"+id, "")
	respDel.Body.Close()
	if respDel.StatusCode != http.StatusOK {
		t.Errorf("delete status = %d, want 200", respDel.StatusCode)
	}

	// メソッド不一致は 405。
	resp405 := rawRequest(t, http.MethodDelete, srv.URL+"/api/games", "")
	resp405.Body.Close()
	if resp405.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("DELETE /api/games status = %d, want 405", resp405.StatusCode)
	}
}
