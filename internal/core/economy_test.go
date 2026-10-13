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

	// 10km × 2800円/km/月 = 28000（Phase 1-7.7 で 20000 → 2800）
	if cost != 28000 {
		t.Errorf("cost = %d, want 28000", cost)
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
	if doubleCost != 42000 {
		t.Errorf("複線維持費 = %d, want 42000", doubleCost)
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
	if elecCost != 36400 {
		t.Errorf("電化維持費 = %d, want 36400", elecCost)
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
	g.Date = Date{Year: 1910, Month: 6} // 季節係数 1.00 → 各都市 Demand=(100)^0.6=15
	c1 := NewCompany("c1", "", "C1")
	c2 := NewCompany("c2", "", "C2")
	g.AddCompany(c1)
	g.AddCompany(c2)

	// score c1 = (100/100)*(10/10) = 1.0
	l1 := ecoLine(g, c1, "l1", "a", "b", 10, 100, 10)
	// score c2 = (100/200)*(10/10) = 0.5
	l2 := ecoLine(g, c2, "l2", "a", "b", 10, 200, 10)

	// 需要 = (15+15)/2 = 15、日次 → 月次 ×30
	// share(c1) = 1.0/1.5 = 2/3 → 15×2/3×30 = 300
	if got := g.linePassengers(l1); got != 300 {
		t.Errorf("linePassengers(l1) = %d, want 300", got)
	}
	// share(c2) = 0.5/1.5 = 1/3 → 15×1/3×30 = 150
	if got := g.linePassengers(l2); got != 150 {
		t.Errorf("linePassengers(l2) = %d, want 150", got)
	}
}

func TestLinePassengers_CapacityLimit(t *testing.T) {
	g := ecoGame()
	// 需要が輸送力を超える巨大都市（べき乗則でも Demand=1000）
	g.AddCity(&City{ID: "big", Name: "Big", Population: 100_000_000})
	c := NewCompany("c", "", "C")
	g.AddCompany(c)
	// 1 本/日 × 500 人/日 × 30 日 = 15000 が月次輸送力
	l := ecoLine(g, c, "l1", "big", "big2", 10, 200, 1)
	g.AddCity(&City{ID: "big2", Name: "Big2", Population: 100_000_000})

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

	// 60km → 価値 60×250000 = 15,000,000 → 月次 25,000
	ecoLine(g, c, "l1", "a", "b", 60, 200, 10)
	if got := g.calculateDepreciation(c); got != 25000 {
		t.Errorf("calculateDepreciation = %d, want 25000", got)
	}

	// 120km を追加 → 30,000,000/600 = 50,000 加算 → 合計 75,000
	ecoLine(g, c, "l2", "a", "b", 120, 200, 10)
	if got := g.calculateDepreciation(c); got != 75000 {
		t.Errorf("calculateDepreciation = %d, want 75000", got)
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
	// 減価償却 250000*10/600 = 4166
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
// 阪急（資金 900 万円）が梅田〜宝塚（19.1km、運賃 200 円）を敷設し、
// 12 ターンで黒字を維持する。
//
// Phase 1-7.7（需要のべき乗則化）後のバランス目標:
//   - 営業係数 70〜90（史実の私鉄、維持）
//   - 建設費 / 初期資金 30〜60%
//   - 路線利益 > 0
//
// 需要を人口のべき乗則（指数 0.6）にしたことで旅客収入が約 1/6 に
// なるため、維持費単価を 20000 → 2800 にスケールし、営業係数 70〜90 を
// 維持した。ただし BaseCostPerKm（建設費）は据え置きのため、Phase 1-7.6
// で定めた投資回収 3〜7 年・資金増加率の各目標は、本フェーズでは
// 満たせなくなった（投資回収は約 22 年、資金増加率は約 1.01〜1.03）。
// これは「需要のべき乗則化に伴う絶対収益の低下」と「建設費単価の据え置き」
// の組み合わせによる既知のクロスフェーズ矛盾であり、建設費単価の再調整は
// 次フェーズ以降の課題とする（本フェーズの完了条件は営業係数の維持）。
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
	hankyu.Cash = 9_000_000
	g.AddCompany(hankyu)

	// 路線敷設
	initialCash := hankyu.Cash
	a := &BuildLineAction{From: "umeda", To: "takarazuka", Gauge: 1067}
	if _, err := a.Apply(g); err != nil {
		t.Fatal(err)
	}

	postCash := hankyu.Cash
	buildCost := g.LineCost("umeda", "takarazuka", 1067, false)
	line := hankyu.Lines[0]
	t.Logf("[梅田〜宝塚] 距離 %.1fkm, 運賃 %d円", line.LengthKm, line.Fare)
	t.Logf("[梅田〜宝塚] 建設費 %d円, 初期資金 %d円, 建設後資金 %d円（建設費/初期資金 %.1f%%）",
		buildCost, initialCash, postCash, float64(buildCost)/float64(initialCash)*100.0)

	// 12 ターン進める（各月の数値を記録）
	var totalLineProfit int64
	for i := 0; i < 12; i++ {
		pax := g.linePassengers(line)
		revenue, maint := g.simulateLine(line, hankyu)
		ratio := 0.0
		if revenue > 0 {
			ratio = float64(maint) / float64(revenue) * 100.0
		}
		totalLineProfit += revenue - maint
		t.Logf("  %s: 旅客 %d人/月, 収入 %d円, 維持費 %d円, 利益 %d円, 営業係数 %.1f",
			g.Date.String(), pax, revenue, maint, revenue-maint, ratio)
		g.EndTurn()
	}

	// 12 ヶ月の実質利益（営業利益 − 減価償却）。資金増加率の分子。
	dep := g.calculateDepreciation(hankyu)
	netProfit := totalLineProfit - dep*12
	// 仕様の定義: G = 1 + 12ヶ月利益 / 基準資金
	growthInit := 1.0 + float64(netProfit)/float64(initialCash)
	growthPost := 1.0 + float64(netProfit)/float64(postCash)
	avgLineProfit := totalLineProfit / 12
	paybackYears := float64(buildCost) / float64(avgLineProfit) / 12.0
	t.Logf("[梅田〜宝塚] 12ターン利益合計 %d円（平均 %d円/月）, 減価償却 %d円/月, 実質利益 %d円",
		totalLineProfit, avgLineProfit, dep, netProfit)
	t.Logf("[梅田〜宝塚] 12ターン後資金 %d円（変化 %d円）。資金増加率（G=1+実質利益/基準資金）: 初期比 %.3f, 建設後比 %.3f",
		hankyu.Cash, hankyu.Cash-initialCash, growthInit, growthPost)
	t.Logf("[梅田〜宝塚] 投資回収期間 %.1f年, 最終月: 利益 %d円, 営業係数 %.1f, 会社月次利益 %d円",
		paybackYears, line.Profit, line.OperatingRatio, hankyu.MonthlyProfit(g))

	// 営業係数 70-90（史実の私鉄、Phase 1-7.7 でも維持）
	if line.OperatingRatio < 70 || line.OperatingRatio > 90 {
		t.Errorf("営業係数 = %.1f, want 70-90", line.OperatingRatio)
	}
	// 建設費 / 初期資金 30-60%
	ratio := float64(buildCost) / float64(initialCash) * 100.0
	if ratio < 30 || ratio > 60 {
		t.Errorf("建設費/初期資金 = %.1f%%, want 30-60%%", ratio)
	}
	// 投資回収期間は BaseCostPerKm 据え置きのため長期化（既知の矛盾）。
	// 参考値として記録し、レンジのみ検証する。
	if paybackYears < 15 || paybackYears > 30 {
		t.Errorf("投資回収期間 = %.1f年, want 15-30（Phase 1-7.7 の暫定）", paybackYears)
	}
	// 資金増加率は需要減により縮小（既知の矛盾）。
	if growthInit < 1.00 || growthInit > 1.06 {
		t.Errorf("資金増加率（初期比） = %.3f, want 1.00-1.06（Phase 1-7.7 の暫定）", growthInit)
	}
	if growthPost < 1.00 || growthPost > 1.06 {
		t.Errorf("資金増加率（建設後比） = %.3f, want 1.00-1.06（Phase 1-7.7 の暫定）", growthPost)
	}
	if line.Profit <= 0 {
		t.Errorf("路線利益 = %d, want > 0", line.Profit)
	}
}

// TestEconomy_Scenario_Deficit は赤字路線のシナリオを検証する。
//
// 資金 1,000 万円で赤字路線（36.7km、運賃 340 円）を敷設し、
// 6 ターンで line_deficit、24 ターンで line_critical が生成される。
// Phase 1-7.7 のバランス目標: 営業係数 150〜200（= コストが収入の 1.5〜2 倍）。
//
// 需要のべき乗則化で旅客収入が大きく下がるため、田舎路線の人口は
// 従来の 5 万/3 万から 3 万/2 万に下げ、営業係数 150-200 を維持する。
func TestEconomy_Scenario_Deficit(t *testing.T) {
	g := NewGame(1907, "rural")
	g.Date = Date{Year: 1910, Month: 1}

	g.AddCity(&City{
		ID: "ruralA", Name: "地方都市", Population: 30000,
		Lat: 35.0000, Lon: 136.0000,
	})
	g.AddCity(&City{
		ID: "ruralB", Name: "近郊町", Population: 20000,
		Lat: 35.3000, Lon: 136.0000,
	})

	co := NewCompany("rural", "rural", "田舎鉄道")
	co.Cash = 10_000_000
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

// TestDemand_PowerLaw は需要モデルが人口のべき乗則（指数 0.6）に
// したがっていることを検証する。
//
// Phase 1-7.7: Demand = (Population / DemandBase)^0.6
// 季節係数を 0.90（1 月）として、以下のレンジに収まることを確認する。
//
//	人口 3 万  → 需要 5〜10
//	人口 12 万 → 需要 15〜22
//	人口 40 万 → 需要 30〜40
//
// 旧線形モデルではそれぞれ 30 / 120 / 400 だったのに対し、べき乗則では
// 6 / 15 / 32 となり、大都市の需要が大幅に圧縮される。
func TestDemand_PowerLaw(t *testing.T) {
	g := NewGame(1907, "hankyu")
	g.Date = Date{Year: 1910, Month: 1}

	tests := []struct {
		population int
		minDemand  int
		maxDemand  int
	}{
		{30000, 5, 10},
		{120000, 15, 22},
		{400000, 30, 40},
	}
	for _, tt := range tests {
		city := &City{ID: "test", Population: tt.population}
		d := g.Demand(city)
		if d < tt.minDemand || d > tt.maxDemand {
			t.Errorf("Population=%d: Demand=%d, want %d〜%d",
				tt.population, d, tt.minDemand, tt.maxDemand)
		}
	}
}

// TestEconomy_KobeOsaka は神戸〜梅田（大都市間、30.5km、運賃 290 円）の
// 統合シナリオを検証する。
//
// Phase 1-7.7 のバランス目標: 営業係数 90〜120（黒字ギリギリ）。
//
// 人口べき乗則（指数 0.6）により、神戸（40 万）・梅田（12 万）の日次需要は
// それぞれ 32 / 15（1 月）となり、旧線形モデルの 400 / 120 から大幅に
// 圧縮される。旧モデルでは神戸〜梅田の営業係数が約 29（過剰に黒字）だった。
//
// 神戸〜梅田は阪急・阪神・国鉄が競合する路線であり、競合線（運賃 200 円）
// の存在でシェアが約 0.41 に下がるため、営業係数は 90〜120 に収まる。
// 需要のべき乗則化と競合によるシェア低下の両方が、大都市間路線を
// 現実的な採算水準に押し上げる。
func TestEconomy_KobeOsaka(t *testing.T) {
	g := NewGame(1907, "hankyu")
	g.Date = Date{Year: 1910, Month: 1}

	g.AddCity(&City{ID: "umeda", Name: "梅田", Population: 120000})
	g.AddCity(&City{ID: "kobe", Name: "神戸", Population: 400000})

	hankyu := NewCompany("hankyu", "hankyu", "阪急")
	hankyu.Cash = 20_000_000
	g.AddCompany(hankyu)

	hanshin := NewCompany("hanshin", "hanshin", "阪神")
	hanshin.Cash = 20_000_000
	g.AddCompany(hanshin)

	// 阪急線（自社、運賃 290 円）
	line := NewLine("kobe-line", hankyu.ID, "kobe", "umeda")
	line.LengthKm = 30.5
	line.Fare = 290
	line.TrainsPerDay = 10
	hankyu.AddLine(line)

	// 競合線（阪神、運賃 200 円で競争力が高い）
	rival := NewLine("rival-line", hanshin.ID, "kobe", "umeda")
	rival.LengthKm = 30.5
	rival.Fare = 200
	rival.TrainsPerDay = 10
	hanshin.AddLine(rival)

	// 12 ヶ月シミュレート
	for i := 0; i < 12; i++ {
		pax := g.linePassengers(line)
		revenue, maint := g.simulateLine(line, hankyu)
		ratio := 0.0
		if revenue > 0 {
			ratio = float64(maint) / float64(revenue) * 100.0
		}
		t.Logf("  %s: 旅客 %d人/月（シェア %.3f）, 収入 %d円, 維持費 %d円, 営業係数 %.1f",
			g.Date.String(), pax, g.lineShare(line), revenue, maint, ratio)
		g.EndTurn()
	}

	t.Logf("[神戸〜梅田] 距離 %.1fkm, 運賃 %d円, シェア %.3f, 最終月 営業係数 %.1f",
		line.LengthKm, line.Fare, g.lineShare(line), line.OperatingRatio)

	// 営業係数 90-120（黒字ギリギリ）
	if line.OperatingRatio < 90 || line.OperatingRatio > 120 {
		t.Errorf("神戸〜梅田の営業係数 = %.1f, want 90-120", line.OperatingRatio)
	}
}
