package core

import "testing"

// newDevelopGame は開発 AI テスト用のゲームを返す。
//
// 2 都市（a, b）と路線 1 本を持つ会社を作る。
func newDevelopGame(t *testing.T, cash int64, realEstate float64) (*Game, *Company, *City) {
	t.Helper()
	g := NewGame(1907, "player")
	g.Date = Date{Year: 1910, Month: 6}
	cityA := &City{ID: "a", Name: "A", Population: 1_000_000, LandValue: 0}
	cityB := &City{ID: "b", Name: "B", Population: 1_000_000, LandValue: 0}
	g.AddCity(cityA)
	g.AddCity(cityB)

	c := NewCompany("hankyu", "hankyu", "阪急")
	c.Cash = cash
	c.AddLine(NewLine("l1", c.ID, "a", "b"))
	g.AddCompany(c)
	return g, c, cityA
}

func TestAIDevelop_SkipsLowRealEstate(t *testing.T) {
	g, c, _ := newDevelopGame(t, 100_000_000, 0.1)
	ai := &AI{Personality: Personality{RealEstate: 0.1}}

	if evs := g.aiDevelop(c, ai); evs != nil {
		t.Errorf("不動産志向が低いのに建設された: %d", len(evs))
	}
}

func TestAIDevelop_SkipsLowCash(t *testing.T) {
	g, c, _ := newDevelopGame(t, 1_000_000, 0.9)
	ai := &AI{Personality: Personality{RealEstate: 0.9}}

	if evs := g.aiDevelop(c, ai); evs != nil {
		t.Errorf("資金不足なのに建設された: %d", len(evs))
	}
}

func TestAIDevelop_SkipsOtherOwner(t *testing.T) {
	g, c, cityA := newDevelopGame(t, 100_000_000, 0.9)
	cityA.OwnerID = "other"
	cityA.Developed = 0.0
	// city b は自社所有可能だが heuristics 上 a が先に評価される。
	// a が他社所有なら b を開発するため、b も他社所有にする。
	g.Cities["b"].OwnerID = "other"

	ai := &AI{Personality: Personality{RealEstate: 0.9}}
	if evs := g.aiDevelop(c, ai); evs != nil {
		t.Errorf("他社所有都市で建設された: %d", len(evs))
	}
}

func TestAIDevelop_SkipsDevelopedCity(t *testing.T) {
	g, c, cityA := newDevelopGame(t, 100_000_000, 0.9)
	cityA.Developed = 0.8
	g.Cities["b"].Developed = 0.8
	ai := &AI{Personality: Personality{RealEstate: 0.9}}

	if evs := g.aiDevelop(c, ai); evs != nil {
		t.Errorf("開発済み都市で建設された: %d", len(evs))
	}
}

func TestAIDevelop_SkipsInsufficientTriple(t *testing.T) {
	g, c, cityA := newDevelopGame(t, 15_000_000, 0.9)
	cityA.LandValue = 100 // cost = 3M * 2 * 1.5 = 9M、*3 = 27M > 15M
	g.Cities["b"].LandValue = 100
	ai := &AI{Personality: Personality{RealEstate: 0.9}}

	if evs := g.aiDevelop(c, ai); evs != nil {
		t.Errorf("資金が建設費の3倍未満なのに建設された: %d", len(evs))
	}
}

func TestAIDevelop_BuildsFacility(t *testing.T) {
	g, c, cityA := newDevelopGame(t, 100_000_000, 0.9)
	ai := &AI{Personality: Personality{RealEstate: 0.9}}

	evs := g.aiDevelop(c, ai)

	if len(evs) != 1 {
		t.Fatalf("events = %d, want 1", len(evs))
	}
	if evs[0].Type != EventFacilityBuilt {
		t.Errorf("type = %q, want %q", evs[0].Type, EventFacilityBuilt)
	}
	if len(c.Facilities) != 1 {
		t.Fatalf("Facilities = %d, want 1", len(c.Facilities))
	}
	if cityA.OwnerID != c.ID {
		t.Errorf("OwnerID = %q, want %q", cityA.OwnerID, c.ID)
	}
	if cityA.Developed <= 0 {
		t.Errorf("Developed = %v, want > 0", cityA.Developed)
	}
}

func TestAIDevelop_NilCity(t *testing.T) {
	g := NewGame(1, "p")
	c := NewCompany("hankyu", "hankyu", "阪急")
	c.Cash = 100_000_000
	c.AddLine(NewLine("l1", c.ID, "missing1", "missing2"))
	g.AddCompany(c)

	ai := &AI{Personality: Personality{RealEstate: 0.9}}
	if evs := g.aiDevelop(c, ai); evs != nil {
		t.Errorf("存在しない都市で建設された: %d", len(evs))
	}
}

func TestExecuteBuildFacility_UnknownKind(t *testing.T) {
	g, c, cityA := newDevelopGame(t, 100_000_000, 0.9)
	if evs := g.executeBuildFacility(c, cityA, "unknown", 2); evs != nil {
		t.Errorf("未知の種類で建設された: %d", len(evs))
	}
}

func TestExecuteBuildFacility_InsufficientCash(t *testing.T) {
	g, c, cityA := newDevelopGame(t, 0, 0.9)
	if evs := g.executeBuildFacility(c, cityA, FacilityHousing, 2); evs != nil {
		t.Errorf("資金不足で建設された: %d", len(evs))
	}
}
