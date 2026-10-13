package core

import "testing"

// demandGame は需要テスト用のゲームを返す（都市 3 つ、会社 2 つ）。
func demandGame() *Game {
	g := NewGame(1907, "hankyu")
	g.Date = Date{Year: 1910, Month: 6} // 季節係数 1.00

	g.AddCity(&City{ID: "a", Name: "A", Population: 100000})
	g.AddCity(&City{ID: "b", Name: "B", Population: 50000})
	g.AddCity(&City{ID: "c", Name: "C", Population: 20000})

	hankyu := NewCompany("hankyu", "hankyu", "阪急")
	nankai := NewCompany("nankai", "nankai", "南海")
	g.AddCompany(hankyu)
	g.AddCompany(nankai)
	return g
}

func TestRecalcCityDemand(t *testing.T) {
	g := demandGame()
	c := g.Cities["a"]

	// nil は何もしない
	g.recalcCityDemand(nil)

	// 接続なし: Demand と一致
	g.recalcCityDemand(c)
	if c.Demand != g.Demand(c) {
		t.Errorf("接続なしの Demand = %d, want %d", c.Demand, g.Demand(c))
	}
	before := c.Demand

	// 路線を追加すると接続数ボーナスで増加する
	hankyu := g.PlayerCompany()
	hankyu.AddLine(NewLine("L1", "hankyu", "a", "b"))
	g.recalcCityDemand(c)

	if c.Demand <= before {
		t.Errorf("接続後の Demand = %d, 接続前 %d 以下", c.Demand, before)
	}
	// 接続数 1 → ボーナス 1.1
	want := int(float64(g.Demand(c)) * 1.1)
	if c.Demand != want {
		t.Errorf("接続後の Demand = %d, want %d", c.Demand, want)
	}
}

// TestRecalcCityDemand_Exported は公開版 RecalcCityDemand が内部版と
// 同じ結果を返すことを検証する。
func TestRecalcCityDemand_Exported(t *testing.T) {
	g := demandGame()
	c := g.Cities["a"]

	// nil でも panic しない
	g.RecalcCityDemand(nil)

	hankyu := g.PlayerCompany()
	hankyu.AddLine(NewLine("L1", "hankyu", "a", "b"))
	g.RecalcCityDemand(c)

	// 公開版と内部版は同一結果
	want := c.Demand
	c.Demand = 0
	g.recalcCityDemand(c)
	if c.Demand != want {
		t.Errorf("公開版と内部版の結果が異なる: %d vs %d", want, c.Demand)
	}
	if c.Demand != int(float64(g.Demand(c))*1.1) {
		t.Errorf("接続ボーナスが反映されていない: %d", c.Demand)
	}
}

func TestCountConnections(t *testing.T) {
	g := demandGame()

	if got := g.countConnections("a"); got != 0 {
		t.Errorf("初期の接続数 = %d, want 0", got)
	}

	hankyu := g.PlayerCompany()
	hankyu.AddLine(NewLine("L1", "hankyu", "a", "b"))
	if got := g.countConnections("a"); got != 1 {
		t.Errorf("a の接続数 = %d, want 1", got)
	}
	if got := g.countConnections("b"); got != 1 {
		t.Errorf("b の接続数 = %d, want 1", got)
	}

	// 別会社の路線も数える
	nankai := g.Companies["nankai"]
	nankai.AddLine(NewLine("L2", "nankai", "b", "c"))
	if got := g.countConnections("b"); got != 2 {
		t.Errorf("b の接続数 = %d, want 2", got)
	}
	if got := g.countConnections("c"); got != 1 {
		t.Errorf("c の接続数 = %d, want 1", got)
	}

	// 接続のない都市
	if got := g.countConnections("missing"); got != 0 {
		t.Errorf("未知都市の接続数 = %d, want 0", got)
	}
}
