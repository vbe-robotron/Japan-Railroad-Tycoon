package core

import (
	"errors"
	"testing"
)

func TestEndTurnAction_Type(t *testing.T) {
	a := &EndTurnAction{}
	if got := a.Type(); got != "end_turn" {
		t.Errorf("Type() = %q, want %q", got, "end_turn")
	}
}

func TestEndTurnAction_Validate_NoPlayer(t *testing.T) {
	g := NewGame(1907, "hankyu") // 会社を追加しない
	a := &EndTurnAction{}

	err := a.Validate(g)
	if err == nil {
		t.Fatal("プレイヤー会社なしでエラーにならなかった")
	}
	if !IsValidationError(err) {
		t.Errorf("ValidationError ではない: %v", err)
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("型が *ValidationError でない: %T", err)
	}
	if ve.Code != "NO_PLAYER" {
		t.Errorf("Code = %q, want %q", ve.Code, "NO_PLAYER")
	}
	if !errors.Is(ve, ErrNoPlayer) {
		t.Errorf("ErrNoPlayer をラップしていない: %v", err)
	}
}

func TestEndTurnAction_Validate_NilGame(t *testing.T) {
	a := &EndTurnAction{}
	err := a.Validate(nil)
	if err == nil {
		t.Fatal("nil ゲームでエラーにならなかった")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("型が *ValidationError でない: %T", err)
	}
	if ve.Code != "GAME_NIL" {
		t.Errorf("Code = %q, want %q", ve.Code, "GAME_NIL")
	}
}

func TestEndTurnAction_Validate_OK(t *testing.T) {
	g := NewGame(1907, "hankyu")
	g.AddCompany(NewCompany("hankyu", "hankyu", "阪急"))
	a := &EndTurnAction{}

	if err := a.Validate(g); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
	// Validate は状態を変更しない
	if g.Turn != 0 {
		t.Errorf("Validate が Turn を変更した: %d", g.Turn)
	}
	if !g.Date.IsZero() {
		t.Errorf("Validate が Date を変更した: %v", g.Date)
	}
}

func TestEndTurnAction_Apply(t *testing.T) {
	g := NewGame(1907, "hankyu")
	g.Date = Date{Year: 1910, Month: 3}
	g.AddCompany(NewCompany("hankyu", "hankyu", "阪急"))
	a := &EndTurnAction{}

	events, err := a.Apply(g)
	if err != nil {
		t.Fatalf("Apply() = %v, want nil", err)
	}
	if len(events) == 0 {
		t.Fatal("Apply() のイベントが空")
	}
	// 1ターン進む
	if g.Turn != 1 {
		t.Errorf("Turn = %d, want 1", g.Turn)
	}
	if !g.Date.Equal(Date{Year: 1910, Month: 4}) {
		t.Errorf("Date = %v, want 1910年4月", g.Date)
	}
}

func TestEndTurnAction_Apply_NoPlayer(t *testing.T) {
	g := NewGame(1907, "hankyu") // 会社なし
	a := &EndTurnAction{}

	events, err := a.Apply(g)
	if err == nil {
		t.Fatal("プレイヤー会社なしでエラーにならなかった")
	}
	if events != nil {
		t.Errorf("events = %v, want nil", events)
	}
	// 状態が変更されていない
	if g.Turn != 0 {
		t.Errorf("Turn = %d, want 0", g.Turn)
	}
}

func TestEndTurnAction_Apply_ViaRegistry(t *testing.T) {
	g := NewGame(1907, "hankyu")
	g.AddCompany(NewCompany("hankyu", "hankyu", "阪急"))

	a, err := DecodeAction("end_turn", []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	events, err := a.Apply(g)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 {
		t.Error("no events")
	}
	if g.Turn != 1 {
		t.Errorf("Turn = %d, want 1", g.Turn)
	}
}
