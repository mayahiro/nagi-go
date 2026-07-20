package text

// Width returns the total terminal cell width of input
func Width(input string, profile WidthProfile) int {
	input = NormalizeUTF8(input)
	total := 0
	for _, grapheme := range Graphemes(input) {
		total += clusterWidth(grapheme.Text, profile)
	}
	return total
}

// Truncate returns the longest grapheme-aligned prefix within maxCells
func Truncate(input string, maxCells int, profile WidthProfile) string {
	input = NormalizeUTF8(input)
	if maxCells < 0 {
		maxCells = 0
	}
	cells := 0
	end := 0
	for _, grapheme := range Graphemes(input) {
		next := cells + clusterWidth(grapheme.Text, profile)
		if next > maxCells {
			break
		}
		cells = next
		end = grapheme.End
	}
	return input[:end]
}

// Wrap hard-wraps input without splitting extended grapheme clusters
//
// CR, LF, and CRLF force a line boundary and are omitted. A grapheme wider
// than maxCells occupies a line by itself, which guarantees progress when
// maxCells is zero or negative
func Wrap(input string, maxCells int, profile WidthProfile) []string {
	input = NormalizeUTF8(input)
	if input == "" {
		return []string{""}
	}
	if maxCells < 0 {
		maxCells = 0
	}
	var lines []string
	start := 0
	cells := 0
	for _, grapheme := range Graphemes(input) {
		if mandatoryBreak(grapheme.Text) {
			lines = append(lines, input[start:grapheme.Start])
			start = grapheme.End
			cells = 0
			continue
		}
		width := clusterWidth(grapheme.Text, profile)
		next := cells + width
		if width != 0 && grapheme.Start != start && next > maxCells {
			lines = append(lines, input[start:grapheme.Start])
			start = grapheme.Start
			cells = width
		} else {
			cells = next
		}
	}
	return append(lines, input[start:])
}

// CellAtByte converts an exact grapheme byte boundary to its terminal cell
// position
//
// The second result is false for offsets inside a grapheme or outside the
// normalized string
func CellAtByte(input string, byteOffset int, profile WidthProfile) (int, bool) {
	input = NormalizeUTF8(input)
	if byteOffset < 0 || byteOffset > len(input) {
		return 0, false
	}
	if byteOffset == 0 {
		return 0, true
	}
	cells := 0
	for _, grapheme := range Graphemes(input) {
		cells += clusterWidth(grapheme.Text, profile)
		if grapheme.End == byteOffset {
			return cells, true
		}
		if grapheme.End > byteOffset {
			return 0, false
		}
	}
	return 0, false
}

// ByteAtCell converts an exact terminal cell boundary to the earliest matching
// byte boundary
//
// The second result is false for positions inside a wide grapheme or beyond
// the normalized string
func ByteAtCell(input string, cellOffset int, profile WidthProfile) (int, bool) {
	input = NormalizeUTF8(input)
	if cellOffset < 0 {
		return 0, false
	}
	if cellOffset == 0 {
		return 0, true
	}
	cells := 0
	for _, grapheme := range Graphemes(input) {
		cells += clusterWidth(grapheme.Text, profile)
		if cells == cellOffset {
			return grapheme.End, true
		}
		if cells > cellOffset {
			return 0, false
		}
	}
	return 0, false
}

func mandatoryBreak(grapheme string) bool {
	return grapheme == "\r" || grapheme == "\n" || grapheme == "\r\n"
}
