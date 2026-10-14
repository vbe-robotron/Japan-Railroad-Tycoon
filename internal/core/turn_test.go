package core

import (
	"testing"
)

// countByType は events 中の指定タイプの数を数える。
func countByType(events []Event, typ string) int {
	n := 0
	for _, e := range events {
		if e.Type == typ {
			n++
		}
	}
	return n
}

func TestEndTurn_SingleTurn(t *testing.T) {
	g := NewGame(1907, "hankyu")
	g.Date = Date{Year: 1910, Month: 3}

	events := g.EndTurn()

	// 時間が1ヶ月進む
	if g.Date.Year != 1910 || g.Date.Month != 4 {
		t.Errorf("Date = %v, want 1910年4月", g.Date)
	}
	// ターンがインクリメントされる
	if g.Turn != 1 {
		t.Errorf("Turn = %d, want 1", g.Turn)
	}

	// turn_ended が1つ、year_changed は0
	if got := countByType(events, EventTurnEnded); got != 1 {
		t.Errorf("turn_ended count = %d, want 1", got)
	}
	if got := countByType(events, EventYearChanged); got != 0 {
		t.Errorf("year_changed count = %d, want 0（年越しでない）", got)
	}

	// turn_ended は時間進行前の Date/Turn を持つ
	e := events[0]
	if e.Type != EventTurnEnded {
		t.Fatalf("events[0].Type = %q, want %q", e.Type, EventTurnEnded)
	}
	if !e.Date.Equal(Date{Year: 1910, Month: 3}) {
		t.Errorf("turn_ended.Date = %v, want 1910年3月（進行前）", e.Date)
	}
	if e.Turn != 0 {
		t.Errorf("turn_ended.Turn = %d, want 0（進行前）", e.Turn)
	}
	wantMsg := "1910年3月 ターン終了"
	if e.Message != wantMsg {
		t.Errorf("turn_ended.Message = %q, want %q", e.Message, wantMsg)
	}
	if e.Payload["turn"] != 0 {
		t.Errorf("turn_ended.Payload[turn] = %v, want 0", e.Payload["turn"])
	}

	// イベントがゲームに記録されている
	if len(g.Events) != len(events) {
		t.Errorf("g.Events の数 = %d, want %d", len(g.Events), len(events))
	}
}

func TestEndTurn_12Turns(t *testing.T) {
	g := NewGame(1907, "hankyu")
	g.Date = Date{Year: 1910, Month: 1}
	g.AddCompany(NewCompany("hankyu", "hankyu", "阪急"))

	var turnEndedCount, yearChangedCount int
	for i := 0; i < 12; i++ {
		events := g.EndTurn()
		for _, e := range events {
			switch e.Type {
			case EventTurnEnded:
				turnEndedCount++
			case EventYearChanged:
				yearChangedCount++
			}
		}
	}

	if g.Date.Year != 1911 || g.Date.Month != 1 {
		t.Errorf("Date = %v, want 1911年1月", g.Date)
	}
	if turnEndedCount != 12 {
		t.Errorf("turn_ended count = %d, want 12", turnEndedCount)
	}
	if yearChangedCount != 1 {
		t.Errorf("year_changed count = %d, want 1", yearChangedCount)
	}
	if g.Turn != 12 {
		t.Errorf("Turn = %d, want 12", g.Turn)
	}
}

func TestEndTurn_YearChanged(t *testing.T) {
	g := NewGame(1907, "hankyu")
	g.Date = Date{Year: 1910, Month: 12}

	events := g.EndTurn()

	if g.Date.Year != 1911 || g.Date.Month != 1 {
		t.Errorf("Date = %v, want 1911年1月", g.Date)
	}

	var yc *Event
	for i := range events {
		if events[i].Type == EventYearChanged {
			yc = &events[i]
		}
	}
	if yc == nil {
		t.Fatal("year_changed イベントが生成されなかった")
	}
	// year_changed は時間進行後の Date/Turn を持つ
	if !yc.Date.Equal(Date{Year: 1911, Month: 1}) {
		t.Errorf("year_changed.Date = %v, want 1911年1月（進行後）", yc.Date)
	}
	if yc.Turn != 1 {
		t.Errorf("year_changed.Turn = %d, want 1（進行後）", yc.Turn)
	}
	wantMsg := "1911年になりました"
	if yc.Message != wantMsg {
		t.Errorf("year_changed.Message = %q, want %q", yc.Message, wantMsg)
	}
	if yc.Payload["year"] != 1911 {
		t.Errorf("year_changed.Payload[year] = %v, want 1911", yc.Payload["year"])
	}
}

