package main

import (
	"testing"
)

func TestCheckNodeJS(t *testing.T) {
	// checkNodeJS runs `node --version`.
	// This test simply verifies it doesn't crash, 
	// regardless of whether node is installed or not.
	_ = checkNodeJS()
}
