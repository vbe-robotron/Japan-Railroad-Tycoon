package core

import (
	"errors"
	"reflect"
	"testing"
)

// buildLineActionGame は BuildLineAction テスト用のゲームを返す。
func buildLineActionGame(year int, cash int64) *Game {
	return constructionGame(year, cash)
}

func TestBuildLineAction_Type(t *testing.T) {
	a := &BuildLineAction{}
	if got := a.Type(); got != "build_line" {
		t.Errorf("Type() = %q, want %q", got, "build_line")
	}
}

func TestBuildLineAction_Validate_NilGame(t *testing.T) {
	a := &BuildLineAction{From: "umeda", To: "takarazuka", Gauge: 1067}
	err := a.Validate(nil)
	if err == nil {
		t.Fatal("nil ゲームでエラーにならなかった")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("型が *ValidationError でない: %T", err)
	}
	if ve.Code != "GAME_NIL" {
		t.Errorf("Code = %q, want GAME_NIL", ve.Code)
	}
}

func TestBuildLineAction_Validate_NoPlayer(t *testing.T) {
	g := buildLineActionGame(1910, 100_000_000)
	delete(g.Companies, "hankyu") // プレイヤー会社を削除

	a := &BuildLineAction{From: "umeda", To: "takarazuka", Gauge: 1067}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("プレイヤー会社なしでエラーにならなかった")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("型が *ValidationError でない: %T", err)
	}
	if ve.Code != "NO_PLAYER" {
		t.Errorf("Code = %q, want NO_PLAYER", ve.Code)
	}
	if !errors.Is(ve, ErrNoPlayer) {
		t.Errorf("ErrNoPlayer をラップしていない: %v", err)
	}
}

func TestBuildLineAction_Validate_SameCity(t *testing.T) {
	g := buildLineActionGame(1910, 100_000_000)
	a := &BuildLineAction{From: "umeda", To: "umeda", Gauge: 1067}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("同一都市でエラーにならなかった")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("型が *ValidationError でない: %T", err)
	}
	if ve.Code != "SAME_CITY" {
		t.Errorf("Code = %q, want SAME_CITY", ve.Code)
	}
}

func TestBuildLineAction_Validate_CityNotFound(t *testing.T) {
	g := buildLineActionGame(1910, 100_000_000)

	tests := []struct {
		name string
		from string
		to   string
		code string
	}{
		{"出発都市なし", "missing", "takarazuka", "FROM_NOT_FOUND"},
		{"目的地都市なし", "umeda", "missing", "TO_NOT_FOUND"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &BuildLineAction{From: tt.from, To: tt.to, Gauge: 1067}
			err := a.Validate(g)
			if err == nil {
				t.Fatal("都市なしでエラーにならなかった")
			}
			ve, ok := err.(*ValidationError)
			if !ok {
				t.Fatalf("型が *ValidationError でない: %T", err)
			}
			if ve.Code != tt.code {
				t.Errorf("Code = %q, want %q", ve.Code, tt.code)
			}
			if !errors.Is(ve, ErrCityNotFound) {
				t.Errorf("ErrCityNotFound をラップしていない: %v", err)
			}
		})
	}
}

func TestBuildLineAction_Validate_LineExists(t *testing.T) {
	g := buildLineActionGame(1910, 100_000_000)
	// 既存路線を追加（逆向きでも重複とみなす）
	g.PlayerCompany().AddLine(NewLine("L1", "hankyu", "takarazuka", "umeda"))

	a := &BuildLineAction{From: "umeda", To: "takarazuka", Gauge: 1067}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("既存路線でエラーにならなかった")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("型が *ValidationError でない: %T", err)
	}
	if ve.Code != "LINE_EXISTS" {
		t.Errorf("Code = %q, want LINE_EXISTS", ve.Code)
	}
	if !errors.Is(ve, ErrLineExists) {
		t.Errorf("ErrLineExists をラップしていない: %v", err)
	}
}

func TestBuildLineAction_Validate_TechnologyFail(t *testing.T) {
	g := buildLineActionGame(1880, 100_000_000) // 蒸気鉄道は 1887
	a := &BuildLineAction{From: "umeda", To: "takarazuka", Gauge: 1067}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("技術未解禁でエラーにならなかった")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("型が *ValidationError でない: %T", err)
	}
	if ve.Code != "GATE_TECHNOLOGY" {
		t.Errorf("Code = %q, want GATE_TECHNOLOGY", ve.Code)
	}
	if !errors.Is(ve, ErrRestricted) {
		t.Errorf("ErrRestricted をラップしていない: %v", err)
	}
}

