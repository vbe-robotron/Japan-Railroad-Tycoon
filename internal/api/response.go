package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/you/jrt/internal/core"
)

// Response は API の統一レスポンス形式。
//
// 成功時は OK=true と Data / Events、失敗時は OK=false と Errors を返す。
type Response struct {
	OK     bool         `json:"ok"`
	Data   any          `json:"data,omitempty"`
	Events []core.Event `json:"events,omitempty"`
	Errors []APIError   `json:"errors,omitempty"`
}

// APIError は API のエラー詳細。
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
}

// writeJSON は Response を JSON として書き出す。
func writeJSON(w http.ResponseWriter, status int, r Response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(r)
}

// writeOK は成功レスポンス（200）を書き出す。
func writeOK(w http.ResponseWriter, data any, events []core.Event) {
	writeJSON(w, http.StatusOK, Response{OK: true, Data: data, Events: events})
}

// writeError はエラーレスポンスを書き出す。
func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, Response{
		OK:     false,
		Errors: []APIError{{Code: code, Message: msg}},
	})
}

// writeValidationError は core の検証エラーを HTTP ステータスに変換する。
//
// *core.ValidationError 以外のエラーは、core.ErrNotFound なら 404、
// それ以外は 500 として扱う。
func writeValidationError(w http.ResponseWriter, err error) {
	if err == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "不明なエラーです")
		return
	}

	var ve *core.ValidationError
	if !errors.As(err, &ve) {
		if errors.Is(err, core.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}

	writeError(w, statusForCode(ve.Code), ve.Code, ve.Msg)
}

// statusForCode は ValidationError.Code を HTTP ステータスに変換する。
//
//   - GAME_NIL / NO_PLAYER         → 400
//   - NOT_FOUND / *_NOT_FOUND      → 404
//   - UNAUTHORIZED 系 / 所有権エラー → 403
//   - その他                        → 400
func statusForCode(code string) int {
	switch {
	case code == "GAME_NIL" || code == "NO_PLAYER":
		return http.StatusBadRequest
	case code == "NOT_FOUND" || strings.HasSuffix(code, "_NOT_FOUND"):
		return http.StatusNotFound
	case strings.Contains(code, "UNAUTHORIZED") ||
		code == "NOT_OWNED" || code == "TARGET_NATIONAL" ||
		code == "OWNED_BY_OTHER":
		return http.StatusForbidden
	default:
		return http.StatusBadRequest
	}
}
