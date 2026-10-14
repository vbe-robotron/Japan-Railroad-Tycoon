package data

import (
	"testing"

	"github.com/you/jrt/internal/core"
)

func sampleDataset() *Dataset {
	return &Dataset{
		Cities: []CityDef{{
			ID: "umeda", Name: "梅田", Prefecture: "大阪府",
			RepresentativeNode: "umeda", Lat: 34.7055, Lon: 135.4983,
			Population: 120000, LandValue: 60, Tags: []string{"kansai"},
		}},
		Companies: []CompanyDef{{
			ID: "hankyu", GroupID: "hankyu", Name: "阪急電鉄",
			LegalName: "箕面有馬電気軌道株式会社", FoundedYear: 1907,
			InitialCash: 9_000_000, TotalShares: 10000, StockPrice: 1000,
			Region: "kansai", Tags: []string{"kansai"}, Reputation: 50,
		}, {
			ID: "jnr", GroupID: "jnr", Name: "国鉄",
			FoundedYear: 1872, InitialCash: 50_000_000, TotalShares: 100000,
			StockPrice: 0, Region: "kansai", Reputation: 50,
		}},
		HistoricalLines: []HistoricalLineDef{{
			CompanyID: "hankyu", From: "umeda", To: "takarazuka",
			Gauge: 1435, Double: false, OpenedYear: 1910, OpenedMonth: 3,
			DisplayName: "宝塚線",
		}},
	}
}

func TestApplyToGame_Cities(t *testing.T) {
	g := core.NewGame(1, "hankyu")
	sampleDataset().ApplyToGame(g, "hankyu")
	if len(g.Cities) != 1 {
		t.Fatalf("len(Cities) = %d, want 1", len(g.Cities))
	}
	c := g.Cities["umeda"]
	if c == nil {
		t.Fatal("umeda が投入されていない")
	}
	if c.Name != "梅田" || c.Population != 120000 || c.LandValue != 60 {
		t.Errorf("city = %+v", c)
	}
	if c.Lat != 34.7055 || c.Lon != 135.4983 {
		t.Errorf("座標 = %v, %v", c.Lat, c.Lon)
	}
}

func TestApplyToGame_Companies(t *testing.T) {
	g := core.NewGame(1, "hankyu")
	sampleDataset().ApplyToGame(g, "hankyu")
	if len(g.Companies) != 2 {
		t.Fatalf("len(Companies) = %d, want 2", len(g.Companies))
	}
	h := g.Companies["hankyu"]
	if h == nil {
		t.Fatal("hankyu が投入されていない")
	}
	if h.LegalName != "箕面有馬電気軌道株式会社" || h.FoundedYear != 1907 {
		t.Errorf("hankyu = %+v", h)
	}
	if h.Cash != 9_000_000 || h.TotalShares != 10000 || h.StockPrice != 1000 {
		t.Errorf("hankyu 財務 = %+v", h)
	}
	if h.Reputation != 50 {
		t.Errorf("hankyu.Reputation = %d", h.Reputation)
	}
}

func TestApplyToGame_PlayerID(t *testing.T) {
	g := core.NewGame(1, "hankyu")
	sampleDataset().ApplyToGame(g, "hankyu")

	if !g.Companies["hankyu"].IsPlayer {
		t.Error("hankyu.IsPlayer = false, want true")
	}
	if g.Companies["jnr"].IsPlayer {
		t.Error("jnr.IsPlayer = true, want false")
	}
	if !g.Companies["jnr"].IsNational {
		t.Error("jnr.IsNational = false, want true")
	}
	if g.Companies["hankyu"].IsNational {
		t.Error("hankyu.IsNational = true, want false")
	}
}

func TestApplyToGame_PlayerNotInData(t *testing.T) {
	g := core.NewGame(1, "player")
	sampleDataset().ApplyToGame(g, "player")
	for _, c := range g.Companies {
		if c.IsPlayer {
			t.Errorf("%s がプレイヤーに設定された", c.ID)
		}
	}
}

func TestApplyToGame_NilSafe(t *testing.T) {
	var ds *Dataset
	g := core.NewGame(1, "p")
	ds.ApplyToGame(g, "p")
	if len(g.Companies) != 0 || len(g.Cities) != 0 {
		t.Error("nil Dataset でデータが投入された")
	}

	sampleDataset().ApplyToGame(nil, "p")
}

func TestApplyToGame_HistoricalLines(t *testing.T) {
	g := core.NewGame(1, "hankyu")
	sampleDataset().ApplyToGame(g, "hankyu")

	if len(g.HistoricalLines) != 1 {
		t.Fatalf("len(HistoricalLines) = %d, want 1", len(g.HistoricalLines))
	}
	h := g.HistoricalLines[0]
	if h.CompanyID != "hankyu" || h.From != "umeda" || h.To != "takarazuka" {
		t.Errorf("HistoricalLine = %+v", h)
	}
	if h.Gauge != 1435 || h.OpenedYear != 1910 || h.OpenedMonth != 3 {
		t.Errorf("HistoricalLine 詳細 = %+v", h)
	}
	if h.DisplayName != "宝塚線" {
		t.Errorf("DisplayName = %q, want 宝塚線", h.DisplayName)
	}
}

func TestApplyToGame_Vanilla(t *testing.T) {
	ds, err := Load(vanillaDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	g := core.NewGame(7, "hankyu")
	ds.ApplyToGame(g, "hankyu")
	if len(g.Cities) != 36 {
		t.Errorf("len(Cities) = %d, want 36", len(g.Cities))
	}
	if len(g.Companies) != 20 {
		t.Errorf("len(Companies) = %d, want 20", len(g.Companies))
	}
	if !g.Companies["hankyu"].IsPlayer {
		t.Error("hankyu がプレイヤーでない")
	}
}
