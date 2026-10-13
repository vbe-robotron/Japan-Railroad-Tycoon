package core

import (
	"errors"
	"testing"
)

// このファイルは SetFareAction の TDD テストである。
//
// 運賃変更の制約（1〜1000円、±50%以内）と、運賃が需要・利益に
// 与える影響（弾力性 -0.7 の間接反映）を検証する。

// --- Type ---

func TestSetFareAction_Type(t *testing.T) {
	a := &SetFareAction{}
	if got := a.Type(); got != "set_fare" {
		t.Errorf("Type() = %q, want %q", got, "set_fare")
	}
}

// --- Validate ---

func TestSetFareAction_Validate_NilGame(t *testing.T) {
	a := &SetFareAction{LineID: "L1", Fare: 200}
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

func TestSetFareAction_Validate_NoPlayer(t *testing.T) {
	g := NewGame(1907, "hankyu")
	a := &SetFareAction{LineID: "L1", Fare: 200}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("プレイヤーなしでエラーにならなかった")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("型が *ValidationError でない: %T", err)
	}
	if ve.Code != "NO_PLAYER" {
		t.Errorf("Code = %q, want NO_PLAYER", ve.Code)
	}
	if !errors.Is(ve, ErrNoPlayer) {
		t.Errorf("ErrNoPlayer をラップしていない")
	}
}

func TestSetFareAction_Validate_LineNotFound(t *testing.T) {
	g := setupFareTestGame()
	a := &SetFareAction{LineID: "not_exist", Fare: 200}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("存在しない路線でエラーにならなかった")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("型が *ValidationError でない: %T", err)
	}
	if ve.Code != "LINE_NOT_FOUND" {
		t.Errorf("Code = %q, want LINE_NOT_FOUND", ve.Code)
	}
	if !errors.Is(ve, ErrNotFound) {
		t.Errorf("ErrNotFound をラップしていない")
	}
}

func TestSetFareAction_Validate_NotOwned(t *testing.T) {
	g := setupFareTestGame()
	// 他社の路線を追加
	other := NewCompany("other", "other", "他社")
	g.AddCompany(other)
	line := NewLine("L_other", "other", "umeda", "takarazuka")
	line.Fare = 200
	other.AddLine(line)

	a2 := &SetFareAction{LineID: "L_other", Fare: 250}
	err := a2.Validate(g)
	if err == nil {
		t.Fatal("他社路線の運賃変更でエラーにならなかった")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("型が *ValidationError でない: %T", err)
	}
	if ve.Code != "NOT_OWNED" {
		t.Errorf("Code = %q, want NOT_OWNED", ve.Code)
	}
	if !errors.Is(ve, ErrUnauthorized) {
		t.Errorf("ErrUnauthorized をラップしていない")
	}
}

func TestSetFareAction_Validate_FareZeroOrNegative(t *testing.T) {
	tests := []struct {
		name string
		fare int
	}{
		{"zero", 0},
		{"negative", -100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := setupFareTestGame()
			a := &SetFareAction{LineID: g.PlayerCompany().Lines[0].ID, Fare: tt.fare}
			err := a.Validate(g)
			if err == nil {
				t.Fatal("不正な運賃でエラーにならなかった")
			}
			ve, ok := err.(*ValidationError)
			if !ok {
				t.Fatalf("型が *ValidationError でない: %T", err)
			}
			if ve.Code != "INVALID_FARE" {
				t.Errorf("Code = %q, want INVALID_FARE", ve.Code)
			}
		})
	}
}

func TestSetFareAction_Validate_FareTooHigh(t *testing.T) {
	g := setupFareTestGame()
	a := &SetFareAction{LineID: g.PlayerCompany().Lines[0].ID, Fare: 1001}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("運賃1001円でエラーにならなかった")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("型が *ValidationError でない: %T", err)
	}
	if ve.Code != "INVALID_FARE" {
		t.Errorf("Code = %q, want INVALID_FARE", ve.Code)
	}
}

func TestSetFareAction_Validate_TooLargeChange(t *testing.T) {
	g := setupFareTestGame()
	line := g.PlayerCompany().Lines[0]
	line.Fare = 200
	// 200 → 400 は +100% なので不可
	a := &SetFareAction{LineID: line.ID, Fare: 400}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("±50%超の変更でエラーにならなかった")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("型が *ValidationError でない: %T", err)
	}
	if ve.Code != "FARE_CHANGE_TOO_LARGE" {
		t.Errorf("Code = %q, want FARE_CHANGE_TOO_LARGE", ve.Code)
	}
}

func TestSetFareAction_Validate_OK(t *testing.T) {
	g := setupFareTestGame()
	line := g.PlayerCompany().Lines[0]
	line.Fare = 200
	a := &SetFareAction{LineID: line.ID, Fare: 250}

	if err := a.Validate(g); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
	// Validate は状態を変更しない
	if line.Fare != 200 {
		t.Errorf("Validate が Fare を変更した: %d", line.Fare)
	}
}

