package core

import (
	"errors"
	"testing"
)

func TestBuySharesAction_Type(t *testing.T) {
	a := &BuySharesAction{}
	if got := a.Type(); got != "buy_shares" {
		t.Errorf("Type() = %q, want %q", got, "buy_shares")
	}
}

func TestBuySharesAction_Validate_NilGame(t *testing.T) {
	a := &BuySharesAction{TargetCompanyID: "hanshin", Count: 100}
	err := a.Validate(nil)
	if err == nil {
		t.Fatal("nil ゲームでエラーにならなかった")
	}
	ve, _ := err.(*ValidationError)
	if ve.Code != "GAME_NIL" {
		t.Errorf("GAME_NIL でない: %q", ve.Code)
	}
}

func TestBuySharesAction_Validate_NoPlayer(t *testing.T) {
	g := NewGame(1907, "hankyu")
	a := &BuySharesAction{TargetCompanyID: "hanshin", Count: 100}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("プレイヤーなしでエラーにならなかった")
	}
	ve, _ := err.(*ValidationError)
	if ve.Code != "NO_PLAYER" {
		t.Errorf("NO_PLAYER でない: %q", ve.Code)
	}
}

func TestBuySharesAction_Validate_TargetNotFound(t *testing.T) {
	g := setupBuySharesTestGame()
	a := &BuySharesAction{TargetCompanyID: "not_exist", Count: 100}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("対象なしでエラーにならなかった")
	}
	ve, _ := err.(*ValidationError)
	if ve.Code != "TARGET_NOT_FOUND" {
		t.Errorf("TARGET_NOT_FOUND でない: %q", ve.Code)
	}
	if !errors.Is(ve, ErrNotFound) {
		t.Error("ErrNotFound をラップしていない")
	}
}

func TestBuySharesAction_Validate_SelfPurchase(t *testing.T) {
	g := setupBuySharesTestGame()
	a := &BuySharesAction{TargetCompanyID: "hankyu", Count: 100}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("自社株取得でエラーにならなかった")
	}
	ve, _ := err.(*ValidationError)
	if ve.Code != "SELF_PURCHASE" {
		t.Errorf("SELF_PURCHASE でない: %q", ve.Code)
	}
}

func TestBuySharesAction_Validate_InvalidCount(t *testing.T) {
	g := setupBuySharesTestGame()
	a := &BuySharesAction{TargetCompanyID: "hanshin", Count: 0}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("Count 0 でエラーにならなかった")
	}
	ve, _ := err.(*ValidationError)
	if ve.Code != "INVALID_COUNT" {
		t.Errorf("INVALID_COUNT でない: %q", ve.Code)
	}
}

func TestBuySharesAction_Validate_InsufficientFunds(t *testing.T) {
	g := setupBuySharesTestGame()
	g.PlayerCompany().Cash = 1000
	a := &BuySharesAction{TargetCompanyID: "hanshin", Count: 1_000_000}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("資金不足でエラーにならなかった")
	}
	ve, _ := err.(*ValidationError)
	if ve.Code != "INSUFFICIENT_FUNDS" {
		t.Errorf("INSUFFICIENT_FUNDS でない: %q", ve.Code)
	}
	if !errors.Is(ve, ErrInsufficientFunds) {
		t.Error("ErrInsufficientFunds をラップしていない")
	}
}

func TestBuySharesAction_Validate_OK(t *testing.T) {
	g := setupBuySharesTestGame()
	a := &BuySharesAction{TargetCompanyID: "hanshin", Count: 100}
	if err := a.Validate(g); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}

func TestBuySharesAction_Apply(t *testing.T) {
	g := setupBuySharesTestGame()
	c := g.PlayerCompany()
	target := g.Companies["hanshin"]
	beforeCash := c.Cash
	cost := int64(100 * target.StockPrice * 105 / 100)

	a := &BuySharesAction{TargetCompanyID: "hanshin", Count: 100}
	events, err := a.Apply(g)
	if err != nil {
		t.Fatal(err)
	}

	// 資金減算
	if c.Cash != beforeCash-cost {
		t.Errorf("Cash = %d, want %d", c.Cash, beforeCash-cost)
	}
	// 保有株
	if c.Holding["hanshin"] != 100 {
		t.Errorf("Holding = %d, want 100", c.Holding["hanshin"])
	}
	// イベント
	found := false
	for _, e := range events {
		if e.Type == EventSharesAcquired {
			found = true
		}
	}
	if !found {
		t.Error("shares_acquired イベントがない")
	}
}

func TestBuySharesAction_Apply_Subsidiary(t *testing.T) {
	g := setupBuySharesTestGame()
	c := g.PlayerCompany()
	target := g.Companies["hanshin"]
	// 50%超を取得
	count := target.TotalShares/2 + 1

	a := &BuySharesAction{TargetCompanyID: "hanshin", Count: count}
	events, err := a.Apply(g)
	if err != nil {
		t.Fatal(err)
	}

	// 子会社化
	if target.ParentID != c.ID {
		t.Errorf("ParentID = %q, want %q", target.ParentID, c.ID)
	}

	found := false
	for _, e := range events {
		if e.Type == EventSubsidiaryEstablished {
			found = true
		}
	}
	if !found {
		t.Error("subsidiary_established イベントがない")
	}
}

func TestBuySharesAction_Apply_ViaRegistry(t *testing.T) {
	g := setupBuySharesTestGame()
	payload := []byte(`{"target_company_id":"hanshin","count":100}`)
	a, err := DecodeAction("buy_shares", payload)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Apply(g); err != nil {
		t.Fatal(err)
	}
	if g.PlayerCompany().Holding["hanshin"] != 100 {
		t.Errorf("Holding = %d, want 100", g.PlayerCompany().Holding["hanshin"])
	}
}

func TestBuySharesAction_Deterministic(t *testing.T) {
	g1 := setupBuySharesTestGame()
	g2 := setupBuySharesTestGame()
	a1 := &BuySharesAction{TargetCompanyID: "hanshin", Count: 100}
	a2 := &BuySharesAction{TargetCompanyID: "hanshin", Count: 100}
	e1, _ := a1.Apply(g1)
	e2, _ := a2.Apply(g2)
	if e1[0].Message != e2[0].Message {
		t.Errorf("Message が異なる")
	}
}

func setupBuySharesTestGame() *Game {
	g := NewGame(1907, "hankyu")
	g.Date = Date{Year: 1910, Month: 1}

	hankyu := NewCompany("hankyu", "hankyu", "阪急")
	hankyu.Cash = 100_000_000
	hankyu.TotalShares = 10000
	hankyu.StockPrice = 1000
	g.AddCompany(hankyu)

	hanshin := NewCompany("hanshin", "hanshin", "阪神")
	hanshin.Cash = 50_000_000
	hanshin.TotalShares = 10000
	hanshin.StockPrice = 1000
	g.AddCompany(hanshin)

	return g
}
