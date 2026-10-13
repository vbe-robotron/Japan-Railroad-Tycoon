package core

import (
	"encoding/json"
	"testing"
)

func TestNewCompany(t *testing.T) {
	c := NewCompany("hankyu", "hankyu-group", "阪急電鉄")

	if c.ID != "hankyu" {
		t.Errorf("ID = %q, want hankyu", c.ID)
	}
	if c.GroupID != "hankyu-group" {
		t.Errorf("GroupID = %q, want hankyu-group", c.GroupID)
	}
	if c.Name != "阪急電鉄" {
		t.Errorf("Name = %q, want 阪急電鉄", c.Name)
	}
	if c.Holding == nil {
		t.Error("Holding が nil")
	}
	if len(c.Holding) != 0 {
		t.Errorf("Holding の数 = %d, want 0", len(c.Holding))
	}
	if c.Lines == nil {
		t.Error("Lines が nil")
	}
	if len(c.Lines) != 0 {
		t.Errorf("Lines の数 = %d, want 0", len(c.Lines))
	}
	if c.Facilities == nil {
		t.Error("Facilities が nil")
	}
	if len(c.Facilities) != 0 {
		t.Errorf("Facilities の数 = %d, want 0", len(c.Facilities))
	}
	if c.Tags == nil {
		t.Error("Tags が nil")
	}
	if c.Reputation != 50 {
		t.Errorf("Reputation = %d, want 50", c.Reputation)
	}
	if c.IsPlayer || c.IsNational {
		t.Error("IsPlayer/IsNational は false であるべき")
	}
}

