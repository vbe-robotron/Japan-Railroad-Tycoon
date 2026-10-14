package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"

	"github.com/you/jrt/internal/cli"
)

func main() {
	baseURL := flag.String("server", "http://localhost:8080", "server URL")
	flag.Parse()

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "jrt> ",
		HistoryFile:     historyPath(),
		AutoComplete:    cli.NewCompleter(),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	session := cli.NewSession(*baseURL)
	session.SetRefresher(rl)
	defer session.Close()

	fmt.Println("Japan Railroad Tycoon v0.1")
	fmt.Println("'help' でコマンド一覧、'start <company>' で開始")

	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		cmd := cli.Parse(line)
		if err := session.Execute(cmd); err != nil {
			if errors.Is(err, cli.ErrQuit) {
				break
			}
			cli.PrintError(err)
		}
	}
}

// historyPath はヒストリファイルのパスを返す。取得できなければ空を返す。
func historyPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".jrt_history")
}
