package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"
)

// handleEndTurn はターンを 1 つ進める。
func handleEndTurn(s *Session, _ *Command) error {
	ctx, cancel := s.ctx()
	defer cancel()

	events, err := s.Game.EndTurn(ctx)
	if err != nil {
		return err
	}
	s.printEvents(events)
	s.refreshStatus(ctx)
	return nil
}

// handleTurnN はターンを N 回進める。
func handleTurnN(s *Session, cmd *Command) error {
	arg, ok := cmd.Arg(0)
	if !ok {
		return errors.New("ターン数を指定してください: turn <N>")
	}
	n, err := strconv.Atoi(arg)
	if err != nil {
		return fmt.Errorf("ターン数は整数で指定してください: %s", arg)
	}
	if n < 0 {
		return errors.New("ターン数は 0 以上で指定してください")
	}
	for i := 0; i < n; i++ {
		ctx, cancel := s.ctx()
		events, err := s.Game.EndTurn(ctx)
		cancel()
		if err != nil {
			return err
		}
		s.printEvents(events)
	}
	ctx, cancel := s.ctx()
	defer cancel()
	s.refreshStatus(ctx)
	return nil
}

// refreshStatus は状態を取得し直し、1 行で要約を表示する。
func (s *Session) refreshStatus(ctx context.Context) {
	st, err := s.Game.Status(ctx)
	if err != nil {
		return
	}
	s.State = st
	printInfo("→ %s (ターン %d)", st.Date.String(), st.Turn)
}
