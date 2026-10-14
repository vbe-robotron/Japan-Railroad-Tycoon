package cli

import (
	"reflect"
	"testing"
)

func TestParse_Verb(t *testing.T) {
	c := Parse("status")
	if c.Verb != "status" {
		t.Errorf("Verb = %q, want status", c.Verb)
	}
	if len(c.Args) != 0 {
		t.Errorf("Args = %v, want empty", c.Args)
	}
}

func TestParse_Args(t *testing.T) {
	c := Parse("build umeda takarazuka")
	want := []string{"umeda", "takarazuka"}
	if !reflect.DeepEqual(c.Args, want) {
		t.Errorf("Args = %v, want %v", c.Args, want)
	}
}

func TestParse_Options(t *testing.T) {
	c := Parse("start hankyu --seed 1907 --difficulty hard")
	if got := c.Options["seed"]; got != "1907" {
		t.Errorf("seed = %q, want 1907", got)
	}
	if got := c.Options["difficulty"]; got != "hard" {
		t.Errorf("difficulty = %q, want hard", got)
	}
	if len(c.Args) != 1 || c.Args[0] != "hankyu" {
		t.Errorf("Args = %v, want [hankyu]", c.Args)
	}
}

func TestParse_OptionEquals(t *testing.T) {
	c := Parse("build a b --gauge=1435")
	if got := c.Options["gauge"]; got != "1435" {
		t.Errorf("gauge = %q, want 1435", got)
	}
}

func TestParse_Flags(t *testing.T) {
	c := Parse("build a b --double")
	if !c.Flag("double") {
		t.Error("double flag not set")
	}
	if _, ok := c.Options["double"]; ok {
		t.Error("double should be a flag, not option")
	}
	if len(c.Args) != 2 {
		t.Errorf("Args = %v, want 2", c.Args)
	}
}

func TestParse_FlagAfterOptions(t *testing.T) {
	c := Parse("build a b --gauge 762 --double")
	if got := c.Options["gauge"]; got != "762" {
		t.Errorf("gauge = %q, want 762", got)
	}
	if !c.Flag("double") {
		t.Error("double flag not set")
	}
}

func TestParse_QuotedString(t *testing.T) {
	c := Parse(`merge abc "Hankyu Hanshin" --ratio 0.5`)
	if len(c.Args) != 2 {
		t.Fatalf("Args = %v, want 2", c.Args)
	}
	if c.Args[1] != "Hankyu Hanshin" {
		t.Errorf("Args[1] = %q, want %q", c.Args[1], "Hankyu Hanshin")
	}
	if got := c.Options["ratio"]; got != "0.5" {
		t.Errorf("ratio = %q, want 0.5", got)
	}
}

func TestParse_Empty(t *testing.T) {
	c := Parse("   ")
	if c.Verb != "" {
		t.Errorf("Verb = %q, want empty", c.Verb)
	}
}

func TestTokenize(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"a b c", []string{"a", "b", "c"}},
		{"  a   b  ", []string{"a", "b"}},
		{`"a b" c`, []string{"a b", "c"}},
		{`'a b' c`, []string{"a b", "c"}},
		{`a "b\"c"`, []string{`a`, `b"c`}},
		{"", nil},
	}
	for _, tc := range cases {
		got := tokenize(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("tokenize(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestCommand_OptInt(t *testing.T) {
	c := Parse("x --n 5")
	if got, err := c.OptInt("n", 1); err != nil || got != 5 {
		t.Errorf("OptInt = (%d, %v), want (5, nil)", got, err)
	}
	if got, err := c.OptInt("missing", 7); err != nil || got != 7 {
		t.Errorf("OptInt default = (%d, %v), want (7, nil)", got, err)
	}
	if _, err := Parse("x --n bad").OptInt("n", 0); err == nil {
		t.Error("OptInt bad: want error")
	}
}

func TestCommand_OptFloat(t *testing.T) {
	c := Parse("x --ratio 0.51")
	if got, err := c.OptFloat("ratio", 0.5); err != nil || got != 0.51 {
		t.Errorf("OptFloat = (%v, %v), want (0.51, nil)", got, err)
	}
	if _, err := Parse("x --ratio no").OptFloat("ratio", 0); err == nil {
		t.Error("OptFloat bad: want error")
	}
}
