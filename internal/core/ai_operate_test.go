package core

import "testing"

// newOperateLine は運用 AI テスト用の会社と路線を返す。
func newOperateLine(cash int64, fare int, operatingRatio float64, trains int) (*Company, *Line) {
	c := NewCompany("hankyu", "hankyu", "阪急")
	c.Cash = cash
	line := NewLine("l1", c.ID, "a", "b")
	line.DisplayName = "テスト線"
	line.Fare = fare
	line.OperatingRatio = operatingRatio
	line.TrainsPerDay = trains
	c.AddLine(line)
	return c, line
}

func TestAIAdjustFare_Increase(t *testing.T) {
	g := NewGame(1, "p")
	c, line := newOperateLine(10_000_000, 100, 120, 10)

	evs := g.aiAdjustFare(c, line)

	if len(evs) != 1 {
		t.Fatalf("events = %d, want 1", len(evs))
	}
	if evs[0].Type != EventFareChanged {
		t.Errorf("type = %q, want %q", evs[0].Type, EventFareChanged)
	}
	if line.Fare != 110 {
		t.Errorf("Fare = %d, want 110", line.Fare)
	}
}

func TestAIAdjustFare_IncreaseCap(t *testing.T) {
	g := NewGame(1, "p")
	c, line := newOperateLine(10_000_000, 990, 120, 10)

	evs := g.aiAdjustFare(c, line)

	if len(evs) != 1 {
		t.Fatalf("events = %d, want 1", len(evs))
	}
	if line.Fare != MaxFare {
		t.Errorf("Fare = %d, want %d（上限）", line.Fare, MaxFare)
	}
}

func TestAIAdjustFare_Decrease(t *testing.T) {
	g := NewGame(1, "p")
	c, line := newOperateLine(10_000_000, 100, 50, 10)

	evs := g.aiAdjustFare(c, line)

	if len(evs) != 1 {
		t.Fatalf("events = %d, want 1", len(evs))
	}
	if line.Fare != 95 {
		t.Errorf("Fare = %d, want 95", line.Fare)
	}
}

func TestAIAdjustFare_NoChange(t *testing.T) {
	g := NewGame(1, "p")

	// 営業係数が 70-100 の範囲 → 変更なし
	c, line := newOperateLine(10_000_000, 100, 85, 10)
	if evs := g.aiAdjustFare(c, line); evs != nil {
		t.Errorf("営業係数 85 で変更された: %d", len(evs))
	}

	// 運賃 0 → 変更なし
	c2, l2 := newOperateLine(10_000_000, 0, 120, 10)
	if evs := g.aiAdjustFare(c2, l2); evs != nil {
		t.Errorf("運賃 0 で変更された")
	}

	// 運賃が上限 → 変更なし
	c3, l3 := newOperateLine(10_000_000, MaxFare, 120, 10)
	if evs := g.aiAdjustFare(c3, l3); evs != nil {
		t.Errorf("運賃上限で変更された")
	}

	// 変更量が 0 になる小運賃 → 変更なし
	c4, l4 := newOperateLine(10_000_000, 5, 120, 10)
	if evs := g.aiAdjustFare(c4, l4); evs != nil {
		t.Errorf("小運賃（変更量 0）で変更された")
	}
}

func TestAIPurchaseStock(t *testing.T) {
	g := NewGame(1, "p")
	c, line := newOperateLine(1_000_000, 100, 85, 10)

	evs := g.aiPurchaseStock(c, line)

	if len(evs) != 1 {
		t.Fatalf("events = %d, want 1", len(evs))
	}
	if evs[0].Type != EventStockPurchased {
		t.Errorf("type = %q, want %q", evs[0].Type, EventStockPurchased)
	}
	if line.TrainsPerDay != 15 {
		t.Errorf("TrainsPerDay = %d, want 15", line.TrainsPerDay)
	}
	if c.Cash != 900_000 {
		t.Errorf("Cash = %d, want 900000", c.Cash)
	}
}

func TestAIPurchaseStock_InsufficientCash(t *testing.T) {
	g := NewGame(1, "p")
	c, line := newOperateLine(100_000, 100, 85, 10)

	if evs := g.aiPurchaseStock(c, line); evs != nil {
		t.Errorf("資金不足で車両購入された: %d", len(evs))
	}
	if line.TrainsPerDay != 10 {
		t.Errorf("TrainsPerDay = %d, want 10（不変）", line.TrainsPerDay)
	}
}

func TestAIPurchaseStock_AtCap(t *testing.T) {
	g := NewGame(1, "p")
	c, line := newOperateLine(1_000_000, 100, 85, 20)

	if evs := g.aiPurchaseStock(c, line); evs != nil {
		t.Errorf("上限で車両購入された: %d", len(evs))
	}
}

func TestAIOperate(t *testing.T) {
	g := NewGame(1, "p")
	c, _ := newOperateLine(1_000_000, 100, 120, 10)
	ai := &AI{CompanyID: "hankyu", Personality: DefaultPersonality("hankyu")}

	evs := g.aiOperate(c, ai)

	if len(evs) != 2 {
		t.Fatalf("events = %d, want 2（fare + stock）", len(evs))
	}
	if !hasEventType(evs, EventFareChanged) {
		t.Error("fare_changed がない")
	}
	if !hasEventType(evs, EventStockPurchased) {
		t.Error("stock_purchased がない")
	}
}
