package cli

import "github.com/chzyer/readline"

// NewCompleter は Tab 補完器を生成する。
//
// 固定候補のみを扱う（動的な都市 ID 等の補完は将来対応）。コマンドと
// 主要なオプションを候補に含める。
func NewCompleter() *readline.PrefixCompleter {
	items := make([]readline.PrefixCompleterInterface, 0, len(CommandSpecs))
	for _, name := range sortedSpecs() {
		items = append(items, readline.PcItem(name))
	}
	for _, name := range []string{"st", "ls", "bf", "q", "?", "end-turn"} {
		items = append(items, readline.PcItem(name))
	}
	return readline.NewPrefixCompleter(items...)
}

// commandNames は補完候補のコマンド名一覧を返す（テスト用）。
func commandNames() []string {
	names := sortedSpecs()
	for _, name := range []string{"st", "ls", "bf", "q", "?", "end-turn"} {
		names = append(names, name)
	}
	return names
}
