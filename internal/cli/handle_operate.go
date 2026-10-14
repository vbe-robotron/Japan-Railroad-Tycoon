package cli

import (
	"errors"
	"fmt"
	"strconv"
)

// handleSetFare は運賃を変更する。
func handleSetFare(s *Session, cmd *Command) error {
	lineID, ok := cmd.Arg(0)
	if !ok {
		return errors.New("路線 ID を指定してください: set-fare <line_id> <fare>")
	}
	fareStr, ok := cmd.Arg(1)
	if !ok {
		return errors.New("運賃を指定してください: set-fare <line_id> <fare>")
	}
	fare, err := strconv.Atoi(fareStr)
	if err != nil {
		return fmt.Errorf("運賃は整数で指定してください: %s", fareStr)
	}
	ctx, cancel := s.ctx()
	defer cancel()

	events, err := s.Game.SetFare(ctx, lineID, fare)
	if err != nil {
		return err
	}
	s.printEvents(events)
	return nil
}

// handleAbandonLine は路線を廃止する。
func handleAbandonLine(s *Session, cmd *Command) error {
	lineID, ok := cmd.Arg(0)
	if !ok {
		return errors.New("路線 ID を指定してください: abandon-line <line_id> [--reason economic]")
	}
	ctx, cancel := s.ctx()
	defer cancel()

	events, err := s.Game.AbandonLine(ctx, lineID, cmd.Option("reason", "economic"))
	if err != nil {
		return err
	}
	s.printEvents(events)
	return nil
}
