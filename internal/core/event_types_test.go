package core

import (
	"sort"
	"testing"
)

func TestAllEventTypes(t *testing.T) {
	types := AllEventTypes()

	// 仕様上 60 種以上が必要。
	if len(types) < 60 {
		t.Fatalf("イベントタイプの数 = %d, want >= 60", len(types))
	}
	t.Logf("イベントタイプ総数: %d", len(types))

	// 空文字が混ざっていないこと。
	for i, typ := range types {
		if typ == "" {
			t.Errorf("index %d のイベントタイプが空です", i)
		}
	}
}

func TestAllEventTypes_Unique(t *testing.T) {
	types := AllEventTypes()
	seen := make(map[string]struct{}, len(types))
	for _, typ := range types {
		if _, ok := seen[typ]; ok {
			t.Errorf("イベントタイプが重複しています: %q", typ)
		}
		seen[typ] = struct{}{}
	}
}

func TestAllEventTypes_ReturnsCopy(t *testing.T) {
	a := AllEventTypes()
	b := AllEventTypes()
	if len(a) == 0 {
		t.Fatal("AllEventTypes が空")
	}
	// 返り値を書き換えても次の呼び出しに影響しないこと。
	a[0] = "mutated"
	if b[0] == "mutated" {
		t.Error("AllEventTypes が内部スライスを共有しています")
	}
}

func TestAllEventTypes_Sorted(t *testing.T) {
	// 一覧はカテゴリ順だが、検証用にソートしても重複がないことを確認する。
	types := AllEventTypes()
	sorted := append([]string(nil), types...)
	sort.Strings(sorted)
	for i := 1; i < len(sorted); i++ {
		if sorted[i] == sorted[i-1] {
			t.Fatalf("重複: %q", sorted[i])
		}
	}
}

func TestIsValidEventType(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"建設系", EventLineOpened, true},
		{"運用系", EventFareChanged, true},
		{"財務系", EventBondsIssued, true},
		{"M&A系", EventTOBSuccess, true},
		{"廃線系", EventLineAbandoned, true},
		{"開発系", EventCityGrown, true},
		{"政治系", EventSubsidyReceived, true},
		{"歴史系", EventDisaster, true},
		{"ターン系", EventTurnEnded, true},
		{"システム系", EventVictory, true},
		{"未知", "unknown_event", false},
		{"空", "", false},
		{"大文字", "LINE_OPENED", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidEventType(tt.s); got != tt.want {
				t.Errorf("IsValidEventType(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestIsValidEventType_AllListed(t *testing.T) {
	for _, typ := range AllEventTypes() {
		if !IsValidEventType(typ) {
			t.Errorf("AllEventTypes にあるのに IsValidEventType(%q) が false", typ)
		}
	}
}
