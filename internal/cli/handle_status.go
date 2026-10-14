package cli

import (
	"context"
	"errors"
	"sort"

	"github.com/you/jrt/internal/core"
)

// handleStatus は現在状態を取得して表示する。
func handleStatus(s *Session, _ *Command) error {
	ctx, cancel := s.ctx()
	defer cancel()

	st, err := s.Game.Status(ctx)
	if err != nil {
		return err
	}
	s.State = st
	printStatus(st)
	return nil
}

// handleLines は自社路線一覧を表示する。--sort profit で月次利益順に並べる。
func handleLines(s *Session, cmd *Command) error {
	ctx, cancel := s.ctx()
	defer cancel()

	lines, err := s.Game.ListLines(ctx)
	if err != nil {
		return err
	}
	if cmd.Option("sort", "") == "profit" {
		sort.SliceStable(lines, func(i, j int) bool { return lines[i].Profit > lines[j].Profit })
	}
	printLines(lines, s.cityMap(ctx))
	return nil
}

// handleLine は 1 本の路線の詳細を表示する。
func handleLine(s *Session, cmd *Command) error {
	id, ok := cmd.Arg(0)
	if !ok {
		return errors.New("路線 ID を指定してください: line <line_id>")
	}
	ctx, cancel := s.ctx()
	defer cancel()

	line, err := s.Game.GetLine(ctx, id)
	if err != nil {
		return err
	}
	printLine(line, s.cityMap(ctx))
	return nil
}

// handleCompanies は会社一覧を表示する。
func handleCompanies(s *Session, _ *Command) error {
	ctx, cancel := s.ctx()
	defer cancel()

	companies, err := s.Game.ListCompanies(ctx)
	if err != nil {
		return err
	}
	printCompanies(companies, s.playerID())
	return nil
}

// handleCities は都市一覧を表示する。
func handleCities(s *Session, _ *Command) error {
	ctx, cancel := s.ctx()
	defer cancel()

	cities, err := s.Game.ListCities(ctx)
	if err != nil {
		return err
	}
	printCities(cities)
	return nil
}

// handleCanBuild は敷設可否を判定して表示する。
func handleCanBuild(s *Session, cmd *Command) error {
	from, ok := cmd.Arg(0)
	if !ok {
		return errors.New("出発都市を指定してください: can-build <from> <to> [--gauge N]")
	}
	to, ok := cmd.Arg(1)
	if !ok {
		return errors.New("到着都市を指定してください: can-build <from> <to> [--gauge N]")
	}
	gauge, err := cmd.OptInt("gauge", 1067)
	if err != nil {
		return err
	}
	ctx, cancel := s.ctx()
	defer cancel()

	result, err := s.Game.CanBuild(ctx, from, to, gauge)
	if err != nil {
		return err
	}
	printGates(result)
	return nil
}

// playerID は現在のプレイヤー会社 ID を返す。
func (s *Session) playerID() string {
	if s.State != nil {
		return s.State.PlayerID
	}
	return ""
}

// cityMap は都市 ID → 都市の対応を返す。取得に失敗した場合は nil を返す。
func (s *Session) cityMap(ctx context.Context) map[string]*core.City {
	cities, err := s.Game.ListCities(ctx)
	if err != nil {
		return nil
	}
	m := make(map[string]*core.City, len(cities))
	for i := range cities {
		m[cities[i].ID] = &cities[i]
	}
	return m
}
