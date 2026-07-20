package text

import "testing"

func TestCursorMovementAcceptsOffsetsInsideACluster(t *testing.T) {
	t.Parallel()

	if got, ok := NextGraphemeBoundary("A日B", 2); !ok || got != 4 {
		t.Fatalf("NextGraphemeBoundary() = %d, %t; want 4, true", got, ok)
	}
	if got, ok := PreviousGraphemeBoundary("A日B", 3); !ok || got != 1 {
		t.Fatalf("PreviousGraphemeBoundary() = %d, %t; want 1, true", got, ok)
	}
}

func TestCustomWidthOverridesCompleteGrapheme(t *testing.T) {
	t.Parallel()

	profile := CustomWidth(ModernWidth(), func(grapheme string) (int, bool) {
		return 1, grapheme == "日"
	})
	if got := GraphemeWidth("日", profile); got != 1 {
		t.Fatalf("GraphemeWidth(日) = %d, want 1", got)
	}
	if got := GraphemeWidth("本", profile); got != 2 {
		t.Fatalf("GraphemeWidth(本) = %d, want 2", got)
	}
}

func TestEmptyTextHasNoGraphemeToOverride(t *testing.T) {
	t.Parallel()

	profile := CustomWidth(ModernWidth(), func(string) (int, bool) {
		return 1, true
	})
	if got := GraphemeWidth("", profile); got != 0 {
		t.Fatalf("GraphemeWidth(empty) = %d, want 0", got)
	}
}

func TestCellOperationsDoNotSplitWideGraphemes(t *testing.T) {
	t.Parallel()

	if got := Truncate("A日B", 2, ModernWidth()); got != "A" {
		t.Fatalf("Truncate() = %q, want A", got)
	}
	if got := Wrap("A日B", 2, ModernWidth()); len(got) != 3 || got[0] != "A" || got[1] != "日" || got[2] != "B" {
		t.Fatalf("Wrap() = %q, want [A 日 B]", got)
	}
	if _, ok := CellAtByte("A日B", 2, ModernWidth()); ok {
		t.Fatal("CellAtByte() succeeded inside a wide grapheme")
	}
	if _, ok := ByteAtCell("A日B", 2, ModernWidth()); ok {
		t.Fatal("ByteAtCell() succeeded inside a wide grapheme")
	}
}
