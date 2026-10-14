package core

import "testing"

// newFinanceCompany は財務 AI テスト用の会社を返す。
func newFinanceCompany(cash, debt, credit int64) *Company {
	c := NewCompany("hankyu", "hankyu", "阪急")
	c.Cash = cash
	c.Debt = debt
	c.CreditLine = credit
	return c
}

func TestAIFinance_Borrow(t *testing.T) {
	g := NewGame(1, "p")
	c := newFinanceCompany(1_000_000, 0, 10_000_000)
	ai := &AI{CompanyID: "hankyu"}

	evs := g.aiFinance(c, ai)

	if len(evs) != 1 {
		t.Fatalf("events = %d, want 1", len(evs))
	}
	if evs[0].Type != EventBondsIssued {
		t.Errorf("type = %q, want %q", evs[0].Type, EventBondsIssued)
	}
	if c.Debt != 5_000_000 || c.Cash != 6_000_000 {
		t.Errorf("Debt/Cash = %d/%d, want 5000000/6000000", c.Debt, c.Cash)
	}
}

func TestAIFinance_Borrow_NegativeCash(t *testing.T) {
	g := NewGame(1, "p")
	c := newFinanceCompany(-1_000_000, 0, 10_000_000)
	ai := &AI{CompanyID: "hankyu"}

	evs := g.aiFinance(c, ai)

	if len(evs) != 1 {
		t.Fatalf("events = %d, want 1", len(evs))
	}
	// amount = 1,000,000 + 3,000,000 = 4,000,000
	if c.Debt != 4_000_000 || c.Cash != 3_000_000 {
		t.Errorf("Debt/Cash = %d/%d, want 4000000/3000000", c.Debt, c.Cash)
	}
}

func TestAIFinance_Borrow_Capped(t *testing.T) {
	g := NewGame(1, "p")
	c := newFinanceCompany(1_000_000, 8_000_000, 10_000_000)
	ai := &AI{CompanyID: "hankyu"}

	evs := g.aiFinance(c, ai)

	if len(evs) != 1 {
		t.Fatalf("events = %d, want 1", len(evs))
	}
	// 5,000,000 を借りると上限超過 → 2,000,000 に制限
	if c.Debt != 10_000_000 || c.Cash != 3_000_000 {
		t.Errorf("Debt/Cash = %d/%d, want 10000000/3000000", c.Debt, c.Cash)
	}
}

func TestAIFinance_NoCredit(t *testing.T) {
	g := NewGame(1, "p")
	c := newFinanceCompany(0, 0, 0)
	ai := &AI{CompanyID: "hankyu"}

	if evs := g.aiFinance(c, ai); evs != nil {
		t.Errorf("信用枠なしで借入が発生: %d", len(evs))
	}
}

func TestAIFinance_Repay(t *testing.T) {
	g := NewGame(1, "p")
	c := newFinanceCompany(40_000_000, 10_000_000, 100_000_000)
	ai := &AI{CompanyID: "hankyu"}

	evs := g.aiFinance(c, ai)

	if len(evs) != 1 {
		t.Fatalf("events = %d, want 1", len(evs))
	}
	if evs[0].Type != EventDebtRepaid {
		t.Errorf("type = %q, want %q", evs[0].Type, EventDebtRepaid)
	}
	// repay = min(debt/2=5M, cash-10M=30M) = 5M
	if c.Debt != 5_000_000 || c.Cash != 35_000_000 {
		t.Errorf("Debt/Cash = %d/%d, want 5000000/35000000", c.Debt, c.Cash)
	}
}

func TestAIFinance_Repay_Capped(t *testing.T) {
	g := NewGame(1, "p")
	c := newFinanceCompany(31_000_000, 100_000_000, 100_000_000)
	ai := &AI{CompanyID: "hankyu"}

	evs := g.aiFinance(c, ai)

	if len(evs) != 1 {
		t.Fatalf("events = %d, want 1", len(evs))
	}
	// repay = min(50M, 21M) = 21M
	if c.Debt != 79_000_000 || c.Cash != 10_000_000 {
		t.Errorf("Debt/Cash = %d/%d, want 79000000/10000000", c.Debt, c.Cash)
	}
}

func TestAIFinance_NoRepay(t *testing.T) {
	g := NewGame(1, "p")
	ai := &AI{CompanyID: "hankyu"}

	// 借金なし
	c := newFinanceCompany(40_000_000, 0, 0)
	if evs := g.aiFinance(c, ai); evs != nil {
		t.Errorf("借金なしで返済イベント: %d", len(evs))
	}

	// 借金 1 円（返済額 0）
	c2 := newFinanceCompany(40_000_000, 1, 0)
	if evs := g.aiFinance(c2, ai); evs != nil {
		t.Errorf("返済額 0 でイベント: %d", len(evs))
	}
}
