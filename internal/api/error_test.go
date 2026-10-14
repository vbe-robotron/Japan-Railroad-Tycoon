package api

import (
	"context"
	"errors"
	"testing"
)

func TestErrorDecode(t *testing.T) {
	c, _, _, _ := newClientServer(t)
	ctx := context.Background()
	g, _, err := c.CreateGame(ctx, CreateGameRequest{CompanyID: "hankyu"})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}

	_, err = g.BuildLine(ctx, "nope", "nope2", 1067, false)
	if err == nil {
		t.Fatal("err = nil, want validation error")
	}
	var e *Error
	if !errors.As(err, &e) {
		t.Fatalf("err type = %T, want *api.Error", err)
	}
	if e.Status != 404 {
		t.Errorf("status = %d, want 404", e.Status)
	}
	if e.Code != "FROM_NOT_FOUND" {
		t.Errorf("code = %q, want FROM_NOT_FOUND", e.Code)
	}
	if e.Message == "" || e.Error() == "" {
		t.Errorf("message/error string empty: %+v", e)
	}
}

func TestErrorDecode_Validation(t *testing.T) {
	c, _, _, _ := newClientServer(t)
	ctx := context.Background()
	g, _, err := c.CreateGame(ctx, CreateGameRequest{CompanyID: "hankyu"})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}

	_, err = g.BuildLine(ctx, "x", "x", 1067, false)
	if err == nil {
		t.Fatal("err = nil, want SAME_CITY")
	}
	if !IsValidation(err) {
		t.Errorf("IsValidation(%v) = false, want true", err)
	}
	if IsNotFound(err) {
		t.Error("IsNotFound = true, want false")
	}
}

func TestIsNotFound(t *testing.T) {
	cases := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{errors.New("plain"), false},
		{&Error{Status: 404}, true},
		{&Error{Code: "NOT_FOUND"}, true},
		{&Error{Code: "GAME_NOT_FOUND"}, true},
		{&Error{Status: 400, Code: "SAME_CITY"}, false},
	}
	for _, tc := range cases {
		if got := IsNotFound(tc.err); got != tc.want {
			t.Errorf("IsNotFound(%v) = %v, want %v", tc.err, got, tc.want)
		}
	}
}

func TestIsValidation(t *testing.T) {
	cases := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{errors.New("plain"), false},
		{&Error{Status: 400, Code: "SAME_CITY"}, true},
		{&Error{Status: 403, Code: "NOT_OWNED"}, true},
		{&Error{Status: 404, Code: "NOT_FOUND"}, false},
		{&Error{Status: 500, Code: "INTERNAL"}, false},
	}
	for _, tc := range cases {
		if got := IsValidation(tc.err); got != tc.want {
			t.Errorf("IsValidation(%v) = %v, want %v", tc.err, got, tc.want)
		}
	}
}

func TestErrorString(t *testing.T) {
	if got := (&Error{Code: "C", Message: "m"}).Error(); got != "C: m" {
		t.Errorf("Error() = %q", got)
	}
	if got := (&Error{Code: "C", Message: "m", Hint: "h"}).Error(); got != "C: m (h)" {
		t.Errorf("Error() = %q", got)
	}
	var nilErr *Error
	if got := nilErr.Error(); got != "<nil>" {
		t.Errorf("nil Error() = %q", got)
	}
}

func TestErrorFromResponse(t *testing.T) {
	e := errorFromResponse(500, nil)
	if e.Status != 500 || e.Code != "HTTP_ERROR" {
		t.Errorf("e = %+v", e)
	}
	e2 := errorFromResponse(400, []Error{{Code: "X", Message: "m"}})
	if e2.Status != 400 || e2.Code != "X" || e2.Message != "m" {
		t.Errorf("e2 = %+v", e2)
	}
}

func TestError_ConnectionAndTimeoutSentinels(t *testing.T) {
	if !errors.Is(ErrConnection, ErrConnection) {
		t.Error("ErrConnection identity failed")
	}
	if ErrTimeout.Code != "TIMEOUT" || ErrInvalid.Code != "INVALID_RESPONSE" {
		t.Error("sentinel codes wrong")
	}
}
