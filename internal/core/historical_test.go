package core

import "testing"

func TestAutoBuildHistoricalLine(t *testing.T) {
	g := NewGame(1, "player")
	g.Date = Date{Year: 1907, Month: 10}
	g.AddCity(&City{ID: "a", Name: "A", Population: 100000, Lat: 35.0, Lon: 135.0})
	g.AddCity(&City{ID: "b", Name: "B", Population: 100000, Lat: 35.0, Lon: 135.2})

	c := NewCompany("jnr", "jnr", "国鉄")
	g.AddCompany(c)

	h := HistoricalLine{
		CompanyID: "jnr", From: "a", To: "b", Gauge: 1067,
		OpenedYear: 1872, OpenedMonth: 10, DisplayName: "東海道本線",
	}
	g.AutoBuildHistoricalLine(c, h)

	if len(c.Lines) != 1 {
		t.Fatalf("Lines = %d, want 1", len(c.Lines))
	}
	l := c.Lines[0]
	if l.DisplayName != "東海道本線" {
		t.Errorf("DisplayName = %q", l.DisplayName)
	}
	if l.Gauge != 1067 {
		t.Errorf("Gauge = %d, want 1067", l.Gauge)
	}
	if l.Opened.Year != 1872 || l.Opened.Month != 10 {
		t.Errorf("Opened = %+v, want 1872-10", l.Opened)
	}
	if l.TrainsPerDay != 10 {
		t.Errorf("TrainsPerDay = %d, want 10", l.TrainsPerDay)
	}
	if l.Fare <= 0 {
		t.Errorf("Fare = %d, want > 0", l.Fare)
	}
	if g.Cities["a"].Demand == 0 || g.Cities["b"].Demand == 0 {
		t.Error("需要が再計算されていない")
	}
}

func TestAutoBuildHistoricalLine_NoCost(t *testing.T) {
	g, c := newAITestGame(t, "hankyu", 1000, nil)
	g.AutoBuildHistoricalLine(c, HistoricalLine{
		CompanyID: "hankyu", From: "a", To: "b", Gauge: 1067,
		OpenedYear: 1910, OpenedMonth: 3,
	})
	if c.Cash != 1000 {
		t.Errorf("Cash = %d, want 1000（コストなし）", c.Cash)
	}
}

func TestAutoBuildHistoricalLine_SkipsExisting(t *testing.T) {
	g, c := newAITestGame(t, "hankyu", 0, nil)
	c.AddLine(NewLine("l1", "hankyu", "b", "a"))

	g.AutoBuildHistoricalLine(c, HistoricalLine{
		CompanyID: "hankyu", From: "a", To: "b", Gauge: 1067,
		OpenedYear: 1910, OpenedMonth: 3,
	})
	if len(c.Lines) != 1 {
		t.Errorf("Lines = %d, want 1（既存をスキップ）", len(c.Lines))
	}
}

func TestAutoBuildHistoricalLine_NilCompany(t *testing.T) {
	g, _ := newAITestGame(t, "hankyu", 0, nil)
	g.AutoBuildHistoricalLine(nil, HistoricalLine{From: "a", To: "b"})
}

func TestAutoBuildHistoricalLine_UnknownCity(t *testing.T) {
	g, c := newAITestGame(t, "hankyu", 0, nil)
	g.AutoBuildHistoricalLine(c, HistoricalLine{
		CompanyID: "hankyu", From: "a", To: "nowhere", Gauge: 1067,
		OpenedYear: 1910, OpenedMonth: 3,
	})
	if len(c.Lines) != 0 {
		t.Errorf("Lines = %d, want 0（都市なし）", len(c.Lines))
	}
}
