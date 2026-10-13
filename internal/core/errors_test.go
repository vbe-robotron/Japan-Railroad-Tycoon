package core

import (
	"errors"
	"testing"
)

func TestValidationError_Error(t *testing.T) {
	base := errors.New("元エラー")
	tests := []struct {
		name string
		err  *ValidationError
		want string
	}{
		{"code+msg+err", NewValidationError("E001", "不正です", base), "E001: 不正です: 元エラー"},
		{"code+msg", NewValidationError("E001", "不正です", nil), "E001: 不正です"},
		{"msg+err", NewValidationError("", "不正です", base), "不正です: 元エラー"},
		{"msgのみ", NewValidationError("", "不正です", nil), "不正です"},
		{"nil", (*ValidationError)(nil), "<nil>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidationError_Unwrap(t *testing.T) {
	base := errors.New("元エラー")
	if got := NewValidationError("E", "m", base).Unwrap(); !errors.Is(got, base) {
		t.Errorf("Unwrap() = %v, want %v", got, base)
	}
	if got := NewValidationError("E", "m", nil).Unwrap(); got != nil {
		t.Errorf("Unwrap() = %v, want nil", got)
	}
	if got := (*ValidationError)(nil).Unwrap(); got != nil {
		t.Errorf("nil.Unwrap() = %v, want nil", got)
	}
}

func TestIsValidationError(t *testing.T) {
	base := errors.New("元エラー")
	ve := NewValidationError("E", "m", base)

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"直接", ve, true},
		{"ラップ", errors.Join(ve), true},
		{"fmt風ラップ", &wrapped{ve}, true},
		{"通常エラー", base, false},
		{"nil", nil, false},
		{"センチネル", ErrInsufficientFunds, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidationError(tt.err); got != tt.want {
				t.Errorf("IsValidationError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// wrapped は Unwrap を持つラッパー（errors.As の伝播確認用）。
type wrapped struct{ err error }

func (w *wrapped) Error() string { return w.err.Error() }
func (w *wrapped) Unwrap() error { return w.err }

func TestSentinelErrors(t *testing.T) {
	sentinels := []error{
		ErrNoPlayer, ErrInsufficientFunds, ErrLineExists, ErrCityNotFound,
		ErrUnauthorized, ErrRestricted, ErrNotFound, ErrInvalidArgument,
	}
	for _, s := range sentinels {
		if s == nil {
			t.Fatal("センチネルエラーが nil です")
		}
		if s.Error() == "" {
			t.Errorf("センチネルエラーのメッセージが空です: %v", s)
		}
	}
}