// --- Apply ---

func TestSetFareAction_Apply(t *testing.T) {
	g := setupFareTestGame()
	line := g.PlayerCompany().Lines[0]
	line.Fare = 200
	a := &SetFareAction{LineID: line.ID, Fare: 250}

	events, err := a.Apply(g)
	if err != nil {
		t.Fatal(err)
	}
	if line.Fare != 250 {
		t.Errorf("Fare = %d, want 250", line.Fare)
	}
	if len(events) != 1 {
		t.Fatalf("events 数 = %d, want 1", len(events))
	}
	e := events[0]
	if e.Type != EventFareChanged {
		t.Errorf("Type = %q, want %q", e.Type, EventFareChanged)
	}
	if e.Severity != SeverityInfo {
		t.Errorf("Severity = %q, want %q", e.Severity, SeverityInfo)
	}
	if e.Payload["line_id"] != line.ID {
		t.Errorf("Payload[line_id] = %v, want %s", e.Payload["line_id"], line.ID)
	}
	if e.Payload["old_fare"] != 200 {
		t.Errorf("Payload[old_fare] = %v, want 200", e.Payload["old_fare"])
	}
	if e.Payload["new_fare"] != 250 {
		t.Errorf("Payload[new_fare] = %v, want 250", e.Payload["new_fare"])
	}
}

func TestSetFareAction_Apply_SameFare(t *testing.T) {
	g := setupFareTestGame()
	line := g.PlayerCompany().Lines[0]
	line.Fare = 200
	a := &SetFareAction{LineID: line.ID, Fare: 200}

	events, err := a.Apply(g)
	if err != nil {
		t.Fatal(err)
	}
	if line.Fare != 200 {
		t.Errorf("Fare = %d, want 200", line.Fare)
	}
	// 同じ運賃でもイベントは生成する（記録のため）
	if len(events) != 1 {
		t.Errorf("events 数 = %d, want 1", len(events))
	}
}

func TestSetFareAction_Apply_ViaRegistry(t *testing.T) {
	g := setupFareTestGame()
	line := g.PlayerCompany().Lines[0]
	line.Fare = 200

	payload := []byte(`{"line_id":"` + line.ID + `","fare":250}`)
	a, err := DecodeAction("set_fare", payload)
	if err != nil {
		t.Fatal(err)
	}
	events, err := a.Apply(g)
	if err != nil {
		t.Fatal(err)
	}
	if line.Fare != 250 {
		t.Errorf("Fare = %d, want 250", line.Fare)
	}
	if len(events) != 1 {
		t.Errorf("events 数 = %d, want 1", len(events))
	}
}

// --- 効果検証 ---

func TestSetFareAction_FareIncrease_ReducesDemand(t *testing.T) {
	g := setupFareTestGame()
	line := g.PlayerCompany().Lines[0]
	line.Fare = 200

	// 運賃前の収益
	g.EndTurn()
	profitBefore := line.Profit

	// 運賃 +25%（200 → 250）
	a := &SetFareAction{LineID: line.ID, Fare: 250}
	if _, err := a.Apply(g); err != nil {
		t.Fatal(err)
	}

	g.EndTurn()
	profitAfter := line.Profit

	// 運賃上げで利益が増える（需要減 < 運賃増）
	t.Logf("利益: %d → %d", profitBefore, profitAfter)
	if profitAfter <= profitBefore {
		t.Errorf("運賃+25%%で利益が増えなかった: %d → %d", profitBefore, profitAfter)
	}
}

// --- 決定論 ---

func TestSetFareAction_Deterministic(t *testing.T) {
	g1 := setupFareTestGame()
	g2 := setupFareTestGame()
	line1 := g1.PlayerCompany().Lines[0]
	line2 := g2.PlayerCompany().Lines[0]

	a1 := &SetFareAction{LineID: line1.ID, Fare: 250}
	a2 := &SetFareAction{LineID: line2.ID, Fare: 250}

	e1, _ := a1.Apply(g1)
	e2, _ := a2.Apply(g2)

	if e1[0].Message != e2[0].Message {
		t.Errorf("Message が異なる: %q vs %q", e1[0].Message, e2[0].Message)
	}
	if line1.Fare != line2.Fare {
		t.Errorf("Fare が異なる: %d vs %d", line1.Fare, line2.Fare)
	}
}

// --- ヘルパー ---

func setupFareTestGame() *Game {
	g := NewGame(1907, "hankyu")
	g.Date = Date{Year: 1910, Month: 1}

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

	a := &BuildLineAction{From: "umeda", To: "takarazuka", Gauge: 1067}
	if _, err := a.Apply(g); err != nil {
		panic(err)
	}
	return g
}
