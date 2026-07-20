package vt

import (
	"reflect"
	"testing"
)

func TestTextChunkBoundary(t *testing.T) {
	decoder := NewDecoder()
	got := decoder.Feed([]byte{0xE6})
	got = append(got, decoder.Feed([]byte{0x97, 0xA5})...)
	got = append(got, decoder.FlushPending()...)
	want := []Event{{Kind: EventText, Text: "日"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("events = %#v, want %#v", got, want)
	}
}

func TestInvalidRunCrossesChunkBoundary(t *testing.T) {
	decoder := NewDecoder()
	got := decoder.Feed([]byte{0xFF})
	got = append(got, decoder.Feed([]byte{0xFE, 'A'})...)
	want := []Event{{Kind: EventText, Text: "\uFFFD"}, {Kind: EventText, Text: "A"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("events = %#v, want %#v", got, want)
	}
}

func TestOversizedPasteIsBoundedAndRecovers(t *testing.T) {
	input := append([]byte(nil), pasteStart...)
	input = append(input, make([]byte, MaxPasteBytes+1)...)
	input = append(input, pasteEnd...)
	decoder := NewDecoder()

	events := decoder.Feed(input)

	want := []Event{{Kind: EventUnknownSequence, Bytes: append([]byte(nil), pasteStart...)}}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
	if decoder.HasPending() {
		t.Fatal("decoder remained pending after paste terminator")
	}
}

func TestEveryByteValueAndIncompleteSuffixIsNonPanicking(t *testing.T) {
	input := make([]byte, 256)
	for index := range input {
		input[index] = byte(index)
	}
	decoder := NewDecoder()
	_ = decoder.Feed(input)
	_ = decoder.FlushPending()
	if decoder.HasPending() {
		t.Fatal("decoder remained pending after flush")
	}
}
