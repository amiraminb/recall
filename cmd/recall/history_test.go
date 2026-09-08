package main

import (
	"slices"
	"testing"
)

func TestSortedTags(t *testing.T) {
	tags := []string{"zsh", "api", "go"}

	got := sortedTags(tags)
	want := []string{"api", "go", "zsh"}
	if !slices.Equal(got, want) {
		t.Errorf("sortedTags() = %v, want %v", got, want)
	}
	if !slices.Equal(tags, []string{"zsh", "api", "go"}) {
		t.Errorf("sortedTags() mutated the input: %v", tags)
	}
}
