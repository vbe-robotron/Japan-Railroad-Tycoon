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
