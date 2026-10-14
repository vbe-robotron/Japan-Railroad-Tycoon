package core

import "testing"

// newBuildGame は建設 AI テスト用のゲームを返す。
//
// 6 月（季節係数 1.0）で、指定の都市群と会社を追加する。
func newBuildGame(t *testing.T, companyID, region string, cash int64, cities ...*City) (*Game, *Company) {
	t.Helper()
	g := NewGame(1907, "player")
	g.Date = Date{Year: 1910, Month: 6}
	for _, c := range cities {
		g.AddCity(c)
	}
	c := NewCompany(companyID, companyID, companyID)
	c.Cash = cash
	c.Region = region
	g.AddCompany(c)
	return g, c
}

// bigCity は需要が十分大きい都市を返す。
func bigCity(id string, lat, lon float64) *City {
	return &City{ID: id, Name: id, Population: 1_000_000, Lat: lat, Lon: lon}
}

func TestAIBuild_NoCandidates(t *testing.T) {
	// 拠点都市が推定できない（Region 不明）ので候補なし
	g, c := newBuildGame(t, "hankyu", "", 100_000_000,
		bigCity("a", 35.0, 135.0))

	ai := &AI{CompanyID: "hankyu", Personality: DefaultPersonality("hankyu")}
	if evs := g.aiBuild(c, ai); evs != nil {
		t.Errorf("候補なしでイベントが生成された: %d", len(evs))
	}
	if len(c.Lines) != 0 {
		t.Errorf("候補なしで路線が敷設された: %d", len(c.Lines))
	}
}

func TestAIBuild_BuildsLine(t *testing.T) {
	g, c := newBuildGame(t, "hankyu", "kansai", 100_000_000,
		bigCity("umeda", 34.7, 135.5),
		bigCity("kobe", 34.7, 135.2))

	ai := &AI{CompanyID: "hankyu", Personality: DefaultPersonality("hankyu")}
	evs := g.aiBuild(c, ai)

	if len(evs) != 1 {
		t.Fatalf("events = %d, want 1", len(evs))
	}
	if evs[0].Type != EventLineOpened {
		t.Errorf("event type = %q, want %q", evs[0].Type, EventLineOpened)
	}
	if len(c.Lines) != 1 {
		t.Fatalf("Lines = %d, want 1", len(c.Lines))
	}
	if c.Lines[0].Gauge != 1067 || c.Lines[0].Double {
		t.Errorf("敷設軌間/複線 = %d/%v, want 1067/false", c.Lines[0].Gauge, c.Lines[0].Double)
	}
}

func TestAIBuild_InsufficientCash(t *testing.T) {
	g, c := newBuildGame(t, "hankyu", "kansai", MinAIBuildCash-1,
		bigCity("umeda", 34.7, 135.5),
		bigCity("kobe", 34.7, 135.2))

	ai := &AI{CompanyID: "hankyu", Personality: DefaultPersonality("hankyu")}
	if evs := g.aiBuild(c, ai); evs != nil {
		t.Errorf("資金不足でイベントが生成された: %d", len(evs))
	}
	if len(c.Lines) != 0 {
		t.Errorf("資金不足で路線が敷設された")
	}
}

func TestAIBuild_LowScore(t *testing.T) {
	// 小都市間はスコア閾値未満で敷設しない
	g, c := newBuildGame(t, "hankyu", "kansai", 100_000_000,
		&City{ID: "umeda", Name: "umeda", Population: 100, Lat: 34.7, Lon: 135.5},
		&City{ID: "kobe", Name: "kobe", Population: 100, Lat: 34.7, Lon: 135.2})

	ai := &AI{CompanyID: "hankyu", Personality: DefaultPersonality("hankyu")}
	if evs := g.aiBuild(c, ai); evs != nil {
		t.Errorf("低スコアで路線が敷設された: %d events", len(evs))
	}
}

func TestBuildCandidates_FirstLine(t *testing.T) {
	g, c := newBuildGame(t, "hankyu", "kansai", 100_000_000,
		bigCity("umeda", 34.7, 135.5),
		bigCity("kobe", 34.7, 135.2),
		bigCity("kyoto", 35.0, 135.7))

	ai := &AI{CompanyID: "hankyu", Personality: DefaultPersonality("hankyu")}
	cands := g.buildCandidates(c, ai)

	if len(cands) == 0 {
		t.Fatal("初回敷設の候補が生成されなかった")
	}
	for _, cand := range cands {
		if cand.From != "umeda" {
			t.Errorf("候補 From = %q, want umeda", cand.From)
		}
	}
	// From→To 昇順ソートの確認
	for i := 1; i < len(cands); i++ {
		if cands[i-1].From > cands[i].From {
			t.Errorf("候補がソートされていない")
		}
	}
}

