package core

import (
	"errors"
	"testing"
)

func TestTenderOfferAction_Type(t *testing.T) {
	a := &TenderOfferAction{}
	if got := a.Type(); got != "tender_offer" {
		t.Errorf("Type() = %q, want %q", got, "tender_offer")
	}
}

func TestTenderOfferAction_Validate_NilGame(t *testing.T) {
	a := &TenderOfferAction{TargetCompanyID: "hanshin", Premium: 1.1, TargetRatio: 0.51}
	err := a.Validate(nil)
	if err == nil {
		t.Fatal("nil ゲームでエラーにならなかった")
	}
	ve, _ := err.(*ValidationError)
	if ve.Code != "GAME_NIL" {
		t.Errorf("GAME_NIL でない")
	}
}

func TestTenderOfferAction_Validate_NoPlayer(t *testing.T) {
	g := NewGame(1907, "hankyu")
	a := &TenderOfferAction{TargetCompanyID: "hanshin", Premium: 1.1, TargetRatio: 0.51}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("プレイヤーなしでエラーにならなかった")
	}
	ve, _ := err.(*ValidationError)
	if ve.Code != "NO_PLAYER" {
		t.Errorf("NO_PLAYER でない")
	}
}

func TestTenderOfferAction_Validate_TargetNotFound(t *testing.T) {
	g := setupMATestGame()
	a := &TenderOfferAction{TargetCompanyID: "not_exist", Premium: 1.1, TargetRatio: 0.51}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("対象なしでエラーにならなかった")
	}
	ve, _ := err.(*ValidationError)
	if ve.Code != "TARGET_NOT_FOUND" {
		t.Errorf("TARGET_NOT_FOUND でない")
	}
	if !errors.Is(ve, ErrNotFound) {
		t.Error("ErrNotFound をラップしていない")
	}
}

func TestTenderOfferAction_Validate_SelfTarget(t *testing.T) {
	g := setupMATestGame()
	a := &TenderOfferAction{TargetCompanyID: "hankyu", Premium: 1.1, TargetRatio: 0.51}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("自社TOBでエラーにならなかった")
	}
	ve, _ := err.(*ValidationError)
	if ve.Code != "SELF_TARGET" {
		t.Errorf("SELF_TARGET でない")
	}
}

func TestTenderOfferAction_Validate_InvalidPremium(t *testing.T) {
	g := setupMATestGame()
	tests := []float64{0.5, 2.0, 0}
	for _, p := range tests {
		a := &TenderOfferAction{TargetCompanyID: "hanshin", Premium: p, TargetRatio: 0.51}
		err := a.Validate(g)
		if err == nil {
			t.Errorf("Premium=%.2f でエラーにならなかった", p)
		}
		ve, _ := err.(*ValidationError)
		if ve.Code != "INVALID_PREMIUM" {
			t.Errorf("INVALID_PREMIUM でない: %q", ve.Code)
		}
	}
}

func TestTenderOfferAction_Validate_InvalidTargetRatio(t *testing.T) {
	g := setupMATestGame()
	tests := []float64{0.1, 0.3, 0.9}
	for _, r := range tests {
		a := &TenderOfferAction{TargetCompanyID: "hanshin", Premium: 1.1, TargetRatio: r}
		err := a.Validate(g)
		if err == nil {
			t.Errorf("TargetRatio=%.2f でエラーにならなかった", r)
		}
	}
}

func TestTenderOfferAction_Validate_InsufficientFunds(t *testing.T) {
	g := setupMATestGame()
	g.PlayerCompany().Cash = 1000
	a := &TenderOfferAction{TargetCompanyID: "hanshin", Premium: 1.5, TargetRatio: 0.51}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("資金不足でエラーにならなかった")
	}
	ve, _ := err.(*ValidationError)
	if ve.Code != "INSUFFICIENT_FUNDS" {
		t.Errorf("INSUFFICIENT_FUNDS でない")
	}
}

func TestTenderOfferAction_Validate_OK(t *testing.T) {
	g := setupMATestGame()
	a := &TenderOfferAction{TargetCompanyID: "hanshin", Premium: 1.1, TargetRatio: 0.51}
	if err := a.Validate(g); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}

func TestTenderOfferAction_Apply(t *testing.T) {
	g := setupMATestGame()
	c := g.PlayerCompany()
	target := g.Companies["hanshin"]
	beforeCash := c.Cash

	a := &TenderOfferAction{TargetCompanyID: "hanshin", Premium: 1.1, TargetRatio: 0.51}
	events, err := a.Apply(g)
	if err != nil {
		t.Fatal(err)
	}

	// 資金減算
	if c.Cash >= beforeCash {
		t.Errorf("資金が減っていない")
	}
	// 保有株が 51% 以上
	holdingPct := float64(c.Holding[target.ID]) / float64(target.TotalShares)
	if holdingPct < 0.51 {
		t.Errorf("holdingPct = %.2f, want >= 0.51", holdingPct)
	}
	// 子会社化
	if target.ParentID != c.ID {
		t.Errorf("ParentID = %q, want %q", target.ParentID, c.ID)
	}
	// イベント
	types := map[string]bool{}
	for _, e := range events {
		types[e.Type] = true
	}
	if !types[EventTOBLaunched] {
		t.Error("tob_launched がない")
	}
	if !types[EventTOBSuccess] {
		t.Error("tob_success がない")
	}
}

func TestTenderOfferAction_Apply_ViaRegistry(t *testing.T) {
	g := setupMATestGame()
	payload := []byte(`{"target_company_id":"hanshin","premium":1.1,"target_ratio":0.51}`)
	a, err := DecodeAction("tender_offer", payload)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Apply(g); err != nil {
		t.Fatal(err)
	}
	if g.Companies["hanshin"].ParentID != "hankyu" {
		t.Error("子会社化されていない")
	}
}

func TestTenderOfferAction_Deterministic(t *testing.T) {
	g1 := setupMATestGame()
	g2 := setupMATestGame()
	a1 := &TenderOfferAction{TargetCompanyID: "hanshin", Premium: 1.1, TargetRatio: 0.51}
	a2 := &TenderOfferAction{TargetCompanyID: "hanshin", Premium: 1.1, TargetRatio: 0.51}
	e1, _ := a1.Apply(g1)
	e2, _ := a2.Apply(g2)
	if len(e1) != len(e2) {
		t.Fatalf("events 数が異なる")
	}
	for i := range e1 {
		if e1[i].Message != e2[i].Message {
			t.Errorf("Message が異なる")
		}
	}
}
