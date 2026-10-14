package data

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// vanillaDir はリポジトリ同梱の data ディレクトリへの相対パス。
var vanillaDir = filepath.Join("..", "..", "data")

func TestLoad_Vanilla(t *testing.T) {
	ds, err := Load(vanillaDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(ds.Cities) != 36 {
		t.Errorf("len(Cities) = %d, want 36", len(ds.Cities))
	}
	if len(ds.Companies) != 20 {
		t.Errorf("len(Companies) = %d, want 20", len(ds.Companies))
	}

	umeda, ok := findCity(ds, "umeda")
	if !ok {
		t.Fatal("umeda が見つからない")
	}
	if umeda.Name != "梅田" || umeda.Population != 120000 || umeda.LandValue != 60 {
		t.Errorf("umeda = %+v", umeda)
	}
	if umeda.Prefecture != "大阪府" {
		t.Errorf("umeda.Prefecture = %q", umeda.Prefecture)
	}

	hankyu, ok := findCompany(ds, "hankyu")
	if !ok {
		t.Fatal("hankyu が見つからない")
	}
	if hankyu.Name != "阪急電鉄" || hankyu.FoundedYear != 1907 {
		t.Errorf("hankyu = %+v", hankyu)
	}
	if hankyu.InitialCash != 9_000_000 || hankyu.TotalShares != 10000 {
		t.Errorf("hankyu 財務 = %+v", hankyu)
	}
	if !hankyu.IsPlayerCandidate {
		t.Error("hankyu.IsPlayerCandidate = false, want true")
	}
}

func TestLoad_EmptyDir(t *testing.T) {
	if _, err := Load(""); !errors.Is(err, ErrInvalidData) {
		t.Errorf("err = %v, want ErrInvalidData", err)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "nope"))
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("err = %v, want fs.ErrNotExist", err)
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "mods/vanilla/cities.json", "{ this is not json")
	_, err := Load(dir)
	if err == nil {
		t.Fatal("err = nil, want parse error")
	}
}

func TestLoad_ValidationError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "mods/vanilla/cities.json",
		`{"cities":[{"id":"a","name":"A","lat":999,"lon":0,"land_value":0}]}`)
	writeFile(t, dir, "mods/vanilla/companies.json", `{"companies":[]}`)
	if _, err := Load(dir); !errors.Is(err, ErrInvalidData) {
		t.Errorf("err = %v, want ErrInvalidData", err)
	}
}

func TestValidate_Valid(t *testing.T) {
	ds := &Dataset{
		Cities: []CityDef{{
			ID: "a", Name: "A", Lat: 34.7, Lon: 135.5, Population: 100, LandValue: 50,
		}},
		Companies: []CompanyDef{{
			ID: "c", Name: "C", InitialCash: 100, TotalShares: 10, StockPrice: 5, Reputation: 50,
		}},
	}
	if err := ds.Validate(); err != nil {
		t.Errorf("Validate: %v", err)
	}
}

func TestValidate_Nil(t *testing.T) {
	var ds *Dataset
	if err := ds.Validate(); !errors.Is(err, ErrInvalidData) {
		t.Errorf("err = %v, want ErrInvalidData", err)
	}
}

func TestValidate_DuplicateID(t *testing.T) {
	t.Run("都市", func(t *testing.T) {
		ds := &Dataset{Cities: []CityDef{
			{ID: "a", Name: "A", Lat: 0, Lon: 0},
			{ID: "a", Name: "A2", Lat: 0, Lon: 0},
		}}
		if err := ds.Validate(); !errors.Is(err, ErrDuplicateID) {
			t.Errorf("err = %v, want ErrDuplicateID", err)
		}
	})
	t.Run("会社", func(t *testing.T) {
		ds := &Dataset{Companies: []CompanyDef{
			{ID: "c", Name: "C", Reputation: 50},
			{ID: "c", Name: "C2", Reputation: 50},
		}}
		if err := ds.Validate(); !errors.Is(err, ErrDuplicateID) {
			t.Errorf("err = %v, want ErrDuplicateID", err)
		}
	})
}

