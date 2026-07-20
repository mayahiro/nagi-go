package vt

import (
	"fmt"
	"strconv"
	"strings"
)

// SgrColorKind identifies an SGR color representation
type SgrColorKind uint8

const (
	// SgrColorDefault is the terminal default color
	SgrColorDefault SgrColorKind = iota
	// SgrColorIndexed is an indexed palette entry
	SgrColorIndexed
	// SgrColorRGB is a 24-bit RGB color
	SgrColorRGB
)

// SgrColor is a color used by the output encoder
//
// Its zero value is the terminal default color.
type SgrColor struct {
	kind             SgrColorKind
	index            uint8
	red, green, blue uint8
}

// IndexedSgrColor returns an indexed palette color
func IndexedSgrColor(index uint8) SgrColor {
	return SgrColor{kind: SgrColorIndexed, index: index}
}

// RGBSgrColor returns a 24-bit RGB color
func RGBSgrColor(red, green, blue uint8) SgrColor {
	return SgrColor{kind: SgrColorRGB, red: red, green: green, blue: blue}
}

// Kind returns the color representation
func (c SgrColor) Kind() SgrColorKind {
	return c.kind
}

// Index returns the palette index when this is an indexed color
func (c SgrColor) Index() (uint8, bool) {
	return c.index, c.kind == SgrColorIndexed
}

// RGB returns components when this is an RGB color
func (c SgrColor) RGB() (red, green, blue uint8, ok bool) {
	return c.red, c.green, c.blue, c.kind == SgrColorRGB
}

// OptionalSgrColor is an optional SGR color
type OptionalSgrColor struct {
	color SgrColor
	set   bool
}

// SomeSgrColor returns an optional color containing color
func SomeSgrColor(color SgrColor) OptionalSgrColor {
	return OptionalSgrColor{color: color, set: true}
}

// Get returns the color and whether it is present
func (c OptionalSgrColor) Get() (SgrColor, bool) {
	return c.color, c.set
}

// SgrStyle is a complete terminal SGR style
type SgrStyle struct {
	// Foreground is the text color
	Foreground SgrColor
	// Background is the cell background color
	Background SgrColor
	// UnderlineColor optionally overrides the underline color
	UnderlineColor OptionalSgrColor
	// Bold enables bold intensity
	Bold bool
	// Dim enables dim intensity
	Dim bool
	// Italic enables italic text
	Italic bool
	// Underline enables underlined text
	Underline bool
	// Blink enables blinking text
	Blink bool
	// Reverse swaps foreground and background
	Reverse bool
	// Hidden hides text
	Hidden bool
	// Strikethrough strikes through text
	Strikethrough bool
}

// Capabilities contains optional terminal encoder capabilities
type Capabilities struct {
	// TrueColor enables 24-bit SGR colors
	TrueColor bool
	// UnderlineColor enables SGR underline color
	UnderlineColor bool
	// CursorShape enables DECSCUSR cursor shape changes
	CursorShape bool
	// SynchronizedUpdates enables synchronized update mode
	SynchronizedUpdates bool
}

// BaselineCapabilities returns the safe xterm-compatible baseline
func BaselineCapabilities() Capabilities {
	return Capabilities{}
}

// ModernCapabilities returns all Phase 3 optional capabilities
func ModernCapabilities() Capabilities {
	return Capabilities{TrueColor: true, UnderlineColor: true, CursorShape: true, SynchronizedUpdates: true}
}

// EraseMode is an erasure range relative to the cursor
type EraseMode uint8

const (
	// EraseAfter erases from the cursor through the end
	EraseAfter EraseMode = iota
	// EraseBefore erases from the start through the cursor
	EraseBefore
	// EraseAll erases the complete line or display
	EraseAll
)

// CursorShape is a DECSCUSR cursor shape
type CursorShape uint8

const (
	// CursorDefault selects the terminal default
	CursorDefault CursorShape = iota
	// CursorBlinkingBlock selects a blinking block
	CursorBlinkingBlock
	// CursorSteadyBlock selects a steady block
	CursorSteadyBlock
	// CursorBlinkingUnderline selects a blinking underline
	CursorBlinkingUnderline
	// CursorSteadyUnderline selects a steady underline
	CursorSteadyUnderline
	// CursorBlinkingBar selects a blinking vertical bar
	CursorBlinkingBar
	// CursorSteadyBar selects a steady vertical bar
	CursorSteadyBar
)

// MouseTracking is an xterm mouse tracking policy
type MouseTracking uint8

const (
	// MouseTrackingPress reports button press and release
	MouseTrackingPress MouseTracking = iota
	// MouseTrackingButton adds movement while a button is held
	MouseTrackingButton
	// MouseTrackingAny reports all pointer movement
	MouseTrackingAny
)

type operationKind uint8

const (
	opMoveTo operationKind = iota
	opMoveRelative
	opSetStyle
	opResetStyle
	opWriteText
	opEraseLine
	opEraseDisplay
	opShowCursor
	opHideCursor
	opSetCursorShape
	opEnterAlternate
	opLeaveAlternate
	opEnablePaste
	opDisablePaste
	opEnableMouse
	opDisableMouse
	opEnableFocus
	opDisableFocus
	opBeginSync
	opEndSync
)

