package cli

import (
	"errors"
	"strconv"

	"github.com/you/jrt/internal/api"
)

// handleStart は新規ゲームを作成して開始する。
func handleStart(s *Session, cmd *Command) error {
	company, ok := cmd.Arg(0)
	if !ok {
		return errors.New("会社名を指定してください: start <company> [--seed N]")
	}
	seed, err := cmd.OptInt64("seed", 0)
	if err != nil {
		return err
	}
	ctx, cancel := s.ctx()
	defer cancel()

	g, st, err := s.Client.CreateGame(ctx, api.CreateGameRequest{
		CompanyID:  company,
		Seed:       seed,
		Difficulty: cmd.Option("difficulty", ""),
	})
	if err != nil {
		return err
	}
	s.begin(g, st)
	printSuccess("ゲームを開始しました: %s (ID: %s)", company, st.ID)
	printStatus(st)
	return nil
}

// handleLoad は既存ゲームを読み込む。
func handleLoad(s *Session, cmd *Command) error {
	id, ok := cmd.Arg(0)
	if !ok {
		return errors.New("ゲーム ID を指定してください: load <game_id>")
	}
	ctx, cancel := s.ctx()
	defer cancel()

	g, st, err := s.Client.LoadGame(ctx, id)
	if err != nil {
		return err
	}
	s.begin(g, st)
	printSuccess("ゲームを読み込みました: %s", st.ID)
	printStatus(st)
	return nil
}

// handleList はゲーム一覧を表示する。
func handleList(s *Session, _ *Command) error {
	ctx, cancel := s.ctx()
	defer cancel()

	games, err := s.Client.ListGames(ctx)
	if err != nil {
		return err
	}
	if len(games) == 0 {
		printInfo("ゲームがありません。'start <company>' で開始できます。")
		return nil
	}
	rows := make([][]string, 0, len(games))
	for _, g := range games {
		rows = append(rows, []string{
			g.ID,
			g.CompanyName,
			g.Date.String(),
			strconv.Itoa(g.Turn),
		})
	}
	printTable([]string{"ID", "会社", "日付", "ターン"}, rows)
	return nil
}

// handleQuit はセッションを終了する。
func handleQuit(s *Session, _ *Command) error {
	s.Running = false
	s.stopEvents()
	printInfo("終了します。")
	return ErrQuit
}
