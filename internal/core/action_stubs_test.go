package core

import "testing"

// TestActionStubs_ValidateApply はスタブとして残っているアクションの
// Validate/Apply が nil を返すことを検証する。
//
// Phase 1-8-2 時点では EndTurnAction・BuildLineAction・SetFareAction・
// BuildFacilityAction が本実装済みのため対象外とし、残り 21 種を検証する。
// Phase 1-8-3 以降で本実装に置き換わるたびに、このテストも更新する。
func TestActionStubs_ValidateApply(t *testing.T) {
	g := NewGame(1907, "hankyu")
	// スタブとして残っている21種のみを対象
	for _, a := range allActions() {
		switch a.(type) {
		case *EndTurnAction, *BuildLineAction, *SetFareAction, *BuildFacilityAction:
			continue // 本実装済み
		}
		if err := a.Validate(g); err != nil {
			t.Errorf("%s.Validate() = %v, want nil", a.Type(), err)
		}
		events, err := a.Apply(g)
		if err != nil {
			t.Errorf("%s.Apply() = %v, want nil", a.Type(), err)
		}
		if events != nil {
			t.Errorf("%s.Apply() events = %v, want nil", a.Type(), events)
		}
	}
}