func TestEndTurn_EventCount(t *testing.T) {
	g := NewGame(1907, "hankyu")
	g.Date = Date{Year: 1910, Month: 1}

	// 12ターン回すと、turn_ended 12 + year_changed 1 = 13 イベント。
	for i := 0; i < 12; i++ {
		g.EndTurn()
	}

	if len(g.Events) != 13 {
		t.Errorf("len(g.Events) = %d, want 13", len(g.Events))
	}
	if got := countByType(g.Events, EventTurnEnded); got != 12 {
		t.Errorf("g.Events の turn_ended = %d, want 12", got)
	}
	if got := countByType(g.Events, EventYearChanged); got != 1 {
		t.Errorf("g.Events の year_changed = %d, want 1", got)
	}
}

func TestEndTurn_Deterministic(t *testing.T) {
	a := NewGame(1907, "hankyu")
	b := NewGame(1907, "hankyu")
	a.Date = Date{Year: 1910, Month: 1}
	b.Date = Date{Year: 1910, Month: 1}

	for i := 0; i < 24; i++ {
		ea := a.EndTurn()
		eb := b.EndTurn()
		if len(ea) != len(eb) {
			t.Fatalf("iteration %d: events の数が一致しない: %d != %d", i, len(ea), len(eb))
		}
		for j := range ea {
			if ea[j].Type != eb[j].Type {
				t.Errorf("iteration %d event %d: Type = %q, want %q", i, j, ea[j].Type, eb[j].Type)
			}
			if ea[j].Message != eb[j].Message {
				t.Errorf("iteration %d event %d: Message = %q, want %q", i, j, ea[j].Message, eb[j].Message)
			}
			if ea[j].ID != eb[j].ID {
				t.Errorf("iteration %d event %d: ID = %q, want %q", i, j, ea[j].ID, eb[j].ID)
			}
		}
	}
}

// endTurnEcoGame は経済計算の検証用に路線を 1 本持つゲームを返す。
func endTurnEcoGame(t *testing.T) (*Game, *Company, *Line) {
	t.Helper()
	g := NewGame(1907, "hankyu")
	g.Date = Date{Year: 1910, Month: 6}
	g.AddCity(&City{ID: "a", Name: "A", Population: 100000, Lat: 35.0, Lon: 135.0})
	g.AddCity(&City{ID: "b", Name: "B", Population: 100000, Lat: 35.0, Lon: 135.2})

	c := NewCompany("hankyu", "hankyu", "阪急")
	c.Cash = 10_000_000
	g.AddCompany(c)

	act := &BuildLineAction{From: "a", To: "b", Gauge: 1067}
	if _, err := act.Apply(g); err != nil {
		t.Fatalf("BuildLineAction: %v", err)
	}
	return g, c, c.Lines[0]
}

func TestEndTurn_WithEconomy(t *testing.T) {
	g, _, line := endTurnEcoGame(t)

	// 敷設直後はまだ経済計算されていない
	if line.Profit != 0 {
		t.Fatalf("敷設直後の Profit = %d, want 0", line.Profit)
	}

	g.EndTurn()

	// EndTurn で経済計算が走り、Line.Profit が更新される
	if line.Profit == 0 {
		t.Errorf("EndTurn 後に Line.Profit = 0（経済計算が呼ばれていない）")
	}
	if line.OperatingRatio == 0 {
		t.Errorf("EndTurn 後に Line.OperatingRatio = 0（更新されていない）")
	}
	if !hasEventType(g.Events, EventTurnEnded) {
		t.Error("turn_ended イベントが生成されなかった")
	}
}

func TestEndTurn_CashChange(t *testing.T) {
	g, c, _ := endTurnEcoGame(t)

	before := c.Cash
	g.EndTurn()

	if c.Cash == before {
		t.Errorf("資金が変化していない: %d", c.Cash)
	}
	// 黒字路線なので資金は増えるはず
	if c.Cash <= before {
		t.Errorf("黒字路線で資金が増えていない: before=%d, after=%d", before, c.Cash)
	}
}

