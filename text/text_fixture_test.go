package text

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/mayahiro/nagi-go/internal/conformance"
)

func TestGraphemeFixtures(t *testing.T) {
	checkGraphemeFile(t, "text/graphemes.txt", "text-graphemes")
}

func TestUnicodeGraphemeConformance(t *testing.T) {
	checkGraphemeFile(t, "text/grapheme-break-17.0.0.txt", "text-grapheme-break")
}

func TestWidthFixtures(t *testing.T) {
	records := textRecords(t, "text/width.txt", "text-width", "text", "modern", "cjk")
	for _, record := range records {
		input := record.Text("text")
		if got, want := Width(input, ModernWidth()), fixtureNumber(record.Field("modern")); got != want {
			t.Errorf("case %s modern: Width() = %d, want %d", record.ID, got, want)
		}
		if got, want := Width(input, CJKWidth()), fixtureNumber(record.Field("cjk")); got != want {
			t.Errorf("case %s cjk: Width() = %d, want %d", record.ID, got, want)
		}
	}
}

func TestTruncateFixtures(t *testing.T) {
	records := textRecords(t, "text/truncate.txt", "text-truncate", "text", "cells", "profile", "expected")
	for _, record := range records {
		got := Truncate(record.Text("text"), fixtureNumber(record.Field("cells")), fixtureProfile(record.Field("profile")))
		if want := record.Text("expected"); got != want {
			t.Errorf("case %s: Truncate() = %q, want %q", record.ID, got, want)
		}
	}
}

func TestWrapFixtures(t *testing.T) {
	records := textRecords(t, "text/wrap.txt", "text-wrap", "text", "cells", "profile", "expected")
	for _, record := range records {
		rawLines := strings.Split(record.Field("expected"), "|")
		want := make([]string, len(rawLines))
		for index, raw := range rawLines {
			decoded, err := conformance.Decode(raw)
			if err != nil {
				t.Fatalf("case %s: invalid expected line: %v", record.ID, err)
			}
			want[index] = string(decoded)
		}
		got := Wrap(record.Text("text"), fixtureNumber(record.Field("cells")), fixtureProfile(record.Field("profile")))
		if !reflect.DeepEqual(got, want) {
			t.Errorf("case %s: Wrap() = %q, want %q", record.ID, got, want)
		}
	}
}

func TestPositionFixtures(t *testing.T) {
	records := textRecords(t, "text/position.txt", "text-position", "text", "byte", "cell")
	for _, record := range records {
		input := record.Text("text")
		byteOffset, hasByte := fixtureOptionalNumber(record.Field("byte"))
		cellOffset, hasCell := fixtureOptionalNumber(record.Field("cell"))
		if hasByte {
			got, ok := CellAtByte(input, byteOffset, ModernWidth())
			if ok != hasCell || (ok && got != cellOffset) {
				t.Errorf("case %s byte to cell: got %d, %t; want %d, %t", record.ID, got, ok, cellOffset, hasCell)
			}
		}
		if hasCell {
			got, ok := ByteAtCell(input, cellOffset, ModernWidth())
			if ok != hasByte || (ok && got != byteOffset) {
				t.Errorf("case %s cell to byte: got %d, %t; want %d, %t", record.ID, got, ok, byteOffset, hasByte)
			}
		}
	}
}

func TestCursorFixtures(t *testing.T) {
	records := textRecords(t, "text/cursor.txt", "text-cursor", "text", "offset", "previous", "next")
	for _, record := range records {
		input := record.Text("text")
		offset := fixtureNumber(record.Field("offset"))
		wantPrevious, hasPrevious := fixtureOptionalNumber(record.Field("previous"))
		if got, ok := PreviousGraphemeBoundary(input, offset); ok != hasPrevious || (ok && got != wantPrevious) {
			t.Errorf("case %s previous: got %d, %t; want %d, %t", record.ID, got, ok, wantPrevious, hasPrevious)
		}
		wantNext, hasNext := fixtureOptionalNumber(record.Field("next"))
		if got, ok := NextGraphemeBoundary(input, offset); ok != hasNext || (ok && got != wantNext) {
			t.Errorf("case %s next: got %d, %t; want %d, %t", record.ID, got, ok, wantNext, hasNext)
		}
	}
}

func TestInvalidUTF8Fixtures(t *testing.T) {
	records := textRecords(t, "text/invalid-utf8.txt", "text-invalid-utf8", "input", "expected")
	for _, record := range records {
		if got, want := NormalizeUTF8(string(record.Bytes("input"))), record.Text("expected"); got != want {
			t.Errorf("case %s: NormalizeUTF8() = %q, want %q", record.ID, got, want)
		}
	}
}

func checkGraphemeFile(t *testing.T, path, suite string) {
	t.Helper()
	records := textRecords(t, path, suite, "text", "boundaries")
	for _, record := range records {
		got := GraphemeBoundaries(record.Text("text"))
		want := fixtureNumbers(record.Field("boundaries"))
		if !reflect.DeepEqual(got, want) {
			t.Errorf("case %s: GraphemeBoundaries() = %v, want %v", record.ID, got, want)
		}
	}
}

func textRecords(t *testing.T, path, suite string, fields ...string) []conformance.Record {
	t.Helper()
	records, err := conformance.Load(path, suite, fields...)
	if errors.Is(err, conformance.ErrNoFixtureRoot) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	return records
}

func fixtureProfile(value string) WidthProfile {
	switch value {
	case "modern":
		return ModernWidth()
	case "cjk":
		return CJKWidth()
	default:
		panic("unknown width profile " + value)
	}
}

func fixtureNumbers(value string) []int {
	parts := strings.Split(value, ",")
	numbers := make([]int, len(parts))
	for index, part := range parts {
		numbers[index] = fixtureNumber(part)
	}
	return numbers
}

func fixtureNumber(value string) int {
	number, err := strconv.Atoi(value)
	if err != nil {
		panic("invalid fixture number " + value)
	}
	return number
}

func fixtureOptionalNumber(value string) (int, bool) {
	if value == "none" {
		return 0, false
	}
	return fixtureNumber(value), true
}