// TerminalOp is one validated pure terminal output operation
type TerminalOp struct {
	kind          operationKind
	x, y          uint32
	dx, dy        int32
	style         SgrStyle
	text          string
	erase         EraseMode
	cursorShape   CursorShape
	mouseTracking MouseTracking
}

// MoveTo returns an absolute zero-based cursor movement
func MoveTo(x, y uint32) TerminalOp {
	return TerminalOp{kind: opMoveTo, x: x, y: y}
}

// MoveRelative returns a signed relative cursor movement
func MoveRelative(dx, dy int32) TerminalOp {
	return TerminalOp{kind: opMoveRelative, dx: dx, dy: dy}
}

// SetStyle returns a complete SGR style operation
func SetStyle(style SgrStyle) TerminalOp {
	return TerminalOp{kind: opSetStyle, style: style}
}

// ResetStyle returns an SGR reset operation
func ResetStyle() TerminalOp {
	return TerminalOp{kind: opResetStyle}
}

// WriteText returns a safe text write operation
func WriteText(text string) TerminalOp {
	return TerminalOp{kind: opWriteText, text: text}
}

// EraseLine returns a line erasure operation
func EraseLine(mode EraseMode) TerminalOp {
	return TerminalOp{kind: opEraseLine, erase: mode}
}

// EraseDisplay returns a display erasure operation
func EraseDisplay(mode EraseMode) TerminalOp {
	return TerminalOp{kind: opEraseDisplay, erase: mode}
}

// ShowCursor returns a cursor visibility operation
func ShowCursor() TerminalOp {
	return TerminalOp{kind: opShowCursor}
}

// HideCursor returns a cursor visibility operation
func HideCursor() TerminalOp {
	return TerminalOp{kind: opHideCursor}
}

// SetCursorShape returns a cursor shape operation
func SetCursorShape(shape CursorShape) TerminalOp {
	return TerminalOp{kind: opSetCursorShape, cursorShape: shape}
}

// EnterAlternateScreen returns an alternate-screen entry operation
func EnterAlternateScreen() TerminalOp {
	return TerminalOp{kind: opEnterAlternate}
}

// LeaveAlternateScreen returns an alternate-screen exit operation
func LeaveAlternateScreen() TerminalOp {
	return TerminalOp{kind: opLeaveAlternate}
}

// EnableBracketedPaste returns a bracketed-paste enable operation
func EnableBracketedPaste() TerminalOp {
	return TerminalOp{kind: opEnablePaste}
}

// DisableBracketedPaste returns a bracketed-paste disable operation
func DisableBracketedPaste() TerminalOp {
	return TerminalOp{kind: opDisablePaste}
}

// EnableMouse returns an SGR mouse reporting operation
func EnableMouse(tracking MouseTracking) TerminalOp {
	return TerminalOp{kind: opEnableMouse, mouseTracking: tracking}
}

// DisableMouse returns an operation disabling known mouse modes
func DisableMouse() TerminalOp {
	return TerminalOp{kind: opDisableMouse}
}

// EnableFocus returns a focus-reporting enable operation
func EnableFocus() TerminalOp {
	return TerminalOp{kind: opEnableFocus}
}

// DisableFocus returns a focus-reporting disable operation
func DisableFocus() TerminalOp {
	return TerminalOp{kind: opDisableFocus}
}

// BeginSynchronizedUpdate returns a synchronized-update begin operation
func BeginSynchronizedUpdate() TerminalOp {
	return TerminalOp{kind: opBeginSync}
}

// EndSynchronizedUpdate returns a synchronized-update end operation
func EndSynchronizedUpdate() TerminalOp {
	return TerminalOp{kind: opEndSync}
}

// Encode encodes terminal operations deterministically for capabilities
func Encode(operations []TerminalOp, capabilities Capabilities) []byte {
	var output strings.Builder
	for _, operation := range operations {
		encodeOperation(&output, operation, capabilities)
	}
	return []byte(output.String())
}

