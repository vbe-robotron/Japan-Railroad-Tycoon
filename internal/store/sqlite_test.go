package store

import (
	"context"
	"testing"

	"github.com/you/jrt/internal/core"
)

// SQLiteStore が Store インターフェースを満たすことをコンパイル時に確認する。
var _ Store = (*SQLiteStore)(nil)

// setupTestGame はテスト用のゲームを構築する（都市 2 件・会社 1 件・路線 1 本）。
func setupTestGame() *core.Game {
	g := core.NewGame(1907, "hankyu")
	g.Date = core.Date{Year: 1910, Month: 3}

	g.AddCity(&core.City{
		ID: "umeda", Name: "梅田", Prefecture: "大阪府",
		Lat: 34.7055, Lon: 135.4983, Population: 120000, LandValue: 60,
	})
	g.AddCity(&core.City{
		ID: "takarazuka", Name: "宝塚", Prefecture: "兵庫県",
		Lat: 34.7995, Lon: 135.3468, Population: 30000, LandValue: 20,
	})

	hankyu := core.NewCompany("hankyu", "hankyu", "阪急")
	hankyu.Cash = 9_000_000
	hankyu.TotalShares = 10000
	hankyu.StockPrice = 1000
	hankyu.Reputation = 50
	g.AddCompany(hankyu)

	a := &core.BuildLineAction{From: "umeda", To: "takarazuka", Gauge: 1067}
	if _, err := a.Apply(g); err != nil {
		panic(err)
	}
	return g
}

// newTestStore はテスト用のインメモリ SQLiteStore を生成する。
func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	st, err := NewSQLite(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	st := newTestStore(t)
	g := setupTestGame()
	if err := st.SaveGame(context.Background(), g); err != nil {
		t.Fatal(err)
	}

	loaded, err := st.LoadGame(context.Background(), g.ID)
	if err != nil {
		t.Fatal(err)
	}

	if loaded.ID != g.ID {
		t.Errorf("ID")
	}
	if loaded.Seed != g.Seed {
		t.Errorf("Seed")
	}
	if loaded.Date != g.Date {
		t.Errorf("Date: %v vs %v", loaded.Date, g.Date)
	}
	if loaded.Turn != g.Turn {
		t.Errorf("Turn")
	}
	if loaded.IDGen.Counter != g.IDGen.Counter {
		t.Errorf("IDGen.Counter: %d vs %d", loaded.IDGen.Counter, g.IDGen.Counter)
	}
	if len(loaded.Companies) != len(g.Companies) {
		t.Errorf("Companies")
	}
	if len(loaded.Cities) != len(g.Cities) {
		t.Errorf("Cities")
	}

	for id, c := range g.Companies {
		lc := loaded.Companies[id]
		if lc == nil {
			t.Fatalf("company %s missing", id)
		}
		if len(lc.Lines) != len(c.Lines) {
			t.Errorf("Lines for %s: %d vs %d", id, len(lc.Lines), len(c.Lines))
		}
		if len(lc.Facilities) != len(c.Facilities) {
			t.Errorf("Facilities for %s", id)
		}
	}
}

