package text

// WidthOverride can replace the cell width of a complete grapheme
//
// When override is true, cells must be zero, one, or two. Any other value is a
// programmer error and causes a panic
type WidthOverride func(grapheme string) (cells int, override bool)

// WidthProfile is a terminal cell-width policy
type WidthProfile struct {
	ambiguousWide bool
	override      WidthOverride
}

// ModernWidth returns a profile where East Asian Ambiguous characters occupy
// one cell
func ModernWidth() WidthProfile {
	return WidthProfile{}
}

// CJKWidth returns a profile where East Asian Ambiguous characters occupy two
// cells
func CJKWidth() WidthProfile {
	return WidthProfile{ambiguousWide: true}
}

// CustomWidth layers override over one of the base width profiles
func CustomWidth(base WidthProfile, override WidthOverride) WidthProfile {
	base.override = override
	return base
}

// GraphemeWidth returns the terminal cell width of input
//
// A single grapheme is the intended input. More than one grapheme returns the
// total width
func GraphemeWidth(input string, profile WidthProfile) int {
	input = NormalizeUTF8(input)
	graphemes := Graphemes(input)
	if len(graphemes) == 0 {
		return 0
	}
	total := 0
	for _, grapheme := range graphemes {
		total += clusterWidth(grapheme.Text, profile)
	}
	return total
}

func clusterWidth(grapheme string, profile WidthProfile) int {
	if profile.override != nil {
		if cells, overridden := profile.override(grapheme); overridden {
			if cells < 0 || cells > 2 {
				panic("text: custom grapheme width must be zero, one, or two")
			}
			return cells
		}
	}
	characters := []rune(grapheme)
	if len(characters) == 0 {
		return 0
	}
	for _, character := range characters {
		if graphemeProperty(character).isControl() {
			return 0
		}
	}
	if isRGIEmoji(characters) {
		return 2
	}

	textPresentation := false
	for index := 1; index < len(characters); index++ {
		if !isEmojiVariationBase(characters[index-1]) {
			continue
		}
		if characters[index] == 0xFE0F {
			return 2
		}
		if characters[index] == 0xFE0E {
			textPresentation = true
		}
	}
	if textPresentation {
		return 1
	}
	for _, character := range characters {
		if isEmojiPresentation(character) {
			return 2
		}
	}

	width := 0
	for _, character := range characters {
		property := graphemeProperty(character)
		if property == graphemeExtend || property == graphemeZWJ || property == graphemePrepend {
			continue
		}
		switch eastAsianWidthProperty(character) {
		case eastAsianWide:
			width = 2
		case eastAsianAmbiguous:
			if profile.ambiguousWide {
				width = 2
			} else {
				width = max(width, 1)
			}
		default:
			width = max(width, 1)
		}
	}
	return width
}
