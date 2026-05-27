package main

import "testing"

func TestVersion(t *testing.T) {
    if version == "" {
        t.Errorf("version should not be empty")
    }
}
