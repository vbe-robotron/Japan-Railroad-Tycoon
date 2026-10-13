package core

import (
	"errors"
	"testing"
)

// このファイルは BuildFacilityAction の TDD テストである。
//
// 施設建設（コスト支払い・開発度上昇・不動産収入）と、阪急モデル
// （沿線開発）の核となる開発度の効果を検証する。

// --- Type ---

func TestBuildFacilityAction_Type(t *testing.T) {
	a := &BuildFacilityAction{}
	if got := a.Type(); got != "build_facility" {
		t.Errorf("Type() = %q, want %q", got, "build_facility")
	}
}

// --- Validate ---

func TestBuildFacilityAction_Validate_NilGame(t *testing.T) {
	a := &BuildFacilityAction{CityID: "umeda", Kind: "housing", Scale: 1}
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

func TestBuildFacilityAction_Validate_NoPlayer(t *testing.T) {
	g := NewGame(1907, "hankyu")
	a := &BuildFacilityAction{CityID: "umeda", Kind: "housing", Scale: 1}
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
}

func TestBuildFacilityAction_Validate_CityNotFound(t *testing.T) {
	g := setupFacilityTestGame()
	a := &BuildFacilityAction{CityID: "not_exist", Kind: "housing", Scale: 1}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("存在しない都市でエラーにならなかった")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("型が *ValidationError でない: %T", err)
	}
	if ve.Code != "CITY_NOT_FOUND" {
		t.Errorf("Code = %q, want CITY_NOT_FOUND", ve.Code)
	}
	if !errors.Is(ve, ErrCityNotFound) {
		t.Errorf("ErrCityNotFound をラップしていない")
	}
}

func TestBuildFacilityAction_Validate_UnknownKind(t *testing.T) {
	g := setupFacilityTestGame()
	a := &BuildFacilityAction{CityID: "umeda", Kind: "unknown", Scale: 1}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("未知の種類でエラーにならなかった")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("型が *ValidationError でない: %T", err)
	}
	if ve.Code != "UNKNOWN_FACILITY_KIND" {
		t.Errorf("Code = %q, want UNKNOWN_FACILITY_KIND", ve.Code)
	}
}

func TestBuildFacilityAction_Validate_InvalidScale(t *testing.T) {
	tests := []struct {
		name  string
		scale int
	}{
		{"zero", 0},
		{"six", 6},
		{"negative", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := setupFacilityTestGame()
			a := &BuildFacilityAction{CityID: "umeda", Kind: "housing", Scale: tt.scale}
			err := a.Validate(g)
			if err == nil {
				t.Fatal("不正な規模でエラーにならなかった")
			}
			ve, ok := err.(*ValidationError)
			if !ok {
				t.Fatalf("型が *ValidationError でない: %T", err)
			}
			if ve.Code != "INVALID_SCALE" {
				t.Errorf("Code = %q, want INVALID_SCALE", ve.Code)
			}
		})
	}
}

func TestBuildFacilityAction_Validate_AlreadyOwnedByOther(t *testing.T) {
	g := setupFacilityTestGame()
	// 他社が既に所有
	g.Cities["umeda"].OwnerID = "other"
	a := &BuildFacilityAction{CityID: "umeda", Kind: "housing", Scale: 1}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("他社所有都市でエラーにならなかった")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("型が *ValidationError でない: %T", err)
	}
	if ve.Code != "OWNED_BY_OTHER" {
		t.Errorf("Code = %q, want OWNED_BY_OTHER", ve.Code)
	}
	if !errors.Is(ve, ErrUnauthorized) {
		t.Errorf("ErrUnauthorized をラップしていない")
	}
}

func TestBuildFacilityAction_Validate_InsufficientFunds(t *testing.T) {
	g := setupFacilityTestGame()
	g.PlayerCompany().Cash = 100_000 // 10万円
	a := &BuildFacilityAction{CityID: "umeda", Kind: "department_store", Scale: 5}
	err := a.Validate(g)
	if err == nil {
		t.Fatal("資金不足でエラーにならなかった")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("型が *ValidationError でない: %T", err)
	}
	if ve.Code != "INSUFFICIENT_FUNDS" {
		t.Errorf("Code = %q, want INSUFFICIENT_FUNDS", ve.Code)
	}
	if !errors.Is(ve, ErrInsufficientFunds) {
		t.Errorf("ErrInsufficientFunds をラップしていない")
	}
}

