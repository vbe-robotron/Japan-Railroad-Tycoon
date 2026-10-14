-- 001_init.sql はゲーム永続化の初期スキーマを定義する。
--
-- すべての子テーブルは games(id) に外部キーで紐づき、ゲーム削除時に
-- CASCADE で消去される。map / slice などの可変長データは JSON 列
-- （payload, holding, tags）に格納する。

CREATE TABLE IF NOT EXISTS games (
    id          TEXT PRIMARY KEY,
    seed        INTEGER NOT NULL,
    id_counter  INTEGER NOT NULL DEFAULT 0,
    player_id   TEXT NOT NULL,
    year        INTEGER NOT NULL,
    month       INTEGER NOT NULL,
    turn        INTEGER NOT NULL,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS companies (
    game_id      TEXT NOT NULL,
    id           TEXT NOT NULL,
    group_id     TEXT NOT NULL,
    name         TEXT NOT NULL,
    legal_name   TEXT NOT NULL DEFAULT '',
    founded_year INTEGER NOT NULL DEFAULT 0,
    cash         INTEGER NOT NULL DEFAULT 0,
    debt         INTEGER NOT NULL DEFAULT 0,
    stock_price  INTEGER NOT NULL DEFAULT 0,
    total_shares INTEGER NOT NULL DEFAULT 0,
    is_player    INTEGER NOT NULL DEFAULT 0,
    is_national  INTEGER NOT NULL DEFAULT 0,
    region       TEXT NOT NULL DEFAULT '',
    tags         TEXT NOT NULL DEFAULT '[]',
    parent_id    TEXT NOT NULL DEFAULT '',
    reputation   INTEGER NOT NULL DEFAULT 50,
    credit_line  INTEGER NOT NULL DEFAULT 0,
    dissolved_year INTEGER,
    dissolved_month INTEGER,
    holding      TEXT NOT NULL DEFAULT '{}',
    PRIMARY KEY (game_id, id),
    FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS cities (
    game_id      TEXT NOT NULL,
    id           TEXT NOT NULL,
    name         TEXT NOT NULL,
    prefecture   TEXT NOT NULL DEFAULT '',
    representative_node TEXT NOT NULL DEFAULT '',
    lat          REAL NOT NULL DEFAULT 0,
    lon          REAL NOT NULL DEFAULT 0,
    population   INTEGER NOT NULL DEFAULT 0,
    land_value   INTEGER NOT NULL DEFAULT 0,
    demand       INTEGER NOT NULL DEFAULT 0,
    developed    REAL NOT NULL DEFAULT 0,
    owner_id     TEXT NOT NULL DEFAULT '',
    political_pressure REAL NOT NULL DEFAULT 0,
    residents_affected INTEGER NOT NULL DEFAULT 0,
    tags         TEXT NOT NULL DEFAULT '[]',
    PRIMARY KEY (game_id, id)
);

CREATE TABLE IF NOT EXISTS lines (
    game_id      TEXT NOT NULL,
    id           TEXT NOT NULL,
    company_id   TEXT NOT NULL,
    from_city    TEXT NOT NULL,
    to_city      TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    length_km    REAL NOT NULL DEFAULT 0,
    gauge        INTEGER NOT NULL DEFAULT 1067,
    double       INTEGER NOT NULL DEFAULT 0,
    electrified  INTEGER NOT NULL DEFAULT 0,
    fare         INTEGER NOT NULL DEFAULT 0,
    trains_per_day INTEGER NOT NULL DEFAULT 10,
    opened_year  INTEGER NOT NULL DEFAULT 0,
    opened_month INTEGER NOT NULL DEFAULT 1,
    profit       INTEGER NOT NULL DEFAULT 0,
    operating_ratio REAL NOT NULL DEFAULT 0,
    condition    REAL NOT NULL DEFAULT 1.0,
    consecutive_deficit_turns INTEGER NOT NULL DEFAULT 0,
    accumulated_loss INTEGER NOT NULL DEFAULT 0,
    political_pressure REAL NOT NULL DEFAULT 0,
    PRIMARY KEY (game_id, id),
    FOREIGN KEY (game_id, company_id) REFERENCES companies(game_id, id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS facilities (
    game_id      TEXT NOT NULL,
    id           TEXT NOT NULL,
    company_id   TEXT NOT NULL,
    city_id      TEXT NOT NULL,
    kind         TEXT NOT NULL,
    scale        INTEGER NOT NULL DEFAULT 1,
    built_year   INTEGER NOT NULL DEFAULT 0,
    built_month  INTEGER NOT NULL DEFAULT 1,
    income       INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (game_id, id),
    FOREIGN KEY (game_id, company_id) REFERENCES companies(game_id, id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS events (
    game_id    TEXT NOT NULL,
    id         TEXT NOT NULL,
    seq        INTEGER NOT NULL,
    type       TEXT NOT NULL,
    date_year  INTEGER NOT NULL,
    date_month INTEGER NOT NULL,
    turn       INTEGER NOT NULL,
    severity   TEXT NOT NULL,
    message    TEXT NOT NULL,
    payload    TEXT NOT NULL DEFAULT '{}',
    company_id TEXT NOT NULL DEFAULT '',
    city_id    TEXT NOT NULL DEFAULT '',
    line_id    TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (game_id, seq),
    FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS actions_log (
    game_id     TEXT NOT NULL,
    seq         INTEGER NOT NULL,
    turn        INTEGER NOT NULL,
    company_id  TEXT NOT NULL,
    action_type TEXT NOT NULL,
    payload     TEXT NOT NULL,
    PRIMARY KEY (game_id, seq),
    FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_events_game_turn ON events(game_id, turn);
CREATE INDEX IF NOT EXISTS idx_lines_company ON lines(game_id, company_id);
CREATE INDEX IF NOT EXISTS idx_facilities_company ON facilities(game_id, company_id);
