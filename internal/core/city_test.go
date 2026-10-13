package core

import (
	"encoding/json"
	"testing"
)

func TestNewCity(t *testing.T) {
	c := NewCity("osaka", "大阪", "大阪府")

	if c.ID != "osaka" || c.Name != "大阪" || c.Prefecture != "大阪府" {
		t.Errorf("基本情報が不正: %+v", c)
	}
	if c.Developed != 0.0 {
		t.Errorf("Developed = %v, want 0.0", c.Developed)
	}
	if c.PoliticalPressure != 0.0 {
		t.Errorf("PoliticalPressure = %v, want 0.0", c.PoliticalPressure)
	}
	if c.Tags == nil {
		t.Error("Tags が nil")
	}
	if len(c.Tags) != 0 {
		t.Errorf("Tags の数 = %d, want 0", len(c.Tags))
	}
}

func TestCity_JSONRoundTrip(t *testing.T) {
	c := NewCity("osaka", "大阪", "大阪府")
	c.RepresentativeNode = "osaka-station"
	c.Population = 2_000_000
	c.LandValue = 90
	c.Demand = 5000
	c.Developed = 0.75
	c.OwnerID = "hankyu"
	c.PoliticalPressure = 0.3
	c.ResidentsAffected = 1234
	c.Tags = append(c.Tags, "urban", "port")

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var restored City
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if restored.ID != c.ID || restored.Name != c.Name || restored.Prefecture != c.Prefecture {
		t.Errorf("基本情報が異なる: %+v", restored)
	}
	if restored.RepresentativeNode != c.RepresentativeNode {
		t.Errorf("RepresentativeNode = %q", restored.RepresentativeNode)
	}
	if restored.Population != c.Population || restored.LandValue != c.LandValue || restored.Demand != c.Demand {
		t.Errorf("数値が異なる: %+v", restored)
	}
	if restored.Developed != c.Developed || restored.PoliticalPressure != c.PoliticalPressure {
		t.Errorf("割合が異なる: %+v", restored)
	}
	if restored.OwnerID != c.OwnerID || restored.ResidentsAffected != c.ResidentsAffected {
		t.Errorf("所有者/住民が異なる: %+v", restored)
	}
	if len(restored.Tags) != 2 {
		t.Errorf("Tags = %+v", restored.Tags)
	}
}
