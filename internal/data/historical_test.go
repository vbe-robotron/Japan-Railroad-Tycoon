package data

import (
	"testing"
)

func TestHistoricalLines_ByCompany(t *testing.T) {
	ds, err := Load(vanillaDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	byCompany := ds.HistoricalLinesByCompany()
	if len(byCompany["hankyu"]) != 4 {
		t.Errorf("hankyu = %d 路線, want 4", len(byCompany["hankyu"]))
	}
	if len(byCompany["kintetsu"]) != 5 {
		t.Errorf("kintetsu = %d 路線, want 5", len(byCompany["kintetsu"]))
	}
	if len(byCompany["hanshin"]) != 1 {
		t.Errorf("hanshin = %d 路線, want 1", len(byCompany["hanshin"]))
	}

	total := 0
	for _, ls := range byCompany {
		total += len(ls)
	}
	if total != len(ds.HistoricalLines) {
		t.Errorf("合計 = %d, want %d", total, len(ds.HistoricalLines))
	}
}

func TestHistoricalLines_ByYear(t *testing.T) {
	ds, err := Load(vanillaDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	byYear := ds.HistoricalLinesByYear()
	if len(byYear[1877]) != 2 {
		t.Errorf("1877 = %d 路線, want 2", len(byYear[1877]))
	}
	if len(byYear[1910]) == 0 {
		t.Error("1910 の路線が無い")
	}
}

func TestHistoricalLines_NilSafe(t *testing.T) {
	var ds *Dataset
	if len(ds.HistoricalLinesByCompany()) != 0 {
		t.Error("nil で空 map でない")
	}
	if len(ds.HistoricalLinesByYear()) != 0 {
		t.Error("nil で空 map でない")
	}
}

func TestLoad_InvalidHistoricalJSON(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "mods/vanilla/cities.json", `{"cities":[]}`)
	writeFile(t, dir, "mods/vanilla/companies.json", `{"companies":[]}`)
	writeFile(t, dir, "mods/vanilla/historical_lines.json", "{ broken")
	if _, err := Load(dir); err == nil {
		t.Fatal("err = nil, want parse error")
	}
}
