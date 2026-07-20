package vt

import "testing"

func TestStyleAttributesRoundTrip(t *testing.T) {
	t.Parallel()

	attributes := Attributes{Bold: true, Underline: true}
	style := (Style{}).WithAttributes(attributes)

	if got := style.Attributes(); got != attributes {
		t.Fatalf("Attributes() = %+v, want %+v", got, attributes)
	}
	if attributes.Empty() {
		t.Fatal("non-empty attributes reported empty")
	}
}

func TestStyleMergePreservesUnspecifiedValues(t *testing.T) {
	t.Parallel()

	base := Style{Foreground: IndexedColor(2), Italic: true}
	overlay := Style{Background: RGBColor(1, 2, 3), Bold: true}
	merged := base.Merge(overlay)

	if merged.Foreground != base.Foreground {
		t.Fatalf("foreground = %+v, want %+v", merged.Foreground, base.Foreground)
	}
	if merged.Background != overlay.Background {
		t.Fatalf("background = %+v, want %+v", merged.Background, overlay.Background)
	}
	if !merged.Bold || !merged.Italic {
		t.Fatalf("attributes = %+v, want bold and italic", merged.Attributes())
	}
}
