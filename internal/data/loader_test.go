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
	if len(ds.Cities) != 10 {
		t.Errorf("len(Cities) = %d, want 10", len(ds.Cities))
	}
	if len(ds.Companies) != 7 {
		t.Errorf("len(Companies) = %d, want 7", len(ds.Companies))
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
