package cli

import (
	"errors"
	"testing"
)

func TestSession_Execute_Unknown(t *testing.T) {
	s := NewSession("http://example.test")
	err := s.Execute(Parse("bogus"))
	var ue *UnknownCommandError
	if !errors.As(err, &ue) {
		t.Fatalf("err = %v, want UnknownCommandError", err)
	}
	if ue.Verb != "bogus" {
		t.Errorf("Verb = %q, want bogus", ue.Verb)
	}
}

func TestSession_Execute_NotRunning(t *testing.T) {
	s := NewSession("http://example.test")
	for _, line := range []string{"status", "build a b", "et", "lines"} {
		if err := s.Execute(Parse(line)); !errors.Is(err, ErrNotRunning) {
			t.Errorf("Execute(%q) err = %v, want ErrNotRunning", line, err)
		}
	}
}

func TestSession_Execute_HelpBeforeStart(t *testing.T) {
	buf := captureOutput(t)
	s := NewSession("http://example.test")
	if err := s.Execute(Parse("help")); err != nil {
		t.Fatalf("help: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("help produced no output")
	}
}

func TestSession_Execute_Empty(t *testing.T) {
	s := NewSession("http://example.test")
	if err := s.Execute(Parse("   ")); err != nil {
		t.Errorf("empty: %v", err)
	}
	if err := s.Execute(nil); err != nil {
		t.Errorf("nil: %v", err)
	}
}

func TestSession_Alias(t *testing.T) {
	s := NewSession("http://example.test")
	if err := s.Execute(Parse("st")); !errors.Is(err, ErrNotRunning) {
		t.Errorf("alias st err = %v, want ErrNotRunning", err)
	}
}
