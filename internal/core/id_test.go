package core

import (
	"encoding/json"
	"testing"
)

func TestIDGenerator_Deterministic(t *testing.T) {
	gen1 := NewIDGenerator(1907)
	gen2 := NewIDGenerator(1907)

	for i := 0; i < 100; i++ {
		id1 := gen1.Next()
		id2 := gen2.Next()
		if id1 != id2 {
			t.Fatalf("iteration %d: %s != %s", i, id1, id2)
		}
	}
}

func TestIDGenerator_Unique(t *testing.T) {
	gen := NewIDGenerator(1907)
	seen := make(map[string]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		id := gen.Next()
		if _, ok := seen[id]; ok {
			t.Fatalf("重複した ID: %s (iteration %d)", id, i)
		}
		seen[id] = struct{}{}
	}
}

func TestIDGenerator_Length(t *testing.T) {
	gen := NewIDGenerator(42)
	for i := 0; i < 100; i++ {
		id := gen.Next()
		if len(id) != nanoIDLength {
			t.Fatalf("iteration %d: len(%q) = %d, want %d", i, id, len(id), nanoIDLength)
		}
	}
}

func TestIDGenerator_Counter(t *testing.T) {
	gen := NewIDGenerator(1907)
	if gen.Counter != 0 {
		t.Fatalf("初期 Counter = %d, want 0", gen.Counter)
	}
	for i := 1; i <= 10; i++ {
		gen.Next()
		if gen.Counter != int64(i) {
			t.Fatalf("Next 後 Counter = %d, want %d", gen.Counter, i)
		}
	}
}

func TestIDGenerator_JSON(t *testing.T) {
	gen := NewIDGenerator(1907)
	gen.Next()
	gen.Next()

	data, err := json.Marshal(gen)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var restored IDGenerator
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if restored.Seed != gen.Seed || restored.Counter != gen.Counter {
		t.Fatalf("復元後の状態が異なる: %+v vs %+v", restored, *gen)
	}

	// 復元後も同じ ID 列が続く
	if gen.Next() != restored.Next() {
		t.Fatal("復元後に同じ ID が続かない")
	}
}

func TestIDGenerator_SeedDifference(t *testing.T) {
	a := NewIDGenerator(1)
	b := NewIDGenerator(2)
	if a.Next() == b.Next() {
		t.Fatal("異なる seed から同じ ID が生成された")
	}
}
