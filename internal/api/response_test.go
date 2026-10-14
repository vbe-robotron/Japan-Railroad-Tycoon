package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/you/jrt/internal/core"
)

func TestStatusForCode(t *testing.T) {
	cases := map[string]int{
		"GAME_NIL":        http.StatusBadRequest,
		"NO_PLAYER":       http.StatusBadRequest,
		"SAME_CITY":       http.StatusBadRequest,
		"FROM_NOT_FOUND":  http.StatusNotFound,
		"NOT_FOUND":       http.StatusNotFound,
		"LINE_NOT_FOUND":  http.StatusNotFound,
		"UNAUTHORIZED":    http.StatusForbidden,
		"NOT_OWNED":       http.StatusForbidden,
		"TARGET_NATIONAL": http.StatusForbidden,
		"OWNED_BY_OTHER":  http.StatusForbidden,
	}
	for code, want := range cases {
		if got := statusForCode(code); got != want {
			t.Errorf("statusForCode(%q) = %d, want %d", code, got, want)
		}
	}
}

func decodeRecorder(t *testing.T, rec *httptest.ResponseRecorder) Response {
	t.Helper()
	var resp Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return resp
}

func TestWriteValidationError(t *testing.T) {
	// ValidationError（404）
	rec := httptest.NewRecorder()
	writeValidationError(rec, core.NewValidationError("FROM_NOT_FOUND", "みつかりません", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
	if got := decodeRecorder(t, rec); len(got.Errors) == 0 || got.Errors[0].Code != "FROM_NOT_FOUND" {
		t.Errorf("errors = %+v", got.Errors)
	}

	// core.ErrNotFound（404）
	rec = httptest.NewRecorder()
	writeValidationError(rec, core.ErrNotFound)
	if rec.Code != http.StatusNotFound {
		t.Errorf("ErrNotFound status = %d, want 404", rec.Code)
	}

	// その他のエラー（500）
	rec = httptest.NewRecorder()
	writeValidationError(rec, errors.New("boom"))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("generic err status = %d, want 500", rec.Code)
	}

	// nil（500）
	rec = httptest.NewRecorder()
	writeValidationError(rec, nil)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("nil err status = %d, want 500", rec.Code)
	}
}

func TestWithRecover(t *testing.T) {
	h := withRecover(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}

func TestWithCORS_Preflight(t *testing.T) {
	h := withCORS(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodOptions, "/api/games", nil))
	if rec.Code != http.StatusNoContent {
		t.Errorf("preflight status = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("allow-origin = %q, want *", got)
	}
}
