package cli

// handleHelp はコマンド一覧、または指定コマンドの詳細を表示する。
func handleHelp(_ *Session, cmd *Command) error {
	outputMu.Lock()
	defer outputMu.Unlock()
	printHelp(Stdout, cmd)
	return nil
}
