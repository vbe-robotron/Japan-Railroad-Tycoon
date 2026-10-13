package core

import "testing"

// TestActionStubs_ValidateApply は全 25 種の Validate/Apply が nil を返すことを検証する。
//
// Phase 1-3 の一時検証。Phase 1-4 以降で本実装に置き換わるたびに、
// このテストもアクションごとに更新する。
func TestActionStubs_ValidateApply(t *testing.T) {
	g := NewGame(1907, "hankyu")
	for _, a := range allActions() {
		t.Run(a.Type(), func(t *testing.T) {
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
		})
	}
}
