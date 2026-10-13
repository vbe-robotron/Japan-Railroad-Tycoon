package core

import (
	"testing"
)

// ecoGame は 2 都市（a, b）と 1 社（c）を持つテスト用ゲームを返す。
//
// 季節係数 1.00 の 6 月に固定する。
func ecoGame() *Game {
	g := NewGame(1907, "hankyu")
	g.Date = Date{Year: 1910, Month: 6} // 季節係数 1.00
	g.AddCity(&City{ID: "a", Name: "A", Population: 100000})
	g.AddCity(&City{ID: "b", Name: "B", Population: 100000})
	return g
}

// ecoLine はテスト用の路線を生成して会社に追加する。
func ecoLine(g *Game, c *Company, id, from, to string, lengthKm float64, fare int, trains int) *Line {
	l := NewLine(id, c.ID, from, to)
	l.LengthKm = lengthKm
	l.Fare = fare
	l.TrainsPerDay = trains
	c.AddLine(l)
	return l
}

func TestSimulateLine_Revenue(t *testing.T) {
	g := ecoGame()
	c := NewCompany("c", "", "C")
	g.AddCompany(c)
	l := ecoLine(g, c, "l1", "a", "b", 10, 200, 10)

	revenue, _ := g.simulateLine(l, c)

	wantPassengers := g.linePassengers(l)
	if revenue != int64(wantPassengers)*int64(l.Fare) {
		t.Errorf("revenue = %d, want %d（旅客数 %d × 運賃 %d）",
			revenue, int64(wantPassengers)*int64(l.Fare), wantPassengers, l.Fare)
	}
	if revenue <= 0 {
		t.Errorf("revenue = %d, want > 0", revenue)
	}
}

func TestSimulateLine_Cost(t *testing.T) {
	g := ecoGame()
	c := NewCompany("c", "", "C")
	g.AddCompany(c)
	l := ecoLine(g, c, "l1", "a", "b", 10, 200, 10)

	_, cost := g.simulateLine(l, c)

	// 10km × 20000円/km/月 = 200000
	if cost != 200000 {
		t.Errorf("cost = %d, want 200000", cost)
	}
}

func TestSimulateLine_Double_IncreaseCost(t *testing.T) {
	g := ecoGame()
	c := NewCompany("c", "", "C")
	g.AddCompany(c)

	single := ecoLine(g, c, "s", "a", "b", 10, 200, 10)
	_, singleCost := g.simulateLine(single, c)

	dbl := ecoLine(g, c, "d", "a", "b", 10, 200, 10)
	dbl.Double = true
	_, doubleCost := g.simulateLine(dbl, c)

	if doubleCost <= singleCost {
		t.Errorf("複線維持費 %d が単線 %d 以下", doubleCost, singleCost)
	}
	// ×1.5
	if doubleCost != 300000 {
		t.Errorf("複線維持費 = %d, want 300000", doubleCost)
	}
}

func TestSimulateLine_Electrified_IncreaseCost(t *testing.T) {
	g := ecoGame()
	c := NewCompany("c", "", "C")
	g.AddCompany(c)

	plain := ecoLine(g, c, "p", "a", "b", 10, 200, 10)
	_, plainCost := g.simulateLine(plain, c)

	elec := ecoLine(g, c, "e", "a", "b", 10, 200, 10)
	elec.Electrified = true
	_, elecCost := g.simulateLine(elec, c)

	if elecCost <= plainCost {
		t.Errorf("電化維持費 %d が非電化 %d 以下", elecCost, plainCost)
	}
	// ×1.3
	if elecCost != 260000 {
		t.Errorf("電化維持費 = %d, want 260000", elecCost)
	}
}

func TestLinePassengers_NoCities(t *testing.T) {
	g := ecoGame()
	c := NewCompany("c", "", "C")
	g.AddCompany(c)
	l := ecoLine(g, c, "l1", "missing1", "missing2", 10, 200, 10)

	if got := g.linePassengers(l); got != 0 {
		t.Errorf("都市なし linePassengers = %d, want 0", got)
	}
	// 片方だけ存在しない場合
	l2 := ecoLine(g, c, "l2", "a", "missing", 10, 200, 10)
	if got := g.linePassengers(l2); got != 0 {
		t.Errorf("片方なし linePassengers = %d, want 0", got)
	}
}

