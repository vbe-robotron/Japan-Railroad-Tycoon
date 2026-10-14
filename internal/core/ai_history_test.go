package core

import "testing"

func TestAIFollowHistory_BuildsLine(t *testing.T) {
	g, c := newAITestGame(t, "hankyu", 100_000_000, &HistoricalLine{
		CompanyID: "hankyu", From: "a", To: "b", Gauge: 1067,
		OpenedYear: 1910, OpenedMonth: 3, DisplayName: "テスト線",
	})
	ai := &AI{CompanyID: "hankyu", HistoryMode: HistoryStrict}

	events := g.aiFollowHistory(c, ai)

	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}
	if events[0].Type != EventLineOpened {
		t.Errorf("Type = %q, want %q", events[0].Type, EventLineOpened)
	}
	if events[0].CompanyID != "hankyu" {
		t.Errorf("CompanyID = %q, want hankyu", events[0].CompanyID)
	}
	if events[0].Payload["ai"] != true {
		t.Errorf("Payload[ai] = %v, want true", events[0].Payload["ai"])
	}
	if len(c.Lines) != 1 {
		t.Fatalf("Lines = %d, want 1", len(c.Lines))
	}
	if c.Lines[0].DisplayName != "テスト線" {
		t.Errorf("DisplayName = %q, want テスト線", c.Lines[0].DisplayName)
	}
	if c.Lines[0].Gauge != 1067 {
		t.Errorf("Gauge = %d, want 1067", c.Lines[0].Gauge)
	}
}

func TestAIFollowHistory_DefaultDisplayName(t *testing.T) {
	g, c := newAITestGame(t, "hankyu", 100_000_000, &HistoricalLine{
		CompanyID: "hankyu", From: "a", To: "b", Gauge: 1067,
		OpenedYear: 1910, OpenedMonth: 3,
	})
	ai := &AI{CompanyID: "hankyu", HistoryMode: HistoryStrict}

	g.aiFollowHistory(c, ai)

	if c.Lines[0].DisplayName != "A〜B" {
		t.Errorf("DisplayName = %q, want A〜B", c.Lines[0].DisplayName)
	}
}

func TestAIFollowHistory_SkipsIfExists(t *testing.T) {
	g, c := newAITestGame(t, "hankyu", 100_000_000, &HistoricalLine{
		CompanyID: "hankyu", From: "a", To: "b", Gauge: 1067,
		OpenedYear: 1910, OpenedMonth: 3, DisplayName: "テスト線",
	})
	g.InitAIs()
	ai := g.AIs["hankyu"]

	// 1 回目で敷設
	if evs := g.aiFollowHistory(c, ai); len(evs) != 1 {
		t.Fatalf("1 回目の events = %d, want 1", len(evs))
	}
	// 2 回目は既存のためスキップ
	if evs := g.aiFollowHistory(c, ai); len(evs) != 0 {
		t.Errorf("2 回目の events = %d, want 0", len(evs))
	}
	if len(c.Lines) != 1 {
		t.Errorf("Lines = %d, want 1", len(c.Lines))
	}
}

func TestAIFollowHistory_ReversedExisting(t *testing.T) {
	g, c := newAITestGame(t, "hankyu", 100_000_000, &HistoricalLine{
		CompanyID: "hankyu", From: "a", To: "b", Gauge: 1067,
		OpenedYear: 1910, OpenedMonth: 3,
	})
	// 逆向きの路線を既に保有
	c.AddLine(NewLine("l1", "hankyu", "b", "a"))

	ai := &AI{CompanyID: "hankyu", HistoryMode: HistoryStrict}
	if evs := g.aiFollowHistory(c, ai); len(evs) != 0 {
		t.Errorf("逆向き既存で events = %d, want 0", len(evs))
	}
}

func TestAIFollowHistory_NoMatchingYear(t *testing.T) {
	g, c := newAITestGame(t, "hankyu", 100_000_000, &HistoricalLine{
		CompanyID: "hankyu", From: "a", To: "b", Gauge: 1067,
		OpenedYear: 1920, OpenedMonth: 7,
	})
	ai := &AI{CompanyID: "hankyu", HistoryMode: HistoryStrict}

	if evs := g.aiFollowHistory(c, ai); len(evs) != 0 {
		t.Errorf("events = %d, want 0（年不一致）", len(evs))
	}
}

func TestAIFollowHistory_OtherCompany(t *testing.T) {
	g, c := newAITestGame(t, "hankyu", 100_000_000, &HistoricalLine{
		CompanyID: "hanshin", From: "a", To: "b", Gauge: 1067,
		OpenedYear: 1910, OpenedMonth: 3,
	})
	ai := &AI{CompanyID: "hankyu", HistoryMode: HistoryStrict}

	if evs := g.aiFollowHistory(c, ai); len(evs) != 0 {
		t.Errorf("events = %d, want 0（他社）", len(evs))
	}
}

func TestAIFollowHistory_BorrowsForCost(t *testing.T) {
	g, c := newAITestGame(t, "hankyu", 0, &HistoricalLine{
		CompanyID: "hankyu", From: "a", To: "b", Gauge: 1067,
		OpenedYear: 1910, OpenedMonth: 3, DisplayName: "テスト線",
	})
	ai := &AI{CompanyID: "hankyu", HistoryMode: HistoryStrict}

	events := g.aiFollowHistory(c, ai)

	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}
	if c.Debt <= 0 {
		t.Errorf("Debt = %d, want > 0", c.Debt)
	}
	if len(c.Lines) != 1 {
		t.Errorf("Lines = %d, want 1", len(c.Lines))
	}
}

func TestAIFollowHistory_UnknownCity(t *testing.T) {
	g, c := newAITestGame(t, "hankyu", 100_000_000, &HistoricalLine{
		CompanyID: "hankyu", From: "a", To: "nowhere", Gauge: 1067,
		OpenedYear: 1910, OpenedMonth: 3,
	})
	ai := &AI{CompanyID: "hankyu", HistoryMode: HistoryStrict}

	if evs := g.aiFollowHistory(c, ai); len(evs) != 0 {
		t.Errorf("events = %d, want 0（都市なし）", len(evs))
	}
}

func TestLineExists(t *testing.T) {
	g, c := newAITestGame(t, "hankyu", 0, nil)
	c.AddLine(NewLine("l1", "hankyu", "a", "b"))

	if !g.lineExists("a", "b", "hankyu") {
		t.Error("lineExists(a,b) = false, want true")
	}
	if !g.lineExists("b", "a", "hankyu") {
		t.Error("lineExists(b,a) = false, want true（逆向き）")
	}
	if g.lineExists("a", "b", "hanshin") {
		t.Error("lineExists(他社) = true, want false")
	}
	if g.lineExists("a", "b", "unknown") {
		t.Error("lineExists(未知会社) = true, want false")
	}
}
