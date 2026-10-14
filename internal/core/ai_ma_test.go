package core

import (
	"math/rand"
	"testing"
)

// zeroSource は常に 0 を返す決定論的な乱数源（テスト用）。
//
// M&A の確率判定（g.rng.Float64()）を必ず通過させるために使う。
type zeroSource struct{}

func (zeroSource) Int63() int64 { return 0 }
func (zeroSource) Seed(int64)   {}

// forceMAProbability は M&A の確率判定を確実に通過させる乱数源に差し替える。
func forceMAProbability(g *Game) {
	g.rng = rand.New(zeroSource{})
}

// aggressiveAI は M&A を実行できる高攻撃性の AI を返す。
func aggressiveAI(companyID string) *AI {
	p := DefaultPersonality(companyID)
	p.Aggression = 0.9
	return &AI{CompanyID: companyID, Personality: p}
}

// setupAIMAGame は M&A AI テスト用のゲームを返す。
//
// プレイヤー（player）、買収側（hankyu）、対象（hanshin）を用意する。
// 乱数源は決定論的な zeroSource に固定する。
func setupAIMAGame(t *testing.T) (*Game, *Company, *Company) {
	t.Helper()
	g := NewGame(1907, "player")
	g.Date = Date{Year: 1910, Month: 1}
	forceMAProbability(g)

	player := NewCompany("player", "player", "プレイヤー")
	player.IsPlayer = true
	g.AddCompany(player)

	hankyu := NewCompany("hankyu", "hankyu", "阪急")
	hankyu.Cash = 200_000_000
	hankyu.TotalShares = 10000
	hankyu.StockPrice = 1000
	hankyu.Region = "kansai"
	g.AddCompany(hankyu)

	hanshin := NewCompany("hanshin", "hanshin", "阪神")
	hanshin.Cash = 50_000_000
	hanshin.TotalShares = 10000
	hanshin.StockPrice = 1000
	hanshin.Region = "kansai"
	g.AddCompany(hanshin)

	return g, hankyu, hanshin
}

func TestAIMA_LowAggression(t *testing.T) {
	g, hankyu, _ := setupAIMAGame(t)
	ai := &AI{CompanyID: "hankyu", Personality: Personality{Aggression: MAMinAggression - 0.1}}

	if evs := g.aiMA(hankyu, ai); evs != nil {
		t.Errorf("低攻撃性でイベントが生成された: %d", len(evs))
	}
	if hankyu.Holding["hanshin"] != 0 {
		t.Error("低攻撃性で株を取得した")
	}
}

func TestAIMA_InsufficientCash(t *testing.T) {
	g, hankyu, _ := setupAIMAGame(t)
	hankyu.Cash = MAMinCash - 1
	ai := &AI{CompanyID: "hankyu", Personality: DefaultPersonality("hankyu")}

	if evs := g.aiMA(hankyu, ai); evs != nil {
		t.Errorf("資金不足でイベントが生成された: %d", len(evs))
	}
}

func TestAIMA_BuysShares(t *testing.T) {
	g, hankyu, hanshin := setupAIMAGame(t)
	ai := aggressiveAI("hankyu")

	evs := g.aiMA(hankyu, ai)

	if !hasEventType(evs, EventSharesAcquired) {
		t.Error("shares_acquired が生成されなかった")
	}
	if hankyu.Holding["hanshin"] != 2000 {
		t.Errorf("Holding = %d, want 2000", hankyu.Holding["hanshin"])
	}
	if hanshin.ParentID == hankyu.ID {
		t.Error("20% 取得で子会社化された")
	}
}

func TestAIMA_TenderOffer(t *testing.T) {
	g, hankyu, hanshin := setupAIMAGame(t)
	// 35% を保有済み（TOB 閾値超）
	hankyu.Holding["hanshin"] = 3500
	ai := aggressiveAI("hankyu")

	evs := g.aiMA(hankyu, ai)

	if !hasEventType(evs, EventTOBLaunched) {
		t.Error("tob_launched が生成されなかった")
	}
	if !hasEventType(evs, EventTOBSuccess) {
		t.Error("tob_success が生成されなかった")
	}
	if !hasEventType(evs, EventSubsidiaryEstablished) {
		t.Error("subsidiary_established が生成されなかった")
	}
	if hanshin.ParentID != hankyu.ID {
		t.Errorf("ParentID = %q, want %q", hanshin.ParentID, hankyu.ID)
	}
}