func TestBuildCandidates_FromLines(t *testing.T) {
	g, c := newBuildGame(t, "hankyu", "kansai", 100_000_000,
		bigCity("a", 35.0, 135.0),
		bigCity("b", 35.0, 135.2),
		bigCity("c", 35.0, 135.4))

	line := NewLine("l1", c.ID, "a", "b")
	line.LengthKm = g.DistanceKm("a", "b")
	c.AddLine(line)

	ai := &AI{CompanyID: "hankyu", Personality: DefaultPersonality("hankyu")}
	cands := g.buildCandidates(c, ai)

	if len(cands) == 0 {
		t.Fatal("延伸候補が生成されなかった")
	}
	found := false
	for _, cand := range cands {
		if cand.To == "c" {
			found = true
		}
		if cand.From == cand.To {
			t.Errorf("同一都市への候補: %q", cand.From)
		}
	}
	if !found {
		t.Error("c への延伸候補が含まれない")
	}
}

func TestScoreBuildCandidate(t *testing.T) {
	g, c := newBuildGame(t, "hankyu", "kansai", 100_000_000,
		bigCity("umeda", 34.7, 135.5),
		bigCity("kobe", 34.7, 135.2))
	ai := &AI{CompanyID: "hankyu", Personality: Personality{Aggression: 0.5}}

	// 通常ケース: 正のスコア
	sc := g.scoreBuildCandidate(c, "umeda", g.Cities["kobe"], ai)
	if sc <= 0 {
		t.Errorf("score = %v, want > 0", sc)
	}

	// from が存在しない → 0
	if got := g.scoreBuildCandidate(c, "nowhere", g.Cities["kobe"], ai); got != 0 {
		t.Errorf("存在しない from の score = %v, want 0", got)
	}
	// to が nil → 0
	if got := g.scoreBuildCandidate(c, "umeda", nil, ai); got != 0 {
		t.Errorf("to=nil の score = %v, want 0", got)
	}
}

func TestScoreBuildCandidate_CompetitorsPenalty(t *testing.T) {
	g, c := newBuildGame(t, "hankyu", "kansai", 100_000_000,
		bigCity("umeda", 34.7, 135.5),
		bigCity("kobe", 34.7, 135.2))
	ai := &AI{CompanyID: "hankyu", Personality: Personality{Aggression: 0.5}}

	base := g.scoreBuildCandidate(c, "umeda", g.Cities["kobe"], ai)

	// 競合会社が kobe に接続
	other := NewCompany("other", "other", "他社")
	other.AddLine(NewLine("x", "other", "kobe", "osaka"))
	g.AddCompany(other)

	after := g.scoreBuildCandidate(c, "umeda", g.Cities["kobe"], ai)
	if after >= base {
		t.Errorf("競合ありでスコアが下がらない: base=%v after=%v", base, after)
	}
}

func TestScoreBuildCandidate_LongDistancePenalty(t *testing.T) {
	g, c := newBuildGame(t, "hankyu", "kansai", 100_000_000,
		bigCity("a", 31.0, 130.0),
		bigCity("b", 45.0, 145.0),
		bigCity("near", 31.0, 130.1))
	ai := &AI{CompanyID: "hankyu", Personality: Personality{Aggression: 0.5}}

	// 遠距離のスコアは 0.7 補正で小さくなる
	far := g.scoreBuildCandidate(c, "a", g.Cities["b"], ai)
	near := g.scoreBuildCandidate(c, "a", g.Cities["near"], ai)
	if g.DistanceKm("a", "b") <= 50 {
		t.Skip("距離が 50km 以下なのでこのテストは対象外")
	}
	if far >= near {
		t.Errorf("遠距離のスコアが補正されていない: far=%v near=%v", far, near)
	}
}

func TestCompanyHomeCity(t *testing.T) {
	g := NewGame(1, "p")
	cases := []struct {
		region string
		want   string
	}{
		{"kansai", "umeda"},
		{"kanto", "tokyo"},
		{"chubu", "nagoya"},
		{"shikoku", "matsuyama"},
		{"unknown", ""},
		{"", ""},
	}
	for _, tc := range cases {
		c := NewCompany("x", "x", "X")
		c.Region = tc.region
		if got := g.companyHomeCity(c); got != tc.want {
			t.Errorf("region %q: home = %q, want %q", tc.region, got, tc.want)
		}
	}
}

func TestCountCompetitorsAt(t *testing.T) {
	g := NewGame(1, "p")
	me := NewCompany("me", "me", "自社")
	other := NewCompany("other", "other", "他社")
	other.AddLine(NewLine("x", "other", "a", "b"))
	other.AddLine(NewLine("y", "other", "b", "c"))
	g.AddCompany(me)
	g.AddCompany(other)

	// 自社を除くと他社の 2 本が数えられる
	if n := g.countCompetitorsAt("b", "me"); n != 2 {
		t.Errorf("countCompetitorsAt(b) = %d, want 2", n)
	}
	// 他社を除外すると 0
	if n := g.countCompetitorsAt("b", "other"); n != 0 {
		t.Errorf("countCompetitorsAt(b, other) = %d, want 0", n)
	}
	// 接続なし
	if n := g.countCompetitorsAt("z", "me"); n != 0 {
		t.Errorf("countCompetitorsAt(z) = %d, want 0", n)
	}
}
