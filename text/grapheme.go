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

// GraphemeIterator traverses normalized extended grapheme clusters without
// materializing a result slice
type GraphemeIterator struct {
	input string
	next  int
}

// IterateGraphemes returns an allocation-free iterator for valid UTF-8 input
//
// Invalid UTF-8 runs are normalized once when the iterator is created
func IterateGraphemes(input string) GraphemeIterator {
	return GraphemeIterator{input: NormalizeUTF8(input)}
}

// Next returns the next grapheme and whether one was available
func (i *GraphemeIterator) Next() (Grapheme, bool) {
	if i.next == len(i.input) {
		return Grapheme{}, false
	}
	start := i.next
	end := nextBoundaryFrom(i.input, start)
	i.next = end
	return Grapheme{Text: i.input[start:end], Start: start, End: end}, true
}

// Graphemes returns the Unicode extended grapheme clusters in input
//
// Invalid UTF-8 runs are replaced by one U+FFFD before byte ranges are
// calculated
func Graphemes(input string) []Grapheme {
	iterator := IterateGraphemes(input)
	if iterator.input == "" {
		return nil
	}
	graphemes := make([]Grapheme, 0, utf8.RuneCountInString(iterator.input))
	for grapheme, ok := iterator.Next(); ok; grapheme, ok = iterator.Next() {
		graphemes = append(graphemes, grapheme)
	}
	return graphemes
}

// GraphemeBoundaries returns all extended grapheme cluster boundaries,
// including zero and the normalized UTF-8 length
func GraphemeBoundaries(input string) []int {
	iterator := IterateGraphemes(input)
	boundaries := make([]int, 1, utf8.RuneCountInString(iterator.input)+1)
	for grapheme, ok := iterator.Next(); ok; grapheme, ok = iterator.Next() {
		boundaries = append(boundaries, grapheme.End)
	}
	return boundaries
}

// IsGraphemeBoundary reports whether byteOffset is an extended grapheme
// cluster boundary in the normalized input
func IsGraphemeBoundary(input string, byteOffset int) bool {
	iterator := IterateGraphemes(input)
	if byteOffset < 0 || byteOffset > len(iterator.input) {
		return false
	}
	if byteOffset == 0 {
		return true
	}
	for grapheme, ok := iterator.Next(); ok; grapheme, ok = iterator.Next() {
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
	iterator := IterateGraphemes(input)
	if byteOffset < 0 || byteOffset >= len(iterator.input) {
		return 0, false
	}
	for grapheme, ok := iterator.Next(); ok; grapheme, ok = iterator.Next() {
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
	iterator := IterateGraphemes(input)
	if byteOffset <= 0 || byteOffset > len(iterator.input) {
		return 0, false
	}
	previous := 0
	for grapheme, ok := iterator.Next(); ok; grapheme, ok = iterator.Next() {
		if grapheme.End >= byteOffset {
			break
		}
		previous = grapheme.End
	}
	return previous, true
}

func nextBoundaryFrom(input string, start int) int {
	first, size := utf8.DecodeRuneInString(input[start:])
	state := newGraphemeBoundaryState(first)
	for offset := start + size; offset < len(input); {
		right, rightSize := utf8.DecodeRuneInString(input[offset:])
		if state.breaksBefore(right) {
			return offset
		}
		state.append(right)
		offset += rightSize
	}
	return len(input)
}

type graphemeBoundaryState struct {
	left                       graphemeBreak
	trailingRegionalIndicators int
	indicConsonant             bool
	indicLinker                bool
	emojiBase                  bool
	emojiZWJ                   bool
}

func newGraphemeBoundaryState(first rune) graphemeBoundaryState {
	var state graphemeBoundaryState
	state.append(first)
	return state
}

func (s graphemeBoundaryState) breaksBefore(right rune) bool {
	left := s.left
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
	if indicConjunctProperty(right) == indicConsonant && s.indicConsonant && s.indicLinker {
		return false
	}
	if isExtendedPictographic(right) && left == graphemeZWJ && s.emojiZWJ {
		return false
	}
	if rightProperty == graphemeRegionalIndicator && s.trailingRegionalIndicators%2 == 1 {
		return false
	}
	return true
}

func (s *graphemeBoundaryState) append(character rune) {
	property := graphemeProperty(character)
	if property == graphemeRegionalIndicator {
		s.trailingRegionalIndicators++
	} else {
		s.trailingRegionalIndicators = 0
	}

	switch indicConjunctProperty(character) {
	case indicConsonant:
		s.indicConsonant = true
		s.indicLinker = false
	case indicLinker:
		if s.indicConsonant {
			s.indicLinker = true
		}
	case indicExtend:
	default:
		s.indicConsonant = false
		s.indicLinker = false
	}

	s.emojiZWJ = property == graphemeZWJ && s.emojiBase
	if property != graphemeExtend {
		s.emojiBase = isExtendedPictographic(character)
	}
	s.left = property
}
