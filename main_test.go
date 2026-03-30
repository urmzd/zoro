package main

import (
	"testing"
)

func TestPrintUsage(t *testing.T) {
	// printUsage writes to stderr; just verify it doesn't panic
	printUsage()
}

func TestVersionVar(t *testing.T) {
	if version == "" {
		t.Error("version should not be empty")
	}
}
