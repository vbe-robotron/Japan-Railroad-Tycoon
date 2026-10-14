package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ErrInvalidData はデータの必須項目不足や値域違反を表す。
var ErrInvalidData = errors.New("invalid data")

// ErrDuplicateID は同一種別内での ID 重複を表す。
var ErrDuplicateID = errors.New("duplicate id")

// Dataset は 1 つの mod から読み込んだゲームデータの集合。
type Dataset struct {
	Cities    []CityDef
	Companies []CompanyDef
}

// CityDef は都市定義（cities.json の 1 要素）。
type CityDef struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Prefecture         string   `json:"prefecture"`
	RepresentativeNode string   `json:"representative_node"`
	Lat                float64  `json:"lat"`
	Lon                float64  `json:"lon"`
	Population         int      `json:"population"`
	LandValue          int      `json:"land_value"`
	Tags               []string `json:"tags"`
}

// CompanyDef は会社定義（companies.json の 1 要素）。
type CompanyDef struct {
	ID                string   `json:"id"`
	GroupID           string   `json:"group_id"`
	Name              string   `json:"name"`
	LegalName         string   `json:"legal_name"`
	FoundedYear       int      `json:"founded_year"`
	InitialCash       int64    `json:"initial_cash"`
	TotalShares       int      `json:"total_shares"`
	StockPrice        int      `json:"stock_price"`
	Region            string   `json:"region"`
	IsPlayerCandidate bool     `json:"is_player_candidate"`
	Tags              []string `json:"tags"`
	Reputation        int      `json:"reputation"`
}

// citiesFile / companiesFile は JSON ファイルのトップレベル構造。
type citiesFile struct {
	SchemaVersion string    `json:"schema_version"`
	Cities        []CityDef `json:"cities"`
}

type companiesFile struct {
	SchemaVersion string       `json:"schema_version"`
	Companies     []CompanyDef `json:"companies"`
}

// Load は dataDir 配下の mods/vanilla からデータを読み込む。
//
// cities.json と companies.json を読み込み、Validate で整合性を確認する。
func Load(dataDir string) (*Dataset, error) {
	if dataDir == "" {
		return nil, fmt.Errorf("%w: データディレクトリが空です", ErrInvalidData)
	}

	var cf citiesFile
	if err := loadJSON(filepath.Join(dataDir, "mods/vanilla/cities.json"), &cf); err != nil {
		return nil, err
	}
	var comp companiesFile
	if err := loadJSON(filepath.Join(dataDir, "mods/vanilla/companies.json"), &comp); err != nil {
		return nil, err
	}

	ds := &Dataset{Cities: cf.Cities, Companies: comp.Companies}
	if err := ds.Validate(); err != nil {
		return nil, err
	}
	return ds, nil
}

// loadJSON は 1 つの JSON ファイルを読み込んで out へデコードする。
func loadJSON(path string, out any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}

// Validate はデータの整合性を確認する。
//
// 都市・会社それぞれについて、必須フィールド・ID 重複・値域を検証する。
// 都市と会社は別名前空間のため、ID 重複は種別ごとに判定する。
func (d *Dataset) Validate() error {
	if d == nil {
		return fmt.Errorf("%w: dataset が nil です", ErrInvalidData)
	}

	seenCity := make(map[string]bool, len(d.Cities))
	for i, c := range d.Cities {
		if c.ID == "" {
			return fmt.Errorf("%w: cities[%d] の ID が空です", ErrInvalidData, i)
		}
		if seenCity[c.ID] {
			return fmt.Errorf("%w: city id %q が重複しています", ErrDuplicateID, c.ID)
		}
		seenCity[c.ID] = true
		if c.Name == "" {
			return fmt.Errorf("%w: city %q の名前が空です", ErrInvalidData, c.ID)
		}
		if c.Lat < -90 || c.Lat > 90 {
			return fmt.Errorf("%w: city %q の緯度が範囲外です: %v", ErrInvalidData, c.ID, c.Lat)
		}
		if c.Lon < -180 || c.Lon > 180 {
			return fmt.Errorf("%w: city %q の経度が範囲外です: %v", ErrInvalidData, c.ID, c.Lon)
		}
		if c.Population < 0 {
			return fmt.Errorf("%w: city %q の人口が負です: %d", ErrInvalidData, c.ID, c.Population)
		}
		if c.LandValue < 0 || c.LandValue > 100 {
			return fmt.Errorf("%w: city %q の地価が範囲外です: %d", ErrInvalidData, c.ID, c.LandValue)
		}
	}

	seenCompany := make(map[string]bool, len(d.Companies))
	for i, c := range d.Companies {
		if c.ID == "" {
			return fmt.Errorf("%w: companies[%d] の ID が空です", ErrInvalidData, i)
		}
		if seenCompany[c.ID] {
			return fmt.Errorf("%w: company id %q が重複しています", ErrDuplicateID, c.ID)
		}
		seenCompany[c.ID] = true
		if c.Name == "" {
			return fmt.Errorf("%w: company %q の名前が空です", ErrInvalidData, c.ID)
		}
		if c.InitialCash < 0 {
			return fmt.Errorf("%w: company %q の初期資金が負です: %d", ErrInvalidData, c.ID, c.InitialCash)
		}
		if c.TotalShares < 0 {
			return fmt.Errorf("%w: company %q の株数が負です: %d", ErrInvalidData, c.ID, c.TotalShares)
		}
		if c.StockPrice < 0 {
			return fmt.Errorf("%w: company %q の株価が負です: %d", ErrInvalidData, c.ID, c.StockPrice)
		}
		if c.Reputation < 0 || c.Reputation > 100 {
			return fmt.Errorf("%w: company %q の評判が範囲外です: %d", ErrInvalidData, c.ID, c.Reputation)
		}
	}

	return nil
}