func TestBuildLineAction_Validate_RegulationFail(t *testing.T) {
	g := buildLineActionGame(1886, 100_000_000) // 私設鉄道条例は 1887
	a := &BuildLineAction{From: "umeda", To: "takarazuka", Gauge: 762}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("法規未制定でエラーにならなかった")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("型が *ValidationError でない: %T", err)
	}
	if ve.Code != "GATE_REGULATION" {
		t.Errorf("Code = %q, want GATE_REGULATION", ve.Code)
	}
}

func TestBuildLineAction_Validate_EconomyFail(t *testing.T) {
	g := buildLineActionGame(1910, 0) // 資金 0
	a := &BuildLineAction{From: "umeda", To: "takarazuka", Gauge: 1067}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("資金不足でエラーにならなかった")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("型が *ValidationError でない: %T", err)
	}
	if ve.Code != "GATE_ECONOMY" {
		t.Errorf("Code = %q, want GATE_ECONOMY", ve.Code)
	}
}

func TestBuildLineAction_Validate_OK(t *testing.T) {
	g := buildLineActionGame(1910, 100_000_000)
	a := &BuildLineAction{From: "umeda", To: "takarazuka", Gauge: 1067}
	if err := a.Validate(g); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}

func TestBuildLineAction_Validate_NoSideEffect(t *testing.T) {
	g := buildLineActionGame(1910, 100_000_000)
	cashBefore := g.PlayerCompany().Cash
	linesBefore := len(g.PlayerCompany().Lines)

	a := &BuildLineAction{From: "umeda", To: "takarazuka", Gauge: 1067}
	if err := a.Validate(g); err != nil {
		t.Fatalf("Validate() = %v, want nil", err)
	}

	if g.PlayerCompany().Cash != cashBefore {
		t.Errorf("Validate が資金を変更した: %d → %d", cashBefore, g.PlayerCompany().Cash)
	}
	if len(g.PlayerCompany().Lines) != linesBefore {
		t.Errorf("Validate が路線を追加した")
	}
	if g.Turn != 0 || !g.Date.Equal(Date{Year: 1910, Month: 6}) {
		t.Errorf("Validate がゲーム状態を変更した: Turn=%d Date=%v", g.Turn, g.Date)
	}
}

func TestBuildLineAction_Apply(t *testing.T) {
	const fund = 100_000_000
	g := buildLineActionGame(1910, fund)
	a := &BuildLineAction{From: "umeda", To: "takarazuka", Gauge: 1067, Double: false}

	events, err := a.Apply(g)
	if err != nil {
		t.Fatalf("Apply() = %v, want nil", err)
	}

	// 資金減算
	wantCash := fund - g.LineCost("umeda", "takarazuka", 1067, false)
	if g.PlayerCompany().Cash != wantCash {
		t.Errorf("Cash = %d, want %d", g.PlayerCompany().Cash, wantCash)
	}

	// Line 追加
	if len(g.PlayerCompany().Lines) != 1 {
		t.Fatalf("Lines = %d, want 1", len(g.PlayerCompany().Lines))
	}

	// イベント生成
	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}
	ev := events[0]
	if ev.Type != EventLineOpened {
		t.Errorf("Type = %q, want %q", ev.Type, EventLineOpened)
	}
	if ev.Severity != SeveritySuccess {
		t.Errorf("Severity = %q, want %q", ev.Severity, SeveritySuccess)
	}
	if ev.CompanyID != "hankyu" {
		t.Errorf("CompanyID = %q, want hankyu", ev.CompanyID)
	}
	if ev.LineID != g.PlayerCompany().Lines[0].ID {
		t.Errorf("LineID = %q, want %q", ev.LineID, g.PlayerCompany().Lines[0].ID)
	}
	if ev.Message == "" {
		t.Error("Message が空")
	}
}

