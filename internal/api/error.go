package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Error は API 呼び出しのエラーを表す。
//
// サーバーが返す {code, message, hint} をそのまま保持し、HTTP ステータスを
// Status に付与する。Status は JSON には含めない（クライアント専用）。
type Error struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
}

// Error はエラーメッセージを返す。
func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Hint != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Hint)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// クライアント側で発生するセンチネルエラー。
var (
	// ErrConnection はサーバーへ接続できなかったことを表す。
	ErrConnection = &Error{Code: "CONNECTION_FAILED", Message: "サーバーに接続できません"}
	// ErrTimeout はサーバーが応答しなかったことを表す。
	ErrTimeout = &Error{Code: "TIMEOUT", Message: "サーバーが応答しません"}
	// ErrInvalid は応答が JSON として不正だったことを表す。
	ErrInvalid = &Error{Code: "INVALID_RESPONSE", Message: "応答が不正です"}
)

// IsNotFound は err が 404（NOT_FOUND）系の *Error かどうかを返す。
func IsNotFound(err error) bool {
	var e *Error
	if !errors.As(err, &e) {
		return false
	}
	if e.Status == http.StatusNotFound {
		return true
	}
	return e.Code == "NOT_FOUND" || strings.HasSuffix(e.Code, "_NOT_FOUND")
}

// IsValidation は err が 404 以外のクライアントエラー（4xx）かどうかを返す。
func IsValidation(err error) bool {
	var e *Error
	if !errors.As(err, &e) {
		return false
	}
	if IsNotFound(e) {
		return false
	}
	return e.Status >= 400 && e.Status < 500
}

// errorFromResponse は HTTP ステータスとエラー一覧から *Error を生成する。
func errorFromResponse(status int, errs []Error) *Error {
	if len(errs) > 0 {
		e := errs[0]
		e.Status = status
		return &e
	}
	return &Error{Status: status, Code: "HTTP_ERROR", Message: http.StatusText(status)}
}
