package core

// Company は鉄道会社を表す。
//
// Holding は他社株の保有数（会社ID → 株数）、Lines/Facilities は保有する
// 路線と施設を表す。GroupID は系列グループ、ParentID は統廃合グラフの
// 親会社を指す。
type Company struct {
	ID          string         `json:"id"`
	GroupID     string         `json:"group_id"`
	Name        string         `json:"name"`
	LegalName   string         `json:"legal_name"`
	FoundedYear int            `json:"founded_year"`
	Cash        int64          `json:"cash"`
	Debt        int64          `json:"debt"`
	StockPrice  int            `json:"stock_price"`
	TotalShares int            `json:"total_shares"`
	Holding     map[string]int `json:"holding"` // 他社株保有
	Lines       []*Line        `json:"lines"`
	Facilities  []*Facility    `json:"facilities"`
	IsPlayer    bool           `json:"is_player"`
	IsNational  bool           `json:"is_national"`
	Region      string         `json:"region"` // "kanto"|"kansai"|"chubu"|"national"|...
	Tags        []string       `json:"tags"`
	ParentID    string         `json:"parent_id,omitempty"`
	Reputation  int            `json:"reputation"` // 0-100
	CreditLine  int64          `json:"credit_line"`
}

// NewCompany は会社を生成する。
//
// Holding は空 map、Lines/Facilities/Tags は空スライスで初期化する。
// 評判（Reputation）の初期値は 50。
func NewCompany(id, groupID, name string) *Company {
	return &Company{
		ID:         id,
		GroupID:    groupID,
		Name:       name,
		Holding:    map[string]int{},
		Lines:      []*Line{},
		Facilities: []*Facility{},
		Tags:       []string{},
		Reputation: 50,
	}
}

// NetAssets は純資産（Cash - Debt）の簡易値を返す。
//
// Phase 1-7 で資産・減価償却を加味した精緻な計算に置き換える予定。
func (c *Company) NetAssets() int64 {
	return c.Cash - c.Debt
}

// MonthlyProfit は直近月次の利益を返す。
//
// Phase 1-7 で実装する。現時点では 0 を返すスタブ。
func (c *Company) MonthlyProfit() int64 {
	return 0
}

// AddLine は路線を追加する。
func (c *Company) AddLine(l *Line) {
	c.Lines = append(c.Lines, l)
}

// RemoveLine は指定した ID の路線を削除する。成功したら true を返す。
func (c *Company) RemoveLine(lineID string) bool {
	for i, l := range c.Lines {
		if l.ID == lineID {
			c.Lines = append(c.Lines[:i], c.Lines[i+1:]...)
			return true
		}
	}
	return false
}

// FindLine は指定した ID の路線を返す。存在しなければ nil を返す。
func (c *Company) FindLine(lineID string) *Line {
	for _, l := range c.Lines {
		if l.ID == lineID {
			return l
		}
	}
	return nil
}