func TestValidate_InvalidCoordinate(t *testing.T) {
	cases := []CityDef{
		{ID: "lat", Name: "L", Lat: 91, Lon: 0},
		{ID: "lat2", Name: "L", Lat: -91, Lon: 0},
		{ID: "lon", Name: "L", Lat: 0, Lon: 181},
		{ID: "lon2", Name: "L", Lat: 0, Lon: -181},
	}
	for _, c := range cases {
		ds := &Dataset{Cities: []CityDef{c}}
		if err := ds.Validate(); !errors.Is(err, ErrInvalidData) {
			t.Errorf("city %q: err = %v, want ErrInvalidData", c.ID, err)
		}
	}
}

func TestValidate_InvalidFields(t *testing.T) {
	t.Run("都市ID空", func(t *testing.T) {
		ds := &Dataset{Cities: []CityDef{{Name: "A"}}}
		if err := ds.Validate(); !errors.Is(err, ErrInvalidData) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("都市名空", func(t *testing.T) {
		ds := &Dataset{Cities: []CityDef{{ID: "a"}}}
		if err := ds.Validate(); !errors.Is(err, ErrInvalidData) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("人気負", func(t *testing.T) {
		ds := &Dataset{Cities: []CityDef{{ID: "a", Name: "A", Population: -1}}}
		if err := ds.Validate(); !errors.Is(err, ErrInvalidData) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("地価範囲外", func(t *testing.T) {
		ds := &Dataset{Cities: []CityDef{{ID: "a", Name: "A", LandValue: 101}}}
		if err := ds.Validate(); !errors.Is(err, ErrInvalidData) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("会社ID空", func(t *testing.T) {
		ds := &Dataset{Companies: []CompanyDef{{Name: "C", Reputation: 50}}}
		if err := ds.Validate(); !errors.Is(err, ErrInvalidData) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("会社名空", func(t *testing.T) {
		ds := &Dataset{Companies: []CompanyDef{{ID: "c", Reputation: 50}}}
		if err := ds.Validate(); !errors.Is(err, ErrInvalidData) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("初期資金負", func(t *testing.T) {
		ds := &Dataset{Companies: []CompanyDef{{ID: "c", Name: "C", InitialCash: -1, Reputation: 50}}}
		if err := ds.Validate(); !errors.Is(err, ErrInvalidData) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("株数負", func(t *testing.T) {
		ds := &Dataset{Companies: []CompanyDef{{ID: "c", Name: "C", TotalShares: -1, Reputation: 50}}}
		if err := ds.Validate(); !errors.Is(err, ErrInvalidData) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("株価負", func(t *testing.T) {
		ds := &Dataset{Companies: []CompanyDef{{ID: "c", Name: "C", StockPrice: -1, Reputation: 50}}}
		if err := ds.Validate(); !errors.Is(err, ErrInvalidData) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("評判範囲外", func(t *testing.T) {
		ds := &Dataset{Companies: []CompanyDef{{ID: "c", Name: "C", Reputation: 101}}}
		if err := ds.Validate(); !errors.Is(err, ErrInvalidData) {
			t.Errorf("err = %v", err)
		}
	})
}

func TestLoad_HistoricalLines(t *testing.T) {
	ds, err := Load(vanillaDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(ds.HistoricalLines) < 11 {
		t.Fatalf("len(HistoricalLines) = %d, want >= 11", len(ds.HistoricalLines))
	}

	takara, ok := findHistoricalLine(ds, "hankyu", "umeda", "takarazuka")
	if !ok {
		t.Fatal("阪急 梅田〜宝塚 が見つからない")
	}
	if takara.Gauge != 1435 || takara.OpenedYear != 1910 || takara.OpenedMonth != 3 {
		t.Errorf("宝塚線 = %+v", takara)
	}
	if takara.DisplayName != "宝塚線" {
		t.Errorf("DisplayName = %q", takara.DisplayName)
	}
}

// TestLoad_AllCompanies は全 20 社が読み込まれることを確認する。
func TestLoad_AllCompanies(t *testing.T) {
	ds, err := Load(vanillaDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(ds.Companies) != 20 {
		t.Fatalf("len(Companies) = %d, want 20", len(ds.Companies))
	}
	for _, id := range []string{"tokyu", "seibu", "odakyu", "meitetsu", "iyotetsu"} {
		if _, ok := findCompany(ds, id); !ok {
			t.Errorf("会社 %q が見つからない", id)
		}
	}
}

// TestLoad_AllCities は全 36 都市が読み込まれることを確認する。
func TestLoad_AllCities(t *testing.T) {
	ds, err := Load(vanillaDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(ds.Cities) != 36 {
		t.Fatalf("len(Cities) = %d, want 36", len(ds.Cities))
	}
	for _, id := range []string{"tokyo", "yokohama", "nagoya", "matsuyama"} {
		if _, ok := findCity(ds, id); !ok {
			t.Errorf("都市 %q が見つからない", id)
		}
	}
}

// TestLoad_AllHistoricalLines は全 38 路線が読み込まれることを確認する。
func TestLoad_AllHistoricalLines(t *testing.T) {
	ds, err := Load(vanillaDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(ds.HistoricalLines) != 38 {
		t.Fatalf("len(HistoricalLines) = %d, want 38", len(ds.HistoricalLines))
	}
}

// TestValidate_AllRefsExist は全路線の会社・都市参照が解決できることを確認する。
func TestValidate_AllRefsExist(t *testing.T) {
	ds, err := Load(vanillaDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	companies := make(map[string]bool, len(ds.Companies))
	for _, c := range ds.Companies {
		companies[c.ID] = true
	}
	cities := make(map[string]bool, len(ds.Cities))
	for _, c := range ds.Cities {
		cities[c.ID] = true
	}
	for i, l := range ds.HistoricalLines {
		if !companies[l.CompanyID] {
			t.Errorf("lines[%d] 会社 %q が存在しない", i, l.CompanyID)
		}
		if !cities[l.From] {
			t.Errorf("lines[%d] 起点 %q が存在しない", i, l.From)
		}
		if !cities[l.To] {
			t.Errorf("lines[%d] 終点 %q が存在しない", i, l.To)
		}
	}
}

// findHistoricalLine は会社・区間で史実路線を検索する。
func findHistoricalLine(ds *Dataset, company, from, to string) (HistoricalLineDef, bool) {
	for _, l := range ds.HistoricalLines {
		if l.CompanyID == company && l.From == from && l.To == to {
			return l, true
		}
	}
	return HistoricalLineDef{}, false
}

func TestValidate_HistoricalLine_UnknownCompany(t *testing.T) {
	ds := &Dataset{
		Cities:          []CityDef{{ID: "a", Name: "A", Lat: 0, Lon: 0}},
		Companies:       []CompanyDef{{ID: "c", Name: "C", Reputation: 50}},
		HistoricalLines: []HistoricalLineDef{{CompanyID: "nope", From: "a", To: "a", Gauge: 1435, OpenedYear: 1900}},
	}
	if err := ds.Validate(); !errors.Is(err, ErrInvalidData) {
		t.Errorf("err = %v, want ErrInvalidData", err)
	}
}

func TestValidate_HistoricalLine_UnknownCity(t *testing.T) {
	ds := &Dataset{
		Cities:          []CityDef{{ID: "a", Name: "A", Lat: 0, Lon: 0}},
		Companies:       []CompanyDef{{ID: "c", Name: "C", Reputation: 50}},
		HistoricalLines: []HistoricalLineDef{{CompanyID: "c", From: "a", To: "z", Gauge: 1435, OpenedYear: 1900}},
	}
	if err := ds.Validate(); !errors.Is(err, ErrInvalidData) {
		t.Errorf("err = %v, want ErrInvalidData", err)
	}
}

func TestValidate_HistoricalLine_SameCity(t *testing.T) {
	ds := &Dataset{
		Cities:          []CityDef{{ID: "a", Name: "A", Lat: 0, Lon: 0}},
		Companies:       []CompanyDef{{ID: "c", Name: "C", Reputation: 50}},
		HistoricalLines: []HistoricalLineDef{{CompanyID: "c", From: "a", To: "a", Gauge: 1435, OpenedYear: 1900}},
	}
	if err := ds.Validate(); !errors.Is(err, ErrInvalidData) {
		t.Errorf("err = %v, want ErrInvalidData", err)
	}
}

func TestValidate_HistoricalLine_InvalidGauge(t *testing.T) {
	ds := &Dataset{
		Cities:          []CityDef{{ID: "a", Name: "A", Lat: 0, Lon: 0}, {ID: "b", Name: "B", Lat: 0, Lon: 0}},
		Companies:       []CompanyDef{{ID: "c", Name: "C", Reputation: 50}},
		HistoricalLines: []HistoricalLineDef{{CompanyID: "c", From: "a", To: "b", Gauge: 999, OpenedYear: 1900}},
	}
	if err := ds.Validate(); !errors.Is(err, ErrInvalidData) {
		t.Errorf("err = %v, want ErrInvalidData", err)
	}
}

func TestValidate_HistoricalLine_InvalidYear(t *testing.T) {
	ds := &Dataset{
		Cities:          []CityDef{{ID: "a", Name: "A", Lat: 0, Lon: 0}, {ID: "b", Name: "B", Lat: 0, Lon: 0}},
		Companies:       []CompanyDef{{ID: "c", Name: "C", Reputation: 50}},
		HistoricalLines: []HistoricalLineDef{{CompanyID: "c", From: "a", To: "b", Gauge: 1435, OpenedYear: 1869}},
	}
	if err := ds.Validate(); !errors.Is(err, ErrInvalidData) {
		t.Errorf("err = %v, want ErrInvalidData", err)
	}
}

func TestValidate_HistoricalLine_EmptyFields(t *testing.T) {
	cases := []HistoricalLineDef{
		{CompanyID: "", From: "a", To: "b", Gauge: 1435, OpenedYear: 1900},
		{CompanyID: "c", From: "", To: "b", Gauge: 1435, OpenedYear: 1900},
		{CompanyID: "c", From: "a", To: "", Gauge: 1435, OpenedYear: 1900},
	}
	for i, l := range cases {
		ds := &Dataset{
			Cities:          []CityDef{{ID: "a", Name: "A", Lat: 0, Lon: 0}, {ID: "b", Name: "B", Lat: 0, Lon: 0}},
			Companies:       []CompanyDef{{ID: "c", Name: "C", Reputation: 50}},
			HistoricalLines: []HistoricalLineDef{l},
		}
		if err := ds.Validate(); !errors.Is(err, ErrInvalidData) {
			t.Errorf("case %d: err = %v, want ErrInvalidData", i, err)
		}
	}
}

func TestLoadJSON_NotFound(t *testing.T) {
	var out struct{}
	if err := loadJSON(filepath.Join(t.TempDir(), "x.json"), &out); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("err = %v, want fs.ErrNotExist", err)
	}
}

func findCity(ds *Dataset, id string) (CityDef, bool) {
	for _, c := range ds.Cities {
		if c.ID == id {
			return c, true
		}
	}
	return CityDef{}, false
}

func findCompany(ds *Dataset, id string) (CompanyDef, bool) {
	for _, c := range ds.Companies {
		if c.ID == id {
			return c, true
		}
	}
	return CompanyDef{}, false
}

func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}
