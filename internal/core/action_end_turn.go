package core

// Validate は EndTurnAction が実行可能かどうかを検証する。
//
// 検証内容:
//   - プレイヤー会社が存在する
//   - ゲームが既に敗北していない
//   - ゲームが既に終了していない
func (a *EndTurnAction) Validate(g *Game) error {
	if g == nil {
		return NewValidationError("GAME_NIL", "ゲームが初期化されていません", nil)
	}
	if g.PlayerCompany() == nil {
		return NewValidationError("NO_PLAYER", "プレイヤー会社が存在しません", ErrNoPlayer)
	}
	return nil
}

// Apply は EndTurnAction を実行する。
//
// g.EndTurn() を呼び、発生したイベントを返す。
func (a *EndTurnAction) Apply(g *Game) ([]Event, error) {
	if err := a.Validate(g); err != nil {
		return nil, err
	}
	events := g.EndTurn()
	return events, nil
}
