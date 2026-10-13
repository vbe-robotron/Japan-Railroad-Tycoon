package core

import (
	"errors"
	"fmt"
)

// ValidationError は入力検証やゲームルール違反を表す共通エラー型。
type ValidationError struct {
	Code string // 機械可読なエラーコード
	Msg  string // 人間可読なメッセージ（日本語）
	Err  error  // ラップする元エラー（任意）
}

// Error はエラーメッセージを返す。
func (e *ValidationError) Error() string {
	if e == nil {
		return "<nil>"
	}
	switch {
	case e.Code != "" && e.Err != nil:
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Msg, e.Err)
	case e.Code != "":
		return fmt.Sprintf("%s: %s", e.Code, e.Msg)
	case e.Err != nil:
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	default:
		return e.Msg
	}
}

// Unwrap はラップした元エラーを返す。
func (e *ValidationError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// ゲーム全体で共有するセンチネルエラー。
var (
	ErrNoPlayer          = errors.New("プレイヤー会社がありません")
	ErrInsufficientFunds = errors.New("資金が不足しています")
	ErrLineExists        = errors.New("既に路線があります")
	ErrCityNotFound      = errors.New("都市が見つかりません")
	ErrUnauthorized      = errors.New("権限がありません")
	ErrRestricted        = errors.New("現時点では実行できません")
	ErrNotFound          = errors.New("見つかりません")
	ErrInvalidArgument   = errors.New("不正な引数です")
)

// NewValidationError は ValidationError を生成する。
func NewValidationError(code, msg string, err error) *ValidationError {
	return &ValidationError{Code: code, Msg: msg, Err: err}
}

// IsValidationError は err が *ValidationError かどうかを返す。
func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}
