package store

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/you/jrt/internal/core"
	_ "modernc.org/sqlite"
)

// schema は初期マイグレーション SQL を埋め込んだもの。
//
//go:embed migrations/001_init.sql
var schema string

// SQLiteStore は Store の SQLite 実装。
//
// 単一接続（MaxOpenConns=1）に固定することで、:memory: データベースでも
// 同一接続が使われ、かつ書き込みの直列化を保証する。
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLite は SQLite ストアを開き、マイグレーションを適用する。
func NewSQLite(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &SQLiteStore{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// Close はデータベース接続を閉じる。
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// migrate はスキーマを適用する。CREATE TABLE IF NOT EXISTS のため冪等。
func (s *SQLiteStore) migrate() error {
	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

// boolToInt は bool を SQLite の INTEGER（0/1）へ変換する。
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// SaveGame はゲーム状態をトランザクション内で保存する。
//
// games を upsert した後、子テーブルを DELETE → INSERT で置き換える。
// これにより削除された路線・施設・会社も正しく反映される。
func (s *SQLiteStore) SaveGame(ctx context.Context, g *core.Game) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// games upsert
	_, err = tx.ExecContext(ctx, `
        INSERT INTO games(id, seed, id_counter, player_id, year, month, turn, updated_at)
        VALUES(?,?,?,?,?,?,?,CURRENT_TIMESTAMP)
        ON CONFLICT(id) DO UPDATE SET
            seed=excluded.seed,
            id_counter=excluded.id_counter,
            player_id=excluded.player_id,
            year=excluded.year,
            month=excluded.month,
            turn=excluded.turn,
            updated_at=CURRENT_TIMESTAMP`,
		g.ID, g.Seed, g.IDGen.Counter, g.PlayerID,
		g.Date.Year, g.Date.Month, g.Turn)
	if err != nil {
		return err
	}

	// 子テーブルを DELETE → INSERT（FK 違反を避けるため子から先に削除）
	for _, table := range []string{"lines", "facilities", "companies", "cities"} {
		if _, err := tx.ExecContext(ctx,
			"DELETE FROM "+table+" WHERE game_id=?", g.ID); err != nil {
			return err
		}
	}

	// companies
	for _, c := range g.Companies {
		tags, err := json.Marshal(c.Tags)
		if err != nil {
			return err
		}
		holding, err := json.Marshal(c.Holding)
		if err != nil {
			return err
		}
		var dy, dm sql.NullInt64
		if c.Dissolved != nil {
			dy = sql.NullInt64{Int64: int64(c.Dissolved.Year), Valid: true}
			dm = sql.NullInt64{Int64: int64(c.Dissolved.Month), Valid: true}
		}
		_, err = tx.ExecContext(ctx, `
            INSERT INTO companies(game_id, id, group_id, name, legal_name,
                founded_year, cash, debt, stock_price, total_shares,
                is_player, is_national, region, tags, parent_id,
                reputation, credit_line, dissolved_year, dissolved_month, holding)
            VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			g.ID, c.ID, c.GroupID, c.Name, c.LegalName,
			c.FoundedYear, c.Cash, c.Debt, c.StockPrice, c.TotalShares,
			boolToInt(c.IsPlayer), boolToInt(c.IsNational), c.Region,
			string(tags), c.ParentID, c.Reputation, c.CreditLine,
			dy, dm, string(holding))
		if err != nil {
			return err
		}
	}

	// cities
	for _, city := range g.Cities {
		tags, err := json.Marshal(city.Tags)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `
            INSERT INTO cities(game_id, id, name, prefecture, representative_node,
                lat, lon, population, land_value, demand, developed,
                owner_id, political_pressure, residents_affected, tags)
            VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			g.ID, city.ID, city.Name, city.Prefecture, city.RepresentativeNode,
			city.Lat, city.Lon, city.Population, city.LandValue, city.Demand,
			city.Developed, city.OwnerID, city.PoliticalPressure,
			city.ResidentsAffected, string(tags))
		if err != nil {
			return err
		}
	}

	// lines
	for _, c := range g.Companies {
		for _, l := range c.Lines {
			_, err := tx.ExecContext(ctx, `
                INSERT INTO lines(game_id, id, company_id, from_city, to_city,
                    display_name, length_km, gauge, double, electrified,
                    fare, trains_per_day, opened_year, opened_month,
                    profit, operating_ratio, condition,
                    consecutive_deficit_turns, accumulated_loss, political_pressure)
                VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
				g.ID, l.ID, l.CompanyID, l.From, l.To,
				l.DisplayName, l.LengthKm, l.Gauge, boolToInt(l.Double),
				boolToInt(l.Electrified), l.Fare, l.TrainsPerDay,
				l.Opened.Year, l.Opened.Month,
				l.Profit, l.OperatingRatio, l.Condition,
				l.ConsecutiveDeficitTurns, l.AccumulatedLoss, l.PoliticalPressure)
			if err != nil {
				return err
			}
		}
	}

	// facilities
	for _, c := range g.Companies {
		for _, f := range c.Facilities {
			_, err := tx.ExecContext(ctx, `
                INSERT INTO facilities(game_id, id, company_id, city_id, kind, scale,
                    built_year, built_month, income)
                VALUES(?,?,?,?,?,?,?,?,?)`,
				g.ID, f.ID, f.CompanyID, f.CityID, f.Kind, f.Scale,
				f.BuiltAt.Year, f.BuiltAt.Month, f.Income)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// LoadGame は指定した ID のゲームを復元する。
func (s *SQLiteStore) LoadGame(ctx context.Context, id string) (*core.Game, error) {
	g := &core.Game{
		ID:        id,
		Companies: map[string]*core.Company{},
		Cities:    map[string]*core.City{},
	}

	// games
	var idCounter int64
	err := s.db.QueryRowContext(ctx, `
        SELECT seed, id_counter, player_id, year, month, turn
        FROM games WHERE id=?`, id).Scan(
		&g.Seed, &idCounter, &g.PlayerID,
		&g.Date.Year, &g.Date.Month, &g.Turn)
	if err != nil {
		return nil, err
	}

	// IDGenerator を復元
	g.IDGen = core.NewIDGenerator(g.Seed)
	g.IDGen.Counter = idCounter

	if err := s.loadCities(ctx, g); err != nil {
		return nil, err
	}
	if err := s.loadCompanies(ctx, g); err != nil {
		return nil, err
	}
	if err := s.loadLines(ctx, g); err != nil {
		return nil, err
	}
	if err := s.loadFacilities(ctx, g); err != nil {
		return nil, err
	}

	// RNG を復元
	g.EnsureRng()

	return g, nil
}

// loadCities は cities テーブルから都市を読み込む。
func (s *SQLiteStore) loadCities(ctx context.Context, g *core.Game) error {
	rows, err := s.db.QueryContext(ctx, `
        SELECT id, name, prefecture, representative_node,
               lat, lon, population, land_value, demand, developed,
               owner_id, political_pressure, residents_affected, tags
        FROM cities WHERE game_id=?`, g.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var c core.City
		var tagsJSON string
		if err := rows.Scan(&c.ID, &c.Name, &c.Prefecture, &c.RepresentativeNode,
			&c.Lat, &c.Lon, &c.Population, &c.LandValue, &c.Demand, &c.Developed,
			&c.OwnerID, &c.PoliticalPressure, &c.ResidentsAffected, &tagsJSON); err != nil {
			return err
		}
		if err := json.Unmarshal([]byte(tagsJSON), &c.Tags); err != nil {
			return err
		}
		g.Cities[c.ID] = &c
	}
	return rows.Err()
}

// loadCompanies は companies テーブルから会社を読み込む。
func (s *SQLiteStore) loadCompanies(ctx context.Context, g *core.Game) error {
	rows, err := s.db.QueryContext(ctx, `
        SELECT id, group_id, name, legal_name, founded_year, cash, debt,
               stock_price, total_shares, is_player, is_national, region,
               tags, parent_id, reputation, credit_line,
               dissolved_year, dissolved_month, holding
        FROM companies WHERE game_id=?`, g.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var c core.Company
		var tagsJSON, holdingJSON string
		var isPlayer, isNational int
		var dy, dm sql.NullInt64
		if err := rows.Scan(&c.ID, &c.GroupID, &c.Name, &c.LegalName,
			&c.FoundedYear, &c.Cash, &c.Debt, &c.StockPrice, &c.TotalShares,
			&isPlayer, &isNational, &c.Region, &tagsJSON, &c.ParentID,
			&c.Reputation, &c.CreditLine, &dy, &dm, &holdingJSON); err != nil {
			return err
		}
		c.IsPlayer = isPlayer != 0
		c.IsNational = isNational != 0
		if err := json.Unmarshal([]byte(tagsJSON), &c.Tags); err != nil {
			return err
		}
		if err := json.Unmarshal([]byte(holdingJSON), &c.Holding); err != nil {
			return err
		}
		if c.Holding == nil {
			c.Holding = map[string]int{}
		}
		if c.Lines == nil {
			c.Lines = []*core.Line{}
		}
		if c.Facilities == nil {
			c.Facilities = []*core.Facility{}
		}
		if dy.Valid && dm.Valid {
			c.Dissolved = &core.Date{Year: int(dy.Int64), Month: int(dm.Int64)}
		}
		g.Companies[c.ID] = &c
	}
	return rows.Err()
}

// loadLines は lines テーブルから路線を読み込み、会社へ紐づける。
func (s *SQLiteStore) loadLines(ctx context.Context, g *core.Game) error {
	rows, err := s.db.QueryContext(ctx, `
        SELECT id, company_id, from_city, to_city, display_name, length_km,
               gauge, double, electrified, fare, trains_per_day,
               opened_year, opened_month, profit, operating_ratio, condition,
               consecutive_deficit_turns, accumulated_loss, political_pressure
        FROM lines WHERE game_id=?`, g.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var l core.Line
		var isDouble, isElectrified int
		if err := rows.Scan(&l.ID, &l.CompanyID, &l.From, &l.To, &l.DisplayName,
			&l.LengthKm, &l.Gauge, &isDouble, &isElectrified, &l.Fare,
			&l.TrainsPerDay, &l.Opened.Year, &l.Opened.Month,
			&l.Profit, &l.OperatingRatio, &l.Condition,
			&l.ConsecutiveDeficitTurns, &l.AccumulatedLoss, &l.PoliticalPressure); err != nil {
			return err
		}
		l.Double = isDouble != 0
		l.Electrified = isElectrified != 0
		if c, ok := g.Companies[l.CompanyID]; ok {
			c.Lines = append(c.Lines, &l)
		}
	}
	return rows.Err()
}

// loadFacilities は facilities テーブルから施設を読み込み、会社へ紐づける。
func (s *SQLiteStore) loadFacilities(ctx context.Context, g *core.Game) error {
	rows, err := s.db.QueryContext(ctx, `
        SELECT id, company_id, city_id, kind, scale, built_year, built_month, income
        FROM facilities WHERE game_id=?`, g.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var f core.Facility
		if err := rows.Scan(&f.ID, &f.CompanyID, &f.CityID, &f.Kind, &f.Scale,
			&f.BuiltAt.Year, &f.BuiltAt.Month, &f.Income); err != nil {
			return err
		}
		if c, ok := g.Companies[f.CompanyID]; ok {
			c.Facilities = append(c.Facilities, &f)
		}
	}
	return rows.Err()
}

// ListGames は保存済みゲームの ID を更新日時の新しい順に返す。
func (s *SQLiteStore) ListGames(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id FROM games ORDER BY updated_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// DeleteGame は指定したゲームを削除する（子テーブルは CASCADE）。
func (s *SQLiteStore) DeleteGame(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM games WHERE id=?", id)
	return err
}

// AppendEvents はイベント列を seq を採番して追記する。
func (s *SQLiteStore) AppendEvents(ctx context.Context, gameID string, events []core.Event) error {
	if len(events) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var maxSeq int
	err = tx.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(seq), 0) FROM events WHERE game_id=?", gameID).Scan(&maxSeq)
	if err != nil {
		return err
	}

	for i, e := range events {
		payload, err := json.Marshal(e.Payload)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `
            INSERT INTO events(game_id, id, seq, type, date_year, date_month,
                turn, severity, message, payload, company_id, city_id, line_id)
            VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			gameID, e.ID, maxSeq+i+1, e.Type, e.Date.Year, e.Date.Month,
			e.Turn, e.Severity, e.Message, string(payload),
			e.CompanyID, e.CityID, e.LineID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// AppendAction はアクションログを seq を採番して追記する。
func (s *SQLiteStore) AppendAction(ctx context.Context, gameID, companyID, actionType string, payload []byte) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var maxSeq int
	err = tx.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(seq), 0) FROM actions_log WHERE game_id=?", gameID).Scan(&maxSeq)
	if err != nil {
		return err
	}

	var turn int
	err = tx.QueryRowContext(ctx, "SELECT turn FROM games WHERE id=?", gameID).Scan(&turn)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
        INSERT INTO actions_log(game_id, seq, turn, company_id, action_type, payload)
        VALUES(?,?,?,?,?,?)`,
		gameID, maxSeq+1, turn, companyID, actionType, string(payload))
	if err != nil {
		return err
	}
	return tx.Commit()
}
