package core

// Action はプレイヤーまたは AI の行動を表す。
//
// Validate は状態を変更せず、実行可能かどうかを検証する。
// Apply は状態を変更し、発生したイベントを返す。
//
// すべての Action は JSON で往復可能で、Type() の文字列をキーに
// DecodeAction で復元できる。
type Action interface {
	Type() string
	Validate(g *Game) error
	Apply(g *Game) ([]Event, error)
}
