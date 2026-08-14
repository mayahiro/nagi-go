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
	if cap(decoder.sequence) > MaxSequenceBytes {
		t.Fatalf("retained sequence capacity = %d", cap(decoder.sequence))
	}
}

func TestKittyKeyWarmedAllocationsAreOwnedResultsOnly(t *testing.T) {
	input := []byte("\x1B[97;2:1;65u")
	decoder := NewDecoder()
	if events := decoder.Feed(input); len(events) != 1 {
		t.Fatalf("warm-up events = %#v", events)
	}
	var events []Event
	allocations := testing.AllocsPerRun(1_000, func() {
		events = decoder.Feed(input)
	})
	if len(events) != 1 || events[0].Kind != EventKey {
		t.Fatalf("decoded events = %#v", events)
	}
	if allocations > 2 {
		t.Fatalf("warmed Kitty key allocations = %f, want at most 2", allocations)
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