func TestBuildFacilityAction_Validate_OK(t *testing.T) {
	g := setupFacilityTestGame()
	a := &BuildFacilityAction{CityID: "umeda", Kind: "housing", Scale: 2}
	if err := a.Validate(g); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
	// Validate は状態を変更しない
	if g.Cities["umeda"].Developed != 0 {
		t.Errorf("Validate が Developed を変更した")
	}
	if g.PlayerCompany().Cash != 100_000_000 {
		t.Errorf("Validate が Cash を変更した")
	}
}

// --- Apply ---

func TestBuildFacilityAction_Apply(t *testing.T) {
	g := setupFacilityTestGame()
	beforeCash := g.PlayerCompany().Cash
	a := &BuildFacilityAction{CityID: "umeda", Kind: "housing", Scale: 2}

	events, err := a.Apply(g)
	if err != nil {
		t.Fatal(err)
	}

	// 資金減算
	if g.PlayerCompany().Cash >= beforeCash {
		t.Errorf("資金が減っていない: %d → %d", beforeCash, g.PlayerCompany().Cash)
	}

	// Facility 追加
	if len(g.PlayerCompany().Facilities) != 1 {
		t.Fatalf("Facilities 数 = %d, want 1", len(g.PlayerCompany().Facilities))
	}
	f := g.PlayerCompany().Facilities[0]
	if f.CityID != "umeda" {
		t.Errorf("Facility.CityID = %q, want umeda", f.CityID)
	}
	if f.Kind != "housing" {
		t.Errorf("Facility.Kind = %q, want housing", f.Kind)
	}
	if f.Scale != 2 {
		t.Errorf("Facility.Scale = %d, want 2", f.Scale)
	}

	// 開発度上昇（housing: 0.20 × 1.3 = 0.26）
	if g.Cities["umeda"].Developed < 0.25 || g.Cities["umeda"].Developed > 0.27 {
		t.Errorf("Developed = %.2f, want 0.26 前後", g.Cities["umeda"].Developed)
	}

	// OwnerID
	if g.Cities["umeda"].OwnerID != "hankyu" {
		t.Errorf("OwnerID = %q, want hankyu", g.Cities["umeda"].OwnerID)
	}

	// イベント
	if len(events) == 0 {
		t.Fatal("events が空")
	}
	found := false
	for _, e := range events {
		if e.Type == EventFacilityBuilt {
			found = true
			if e.Payload["facility_id"] != f.ID {
				t.Errorf("Payload[facility_id] = %v, want %s", e.Payload["facility_id"], f.ID)
			}
			if e.Payload["kind"] != "housing" {
				t.Errorf("Payload[kind] = %v, want housing", e.Payload["kind"])
			}
		}
	}
	if !found {
		t.Error("facility_built イベントがない")
	}
}

func TestBuildFacilityAction_Apply_CashDecrease(t *testing.T) {
	g := setupFacilityTestGame()
	beforeCash := g.PlayerCompany().Cash
	cost := g.FacilityCost(g.Cities["umeda"], "housing", 2)

	a := &BuildFacilityAction{CityID: "umeda", Kind: "housing", Scale: 2}
	if _, err := a.Apply(g); err != nil {
		t.Fatal(err)
	}

	expected := beforeCash - cost
	if g.PlayerCompany().Cash != expected {
		t.Errorf("Cash = %d, want %d", g.PlayerCompany().Cash, expected)
	}
}

func TestBuildFacilityAction_Apply_RecalcDemand(t *testing.T) {
	g := setupFacilityTestGame()
	beforeDemand := g.Cities["umeda"].Demand

	a := &BuildFacilityAction{CityID: "umeda", Kind: "housing", Scale: 3}
	if _, err := a.Apply(g); err != nil {
		t.Fatal(err)
	}

	afterDemand := g.Cities["umeda"].Demand
	if afterDemand <= beforeDemand {
		t.Errorf("需要が増えていない: %d → %d", beforeDemand, afterDemand)
	}
}

func TestBuildFacilityAction_Apply_ViaRegistry(t *testing.T) {
	g := setupFacilityTestGame()
	payload := []byte(`{"city_id":"umeda","kind":"housing","scale":2}`)
	a, err := DecodeAction("build_facility", payload)
	if err != nil {
		t.Fatal(err)
	}
	events, err := a.Apply(g)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 {
		t.Error("no events")
	}
	if len(g.PlayerCompany().Facilities) != 1 {
		t.Errorf("Facilities 数 = %d, want 1", len(g.PlayerCompany().Facilities))
	}
}