func TestSaveLoad_FieldRoundTrip(t *testing.T) {
	st := newTestStore(t)
	g := setupTestGame()
	c := g.Companies["hankyu"]
	c.LegalName = "阪急電鉄株式会社"
	c.FoundedYear = 1907
	c.Debt = 1234
	c.IsPlayer = true
	c.IsNational = false
	c.Region = "kansai"
	c.Tags = []string{"private", "major"}
	c.ParentID = "parent"
	c.CreditLine = 500
	c.Holding = map[string]int{"keihan": 300}
	c.Facilities = append(c.Facilities, core.NewFacility("fac1", "hankyu", "umeda", core.FacilityHotel, 3))
	line := c.Lines[0]
	line.Double = true
	line.Electrified = true
	line.Fare = 200
	line.TrainsPerDay = 24
	line.Profit = 999
	line.OperatingRatio = 1.25
	line.Condition = 0.8
	line.ConsecutiveDeficitTurns = 4
	line.AccumulatedLoss = 5555
	line.PoliticalPressure = 0.3

	if err := st.SaveGame(context.Background(), g); err != nil {
		t.Fatal(err)
	}
	loaded, err := st.LoadGame(context.Background(), g.ID)
	if err != nil {
		t.Fatal(err)
	}
	lc := loaded.Companies["hankyu"]
	if lc.LegalName != c.LegalName || lc.FoundedYear != c.FoundedYear || lc.Debt != c.Debt {
		t.Errorf("company scalar mismatch: %+v", lc)
	}
	if !lc.IsPlayer || lc.IsNational || lc.Region != "kansai" || lc.ParentID != "parent" {
		t.Errorf("company flags mismatch: %+v", lc)
	}
	if lc.CreditLine != 500 || lc.Holding["keihan"] != 300 {
		t.Errorf("credit/holding mismatch: %+v", lc)
	}
	if len(lc.Tags) != 2 || lc.Tags[0] != "private" {
		t.Errorf("tags mismatch: %v", lc.Tags)
	}
	if len(lc.Facilities) != 1 || lc.Facilities[0].Kind != core.FacilityHotel || lc.Facilities[0].Scale != 3 {
		t.Errorf("facilities mismatch: %+v", lc.Facilities)
	}
	ll := lc.Lines[0]
	if !ll.Double || !ll.Electrified || ll.Fare != 200 || ll.TrainsPerDay != 24 {
		t.Errorf("line flags mismatch: %+v", ll)
	}
	if ll.Profit != 999 || ll.OperatingRatio != 1.25 || ll.Condition != 0.8 {
		t.Errorf("line numbers mismatch: %+v", ll)
	}
	if ll.ConsecutiveDeficitTurns != 4 || ll.AccumulatedLoss != 5555 || ll.PoliticalPressure != 0.3 {
		t.Errorf("line deficit mismatch: %+v", ll)
	}
}

func TestLoadGame_EnsureRng(t *testing.T) {
	st := newTestStore(t)
	g := setupTestGame()
	if err := st.SaveGame(context.Background(), g); err != nil {
		t.Fatal(err)
	}

	loaded, err := st.LoadGame(context.Background(), g.ID)
	if err != nil {
		t.Fatal(err)
	}

	events := loaded.EndTurn()
	if len(events) == 0 {
		t.Error("EndTurn failed")
	}
}

func TestListGames_DeleteGame(t *testing.T) {
	st := newTestStore(t)
	g1 := setupTestGame()
	g2 := setupTestGame()
	g2.ID = g2.IDGen.Next()

	if err := st.SaveGame(context.Background(), g1); err != nil {
		t.Fatal(err)
	}
	if err := st.SaveGame(context.Background(), g2); err != nil {
		t.Fatal(err)
	}

	ids, err := st.ListGames(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Errorf("ListGames = %d, want 2", len(ids))
	}

	if err := st.DeleteGame(context.Background(), g1.ID); err != nil {
		t.Fatal(err)
	}
	ids, _ = st.ListGames(context.Background())
	if len(ids) != 1 {
		t.Errorf("after delete = %d, want 1", len(ids))
	}

	// 削除後の LoadGame はエラー
	if _, err := st.LoadGame(context.Background(), g1.ID); err == nil {
		t.Error("LoadGame after delete should error")
	}
}

func TestDeleteGame_Nonexistent(t *testing.T) {
	st := newTestStore(t)
	if err := st.DeleteGame(context.Background(), "nope"); err != nil {
		t.Errorf("DeleteGame nonexistent = %v", err)
	}
}

func TestAppendEvents(t *testing.T) {
	st := newTestStore(t)
	g := setupTestGame()
	if err := st.SaveGame(context.Background(), g); err != nil {
		t.Fatal(err)
	}

	events := g.EndTurn()
	if err := st.AppendEvents(context.Background(), g.ID, events); err != nil {
		t.Fatal(err)
	}
	// 2 回目は seq が継続する
	if err := st.AppendEvents(context.Background(), g.ID, g.EndTurn()); err != nil {
		t.Fatal(err)
	}
}

func TestAppendEvents_Empty(t *testing.T) {
	st := newTestStore(t)
	if err := st.AppendEvents(context.Background(), "game", nil); err != nil {
		t.Errorf("empty AppendEvents = %v", err)
	}
}

