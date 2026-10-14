package cli

import (
	"errors"
	"fmt"
	"strconv"
)

// handleBuyShares は他社株を取得する。
func handleBuyShares(s *Session, cmd *Command) error {
	target, ok := cmd.Arg(0)
	if !ok {
		return errors.New("会社 ID を指定してください: buy-shares <company_id> <count>")
	}
	countStr, ok := cmd.Arg(1)
	if !ok {
		return errors.New("株数を指定してください: buy-shares <company_id> <count>")
	}
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return fmt.Errorf("株数は整数で指定してください: %s", countStr)
	}
	ctx, cancel := s.ctx()
	defer cancel()

	events, err := s.Game.BuyShares(ctx, target, count)
	if err != nil {
		return err
	}
	s.printEvents(events)
	return nil
}

// handleTOB は TOB を仕掛ける。
func handleTOB(s *Session, cmd *Command) error {
	target, ok := cmd.Arg(0)
	if !ok {
		return errors.New("会社 ID を指定してください: tob <company_id> [--premium 1.1] [--ratio 0.51]")
	}
	premium, err := cmd.OptFloat("premium", 1.1)
	if err != nil {
		return err
	}
	ratio, err := cmd.OptFloat("ratio", 0.51)
	if err != nil {
		return err
	}
	ctx, cancel := s.ctx()
	defer cancel()

	events, err := s.Game.TenderOffer(ctx, target, premium, ratio)
	if err != nil {
		return err
	}
	s.printEvents(events)
	return nil
}

// handleMerge は他社と合併する。
func handleMerge(s *Session, cmd *Command) error {
	target, ok := cmd.Arg(0)
	if !ok {
		return errors.New("会社 ID を指定してください: merge <company_id> <new_name> [--ratio 0.5]")
	}
	name, ok := cmd.Arg(1)
	if !ok {
		return errors.New("新会社名を指定してください: merge <company_id> <new_name> [--ratio 0.5]")
	}
	ratio, err := cmd.OptFloat("ratio", 0.5)
	if err != nil {
		return err
	}
	ctx, cancel := s.ctx()
	defer cancel()

	events, err := s.Game.Merge(ctx, target, name, ratio)
	if err != nil {
		return err
	}
	s.printEvents(events)
	return nil
}