func TestLinePassengers_DemandShare(t *testing.T) {
	g := ecoGame()
	g.Date = Date{Year: 1910, Month: 6} // 季節係数 1.00 → 各都市 Demand=100
	c1 := NewCompany("c1", "", "C1")
	c2 := NewCompany("c2", "", "C2")
	g.AddCompany(c1)
	g.AddCompany(c2)

	// score c1 = (100/100)*(10/10) = 1.0
	l1 := ecoLine(g, c1, "l1", "a", "b", 10, 100, 10)
	// score c2 = (100/200)*(10/10) = 0.5
	l2 := ecoLine(g, c2, "l2", "a", "b", 10, 200, 10)

	// 需要 = (100+100)/2 = 100、日次 → 月次 ×30
	// share(c1) = 1.0/1.5 = 2/3 → 100×2/3×30 = 2000
	if got := g.linePassengers(l1); got != 2000 {
		t.Errorf("linePassengers(l1) = %d, want 2000", got)
	}
	// share(c2) = 0.5/1.5 = 1/3 → 100×1/3×30 = 1000
	if got := g.linePassengers(l2); got != 1000 {
		t.Errorf("linePassengers(l2) = %d, want 1000", got)
	}
}

func TestLinePassengers_CapacityLimit(t *testing.T) {
	g := ecoGame()
	g.AddCity(&City{ID: "big", Name: "Big", Population: 10_000_000})
	c := NewCompany("c", "", "C")
	g.AddCompany(c)
	// 1 本/日 × 500 人/日 × 30 日 = 15000 が月次輸送力
	l := ecoLine(g, c, "l1", "big", "big2", 10, 200, 1)
	g.AddCity(&City{ID: "big2", Name: "Big2", Population: 10_000_000})

	if got := g.linePassengers(l); got != 15000 {
		t.Errorf("輸送力制約 linePassengers = %d, want 15000", got)
	}
}

func TestLineShare_SingleCompany(t *testing.T) {
	g := ecoGame()
	c := NewCompany("c", "", "C")
	g.AddCompany(c)
	l := ecoLine(g, c, "l1", "a", "b", 10, 200, 10)

	if got := g.lineShare(l); got != 1.0 {
		t.Errorf("独占 lineShare = %v, want 1.0", got)
	}

	// スコア 0（Condition=0）でもゼロ除算せず 1.0 を返す
	l.Condition = 0
	if got := g.lineShare(l); got != 1.0 {
		t.Errorf("スコア0 lineShare = %v, want 1.0", got)
	}
}

func TestLineShare_TwoCompanies(t *testing.T) {
	g := ecoGame()
	c1 := NewCompany("c1", "", "C1")
	c2 := NewCompany("c2", "", "C2")
	g.AddCompany(c1)
	g.AddCompany(c2)

	l1 := ecoLine(g, c1, "l1", "a", "b", 10, 100, 10)
	l2 := ecoLine(g, c2, "l2", "b", "a", 10, 200, 10) // 逆方向も同一区間

	want := 1.0 / 1.5
	if got := g.lineShare(l1); got != want {
		t.Errorf("lineShare(l1) = %v, want %v", got, want)
	}
	if got := g.lineShare(l2); got != 0.5/1.5 {
		t.Errorf("lineShare(l2) = %v, want %v", got, 0.5/1.5)
	}
}

func TestLineScore(t *testing.T) {
	g := ecoGame()
	c := NewCompany("c", "", "C")
	g.AddCompany(c)

	// (100/200) × (20/10) × 1.3 × 0.5 = 0.65
	l := ecoLine(g, c, "l1", "a", "b", 10, 200, 20)
	l.Electrified = true
	l.Condition = 0.5
	if got := g.lineScore(l); got != 0.65 {
		t.Errorf("lineScore = %v, want 0.65", got)
	}

	// Fare=0 は運賃補正なし、Condition=1.0
	l2 := ecoLine(g, c, "l2", "a", "b", 10, 0, 10)
	l2.Condition = 1.0
	if got := g.lineScore(l2); got != 1.0 {
		t.Errorf("Fare=0 lineScore = %v, want 1.0", got)
	}
}

