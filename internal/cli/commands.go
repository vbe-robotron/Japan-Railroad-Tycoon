package cli

import (
	"fmt"
	"io"
	"sort"
)

// CommandSpec は 1 コマンドの仕様（ヘルプ表示用）。
type CommandSpec struct {
	Name    string // 正式な動詞
	Usage   string // 引数を含む使い方
	Summary string // 説明
	Group   string // グループ名
}

// CommandSpecs は全コマンドの仕様を定義順に並べたもの。
var CommandSpecs = []CommandSpec{
	{Name: "start", Usage: "start <company> [--seed N] [--difficulty normal]", Summary: "新規ゲームを開始する", Group: "セッション"},
	{Name: "load", Usage: "load <game_id>", Summary: "既存ゲームを読み込む", Group: "セッション"},
	{Name: "list", Usage: "list", Summary: "ゲーム一覧を表示する", Group: "セッション"},
	{Name: "quit", Usage: "quit", Summary: "終了する", Group: "セッション"},

	{Name: "status", Usage: "status", Summary: "現在の状態を表示する", Group: "状態"},
	{Name: "lines", Usage: "lines [--sort profit]", Summary: "自社路線の一覧を表示する", Group: "状態"},
	{Name: "line", Usage: "line <line_id>", Summary: "路線の詳細を表示する", Group: "状態"},
	{Name: "companies", Usage: "companies", Summary: "会社一覧を表示する", Group: "状態"},
	{Name: "cities", Usage: "cities", Summary: "都市一覧を表示する", Group: "状態"},
	{Name: "can-build", Usage: "can-build <from> <to> [--gauge N]", Summary: "敷設可否を判定する", Group: "状態"},

	{Name: "build", Usage: "build <from> <to> [--gauge N] [--double]", Summary: "路線を敷設する", Group: "建設"},
	{Name: "build-facility", Usage: "build-facility <city> <kind> [--scale N]", Summary: "開発施設を建設する", Group: "建設"},
	{Name: "set-fare", Usage: "set-fare <line_id> <fare>", Summary: "運賃を変更する", Group: "建設"},
	{Name: "abandon-line", Usage: "abandon-line <line_id> [--reason economic]", Summary: "路線を廃止する", Group: "建設"},

	{Name: "buy-shares", Usage: "buy-shares <company_id> <count>", Summary: "他社株を取得する", Group: "M&A"},
	{Name: "tob", Usage: "tob <company_id> [--premium 1.1] [--ratio 0.51]", Summary: "TOB を仕掛ける", Group: "M&A"},
	{Name: "merge", Usage: "merge <company_id> <new_name> [--ratio 0.5]", Summary: "他社と合併する", Group: "M&A"},

	{Name: "et", Usage: "et", Summary: "ターンを終了する", Group: "ターン"},
	{Name: "turn", Usage: "turn <N>", Summary: "N ターン進める", Group: "ターン"},

	{Name: "help", Usage: "help [<command>]", Summary: "コマンド一覧を表示する", Group: "ヘルプ"},
}

// aliases は別名 → 正式動詞の対応。
var aliases = map[string]string{
	"end-turn": "et",
	"st":       "status",
	"ls":       "lines",
	"bf":       "build-facility",
	"q":        "quit",
	"?":        "help",
}

// groupOrder はヘルプ表示のグループ順。
var groupOrder = []string{"セッション", "状態", "建設", "M&A", "ターン", "ヘルプ"}

// resolveVerb は別名を正式動詞へ解決する。別名でなければそのまま返す。
func resolveVerb(verb string) string {
	if canon, ok := aliases[verb]; ok {
		return canon
	}
	return verb
}

// specOf は正式動詞の仕様を返す。見つからなければ nil。
func specOf(verb string) *CommandSpec {
	for i := range CommandSpecs {
		if CommandSpecs[i].Name == verb {
			return &CommandSpecs[i]
		}
	}
	return nil
}

// sessionVerbs はゲーム開始前でも実行できる動詞。
var sessionVerbs = map[string]bool{
	"start": true,
	"load":  true,
	"list":  true,
	"quit":  true,
	"help":  true,
}

// requiresGame はコマンドの実行にゲームが必要かどうかを返す。
func requiresGame(verb string) bool {
	return !sessionVerbs[verb]
}

// printHelp は全コマンド一覧、または指定コマンドの詳細を書き出す。
func printHelp(w io.Writer, cmd *Command) {
	if cmd != nil {
		if arg, ok := cmd.Arg(0); ok {
			printHelpDetail(w, resolveVerb(arg))
			return
		}
	}
	fmt.Fprintln(w, "コマンド一覧:")
	for _, group := range groupOrder {
		specs := specsInGroup(group)
		if len(specs) == 0 {
			continue
		}
		fmt.Fprintf(w, "\n[%s]\n", group)
		for _, s := range specs {
			fmt.Fprintf(w, "  %-40s %s\n", s.Usage, s.Summary)
		}
	}
	fmt.Fprintln(w, "\n別名: et=end-turn, st=status, ls=lines, bf=build-facility, q=quit, ?=help")
}

// printHelpDetail は 1 コマンドの使い方を書き出す。
func printHelpDetail(w io.Writer, verb string) {
	s := specOf(verb)
	if s == nil {
		fmt.Fprintf(w, "不明なコマンド: %s\n", verb)
		return
	}
	fmt.Fprintf(w, "%s\n  %s\n", s.Usage, s.Summary)
}

// specsInGroup は指定グループのコマンドを定義順で返す。
func specsInGroup(group string) []*CommandSpec {
	var specs []*CommandSpec
	for i := range CommandSpecs {
		if CommandSpecs[i].Group == group {
			specs = append(specs, &CommandSpecs[i])
		}
	}
	return specs
}

// sortedSpecs はコマンド名の一覧をソートして返す（補完用）。
func sortedSpecs() []string {
	names := make([]string, 0, len(CommandSpecs))
	for _, s := range CommandSpecs {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return names
}
