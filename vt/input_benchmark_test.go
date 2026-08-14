package vt

import "testing"

func BenchmarkDecoderKittyKey(b *testing.B) {
	input := []byte("\x1B[97;2:1;65u")
	decoder := NewDecoder()
	b.ReportAllocs()
	b.SetBytes(int64(len(input)))
	for b.Loop() {
		events := decoder.Feed(input)
		if len(events) != 1 || events[0].Kind != EventKey {
			b.Fatalf("decoded events = %#v", events)
		}
	}
}
