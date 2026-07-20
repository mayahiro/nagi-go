package text

// WrappedLine is one hard-wrapped line and its terminal cell width
type WrappedLine struct {
	// Text is the line content without a mandatory line-break grapheme
	Text string
	// Width is the line width in terminal cells
	Width int
}

// WrappedLines iterates hard-wrapped lines without materializing a result
// slice
type WrappedLines struct {
	input     string
	graphemes GraphemeIterator
	profile   WidthProfile
	maxCells  int
	start     int
	cells     int
	finished  bool
}

// IterateWrappedLines returns an iterator using the same rules as Wrap
func IterateWrappedLines(input string, maxCells int, profile WidthProfile) WrappedLines {
	if maxCells < 0 {
		maxCells = 0
	}
	graphemes := IterateGraphemes(input)
	return WrappedLines{
		input:     graphemes.input,
		graphemes: graphemes,
		profile:   profile,
		maxCells:  maxCells,
	}
}

// Next returns the next wrapped line and whether one was available
func (i *WrappedLines) Next() (WrappedLine, bool) {
	if i.finished {
		return WrappedLine{}, false
	}
	for grapheme, ok := i.graphemes.Next(); ok; grapheme, ok = i.graphemes.Next() {
		if mandatoryBreak(grapheme.Text) {
			line := WrappedLine{Text: i.input[i.start:grapheme.Start], Width: i.cells}
			i.start = grapheme.End
			i.cells = 0
			return line, true
		}
		width := clusterWidth(grapheme.Text, i.profile)
		next := i.cells + width
		if width != 0 && grapheme.Start != i.start && next > i.maxCells {
			line := WrappedLine{Text: i.input[i.start:grapheme.Start], Width: i.cells}
			i.start = grapheme.Start
			i.cells = width
			return line, true
		}
		i.cells = next
	}
	i.finished = true
	return WrappedLine{Text: i.input[i.start:], Width: i.cells}, true
}

// Width returns the total terminal cell width of input
func Width(input string, profile WidthProfile) int {
	iterator := IterateGraphemes(input)
	total := 0
	for grapheme, ok := iterator.Next(); ok; grapheme, ok = iterator.Next() {
		total += clusterWidth(grapheme.Text, profile)
	}
	return total
}

// Truncate returns the longest grapheme-aligned prefix within maxCells
func Truncate(input string, maxCells int, profile WidthProfile) string {
	iterator := IterateGraphemes(input)
	if maxCells < 0 {
		maxCells = 0
	}
	cells := 0
	end := 0
	for grapheme, ok := iterator.Next(); ok; grapheme, ok = iterator.Next() {
		next := cells + clusterWidth(grapheme.Text, profile)
		if next > maxCells {
			break
		}
		cells = next
		end = grapheme.End
	}
	return iterator.input[:end]
}

// Wrap hard-wraps input without splitting extended grapheme clusters
//
// CR, LF, and CRLF force a line boundary and are omitted. A grapheme wider
// than maxCells occupies a line by itself, which guarantees progress when
// maxCells is zero or negative
func Wrap(input string, maxCells int, profile WidthProfile) []string {
	iterator := IterateWrappedLines(input, maxCells, profile)
	var lines []string
	for line, ok := iterator.Next(); ok; line, ok = iterator.Next() {
		lines = append(lines, line.Text)
	}
	return lines
}

// CellAtByte converts an exact grapheme byte boundary to its terminal cell
// position
//
// The second result is false for offsets inside a grapheme or outside the
// normalized string
func CellAtByte(input string, byteOffset int, profile WidthProfile) (int, bool) {
	iterator := IterateGraphemes(input)
	if byteOffset < 0 || byteOffset > len(iterator.input) {
		return 0, false
	}
	if byteOffset == 0 {
		return 0, true
	}
	cells := 0
	for grapheme, ok := iterator.Next(); ok; grapheme, ok = iterator.Next() {
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
	iterator := IterateGraphemes(input)
	if cellOffset < 0 {
		return 0, false
	}
	if cellOffset == 0 {
		return 0, true
	}
	cells := 0
	for grapheme, ok := iterator.Next(); ok; grapheme, ok = iterator.Next() {
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