func TestAIMA_Merge(t *testing.T) {
	g, hankyu, hanshin := setupAIMAGame(t)
	// 60% を保有済み（子会社）
	hankyu.Holding["hanshin"] = 6000
	hanshin.ParentID = hankyu.ID
	ai := aggressiveAI("hankyu")

	evs := g.aiMA(hankyu, ai)

	if !hasEventType(evs, EventMergeCompleted) {
		t.Error("merge_completed が生成されなかった")
	}
	if hanshin.Dissolved == nil {
		t.Error("対象会社が解散していない")
	}
}

func TestAIMA_NoTarget(t *testing.T) {
	g := NewGame(1907, "player")
	g.Date = Date{Year: 1910, Month: 1}
	c := NewCompany("hankyu", "hankyu", "阪急")
	c.Cash = 200_000_000
	c.TotalShares = 10000
	c.StockPrice = 1000
	g.AddCompany(c)
	ai := &AI{CompanyID: "hankyu", Personality: DefaultPersonality("hankyu")}

	if evs := g.aiMA(c, ai); evs != nil {
		t.Errorf("対象なしでイベントが生成された: %d", len(evs))
	}
}

func TestFindTakeoverTarget(t *testing.T) {
	g, hankyu, _ := setupAIMAGame(t)
	// 国鉄・解散済み会社は対象外
	jnr := NewCompany("jnr", "jnr", "国鉄")
	jnr.IsNational = true
	jnr.TotalShares = 10000
	jnr.StockPrice = 1000
	g.AddCompany(jnr)

	dead := NewCompany("dead", "dead", "旧社")
	dead.TotalShares = 10000
	dead.StockPrice = 1000
	d := Date{Year: 1900, Month: 1}
	dead.Dissolved = &d
	g.AddCompany(dead)

	ai := &AI{CompanyID: "hankyu", Personality: DefaultPersonality("hankyu")}
	target := g.findTakeoverTarget(hankyu, ai)
	if target == nil {
		t.Fatal("買収対象が見つからなかった")
	}
	if target.ID != "hanshin" {
		t.Errorf("target = %q, want hanshin", target.ID)
	}
}

func TestFindTakeoverTarget_SkipsPlayerForLowAggression(t *testing.T) {
	g := NewGame(1907, "player")
	player := NewCompany("player", "player", "プレイヤー")
	player.IsPlayer = true
	player.TotalShares = 10000
	player.StockPrice = 1000
	player.Region = "kansai"
	g.AddCompany(player)

	hankyu := NewCompany("hankyu", "hankyu", "阪急")
	hankyu.Cash = 200_000_000
	hankyu.Region = "kansai"
	g.AddCompany(hankyu)

	low := &AI{CompanyID: "hankyu", Personality: Personality{Aggression: 0.6}}
	if got := g.findTakeoverTarget(hankyu, low); got != nil {
		t.Errorf("低攻撃性でプレイヤーが対象になった: %q", got.ID)
	}

	high := &AI{CompanyID: "hankyu", Personality: Personality{Aggression: 0.95}}
	got := g.findTakeoverTarget(hankyu, high)
	if got == nil || got.ID != "player" {
		t.Errorf("高攻撃性でプレイヤーが対象にならない: %v", got)
	}
}

func TestFindTakeoverTarget_Deterministic(t *testing.T) {
	g1, h1, _ := setupAIMAGame(t)
	g2, h2, _ := setupAIMAGame(t)
	ai := &AI{CompanyID: "hankyu", Personality: DefaultPersonality("hankyu")}
	t1 := g1.findTakeoverTarget(h1, ai)
	t2 := g2.findTakeoverTarget(h2, ai)
	if t1 == nil || t2 == nil || t1.ID != t2.ID {
		t.Errorf("対象選定が非決定: %v vs %v", t1, t2)
	}
}

