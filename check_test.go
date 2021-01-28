package main

import (
	"testing"
)

func TestCheck(t *testing.T) {
	s := check()

	expected := "Hello, Golang\n"

	if s != expected {
		t.Errorf("Tag.String()\n  is:%s\nwant:%s", s, expected)
	}
}
