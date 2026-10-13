package core

import "math/rand"

// Game はゲーム全体の状態を表す。
//
// IDGen は決定論的 ID 生成器であり、Counter を保存するため JSON に含める。
// 一方 rng は再シード可能なため JSON には含めず、ロード後に EnsureRng で
// 復元する。
type Game struct {
	ID        string              `json:"id"`
	Seed      int64               `json:"seed"`
	IDGen     *IDGenerator        `json:"id_gen"`
	Date      Date                `json:"date"`
	Turn      int                 `json:"turn"`
	PlayerID  string              `json:"player_id"`
	Companies map[string]*Company `json:"companies"`
	Cities    map[string]*City    `json:"cities"`
	Events    []Event             `json:"events"`
	rng       *rand.Rand          `json:"-"`
}

// NewGame は新しいゲームを生成する。
//
// Date は開始年を引数に取らないためゼロ値のままとする（呼び出し側で設定）。
// Companies/Cities は空 map、Events は空スライスで初期化する。
func NewGame(seed int64, playerID string) *Game {
	idGen := NewIDGenerator(seed)
	return &Game{
		ID:        idGen.Next(),
		Seed:      seed,
		IDGen:     idGen,
		Turn:      0,
		PlayerID:  playerID,
		Companies: map[string]*Company{},
		Cities:    map[string]*City{},
		Events:    []Event{},
		rng:       rand.New(rand.NewSource(seed)),
	}
}

// PlayerCompany はプレイヤー会社を返す。存在しなければ nil を返す。
func (g *Game) PlayerCompany() *Company {
	return g.Companies[g.PlayerID]
}

// AddCompany は会社を追加する。
func (g *Game) AddCompany(c *Company) {
	g.Companies[c.ID] = c
}

// AddCity は都市を追加する。
func (g *Game) AddCity(c *City) {
	g.Cities[c.ID] = c
}

// Rng はゲームの乱数源を返す。
func (g *Game) Rng() *rand.Rand {
	return g.rng
}

// EnsureRng は rng が nil の場合に Seed から再初期化する。
//
// JSON ロード後は rng が nil になるため、その後に呼ぶ想定。
func (g *Game) EnsureRng() {
	if g.rng == nil {
		g.rng = rand.New(rand.NewSource(g.Seed))
	}
}

// AppendEvents はイベントを追記する。
func (g *Game) AppendEvents(events ...Event) {
	g.Events = append(g.Events, events...)
}