func TestEndTurn_WithAI(t *testing.T) {
	g, c := newAITestGame(t, "hankyu", 100_000_000, &HistoricalLine{
		CompanyID: "hankyu", From: "a", To: "b", Gauge: 1067,
		OpenedYear: 1910, OpenedMonth: 3, DisplayName: "テスト線",
	})
	g.InitAIs()

	events := g.EndTurn()

	if len(c.Lines) != 1 {
		t.Fatalf("AI が路線を敷設していない: Lines = %d, want 1", len(c.Lines))
	}
	if !hasEventType(events, EventLineOpened) {
		t.Error("line_opened イベントが生成されなかった")
	}
	// turn_ended は AI 行動の後にも必ず生成される
	if !hasEventType(events, EventTurnEnded) {
		t.Error("turn_ended イベントが生成されなかった")
	}
}

func TestEndTurn_WithAI_Deterministic(t *testing.T) {
	setup := func() *Game {
		g, _ := newAITestGame(t, "hankyu", 100_000_000, &HistoricalLine{
			CompanyID: "hankyu", From: "a", To: "b", Gauge: 1067,
			OpenedYear: 1910, OpenedMonth: 3, DisplayName: "テスト線",
		})
		g.InitAIs()
		return g
	}
	a := setup()
	b := setup()
	for i := 0; i < 3; i++ {
		ea := a.EndTurn()
		eb := b.EndTurn()
		if len(ea) != len(eb) {
			t.Fatalf("iteration %d: events 数が不一致", i)
		}
		for j := range ea {
			if ea[j].ID != eb[j].ID || ea[j].Message != eb[j].Message {
				t.Errorf("iteration %d event %d: 不一致", i, j)
			}
		}
	}
}

// newAdaptiveGame は適応 AI の 100 ターンテスト用のゲームを返す。
func newAdaptiveGame(seed int64) (*Game, *Company) {
	g := NewGame(seed, "player")
	g.Date = Date{Year: 1910, Month: 1}

	coords := map[string][2]float64{
		"umeda": {34.70, 135.50},
		"kobe":  {34.69, 135.20},
		"kyoto": {35.01, 135.77},
		"osaka": {34.65, 135.50},
		"nara":  {34.68, 135.80},
		"otsu":  {35.00, 135.87},
	}
	for id, ll := range coords {
		g.AddCity(&City{ID: id, Name: id, Population: 1_000_000, LandValue: 50, Lat: ll[0], Lon: ll[1]})
	}

	c := NewCompany("hankyu", "hankyu", "阪急")
	c.Cash = 200_000_000
	c.CreditLine = 50_000_000
	c.Region = "kansai"
	g.AddCompany(c)

	g.AIs["hankyu"] = &AI{
		CompanyID:   "hankyu",
		Personality: DefaultPersonality("hankyu"),
		HistoryMode: HistoryIgnore,
	}
	return g, c
}

func TestEndTurn_WithAdaptiveAI(t *testing.T) {
	g, c := newAdaptiveGame(1907)

	for i := 0; i < 100; i++ {
		g.EndTurn()
	}

	if g.Date.Year != 1918 || g.Date.Month != 5 {
		t.Errorf("Date = %v, want 1918年5月", g.Date)
	}
	if len(c.Lines) < 2 {
		t.Errorf("AI が路線網を拡大していない: Lines = %d", len(c.Lines))
	}
	t.Logf("100ターン後: 路線 %d 本, 施設 %d, 資金 %s, 借金 %s",
		len(c.Lines), len(c.Facilities), FormatYen(c.Cash), FormatYen(c.Debt))
}

func TestEndTurn_WithAdaptiveAI_Deterministic(t *testing.T) {
	a, _ := newAdaptiveGame(1907)
	b, _ := newAdaptiveGame(1907)

	for i := 0; i < 36; i++ {
		ea := a.EndTurn()
		eb := b.EndTurn()
		if len(ea) != len(eb) {
			t.Fatalf("iteration %d: events 数が不一致: %d != %d", i, len(ea), len(eb))
		}
		for j := range ea {
			if ea[j].ID != eb[j].ID || ea[j].Message != eb[j].Message || ea[j].Type != eb[j].Type {
				t.Fatalf("iteration %d event %d: 不一致: %+v vs %+v", i, j, ea[j], eb[j])
			}
		}
	}
}

func TestEndTurn_NoLines(t *testing.T) {
	g := NewGame(1907, "hankyu")
	g.Date = Date{Year: 1910, Month: 6}
	c := NewCompany("hankyu", "hankyu", "阪急")
	c.Cash = 1_000_000
	g.AddCompany(c)

	before := c.Cash
	g.EndTurn()

	if c.Cash != before {
		t.Errorf("路線なしで資金が変化した: before=%d, after=%d", before, c.Cash)
	}
}

