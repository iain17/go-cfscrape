package cfscrape

import (
	"strings"
	"testing"
)

func TestNodeInvocation(t *testing.T) {
	s, err := NodeExecute(`5+5`)
	if err != nil {
		t.Error(err)
	}

	s = strings.TrimSpace(s)

	if s != "10" {
		t.Errorf("Got %s. Expected %s\n", s, "10")
	}
}
