package text

import "unicode/utf8"

// Grapheme is one extended grapheme cluster and its UTF-8 byte range
type Grapheme struct {
	// Text is the cluster text after invalid UTF-8 replacement; Unicode
	// normalization is not applied
	Text string
	// Start is the inclusive UTF-8 byte offset in the normalized input
	Start int
	// End is the exclusive UTF-8 byte offset in the normalized input
	End int
}

// Graphemes returns the Unicode extended grapheme clusters in input
//
// Invalid UTF-8 runs are replaced by one U+FFFD before byte ranges are
// calculated
func Graphemes(input string) []Grapheme {
	input = NormalizeUTF8(input)
	if input == "" {
		return nil
	}
	graphemes := make([]Grapheme, 0, utf8.RuneCountInString(input))
	for start := 0; start < len(input); {
		end := nextBoundaryFrom(input, start)
		graphemes = append(graphemes, Grapheme{Text: input[start:end], Start: start, End: end})
		start = end
	}
	return graphemes
}

// GraphemeBoundaries returns all extended grapheme cluster boundaries,
// including zero and the normalized UTF-8 length
func GraphemeBoundaries(input string) []int {
	input = NormalizeUTF8(input)
	boundaries := make([]int, 1, utf8.RuneCountInString(input)+1)
	for _, grapheme := range Graphemes(input) {
		boundaries = append(boundaries, grapheme.End)
	}
	return boundaries
}

// IsGraphemeBoundary reports whether byteOffset is an extended grapheme
// cluster boundary in the normalized input
func IsGraphemeBoundary(input string, byteOffset int) bool {
	input = NormalizeUTF8(input)
	if byteOffset < 0 || byteOffset > len(input) {
		return false
	}
	if byteOffset == 0 {
		return true
	}
	for _, grapheme := range Graphemes(input) {
		if grapheme.End == byteOffset {
			return true
		}
	}
	return false
}

// NextGraphemeBoundary returns the nearest strict extended grapheme boundary
// after byteOffset
//
// The offset may be inside a UTF-8 sequence or grapheme. The second result is
// false at or beyond the normalized string end
func NextGraphemeBoundary(input string, byteOffset int) (int, bool) {
	input = NormalizeUTF8(input)
	if byteOffset < 0 || byteOffset >= len(input) {
		return 0, false
	}
	for _, grapheme := range Graphemes(input) {
		if grapheme.End > byteOffset {
			return grapheme.End, true
		}
	}
	return 0, false
}

// PreviousGraphemeBoundary returns the nearest strict extended grapheme
// boundary before byteOffset
//
// The offset may be inside a UTF-8 sequence or grapheme. The second result is
// false at zero or outside the normalized string
func PreviousGraphemeBoundary(input string, byteOffset int) (int, bool) {
	input = NormalizeUTF8(input)
	if byteOffset <= 0 || byteOffset > len(input) {
		return 0, false
	}
	previous := 0
	for _, grapheme := range Graphemes(input) {
		if grapheme.End >= byteOffset {
			break
		}
		previous = grapheme.End
	}
	return previous, true
}

func nextBoundaryFrom(input string, start int) int {
	first, size := utf8.DecodeRuneInString(input[start:])
	prefix := []rune{first}
	for offset := start + size; offset < len(input); {
		right, rightSize := utf8.DecodeRuneInString(input[offset:])
		if shouldBreak(prefix, right) {
			return offset
		}
		prefix = append(prefix, right)
		offset += rightSize
	}
	return len(input)
}

func shouldBreak(prefix []rune, right rune) bool {
	left := graphemeProperty(prefix[len(prefix)-1])
	rightProperty := graphemeProperty(right)
	if left == graphemeCR && rightProperty == graphemeLF {
		return false
	}
	if left.isControl() || rightProperty.isControl() {
		return true
	}
	if left == graphemeL && (rightProperty == graphemeL || rightProperty == graphemeV || rightProperty == graphemeLV || rightProperty == graphemeLVT) {
		return false
	}
	if (left == graphemeLV || left == graphemeV) && (rightProperty == graphemeV || rightProperty == graphemeT) {
		return false
	}
	if (left == graphemeLVT || left == graphemeT) && rightProperty == graphemeT {
		return false
	}
	if rightProperty == graphemeExtend || rightProperty == graphemeZWJ {
		return false
	}
	if rightProperty == graphemeSpacingMark || left == graphemePrepend {
		return false
	}
	if indicLinkerBefore(prefix, right) || emojiZWJBefore(prefix, right) {
		return false
	}
	if rightProperty == graphemeRegionalIndicator && trailingRegionalIndicators(prefix)%2 == 1 {
		return false
	}
	return true
}

func indicLinkerBefore(prefix []rune, right rune) bool {
	if indicConjunctProperty(right) != indicConsonant {
		return false
	}
	linkerSeen := false
	for index := len(prefix) - 1; index >= 0; index-- {
		switch indicConjunctProperty(prefix[index]) {
		case indicLinker:
			linkerSeen = true
		case indicExtend:
		case indicConsonant:
			return linkerSeen
		default:
			return false
		}
	}
	return false
}

func emojiZWJBefore(prefix []rune, right rune) bool {
	if !isExtendedPictographic(right) || graphemeProperty(prefix[len(prefix)-1]) != graphemeZWJ {
		return false
	}
	for index := len(prefix) - 2; index >= 0; index-- {
		if graphemeProperty(prefix[index]) != graphemeExtend {
			return isExtendedPictographic(prefix[index])
		}
	}
	return false
}

func trailingRegionalIndicators(prefix []rune) int {
	count := 0
	for index := len(prefix) - 1; index >= 0; index-- {
		if graphemeProperty(prefix[index]) != graphemeRegionalIndicator {
			break
		}
		count++
	}
	return count
}