func TestSameOD(t *testing.T) {
	a := &Line{From: "x", To: "y"}
	b := &Line{From: "x", To: "y"}
	c := &Line{From: "y", To: "x"}
	d := &Line{From: "x", To: "z"}

	if !sameOD(a, b) {
		t.Error("同一方向 sameOD = false, want true")
	}
	if !sameOD(a, c) {
		t.Error("逆方向 sameOD = false, want true")
	}
	if sameOD(a, d) {
		t.Error("異なる区間 sameOD = true, want false")
	}
}

func TestUpdateLineHealth_Deficit(t *testing.T) {
	g := ecoGame()
	c := NewCompany("c", "", "C")
	l := ecoLine(g, c, "l1", "a", "b", 10, 200, 10)
	l.DisplayName = "宝塚線"

	// 6 ターン目に line_deficit が生成される
	var deficit *Event
	for i := 0; i < 6; i++ {
		events := g.updateLineHealth(c, l, -12000)
		for j := range events {
			if events[j].Type == EventLineDeficit {
				deficit = &events[j]
			}
		}
		// 6 ターン目より前はイベントなし
		if i < 5 && len(events) != 0 {
			t.Errorf("turn %d: 予期しないイベント %d 件", i+1, len(events))
		}
	}

	if deficit == nil {
		t.Fatal("6 ターン目に line_deficit が生成されなかった")
	}
	if deficit.Severity != SeverityWarning {
		t.Errorf("severity = %q, want %q", deficit.Severity, SeverityWarning)
	}
	if deficit.LineID != "l1" {
		t.Errorf("LineID = %q, want l1", deficit.LineID)
	}
	if deficit.CompanyID != "c" {
		t.Errorf("CompanyID = %q, want c", deficit.CompanyID)
	}
	if deficit.Payload["monthly_loss"] != int64(-12000) {
		t.Errorf("monthly_loss = %v, want -12000", deficit.Payload["monthly_loss"])
	}
	if deficit.Payload["consecutive_turns"] != 6 {
		t.Errorf("consecutive_turns = %v, want 6", deficit.Payload["consecutive_turns"])
	}
	if l.AccumulatedLoss != -72000 {
		t.Errorf("AccumulatedLoss = %d, want -72000", l.AccumulatedLoss)
	}
	if l.ConsecutiveDeficitTurns != 6 {
		t.Errorf("ConsecutiveDeficitTurns = %d, want 6", l.ConsecutiveDeficitTurns)
	}
}

func TestUpdateLineHealth_Critical(t *testing.T) {
	g := ecoGame()
	c := NewCompany("c", "", "C")
	l := ecoLine(g, c, "l1", "a", "b", 10, 200, 10)
	l.DisplayName = "赤字線"

	var critical *Event
	for i := 0; i < 24; i++ {
		events := g.updateLineHealth(c, l, -5000)
		for j := range events {
			if events[j].Type == EventLineCritical {
				critical = &events[j]
			}
		}
	}

	if critical == nil {
		t.Fatal("24 ターン目に line_critical が生成されなかった")
	}
	if critical.Severity != SeverityDanger {
		t.Errorf("severity = %q, want %q", critical.Severity, SeverityDanger)
	}
	if critical.Payload["consecutive_turns"] != 24 {
		t.Errorf("consecutive_turns = %v, want 24", critical.Payload["consecutive_turns"])
	}
	if critical.Payload["accumulated_loss"] != int64(-120000) {
		t.Errorf("accumulated_loss = %v, want -120000", critical.Payload["accumulated_loss"])
	}
	if l.ConsecutiveDeficitTurns != 24 {
		t.Errorf("ConsecutiveDeficitTurns = %d, want 24", l.ConsecutiveDeficitTurns)
	}
}