func TestTakeoverScore(t *testing.T) {
	g, hankyu, hanshin := setupAIMAGame(t)
	ai := &AI{CompanyID: "hankyu", Personality: Personality{Aggression: 0.5}}

	base := g.takeoverScore(hankyu, hanshin, ai)

	// 時価総額が小さいほど高スコア
	small := NewCompany("small", "small", "小社")
	small.TotalShares = 1000
	small.StockPrice = 1000
	small.Region = "kansai"
	g.AddCompany(small)
	if got := g.takeoverScore(hankyu, small, ai); got <= base {
		t.Errorf("小規模会社のスコアが高くない: small=%v base=%v", got, base)
	}

	// 資金力補正: 時価総額より資金が少ないと 0.3 倍
	hankyu.Cash = 1000
	penalized := g.takeoverScore(hankyu, hanshin, ai)
	if penalized >= base {
		t.Errorf("資金不足でスコアが下がらない: penalized=%v base=%v", penalized, base)
	}
}

func TestAdjacentToCompany(t *testing.T) {
	g, hankyu, hanshin := setupAIMAGame(t)

	hankyu.AddLine(NewLine("l1", "hankyu", "a", "b"))
	hanshin.AddLine(NewLine("l2", "hanshin", "b", "c"))

	if !g.adjacentToCompany(hankyu, hanshin) {
		t.Error("接続都市を共有しているのに隣接と判定されない")
	}

	other := NewCompany("other", "other", "他社")
	other.AddLine(NewLine("l3", "other", "x", "y"))
	g.AddCompany(other)
	if g.adjacentToCompany(hankyu, other) {
		t.Error("接続していないのに隣接と判定された")
	}
}

func TestTakeoverShareCount(t *testing.T) {
	g, hankyu, hanshin := setupAIMAGame(t)
	ai := &AI{CompanyID: "hankyu", Personality: DefaultPersonality("hankyu")}

	// 発行済株の 20% が上限
	if got := g.takeoverShareCount(hankyu, hanshin, ai); got != 2000 {
		t.Errorf("takeoverShareCount = %d, want 2000", got)
	}

	// 資金の半分で買える株数が上限を下回る
	hankyu.Cash = 200_000 // affordable = 100
	if got := g.takeoverShareCount(hankyu, hanshin, ai); got != 100 {
		t.Errorf("takeoverShareCount = %d, want 100", got)
	}

	// 100 株未満は見送り（0）
	hankyu.Cash = 100_000 // affordable = 50
	if got := g.takeoverShareCount(hankyu, hanshin, ai); got != 0 {
		t.Errorf("takeoverShareCount = %d, want 0", got)
	}

	// 株価 0 は 0
	hanshin.StockPrice = 0
	hankyu.Cash = 200_000_000
	if got := g.takeoverShareCount(hankyu, hanshin, ai); got != 0 {
		t.Errorf("株価 0 の takeoverShareCount = %d, want 0", got)
	}
}

func TestTOBCost(t *testing.T) {
	g, hankyu, hanshin := setupAIMAGame(t)
	hankyu.Holding["hanshin"] = 3500
	ai := &AI{CompanyID: "hankyu", Personality: DefaultPersonality("hankyu")}

	want := int64(float64((int(float64(hanshin.TotalShares)*MATOBTargetRatio)-3500)*hanshin.StockPrice) * MATOBPremium)
	if got := g.tobCost(hankyu, hanshin, ai); got != want {
		t.Errorf("tobCost = %d, want %d", got, want)
	}
}

func TestHoldingPct(t *testing.T) {
	g, hankyu, hanshin := setupAIMAGame(t)
	hankyu.Holding["hanshin"] = 2500
	if got := g.holdingPct(hankyu, hanshin); got != 0.25 {
		t.Errorf("holdingPct = %v, want 0.25", got)
	}
	hanshin.TotalShares = 0
	if got := g.holdingPct(hankyu, hanshin); got != 0 {
		t.Errorf("発行済株 0 の holdingPct = %v, want 0", got)
	}
}

func TestApplyAIAction_Invalid(t *testing.T) {
	g, hankyu, _ := setupAIMAGame(t)
	// 資金不足の株取得は Validate で失敗し nil を返す
	hankyu.Cash = 0
	action := &BuySharesAction{ActorCompanyID: "hankyu", TargetCompanyID: "hanshin", Count: 1000}
	if evs, err := g.applyAIAction(action); evs != nil || err == nil {
		t.Errorf("無効アクションでエラーにならない: evs=%d err=%v", len(evs), err)
	}
}