func TestAppendAction(t *testing.T) {
	st := newTestStore(t)
	g := setupTestGame()
	if err := st.SaveGame(context.Background(), g); err != nil {
		t.Fatal(err)
	}

	payload := []byte(`{"from":"umeda","to":"takarazuka"}`)
	if err := st.AppendAction(context.Background(), g.ID, "hankyu", "build_line", payload); err != nil {
		t.Fatal(err)
	}
	if err := st.AppendAction(context.Background(), g.ID, "hankyu", "end_turn", []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
}

func TestAppendAction_GameNotFound(t *testing.T) {
	st := newTestStore(t)
	err := st.AppendAction(context.Background(), "missing", "hankyu", "end_turn", []byte(`{}`))
	if err == nil {
		t.Error("AppendAction for missing game should error")
	}
}

func TestHoldingJSONRoundTrip(t *testing.T) {
	st := newTestStore(t)
	g := setupTestGame()
	hankyu := g.Companies["hankyu"]
	hankyu.Holding = map[string]int{"hanshin": 5000, "keihan": 3000}
	if err := st.SaveGame(context.Background(), g); err != nil {
		t.Fatal(err)
	}

	loaded, err := st.LoadGame(context.Background(), g.ID)
	if err != nil {
		t.Fatal(err)
	}
	lc := loaded.Companies["hankyu"]
	if lc.Holding["hanshin"] != 5000 {
		t.Errorf("Holding = %v", lc.Holding)
	}
}

func TestDissolvedJSONRoundTrip(t *testing.T) {
	st := newTestStore(t)
	g := setupTestGame()
	hankyu := g.Companies["hankyu"]
	hankyu.Dissolved = &core.Date{Year: 1987, Month: 4}
	if err := st.SaveGame(context.Background(), g); err != nil {
		t.Fatal(err)
	}

	loaded, err := st.LoadGame(context.Background(), g.ID)
	if err != nil {
		t.Fatal(err)
	}
	lc := loaded.Companies["hankyu"]
	if lc.Dissolved == nil || lc.Dissolved.Year != 1987 || lc.Dissolved.Month != 4 {
		t.Errorf("Dissolved = %v", lc.Dissolved)
	}
}

func TestNewSQLite_InvalidPath(t *testing.T) {
	if _, err := NewSQLite("/nonexistent-dir-xyz/jrt.db"); err == nil {
		t.Error("NewSQLite with invalid path should error")
	}
}

func TestClose(t *testing.T) {
	st, err := NewSQLite(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Close(); err != nil {
		t.Errorf("Close = %v", err)
	}
}

func TestSaveGame_Idempotent(t *testing.T) {
	st := newTestStore(t)
	g := setupTestGame()
	ctx := context.Background()
	if err := st.SaveGame(ctx, g); err != nil {
		t.Fatal(err)
	}
	// 路線を削除して再保存 → DELETE → INSERT で反映される
	g.Companies["hankyu"].Lines = nil
	if err := st.SaveGame(ctx, g); err != nil {
		t.Fatal(err)
	}
	loaded, err := st.LoadGame(ctx, g.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Companies["hankyu"].Lines) != 0 {
		t.Errorf("lines should be empty, got %d", len(loaded.Companies["hankyu"].Lines))
	}
}

func TestLoadGame_NotFound(t *testing.T) {
	st := newTestStore(t)
	_, err := st.LoadGame(context.Background(), "missing")
	if err == nil {
		t.Fatal("LoadGame missing should error")
	}
}

// TestLoadGame_CorruptJSON は壊れた JSON 列で LoadGame がエラーになることを確認する。
func TestLoadGame_CorruptJSON(t *testing.T) {
	ctx := context.Background()
	queries := map[string]string{
		"city_tags":       "UPDATE cities SET tags='{' WHERE game_id=?",
		"company_tags":    "UPDATE companies SET tags='{' WHERE game_id=?",
		"company_holding": "UPDATE companies SET holding='{' WHERE game_id=?",
	}
	for name, q := range queries {
		t.Run(name, func(t *testing.T) {
			st := newTestStore(t)
			g := setupTestGame()
			if err := st.SaveGame(ctx, g); err != nil {
				t.Fatal(err)
			}
			if _, err := st.db.ExecContext(ctx, q, g.ID); err != nil {
				t.Fatal(err)
			}
			if _, err := st.LoadGame(ctx, g.ID); err == nil {
				t.Errorf("%s: expected error", name)
			}
		})
	}
}

// TestAppendEvents_MarshalError は JSON 化できない payload でエラーになることを確認する。
func TestAppendEvents_MarshalError(t *testing.T) {
	st := newTestStore(t)
	g := setupTestGame()
	if err := st.SaveGame(context.Background(), g); err != nil {
		t.Fatal(err)
	}
	e := core.Event{
		ID: "e1", Type: "test",
		Payload: map[string]any{"bad": make(chan int)},
	}
	if err := st.AppendEvents(context.Background(), g.ID, []core.Event{e}); err == nil {
		t.Error("AppendEvents with unmarshalable payload should error")
	}
}

// TestSaveGame_LineFKError は存在しない会社を参照する路線でエラーになることを確認する。
func TestSaveGame_LineFKError(t *testing.T) {
	st := newTestStore(t)
	g := setupTestGame()
	g.Companies["hankyu"].Lines[0].CompanyID = "ghost"
	if err := st.SaveGame(context.Background(), g); err == nil {
		t.Error("SaveGame with ghost line company should error")
	}
}

// TestSaveGame_FacilityFKError は存在しない会社を参照する施設でエラーになることを確認する。
func TestSaveGame_FacilityFKError(t *testing.T) {
	st := newTestStore(t)
	g := setupTestGame()
	g.Companies["hankyu"].Facilities = append(
		g.Companies["hankyu"].Facilities,
		core.NewFacility("fac", "ghost", "umeda", core.FacilityHotel, 1))
	if err := st.SaveGame(context.Background(), g); err == nil {
		t.Error("SaveGame with ghost facility company should error")
	}
}

// TestClosedDB_Errors は DB クローズ後の各メソッドがエラーを返すことを確認する。
func TestClosedDB_Errors(t *testing.T) {
	st, err := NewSQLite(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Close(); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	g := setupTestGame()

	if err := st.SaveGame(ctx, g); err == nil {
		t.Error("SaveGame on closed DB should error")
	}
	if _, err := st.LoadGame(ctx, "x"); err == nil {
		t.Error("LoadGame on closed DB should error")
	}
	if _, err := st.ListGames(ctx); err == nil {
		t.Error("ListGames on closed DB should error")
	}
	if err := st.DeleteGame(ctx, "x"); err == nil {
		t.Error("DeleteGame on closed DB should error")
	}
	if err := st.AppendEvents(ctx, "x", []core.Event{{ID: "e"}}); err == nil {
		t.Error("AppendEvents on closed DB should error")
	}
	if err := st.AppendAction(ctx, "x", "c", "t", nil); err == nil {
		t.Error("AppendAction on closed DB should error")
	}
}

// TestLoadGame_ScanError は NULL 列により LoadGame がエラーになることを確認する。
func TestLoadGame_ScanError(t *testing.T) {
	ctx := context.Background()
	queries := map[string]string{
		"city_lat":       "UPDATE cities SET lat='abc' WHERE game_id=?",
		"company_cash":   "UPDATE companies SET cash='abc' WHERE game_id=?",
		"line_length_km": "UPDATE lines SET length_km='abc' WHERE game_id=?",
	}
	for name, q := range queries {
		t.Run(name, func(t *testing.T) {
			st := newTestStore(t)
			g := setupTestGame()
			if err := st.SaveGame(ctx, g); err != nil {
				t.Fatal(err)
			}
			if _, err := st.db.ExecContext(ctx, q, g.ID); err != nil {
				t.Fatal(err)
			}
			if _, err := st.LoadGame(ctx, g.ID); err == nil {
				t.Errorf("%s: expected error", name)
			}
		})
	}
}

// TestLoadGame_FacilityScanError は施設の NULL 列で LoadGame がエラーになることを確認する。
func TestLoadGame_FacilityScanError(t *testing.T) {
	ctx := context.Background()
	st := newTestStore(t)
	g := setupTestGame()
	g.Companies["hankyu"].Facilities = append(
		g.Companies["hankyu"].Facilities,
		core.NewFacility("fac", "hankyu", "umeda", core.FacilityHotel, 1))
	if err := st.SaveGame(ctx, g); err != nil {
		t.Fatal(err)
	}
	if _, err := st.db.ExecContext(ctx,
		"UPDATE facilities SET scale='abc' WHERE game_id=?", g.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := st.LoadGame(ctx, g.ID); err == nil {
		t.Error("expected error from facility scan")
	}
}