func TestUpdateLineHealth_Recovery(t *testing.T) {
	g := ecoGame()
	c := NewCompany("c", "", "C")
	l := ecoLine(g, c, "l1", "a", "b", 10, 200, 10)
	l.DisplayName = "回復線"
	l.ConsecutiveDeficitTurns = 3
	l.AccumulatedLoss = -3000

	events := g.updateLineHealth(c, l, 5000)

	if len(events) != 1 || events[0].Type != EventLineProfitable {
		t.Fatalf("黑字転換イベントが生成されなかった: %+v", events)
	}
	if events[0].Severity != SeveritySuccess {
		t.Errorf("severity = %q, want %q", events[0].Severity, SeveritySuccess)
	}
	if events[0].Payload["monthly_profit"] != int64(5000) {
		t.Errorf("monthly_profit = %v, want 5000", events[0].Payload["monthly_profit"])
	}
	if l.ConsecutiveDeficitTurns != 0 {
		t.Errorf("ConsecutiveDeficitTurns = %d, want 0", l.ConsecutiveDeficitTurns)
	}
	if l.AccumulatedLoss != 0 {
		t.Errorf("AccumulatedLoss = %d, want 0", l.AccumulatedLoss)
	}

	// もともと赤字でない（0）の場合、黒字でもイベントなし
	l2 := ecoLine(g, c, "l2", "a", "b", 10, 200, 10)
	if events := g.updateLineHealth(c, l2, 1000); len(events) != 0 {
		t.Errorf("通常黒字でイベント %d 件, want 0", len(events))
	}
}

func TestCalculateDepreciation(t *testing.T) {
	g := ecoGame()
	c := NewCompany("c", "", "C")

	// 路線なし → 0
	if got := g.calculateDepreciation(c); got != 0 {
		t.Errorf("路線なし calculateDepreciation = %d, want 0", got)
	}

	// 60km → 価値 300000 → 月次 500
	ecoLine(g, c, "l1", "a", "b", 60, 200, 10)
	if got := g.calculateDepreciation(c); got != 500 {
		t.Errorf("calculateDepreciation = %d, want 500", got)
	}

	// 120km を追加 → 1000 加算
	ecoLine(g, c, "l2", "a", "b", 120, 200, 10)
	if got := g.calculateDepreciation(c); got != 1500 {
		t.Errorf("calculateDepreciation = %d, want 1500", got)
	}
}

func TestSimulateCompany(t *testing.T) {
	g := ecoGame()
	c := NewCompany("c", "", "C")
	g.AddCompany(c)
	c.Cash = 0

	l := ecoLine(g, c, "l1", "a", "b", 10, 200, 10)
	events := g.simulateCompany(c)

	revenue, cost := g.simulateLine(l, c)
	wantProfit := revenue - cost
	if l.Profit != wantProfit {
		t.Errorf("Line.Profit = %d, want %d", l.Profit, wantProfit)
	}
	if l.OperatingRatio != float64(cost)/float64(revenue)*100.0 {
		t.Errorf("OperatingRatio = %v, want %v", l.OperatingRatio,
			float64(cost)/float64(revenue)*100.0)
	}
	// 減価償却 5000*10/600 = 83
	dep := g.calculateDepreciation(c)
	if c.Cash != revenue-cost-dep {
		t.Errorf("Cash = %d, want %d", c.Cash, revenue-cost-dep)
	}
	if len(events) != 0 {
		t.Errorf("黒字でイベント %d 件, want 0", len(events))
	}

	// 収入 0 の場合 OperatingRatio = 999.9
	l2 := ecoLine(g, c, "l2", "a", "b", 10, 0, 0)
	g.simulateCompany(c)
	if l2.OperatingRatio != 999.9 {
		t.Errorf("収入0 OperatingRatio = %v, want 999.9", l2.OperatingRatio)
	}
}

func TestSimulateEconomy_Deterministic(t *testing.T) {
	setup := func() *Game {
		g := ecoGame()
		c1 := NewCompany("zzz", "", "Z")
		c2 := NewCompany("aaa", "", "A")
		g.AddCompany(c1)
		g.AddCompany(c2)
		ecoline1 := ecoLine(g, c1, "z1", "a", "b", 400, 100, 10)
		ecoline1.DisplayName = "Z線"
		ecoline2 := ecoLine(g, c2, "a1", "a", "b", 400, 100, 10)
		ecoline2.DisplayName = "A線"
		return g
	}

	g1 := setup()
	g2 := setup()

	// 6 ターン分シミュレートしてイベント列が一致することを確認
	collect := func(g *Game) []Event {
		var all []Event
		for i := 0; i < 6; i++ {
			all = append(all, g.simulateEconomy()...)
		}
		return all
	}
	e1 := collect(g1)
	e2 := collect(g2)

	if len(e1) != len(e2) {
		t.Fatalf("イベント数が一致しない: %d != %d", len(e1), len(e2))
	}
	for i := range e1 {
		if e1[i].Type != e2[i].Type {
			t.Errorf("event %d Type = %q, want %q", i, e1[i].Type, e2[i].Type)
		}
		if e1[i].CompanyID != e2[i].CompanyID {
			t.Errorf("event %d CompanyID = %q, want %q", i, e1[i].CompanyID, e2[i].CompanyID)
		}
	}
	// ID順（aaa が先）で処理される
	if len(e1) == 0 || e1[0].CompanyID != "aaa" {
		t.Errorf("イベント順が ID 順でない: %+v", e1)
	}
}