// --- 施設収入 ---

func TestFacility_Income(t *testing.T) {
	g := setupFacilityTestGame()
	a := &BuildFacilityAction{CityID: "umeda", Kind: "department_store", Scale: 3}
	if _, err := a.Apply(g); err != nil {
		t.Fatal(err)
	}
	f := g.PlayerCompany().Facilities[0]
	if f.Income <= 0 {
		t.Errorf("Facility.Income = %d, want > 0", f.Income)
	}
}

// TestBuildFacilityAction_FacilityIncome_EachTurn は施設収入が毎ターン
// 資金に加算されることを検証する。路線がなければ資金増加は施設収入のみ。
func TestBuildFacilityAction_FacilityIncome_EachTurn(t *testing.T) {
	g := setupFacilityTestGame()
	a := &BuildFacilityAction{CityID: "umeda", Kind: "housing", Scale: 2}
	if _, err := a.Apply(g); err != nil {
		t.Fatal(err)
	}
	f := g.PlayerCompany().Facilities[0]
	if f.Income <= 0 {
		t.Fatalf("Income = %d, want > 0", f.Income)
	}

	before := g.PlayerCompany().Cash
	g.EndTurn()
	after := g.PlayerCompany().Cash

	// 路線なし・減価償却なしのため、資金増加は施設収入のみ
	if after-before != f.Income {
		t.Errorf("資金増加 = %d, want 施設収入 %d", after-before, f.Income)
	}
}

// --- テーブル ---

func TestFacilityByKind(t *testing.T) {
	def, ok := FacilityByKind("housing")
	if !ok {
		t.Fatal("housing が見つからない")
	}
	if def.BaseCost != 3_000_000 {
		t.Errorf("housing BaseCost = %d, want 3000000", def.BaseCost)
	}
	if _, ok := FacilityByKind("unknown"); ok {
		t.Error("unknown が見つかった")
	}
}

func TestScaleDevelopFactor(t *testing.T) {
	tests := []struct {
		scale int
		want  float64
	}{
		{0, 1.0}, // 下限クランプ
		{1, 1.0},
		{2, 1.3},
		{3, 1.6},
		{5, 2.2},
		{9, 2.2}, // 上限クランプ
	}
	for _, tt := range tests {
		if got := ScaleDevelopFactor(tt.scale); got < tt.want-0.001 || got > tt.want+0.001 {
			t.Errorf("ScaleDevelopFactor(%d) = %v, want %v", tt.scale, got, tt.want)
		}
	}
}

// --- 決定論 ---

func TestBuildFacilityAction_Deterministic(t *testing.T) {
	g1 := setupFacilityTestGame()
	g2 := setupFacilityTestGame()

	a1 := &BuildFacilityAction{CityID: "umeda", Kind: "housing", Scale: 2}
	a2 := &BuildFacilityAction{CityID: "umeda", Kind: "housing", Scale: 2}

	e1, _ := a1.Apply(g1)
	e2, _ := a2.Apply(g2)

	if len(e1) != len(e2) {
		t.Fatalf("events 数が異なる: %d vs %d", len(e1), len(e2))
	}
	for i := range e1 {
		if e1[i].Message != e2[i].Message {
			t.Errorf("Message が異なる: %q vs %q", e1[i].Message, e2[i].Message)
		}
		if e1[i].ID != e2[i].ID {
			t.Errorf("ID が異なる: %q vs %q", e1[i].ID, e2[i].ID)
		}
	}
}

// --- ヘルパー ---

func setupFacilityTestGame() *Game {
	g := NewGame(1907, "hankyu")
	g.Date = Date{Year: 1910, Month: 1}

	g.AddCity(&City{
		ID: "umeda", Name: "梅田", Population: 120000,
		Lat: 34.7055, Lon: 135.4983, LandValue: 60,
	})
	g.AddCity(&City{
		ID: "takarazuka", Name: "宝塚", Population: 30000,
		Lat: 34.7995, Lon: 135.3468, LandValue: 20,
	})

	hankyu := NewCompany("hankyu", "hankyu", "阪急")
	// 施設建設テストでは高額施設（百貨店 規模5・梅田の地価 60 で最大 3,300 万円。
	// Phase 1-8-2.5 で基本コストを旧比 0.6 倍に削減）を建てるため、
	// 十分な初期資金を持たせる。
	hankyu.Cash = 100_000_000
	g.AddCompany(hankyu)

	return g
}
