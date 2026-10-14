package cli

import (
	"strconv"
	"strings"
)

// Command はパース済みの 1 コマンドを表す。
//
// Verb は先頭トークン、Args は位置引数、Options は `--key value`、
// Flags は `--flag`（値なし）を保持する。
type Command struct {
	Verb    string
	Args    []string
	Options map[string]string
	Flags   map[string]bool
}

// Parse は 1 行の入力文字列を Command へ変換する。
//
// 空行は Verb が空の Command を返す。オプションは次の規則で解釈する。
//   - `--key value` または `--key=value` → Options
//   - `--flag`（次トークンが `--` 始まり、または末尾）→ Flags
func Parse(line string) *Command {
	c := &Command{
		Options: map[string]string{},
		Flags:   map[string]bool{},
	}
	tokens := tokenize(line)
	if len(tokens) == 0 {
		return c
	}
	c.Verb = tokens[0]
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		if !strings.HasPrefix(tok, "--") || len(tok) == 2 {
			c.Args = append(c.Args, tok)
			continue
		}
		key := tok[2:]
		if eq := strings.IndexByte(key, '='); eq >= 0 {
			c.Options[key[:eq]] = key[eq+1:]
			continue
		}
		if i+1 < len(tokens) && !strings.HasPrefix(tokens[i+1], "--") {
			c.Options[key] = tokens[i+1]
			i++
			continue
		}
		c.Flags[key] = true
	}
	return c
}

// tokenize は空白区切りでトークン化する。
//
// ダブルクォート・シングルクォートで囲まれた範囲は 1 トークンとして扱い、
// クォート文字は取り除く。バックスラッシュはクォート内でのエスケープに使う。
func tokenize(line string) []string {
	var tokens []string
	var buf strings.Builder
	inToken := false
	var quote rune

	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch {
		case quote != 0:
			if r == '\\' && i+1 < len(runes) {
				i++
				buf.WriteRune(runes[i])
				continue
			}
			if r == quote {
				quote = 0
				continue
			}
			buf.WriteRune(r)
		case r == '"' || r == '\'':
			quote = r
			inToken = true
		case r == ' ' || r == '\t':
			if inToken {
				tokens = append(tokens, buf.String())
				buf.Reset()
				inToken = false
			}
		default:
			buf.WriteRune(r)
			inToken = true
		}
	}
	if inToken {
		tokens = append(tokens, buf.String())
	}
	return tokens
}

// Arg は i 番目の位置引数と、存在するかどうかを返す。
func (c *Command) Arg(i int) (string, bool) {
	if i < 0 || i >= len(c.Args) {
		return "", false
	}
	return c.Args[i], true
}

// Option はオプション値を返す。未指定なら def を返す。
func (c *Command) Option(key, def string) string {
	if v, ok := c.Options[key]; ok {
		return v
	}
	return def
}

// Flag はフラグが立っているかを返す。
func (c *Command) Flag(key string) bool { return c.Flags[key] }

// OptInt は整数オプションを返す。未指定なら def、不正ならエラーを返す。
func (c *Command) OptInt(key string, def int) (int, error) {
	v, ok := c.Options[key]
	if !ok {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, &ParseError{Key: key, Value: v}
	}
	return n, nil
}

// OptInt64 は int64 オプションを返す。未指定なら def、不正ならエラーを返す。
func (c *Command) OptInt64(key string, def int64) (int64, error) {
	v, ok := c.Options[key]
	if !ok {
		return def, nil
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, &ParseError{Key: key, Value: v}
	}
	return n, nil
}

// OptFloat は浮動小数点オプションを返す。未指定なら def、不正ならエラーを返す。
func (c *Command) OptFloat(key string, def float64) (float64, error) {
	v, ok := c.Options[key]
	if !ok {
		return def, nil
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, &ParseError{Key: key, Value: v}
	}
	return f, nil
}

// ParseError は数値オプションの解釈に失敗したことを表す。
type ParseError struct {
	Key   string
	Value string
}

// Error はエラーメッセージを返す。
func (e *ParseError) Error() string {
	return "--" + e.Key + " に数値を指定してください: " + e.Value
}
