package text

import "sort"

// UnicodeVersion is the Unicode data version used by segmentation and width
// calculations
const UnicodeVersion = unicodeVersion

type graphemeBreak uint8

const (
	graphemeOther graphemeBreak = iota
	graphemeCR
	graphemeLF
	graphemeControl
	graphemeExtend
	graphemeZWJ
	graphemeRegionalIndicator
	graphemePrepend
	graphemeSpacingMark
	graphemeL
	graphemeV
	graphemeT
	graphemeLV
	graphemeLVT
)

func (property graphemeBreak) isControl() bool {
	return property == graphemeCR || property == graphemeLF || property == graphemeControl
}

type indicConjunctBreak uint8

const (
	indicNone indicConjunctBreak = iota
	indicConsonant
	indicExtend
	indicLinker
)

type eastAsianWidth uint8

const (
	eastAsianNarrow eastAsianWidth = iota
	eastAsianAmbiguous
	eastAsianWide
)

type codePointRange struct {
	start rune
	end   rune
}

type valueRange struct {
	start rune
	end   rune
	value uint8
}

func graphemeProperty(character rune) graphemeBreak {
	return graphemeBreak(valueAt(graphemeBreakRanges[:], character))
}

func indicConjunctProperty(character rune) indicConjunctBreak {
	return indicConjunctBreak(valueAt(indicConjunctBreakRanges[:], character))
}

func eastAsianWidthProperty(character rune) eastAsianWidth {
	return eastAsianWidth(valueAt(eastAsianWidthRanges[:], character))
}

func isExtendedPictographic(character rune) bool {
	return containsCodePoint(extendedPictographicRanges[:], character)
}

func isEmojiPresentation(character rune) bool {
	return containsCodePoint(emojiPresentationRanges[:], character)
}

func isEmojiVariationBase(character rune) bool {
	index, found := slicesBinarySearch(emojiVariationBases[:], character)
	return found && index < len(emojiVariationBases)
}

func isRGIEmoji(sequence []rune) bool {
	index := sort.Search(len(rgiEmojiSequences), func(index int) bool {
		return compareRunes(rgiEmojiSequences[index], sequence) >= 0
	})
	return index < len(rgiEmojiSequences) && compareRunes(rgiEmojiSequences[index], sequence) == 0
}

func valueAt(ranges []valueRange, character rune) uint8 {
	index := sort.Search(len(ranges), func(index int) bool { return ranges[index].end >= character })
	if index < len(ranges) && ranges[index].start <= character {
		return ranges[index].value
	}
	return 0
}

func containsCodePoint(ranges []codePointRange, character rune) bool {
	index := sort.Search(len(ranges), func(index int) bool { return ranges[index].end >= character })
	return index < len(ranges) && ranges[index].start <= character
}

func slicesBinarySearch(values []rune, wanted rune) (int, bool) {
	index := sort.Search(len(values), func(index int) bool { return values[index] >= wanted })
	return index, index < len(values) && values[index] == wanted
}

func compareRunes(left, right []rune) int {
	for index := 0; index < min(len(left), len(right)); index++ {
		if left[index] < right[index] {
			return -1
		}
		if left[index] > right[index] {
			return 1
		}
	}
	switch {
	case len(left) < len(right):
		return -1
	case len(left) > len(right):
		return 1
	default:
		return 0
	}
}