func encodeOperation(output *strings.Builder, operation TerminalOp, capabilities Capabilities) {
	switch operation.kind {
	case opMoveTo:
		fmt.Fprintf(output, "\x1B[%d;%dH", uint64(operation.y)+1, uint64(operation.x)+1)
	case opMoveRelative:
		if operation.dy < 0 {
			writeCSICount(output, uint32(-int64(operation.dy)), 'A')
		} else if operation.dy > 0 {
			writeCSICount(output, uint32(operation.dy), 'B')
		}
		if operation.dx > 0 {
			writeCSICount(output, uint32(operation.dx), 'C')
		} else if operation.dx < 0 {
			writeCSICount(output, uint32(-int64(operation.dx)), 'D')
		}
	case opSetStyle:
		writeStyle(output, operation.style, capabilities)
	case opResetStyle:
		output.WriteString("\x1B[0m")
	case opWriteText:
		writeSafeText(output, operation.text)
	case opEraseLine:
		writeErase(output, operation.erase, 'K')
	case opEraseDisplay:
		writeErase(output, operation.erase, 'J')
	case opShowCursor:
		output.WriteString("\x1B[?25h")
	case opHideCursor:
		output.WriteString("\x1B[?25l")
	case opSetCursorShape:
		if capabilities.CursorShape && operation.cursorShape <= CursorSteadyBar {
			fmt.Fprintf(output, "\x1B[%d q", uint8(operation.cursorShape))
		}
	case opEnterAlternate:
		output.WriteString("\x1B[?1049h")
	case opLeaveAlternate:
		output.WriteString("\x1B[?1049l")
	case opEnablePaste:
		output.WriteString("\x1B[?2004h")
	case opDisablePaste:
		output.WriteString("\x1B[?2004l")
	case opEnableMouse:
		mode := 0
		switch operation.mouseTracking {
		case MouseTrackingPress:
			mode = 1000
		case MouseTrackingButton:
			mode = 1002
		case MouseTrackingAny:
			mode = 1003
		}
		if mode != 0 {
			fmt.Fprintf(output, "\x1B[?%dh\x1B[?1006h", mode)
		}
	case opDisableMouse:
		output.WriteString("\x1B[?1000l\x1B[?1002l\x1B[?1003l\x1B[?1006l")
	case opEnableFocus:
		output.WriteString("\x1B[?1004h")
	case opDisableFocus:
		output.WriteString("\x1B[?1004l")
	case opBeginSync:
		if capabilities.SynchronizedUpdates {
			output.WriteString("\x1B[?2026h")
		}
	case opEndSync:
		if capabilities.SynchronizedUpdates {
			output.WriteString("\x1B[?2026l")
		}
	}
}

func writeCSICount(output *strings.Builder, count uint32, finalCharacter byte) {
	fmt.Fprintf(output, "\x1B[%d%c", count, finalCharacter)
}

func writeErase(output *strings.Builder, mode EraseMode, finalCharacter byte) {
	fmt.Fprintf(output, "\x1B[%d%c", uint8(mode), finalCharacter)
}

func writeStyle(output *strings.Builder, style SgrStyle, capabilities Capabilities) {
	parameters := []string{"0"}
	pushColor(&parameters, style.Foreground, 38, capabilities)
	pushColor(&parameters, style.Background, 48, capabilities)
	if capabilities.UnderlineColor {
		if color, ok := style.UnderlineColor.Get(); ok {
			pushExtendedColor(&parameters, color, 58, capabilities)
		}
	}
	attributes := []struct {
		enabled bool
		code    string
	}{
		{style.Bold, "1"}, {style.Dim, "2"}, {style.Italic, "3"}, {style.Underline, "4"},
		{style.Blink, "5"}, {style.Reverse, "7"}, {style.Hidden, "8"}, {style.Strikethrough, "9"},
	}
	for _, attribute := range attributes {
		if attribute.enabled {
			parameters = append(parameters, attribute.code)
		}
	}
	output.WriteString("\x1B[")
	output.WriteString(strings.Join(parameters, ";"))
	output.WriteByte('m')
}

func pushColor(parameters *[]string, color SgrColor, prefix int, capabilities Capabilities) {
	if color.Kind() == SgrColorDefault {
		return
	}
	pushExtendedColor(parameters, color, prefix, capabilities)
}

func pushExtendedColor(parameters *[]string, color SgrColor, prefix int, capabilities Capabilities) {
	switch color.Kind() {
	case SgrColorDefault:
		reset := 39
		if prefix == 48 {
			reset = 49
		} else if prefix == 58 {
			reset = 59
		}
		*parameters = append(*parameters, strconv.Itoa(reset))
	case SgrColorIndexed:
		index, _ := color.Index()
		*parameters = append(*parameters, strconv.Itoa(prefix), "5", strconv.Itoa(int(index)))
	case SgrColorRGB:
		red, green, blue, _ := color.RGB()
		if capabilities.TrueColor {
			*parameters = append(*parameters, strconv.Itoa(prefix), "2", strconv.Itoa(int(red)), strconv.Itoa(int(green)), strconv.Itoa(int(blue)))
		} else {
			*parameters = append(*parameters, strconv.Itoa(prefix), "5", strconv.Itoa(int(indexedRGB(red, green, blue))))
		}
	}
}

func indexedRGB(red, green, blue uint8) uint8 {
	level := func(component uint8) uint8 {
		return uint8((uint16(component)*5 + 127) / 255)
	}
	return 16 + 36*level(red) + 6*level(green) + level(blue)
}

func writeSafeText(output *strings.Builder, input string) {
	input = strings.ToValidUTF8(input, "\uFFFD")
	for _, character := range input {
		if character <= 0x1F || (character >= 0x7F && character <= 0x9F) {
			output.WriteRune('\uFFFD')
		} else {
			output.WriteRune(character)
		}
	}
}
