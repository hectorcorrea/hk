package viewModels

import (
	"testing"
)

func TestTextInQuotes(t *testing.T) {
	t1 := "src=\"hello\""
  r1 := viewModels.TextInQuotes(t1)
  if r1 != "hello" {
    t.Errorf("could not find basic text in quotes: %s", t1)
  }

  t2 := "src=\"hello"
  r2 := viewModels.TextInQuotes(t2)
  if r2 != "" {
    t.Errorf("failed to detect unbalanced quotes: %s", t2)
  }

  t3 := "src=hello\""
  r3 := viewModels.TextInQuotes(t3)
  if r3 != "" {
    t.Errorf("failed to detect unbalanced quotes: %s", t3)
  }

  t4 := "src=hello"
  r4 := viewModels.TextInQuotes(t4)
  if r4 != "" {
    t.Errorf("failed to detect lack of quotes: %s", t4)
  }
}
