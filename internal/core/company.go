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
// 路線など固定資産を含めない簡易値。固定資産を含める場合は
// NetAssetsWithLines を使う。
func (c *Company) NetAssets() int64 {
	return c.Cash - c.Debt
}

// NetAssetsWithLines は純資産に路線価値を加えたものを返す。
//
// 路線価値は敷設コスト相当（距離 × BaseCostPerKm）とみなす簡易評価。
// 車両・施設の評価は Phase 1-8 以降で追加する。
func (c *Company) NetAssetsWithLines(g *Game) int64 {
	total := c.NetAssets()
	for _, l := range c.Lines {
		total += int64(l.LengthKm * float64(BaseCostPerKm))
	}
	return total
}

// MonthlyProfit は直近月の営業利益を返す。
//
// 路線の Profit 合計から月次減価償却を差し引いた値。
// 路線がない場合は減価償却 0 のため 0 を返す。
func (c *Company) MonthlyProfit(g *Game) int64 {
	var total int64
	for _, l := range c.Lines {
		total += l.Profit
	}
	total -= g.calculateDepreciation(c)
	return total
}

// AnnualProfit は直近 12 ヶ月の営業利益を推定する。
//
// Phase 1-7 では MonthlyProfit × 12 の簡易版。
// Phase 1-8 以降で履歴ベースに精緻化する。
func (c *Company) AnnualProfit(g *Game) int64 {
	return c.MonthlyProfit(g) * 12
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