// --- 統合シナリオ ---

// TestEconomy_Scenario_Profitable は儲かる路線のシナリオを検証する。
//
// 阪急（資金 100 万円）が梅田〜宝塚を敷設し、12 ターンで資金が増える。
// Phase 1-7.5 のバランス目標:
//   - 営業係数 70〜90（史実の私鉄に近づける）
//   - 12 ターン後の資金増加率 > 1.1（6.6 倍から大幅に圧縮）
//
// 注意（数値設計上の制約）:
//
//	営業係数 r と 12 ターン資金増加率 G は
//	  G = 1 + (12ヶ月の営業利益)/開始資金 = 1 + (12ヶ月収入)(1-r)/開始資金
//	の関係にあり、両者は独立に選べない。本シナリオ（12ヶ月収入 = 5,346,000、
//	開始資金 = 904,606）では
//	  r = 70-90% → G = 1.59-2.65
//	  G = 1.1-1.5 → r = 91.5-98.3%
//	となり、「営業係数 70-90」と「資金増加率 1.1-1.5」は同時に満たせない。
//	史実の営業係数を優先し、資金増加率は 1.1 超（黒字成長）を検証する。
func TestEconomy_Scenario_Profitable(t *testing.T) {
	g := NewGame(1907, "hankyu")
	g.Date = Date{Year: 1910, Month: 1}

	// 都市（実座標付き）
	g.AddCity(&City{
		ID: "umeda", Name: "梅田", Population: 120000,
		Lat: 34.7055, Lon: 135.4983,
	})
	g.AddCity(&City{
		ID: "takarazuka", Name: "宝塚", Population: 30000,
		Lat: 34.7995, Lon: 135.3468,
	})

	hankyu := NewCompany("hankyu", "hankyu", "阪急")
	hankyu.Cash = 1_000_000
	g.AddCompany(hankyu)

	// 路線敷設
	a := &BuildLineAction{From: "umeda", To: "takarazuka", Gauge: 1067}
	if _, err := a.Apply(g); err != nil {
		t.Fatal(err)
	}

	startCash := hankyu.Cash
	line := hankyu.Lines[0]
	t.Logf("[梅田〜宝塚] 距離 %.1fkm, 運賃 %d円, 敷設コスト %d円, 敷設後資金 %d円",
		line.LengthKm, line.Fare, g.LineCost("umeda", "takarazuka", 1067, false), startCash)

	// 12 ターン進める（各月の数値を記録）
	for i := 0; i < 12; i++ {
		pax := g.linePassengers(line)
		revenue, maint := g.simulateLine(line, hankyu)
		ratio := 0.0
		if revenue > 0 {
			ratio = float64(maint) / float64(revenue) * 100.0
		}
		t.Logf("  %s: 旅客 %d人/月, 収入 %d円, 維持費 %d円, 利益 %d円, 営業係数 %.1f",
			g.Date.String(), pax, revenue, maint, revenue-maint, ratio)
		g.EndTurn()
	}

	growth := float64(hankyu.Cash) / float64(startCash)
	dep := g.calculateDepreciation(hankyu)
	t.Logf("[梅田〜宝塚] 12ターン後資金 %d円（%.3f倍, 変化 %d円）",
		hankyu.Cash, growth, hankyu.Cash-startCash)
	t.Logf("[梅田〜宝塚] 最終月: 利益 %d円, 営業係数 %.1f, 減価償却 %d円/月, 会社月次利益 %d円",
		line.Profit, line.OperatingRatio, dep, hankyu.MonthlyProfit(g))

	// 営業係数 70-90（史実の私鉄）
	if line.OperatingRatio < 70 || line.OperatingRatio > 90 {
		t.Errorf("営業係数 = %.1f, want 70-90", line.OperatingRatio)
	}
	// 資金増加率 > 1.1（黒字成長）
	if growth <= 1.1 {
		t.Errorf("資金増加率 = %.3f, want > 1.1", growth)
	}
	// 6.6 倍の過大な黒字から改善していること
	if growth >= 2.0 {
		t.Errorf("資金増加率 = %.3f, want < 2.0（黒字が過大）", growth)
	}
	if line.Profit <= 0 {
		t.Errorf("路線利益 = %d, want > 0", line.Profit)
	}
}