// newMAGame は M&A・廃線 AI の長期テスト用のゲームを返す。
//
// プレイヤー 1 社、高攻撃性の AI 3 社、低攻撃性で赤字路線を持つ対象 1 社を用意する。
func newMAGame(seed int64) *Game {
	g := NewGame(seed, "player")
	g.Date = Date{Year: 1910, Month: 1}

	cities := map[string][2]float64{
		"umeda": {34.70, 135.50},
		"kobe":  {34.69, 135.20},
		"kyoto": {35.01, 135.77},
		"osaka": {34.65, 135.50},
		"nara":  {34.68, 135.80},
		"otsu":  {35.00, 135.87},
	}
	for id, ll := range cities {
		g.AddCity(&City{ID: id, Name: id, Population: 1_000_000, LandValue: 50, Lat: ll[0], Lon: ll[1]})
	}

	player := NewCompany("player", "player", "プレイヤー")
	player.IsPlayer = true
	player.Cash = 100_000_000
	g.AddCompany(player)

	aggressors := []string{"hankyu", "tokyu", "meitetsu"}
	coords := [][2]string{{"umeda", "kobe"}, {"kyoto", "otsu"}, {"osaka", "nara"}}
	for i, id := range aggressors {
		c := NewCompany(id, id, id)
		c.Cash = 500_000_000
		c.CreditLine = 200_000_000
		c.TotalShares = 10000
		c.StockPrice = 1000
		c.Region = "kansai"
		g.AddCompany(c)
		line := NewLine("line_"+id, id, coords[i][0], coords[i][1])
		line.LengthKm = g.DistanceKm(coords[i][0], coords[i][1])
		c.AddLine(line)
	}

	hanshin := NewCompany("hanshin", "hanshin", "阪神")
	hanshin.Cash = 100_000_000
	hanshin.TotalShares = 10000
	hanshin.StockPrice = 1000
	hanshin.Region = "kansai"
	g.AddCompany(hanshin)
	defLine := NewLine("line_hanshin", "hanshin", "osaka", "otsu")
	defLine.LengthKm = g.DistanceKm("osaka", "otsu")
	defLine.ConsecutiveDeficitTurns = AbandonDeficitTurns
	hanshin.AddLine(defLine)

	g.InitAIs()
	for _, ai := range g.AIs {
		if ai != nil {
			ai.HistoryMode = HistoryIgnore
		}
	}
	return g
}

func TestEndTurn_WithMA(t *testing.T) {
	g := newMAGame(1907)
	beforeActive := len(g.Companies)

	for i := 0; i < 200; i++ {
		g.EndTurn()
	}

	dissolved := 0
	for _, c := range g.Companies {
		if c.Dissolved != nil {
			dissolved++
		}
	}
	active := len(g.Companies) - dissolved

	maCount := 0
	for _, e := range g.Events {
		switch e.Type {
		case EventSharesAcquired, EventTOBLaunched, EventMergeCompleted:
			maCount++
		}
	}
	abandonCount := countByType(g.Events, EventLineAbandoned)

	if dissolved == 0 {
		t.Error("200ターンで M&A が発生しなかった（合併による会社消滅なし）")
	}
	if maCount == 0 {
		t.Error("M&A イベントが記録されていない")
	}
	if abandonCount == 0 {
		t.Error("廃線が発生しなかった")
	}
	t.Logf("200ターン後: 稼働会社 %d → %d（解散 %d）, M&A %d 件, 廃線 %d 件",
		beforeActive, active, dissolved, maCount, abandonCount)
}

func TestEndTurn_WithMA_Deterministic(t *testing.T) {
	a := newMAGame(1907)
	b := newMAGame(1907)

	for i := 0; i < 60; i++ {
		ea := a.EndTurn()
		eb := b.EndTurn()
		if len(ea) != len(eb) {
			t.Fatalf("iteration %d: events 数が不一致: %d != %d", i, len(ea), len(eb))
		}
		for j := range ea {
			if ea[j].ID != eb[j].ID || ea[j].Message != eb[j].Message || ea[j].Type != eb[j].Type {
				t.Fatalf("iteration %d event %d: 不一致: %+v vs %+v", i, j, ea[j], eb[j])
			}
		}
	}
}