func TestCompany_NetAssets(t *testing.T) {
	tests := []struct {
		name string
		cash int64
		debt int64
		want int64
	}{
		{"黒字", 1000, 400, 600},
		{"負債超過", 100, 500, -400},
		{"ゼロ", 0, 0, 0},
		{"負債なし", 12345, 0, 12345},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompany("c", "", "C")
			c.Cash = tt.cash
			c.Debt = tt.debt
			if got := c.NetAssets(); got != tt.want {
				t.Errorf("NetAssets() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCompany_MonthlyProfit(t *testing.T) {
	g := NewGame(1907, "hankyu")
	c := NewCompany("c", "", "C")

	// 路線なし → 0
	if got := c.MonthlyProfit(g); got != 0 {
		t.Errorf("路線なし MonthlyProfit = %d, want 0", got)
	}

	// 路線利益のみ（減価償却 0）
	l := NewLine("l1", "c", "a", "b")
	l.Profit = 1000
	c.AddLine(l)
	if got := c.MonthlyProfit(g); got != 1000 {
		t.Errorf("MonthlyProfit = %d, want 1000", got)
	}

	// 減価償却を差し引く: 60km → 価値 15,000,000 → 月次償却 25,000
	l.LengthKm = 60
	if got := c.MonthlyProfit(g); got != -24000 {
		t.Errorf("MonthlyProfit（減価償却後） = %d, want -24000", got)
	}
}

func TestCompany_AnnualProfit(t *testing.T) {
	g := NewGame(1907, "hankyu")
	c := NewCompany("c", "", "C")

	if got := c.AnnualProfit(g); got != 0 {
		t.Errorf("路線なし AnnualProfit = %d, want 0", got)
	}

	l := NewLine("l1", "c", "a", "b")
	l.Profit = 1000
	c.AddLine(l)
	if got := c.AnnualProfit(g); got != 12000 {
		t.Errorf("AnnualProfit = %d, want 12000", got)
	}
}

func TestCompany_NetAssetsWithLines(t *testing.T) {
	g := NewGame(1907, "hankyu")
	c := NewCompany("c", "", "C")
	c.Cash = 1000
	c.Debt = 400

	// 路線なしは NetAssets と同じ
	if got := c.NetAssetsWithLines(g); got != 600 {
		t.Errorf("路線なし NetAssetsWithLines = %d, want 600", got)
	}

	// 路線価値: 10km × 250000 = 2,500,000
	l := NewLine("l1", "c", "a", "b")
	l.LengthKm = 10
	c.AddLine(l)
	if got := c.NetAssetsWithLines(g); got != 2_500_600 {
		t.Errorf("NetAssetsWithLines = %d, want 2500600", got)
	}
}

func TestCompany_AddLine_RemoveLine(t *testing.T) {
	c := NewCompany("c", "", "C")
	if len(c.Lines) != 0 {
		t.Fatalf("初期 Lines の数 = %d, want 0", len(c.Lines))
	}

	l1 := NewLine("l1", "c", "a", "b")
	l2 := NewLine("l2", "c", "b", "c")
	c.AddLine(l1)
	c.AddLine(l2)
	if len(c.Lines) != 2 {
		t.Fatalf("Lines の数 = %d, want 2", len(c.Lines))
	}

	tests := []struct {
		name   string
		lineID string
		want   bool
		after  int
	}{
		{"存在する路線", "l1", true, 1},
		{"存在しない路線", "nope", false, 1},
		{"残りを削除", "l2", true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.RemoveLine(tt.lineID)
			if got != tt.want {
				t.Errorf("RemoveLine(%q) = %v, want %v", tt.lineID, got, tt.want)
			}
			if len(c.Lines) != tt.after {
				t.Errorf("Lines の数 = %d, want %d", len(c.Lines), tt.after)
			}
		})
	}
}

func TestCompany_FindLine(t *testing.T) {
	c := NewCompany("c", "", "C")
	l := NewLine("l1", "c", "a", "b")
	c.AddLine(l)

	if got := c.FindLine("l1"); got != l {
		t.Errorf("FindLine(l1) = %+v, want %+v", got, l)
	}
	if got := c.FindLine("missing"); got != nil {
		t.Errorf("FindLine(missing) = %+v, want nil", got)
	}
}

func TestCompany_JSONRoundTrip(t *testing.T) {
	c := NewCompany("hankyu", "hankyu-group", "阪急電鉄")
	c.LegalName = "阪急電鉄株式会社"
	c.FoundedYear = 1907
	c.Cash = 1_000_000
	c.Debt = 500_000
	c.StockPrice = 1200
	c.TotalShares = 1_000_000
	c.Holding["hanshin"] = 100
	c.IsPlayer = true
	c.Region = "kansai"
	c.Tags = append(c.Tags, "private")
	c.ParentID = "parent"
	c.Reputation = 75
	c.CreditLine = 10_000_000
	c.AddLine(NewLine("l1", "hankyu", "osaka", "kobe"))
	c.Facilities = append(c.Facilities, NewFacility("f1", "hankyu", "osaka", "department_store", 3))

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var restored Company
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if restored.ID != c.ID || restored.GroupID != c.GroupID || restored.Name != c.Name {
		t.Errorf("基本情報が異なる: %+v", restored)
	}
	if restored.LegalName != c.LegalName || restored.FoundedYear != c.FoundedYear {
		t.Errorf("法人情報が異なる: %+v", restored)
	}
	if restored.Cash != c.Cash || restored.Debt != c.Debt {
		t.Errorf("資産情報が異なる: %+v", restored)
	}
	if restored.StockPrice != c.StockPrice || restored.TotalShares != c.TotalShares {
		t.Errorf("株式情報が異なる: %+v", restored)
	}
	if restored.Holding["hanshin"] != 100 {
		t.Errorf("Holding が異なる: %+v", restored.Holding)
	}
	if len(restored.Lines) != 1 || restored.Lines[0].ID != "l1" {
		t.Errorf("Lines が異なる: %+v", restored.Lines)
	}
	if len(restored.Facilities) != 1 || restored.Facilities[0].ID != "f1" {
		t.Errorf("Facilities が異なる: %+v", restored.Facilities)
	}
	if !restored.IsPlayer || restored.Region != "kansai" {
		t.Errorf("フラグが異なる: %+v", restored)
	}
	if restored.ParentID != "parent" || restored.Reputation != 75 || restored.CreditLine != c.CreditLine {
		t.Errorf("その他が異なる: %+v", restored)
	}
}

func TestCompany_JSONHoldingNotNil(t *testing.T) {
	// 空 map が null にシリアライズされないこと
	c := NewCompany("c", "", "C")
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if string(m["holding"]) == "null" {
		t.Error("holding が null になっている")
	}
	if string(m["lines"]) == "null" {
		t.Error("lines が null になっている")
	}
}
