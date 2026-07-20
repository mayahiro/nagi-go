package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRangesMergeDeterministically(t *testing.T) {
	t.Parallel()

	if got, want := mergeRanges([]codePointRange{{start: 3, end: 4}, {start: 1, end: 2}}), []codePointRange{{start: 1, end: 4}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("mergeRanges() = %+v, want %+v", got, want)
	}
	input := []valueRange{
		{start: 2, end: 2, value: 1},
		{start: 1, end: 1, value: 1},
		{start: 3, end: 3, value: 2},
	}
	want := []valueRange{{start: 1, end: 2, value: 1}, {start: 3, end: 3, value: 2}}
	if got := mergeValueRanges(input); !reflect.DeepEqual(got, want) {
		t.Fatalf("mergeValueRanges() = %+v, want %+v", got, want)
	}
}

func TestSourceRejectsVersionMarkerAfterHeader(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	path := filepath.Join(directory, "data.txt")
	if err := os.WriteFile(path, []byte(strings.Repeat("\n", 30)+unicodeVersion+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := source(directory, "data.txt", unicodeVersion); err == nil {
		t.Fatal("source() accepted a version marker after the first 30 lines")
	}
}