// TestEconomy_Scenario_Deficit は赤字路線のシナリオを検証する。
//
// 資金 100 万円で赤字路線（36.7km、運賃 340 円）を敷設し、
// 6 ターンで line_deficit、24 ターンで line_critical が生成される。
// Phase 1-7.5 のバランス目標: 営業係数 150〜200（= コストが収入の 1.5〜2 倍）。
func TestEconomy_Scenario_Deficit(t *testing.T) {
	g := NewGame(1907, "rural")
	g.Date = Date{Year: 1910, Month: 1}

	g.AddCity(&City{
		ID: "ruralA", Name: "地方都市", Population: 50000,
		Lat: 35.0000, Lon: 136.0000,
	})
	g.AddCity(&City{
		ID: "ruralB", Name: "近郊町", Population: 30000,
		Lat: 35.3000, Lon: 136.0000,
	})

	co := NewCompany("rural", "rural", "田舎鉄道")
	co.Cash = 1_000_000
	g.AddCompany(co)

	a := &BuildLineAction{From: "ruralA", To: "ruralB", Gauge: 1067}
	if _, err := a.Apply(g); err != nil {
		t.Fatal(err)
	}
	line := co.Lines[0]
	t.Logf("[田舎路線] 距離 %.1fkm, 運賃 %d円, 敷設コスト %d円, 敷設後資金 %d円",
		line.LengthKm, line.Fare, g.LineCost("ruralA", "ruralB", 1067, false), co.Cash)

	// 6 ターン（各月の数値を記録）
	for i := 0; i < 6; i++ {
		pax := g.linePassengers(line)
		revenue, maint := g.simulateLine(line, co)
		ratio := 0.0
		if revenue > 0 {
			ratio = float64(maint) / float64(revenue) * 100.0
		}
		t.Logf("  %s: 旅客 %d人/月, 収入 %d円, 維持費 %d円, 営業係数 %.1f",
			g.Date.String(), pax, revenue, maint, ratio)
		g.EndTurn()
	}
	if !hasEventType(g.Events, EventLineDeficit) {
		t.Errorf("6 ターンで line_deficit が生成されなかった")
	}
	t.Logf("[田舎路線] 6ターン後: 利益 %d円, 営業係数 %.1f, 累積損失 %d円",
		line.Profit, line.OperatingRatio, line.AccumulatedLoss)

	if line.OperatingRatio < 150 || line.OperatingRatio > 200 {
		t.Errorf("赤字路線の営業係数 = %.1f, want 150-200", line.OperatingRatio)
	}

	// さらに 18 ターン（合計 24）
	for i := 0; i < 18; i++ {
		g.EndTurn()
	}
	if !hasEventType(g.Events, EventLineCritical) {
		t.Errorf("24 ターンで line_critical が生成されなかった")
	}
	t.Logf("[田舎路線] 24ターン後: 連続赤字 %d, 累積損失 %d円, 資金 %d円",
		line.ConsecutiveDeficitTurns, line.AccumulatedLoss, co.Cash)

	if line.ConsecutiveDeficitTurns != 24 {
		t.Errorf("ConsecutiveDeficitTurns = %d, want 24", line.ConsecutiveDeficitTurns)
	}
}

// hasEventType は events 中に指定タイプが存在するかを返す。
func hasEventType(events []Event, typ string) bool {
	for _, e := range events {
		if e.Type == typ {
			return true
		}
	}
	return false
}
