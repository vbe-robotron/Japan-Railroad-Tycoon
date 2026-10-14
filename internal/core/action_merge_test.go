package core

import (
	"testing"
)

func TestMergeAction_Type(t *testing.T) {
	a := &MergeAction{}
	if got := a.Type(); got != "merge" {
		t.Errorf("Type() = %q, want %q", got, "merge")
	}
}

func TestMergeAction_Validate_NilGame(t *testing.T) {
	a := &MergeAction{TargetCompanyID: "hanshin", Ratio: 0.5, Name: "新阪急"}
	err := a.Validate(nil)
	if err == nil {
		t.Fatal("nil ゲームでエラーにならなかった")
	}
}

func TestMergeAction_Validate_NoPlayer(t *testing.T) {
	g := NewGame(1907, "hankyu")
	a := &MergeAction{TargetCompanyID: "hanshin", Ratio: 0.5, Name: "新阪急"}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("プレイヤーなしでエラーにならなかった")
	}
}

func TestMergeAction_Validate_TargetNotFound(t *testing.T) {
	g := setupMATestGame()
	a := &MergeAction{TargetCompanyID: "not_exist", Ratio: 0.5, Name: "新"}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("対象なしでエラーにならなかった")
	}
}

func TestMergeAction_Validate_SelfMerge(t *testing.T) {
	g := setupMATestGame()
	a := &MergeAction{TargetCompanyID: "hankyu", Ratio: 0.5, Name: "新"}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("自社合併でエラーにならなかった")
	}
}

func TestMergeAction_Validate_InvalidRatio(t *testing.T) {
	g := setupMATestGame()
	tests := []float64{-0.1, 1.5}
	for _, r := range tests {
		a := &MergeAction{TargetCompanyID: "hanshin", Ratio: r, Name: "新"}
		err := a.Validate(g)
		if err == nil {
			t.Errorf("Ratio=%.2f でエラーにならなかった", r)
		}
	}
}

func TestMergeAction_Validate_EmptyName(t *testing.T) {
	g := setupMATestGame()
	a := &MergeAction{TargetCompanyID: "hanshin", Ratio: 0.5, Name: ""}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("空の社名でエラーにならなかった")
	}
}

func TestMergeAction_Validate_InsufficientFunds(t *testing.T) {
	g := setupMATestGame()
	g.PlayerCompany().Cash = 1000
	a := &MergeAction{TargetCompanyID: "hanshin", Ratio: 0.5, Name: "新"}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("資金不足でエラーにならなかった")
	}
}

func TestMergeAction_Validate_OK(t *testing.T) {
	g := setupMATestGame()
	a := &MergeAction{TargetCompanyID: "hanshin", Ratio: 0.5, Name: "新阪急"}
	if err := a.Validate(g); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}

func TestMergeAction_Apply(t *testing.T) {
	g := setupMATestGame()
	c := g.PlayerCompany()
	target := g.Companies["hanshin"]
	beforeCash := c.Cash
	targetLines := len(target.Lines)

	a := &MergeAction{TargetCompanyID: "hanshin", Ratio: 0.5, Name: "新阪急"}
	events, err := a.Apply(g)
	if err != nil {
		t.Fatal(err)
	}

	// 資金減算
	if c.Cash >= beforeCash {
		t.Errorf("資金が減っていない")
	}
	// 社名変更
	if c.Name != "新阪急" {
		t.Errorf("Name = %q, want 新阪急", c.Name)
	}
	// 対象会社解散
	if target.Dissolved == nil {
		t.Error("対象会社が解散していない")
	}
	// 路線引継ぎ
	if len(c.Lines) < targetLines {
		t.Errorf("Lines 数 = %d, want >= %d", len(c.Lines), targetLines)
	}
	// イベント
	types := map[string]bool{}
	for _, e := range events {
		types[e.Type] = true
	}
	if !types[EventMergeCompleted] {
		t.Error("merge_completed がない")
	}
}

func TestMergeAction_Apply_ViaRegistry(t *testing.T) {
	g := setupMATestGame()
	payload := []byte(`{"target_company_id":"hanshin","ratio":0.5,"name":"新阪急"}`)
	a, err := DecodeAction("merge", payload)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Apply(g); err != nil {
		t.Fatal(err)
	}
	if g.PlayerCompany().Name != "新阪急" {
		t.Error("社名が変わっていない")
	}
}

func TestMergeAction_Deterministic(t *testing.T) {
	g1 := setupMATestGame()
	g2 := setupMATestGame()
	a1 := &MergeAction{TargetCompanyID: "hanshin", Ratio: 0.5, Name: "新阪急"}
	a2 := &MergeAction{TargetCompanyID: "hanshin", Ratio: 0.5, Name: "新阪急"}
	e1, _ := a1.Apply(g1)
	e2, _ := a2.Apply(g2)
	if e1[0].Message != e2[0].Message {
		t.Errorf("Message が異なる")
	}
}

func setupMATestGame() *Game {
	g := NewGame(1907, "hankyu")
	g.Date = Date{Year: 1910, Month: 1}

	hankyu := NewCompany("hankyu", "hankyu", "阪急")
	hankyu.Cash = 200_000_000
	hankyu.TotalShares = 10000
	hankyu.StockPrice = 1000
	hankyu.Reputation = 50
	g.AddCompany(hankyu)

	hanshin := NewCompany("hanshin", "hanshin", "阪神")
	hanshin.Cash = 50_000_000
	hanshin.TotalShares = 10000
	hanshin.StockPrice = 1000
	hanshin.Reputation = 50
	g.AddCompany(hanshin)

	return g
}