func TestBuildLineAction_Apply_Invalid(t *testing.T) {
	g := buildLineActionGame(1910, 100_000_000)
	// 同一都市なので Validate が失敗する
	a := &BuildLineAction{From: "umeda", To: "umeda", Gauge: 1067}
	events, err := a.Apply(g)
	if err == nil {
		t.Fatal("Apply() = nil, want error")
	}
	if events != nil {
		t.Errorf("events = %v, want nil", events)
	}
	// 状態が変化していない
	if len(g.PlayerCompany().Lines) != 0 {
		t.Error("失敗時に Line が追加された")
	}
	if g.PlayerCompany().Cash != 100_000_000 {
		t.Error("失敗時に資金が変化した")
	}
}

func TestBuildLineAction_Apply_Cost(t *testing.T) {
	g := buildLineActionGame(1910, 100_000_000)
	a := &BuildLineAction{From: "umeda", To: "takarazuka", Gauge: 1435, Double: true}

	wantCost := g.LineCost("umeda", "takarazuka", 1435, true)
	events, err := a.Apply(g)
	if err != nil {
		t.Fatalf("Apply() = %v", err)
	}
	if events[0].Payload["cost"] != wantCost {
		t.Errorf("Payload[cost] = %v, want %d", events[0].Payload["cost"], wantCost)
	}
	if events[0].Payload["months"] != g.buildMonths("umeda", "takarazuka") {
		t.Errorf("Payload[months] = %v, want %d",
			events[0].Payload["months"], g.buildMonths("umeda", "takarazuka"))
	}
}

func TestBuildLineAction_Apply_LineFields(t *testing.T) {
	g := buildLineActionGame(1910, 100_000_000)
	a := &BuildLineAction{From: "umeda", To: "takarazuka", Gauge: 1435, Double: true}

	if _, err := a.Apply(g); err != nil {
		t.Fatalf("Apply() = %v", err)
	}
	line := g.PlayerCompany().Lines[0]

	if line.CompanyID != "hankyu" {
		t.Errorf("CompanyID = %q, want hankyu", line.CompanyID)
	}
	if line.From != "umeda" || line.To != "takarazuka" {
		t.Errorf("From/To = %q/%q, want umeda/takarazuka", line.From, line.To)
	}
	if line.DisplayName != "梅田〜宝塚" {
		t.Errorf("DisplayName = %q, want 梅田〜宝塚", line.DisplayName)
	}
	if line.LengthKm != g.DistanceKm("umeda", "takarazuka") {
		t.Errorf("LengthKm = %v, want %v", line.LengthKm, g.DistanceKm("umeda", "takarazuka"))
	}
	if line.Gauge != 1435 {
		t.Errorf("Gauge = %d, want 1435", line.Gauge)
	}
	if !line.Double {
		t.Error("Double = false, want true")
	}
	if !line.Opened.Equal(Date{Year: 1910, Month: 6}) {
		t.Errorf("Opened = %v, want 1910年6月", line.Opened)
	}
	if line.Fare != g.DefaultFare("umeda", "takarazuka") {
		t.Errorf("Fare = %d, want %d", line.Fare, g.DefaultFare("umeda", "takarazuka"))
	}
	if line.TrainsPerDay != 10 {
		t.Errorf("TrainsPerDay = %d, want 10", line.TrainsPerDay)
	}
	// 健全度は NewLine の既定値
	if line.Condition != 1.0 {
		t.Errorf("Condition = %v, want 1.0", line.Condition)
	}
}

func TestBuildLineAction_Apply_DemandRecalc(t *testing.T) {
	g := buildLineActionGame(1910, 100_000_000)
	from := g.Cities["umeda"]
	to := g.Cities["takarazuka"]
	// 事前に需要をゼロにしておき、再計算されたことを確認する
	from.Demand = 0
	to.Demand = 0

	a := &BuildLineAction{From: "umeda", To: "takarazuka", Gauge: 1067}
	if _, err := a.Apply(g); err != nil {
		t.Fatalf("Apply() = %v", err)
	}

	// 接続数 1 → ボーナス 1.1
	wantFrom := int(float64(g.Demand(from)) * 1.1)
	wantTo := int(float64(g.Demand(to)) * 1.1)
	if from.Demand != wantFrom {
		t.Errorf("from.Demand = %d, want %d", from.Demand, wantFrom)
	}
	if to.Demand != wantTo {
		t.Errorf("to.Demand = %d, want %d", to.Demand, wantTo)
	}
}

func TestBuildLineAction_Apply_ViaRegistry(t *testing.T) {
	g := buildLineActionGame(1910, 100_000_000)

	a, err := DecodeAction("build_line",
		[]byte(`{"from":"umeda","to":"takarazuka","gauge":1067,"double":true}`))
	if err != nil {
		t.Fatalf("DecodeAction: %v", err)
	}
	events, err := a.Apply(g)
	if err != nil {
		t.Fatalf("Apply() = %v", err)
	}
	if len(events) != 1 || events[0].Type != EventLineOpened {
		t.Errorf("イベントが不正: %+v", events)
	}
	if len(g.PlayerCompany().Lines) != 1 {
		t.Errorf("Lines = %d, want 1", len(g.PlayerCompany().Lines))
	}
}

func TestBuildLineAction_Deterministic(t *testing.T) {
	run := func() (*Game, []Event) {
		g := buildLineActionGame(1910, 100_000_000)
		a := &BuildLineAction{From: "umeda", To: "takarazuka", Gauge: 1067, Double: true}
		events, err := a.Apply(g)
		if err != nil {
			t.Fatalf("Apply() = %v", err)
		}
		return g, events
	}

	g1, ev1 := run()
	g2, ev2 := run()

	if g1.PlayerCompany().Cash != g2.PlayerCompany().Cash {
		t.Errorf("Cash が一致しない: %d != %d", g1.PlayerCompany().Cash, g2.PlayerCompany().Cash)
	}
	if g1.PlayerCompany().Lines[0].ID != g2.PlayerCompany().Lines[0].ID {
		t.Errorf("LineID が一致しない: %q != %q",
			g1.PlayerCompany().Lines[0].ID, g2.PlayerCompany().Lines[0].ID)
	}
	if !reflect.DeepEqual(ev1, ev2) {
		t.Errorf("イベントが一致しない:\n%+v\n%+v", ev1, ev2)
	}
}

// TestBuildLine_Scenario は阪急が梅田〜宝塚を敷設する統合シナリオを検証する。
//
// Phase 1-7.6 では実座標（梅田・宝塚）を与え、19.1km の現実的な距離で
// 敷設コスト（BaseCostPerKm=250000 で約 477 万円）を検証する。
// 資金は建設費を賄い、なお余裕が残る 1,000 万円とする
// （建設費 / 初期資金 ≒ 48%）。
func TestBuildLine_Scenario(t *testing.T) {
	const (
		seed = int64(1907)
		fund = int64(10_000_000)
	)
	g := NewGame(seed, "hankyu")
	umeda := NewCity("umeda", "梅田", "大阪")
	umeda.Lat, umeda.Lon = 34.7055, 135.4983
	takarazuka := NewCity("takarazuka", "宝塚", "兵庫")
	takarazuka.Lat, takarazuka.Lon = 34.7995, 135.3468
	g.AddCity(umeda)
	g.AddCity(takarazuka)

	hankyu := NewCompany("hankyu", "hankyu", "阪急")
	hankyu.Cash = fund
	g.AddCompany(hankyu)

	g.Date = Date{Year: 1910, Month: 3}

	cost := g.LineCost("umeda", "takarazuka", 1067, false)
	a := &BuildLineAction{From: "umeda", To: "takarazuka", Gauge: 1067, Double: false}

	events, err := a.Apply(g)
	if err != nil {
		t.Fatalf("Apply() = %v", err)
	}

	// 資金減算
	if hankyu.Cash != fund-cost {
		t.Errorf("Cash = %d, want %d", hankyu.Cash, fund-cost)
	}
	// Line 追加
	if len(hankyu.Lines) != 1 {
		t.Fatalf("Lines = %d, want 1", len(hankyu.Lines))
	}
	line := hankyu.Lines[0]
	if line.From != "umeda" || line.To != "takarazuka" {
		t.Errorf("From/To = %q/%q", line.From, line.To)
	}
	if line.Opened != g.Date {
		t.Errorf("Opened = %v, want %v", line.Opened, g.Date)
	}
	// イベント生成
	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}
	if events[0].Type != EventLineOpened {
		t.Errorf("Type = %q, want %q", events[0].Type, EventLineOpened)
	}
	if events[0].Payload["cost"] != cost {
		t.Errorf("Payload[cost] = %v, want %d", events[0].Payload["cost"], cost)
	}
	// 地名を含むメッセージ
	if events[0].Message == "" {
		t.Error("Message が空")
	}
	t.Logf("コスト=%s イベント=%q", FormatYen(cost), events[0].Message)
}
