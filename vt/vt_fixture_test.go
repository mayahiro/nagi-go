package vt

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/mayahiro/nagi-go/internal/conformance"
)

func TestInputFixturesAndEverySingleSplit(t *testing.T) {
	records := vtRecords(t, "vt/input.txt", "vt-input", "input", "expected")
	for _, record := range records {
		input := record.Bytes("input")
		whole := decodeChunks([][]byte{input})
		if got, want := canonicalEvents(whole), record.Field("expected"); got != want {
			t.Errorf("case %s whole: events = %s, want %s", record.ID, got, want)
		}

		byteChunks := make([][]byte, len(input))
		for index := range input {
			byteChunks[index] = input[index : index+1]
		}
		if got := decodeChunks(byteChunks); !reflect.DeepEqual(got, whole) {
			t.Errorf("case %s byte chunks differ: %#v, want %#v", record.ID, got, whole)
		}
		for split := 0; split <= len(input); split++ {
			got := decodeChunks([][]byte{input[:split], input[split:]})
			if !reflect.DeepEqual(got, whole) {
				t.Errorf("case %s split %d differs: %#v, want %#v", record.ID, split, got, whole)
			}
		}
	}
}

func TestOutputFixtures(t *testing.T) {
	records := vtRecords(t, "vt/output.txt", "vt-output", "capabilities", "operations", "expected")
	for _, record := range records {
		var capabilities Capabilities
		switch record.Field("capabilities") {
		case "modern":
			capabilities = ModernCapabilities()
		case "baseline":
			capabilities = BaselineCapabilities()
		default:
			t.Fatalf("case %s has unknown capabilities", record.ID)
		}
		got := Encode(fixtureOperations(record.Field("operations")), capabilities)
		want := record.Bytes("expected")
		if !bytes.Equal(got, want) {
			t.Errorf("case %s: Encode() = %q, want %q", record.ID, got, want)
		}
	}
}

func vtRecords(t *testing.T, path, suite string, fields ...string) []conformance.Record {
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

func decodeChunks(chunks [][]byte) []Event {
	decoder := NewDecoder()
	var events []Event
	for _, chunk := range chunks {
		events = append(events, decoder.Feed(chunk)...)
	}
	return append(events, decoder.FlushPending()...)
}

func canonicalEvents(events []Event) string {
	values := make([]string, len(events))
	for index, event := range events {
		values[index] = canonicalEvent(event)
	}
	return strings.Join(values, "|")
}

func canonicalEvent(event Event) string {
	switch event.Kind {
	case EventText:
		return "text:" + scalarText(event.Text)
	case EventPaste:
		return "paste:" + scalarText(event.Text)
	case EventKey:
		text := "-"
		if event.Key.HasText {
			text = scalarText(event.Key.Text)
		}
		return fmt.Sprintf("key:%s:%s:%s:%s:%s",
			keyCodeText(event.Key), modifierText(event.Key.Modifiers), text,
			keyActionText(event.Key.Action), keyProtocolText(event.Key.Protocol))
	case EventMouse:
		return fmt.Sprintf("mouse:%s:%s:%d,%d:%s",
			mouseKindText(event.Mouse.Kind), mouseButtonText(event.Mouse),
			event.Mouse.X, event.Mouse.Y, modifierText(event.Mouse.Modifiers))
	case EventFocusIn:
		return "focus-in"
	case EventFocusOut:
		return "focus-out"
	case EventTerminalResponse:
		return fmt.Sprintf("response:%X", event.Bytes)
	case EventUnknownSequence:
		return fmt.Sprintf("unknown:%X", event.Bytes)
	default:
		panic("unknown event kind")
	}
}

func keyCodeText(key KeyEvent) string {
	switch key.Code {
	case KeyCharacter:
		return fmt.Sprintf("char-U+%04X", key.Character)
	case KeyEnter:
		return "enter"
	case KeyTab:
		return "tab"
	case KeyBackspace:
		return "backspace"
	case KeyEscape:
		return "escape"
	case KeyUp:
		return "up"
	case KeyDown:
		return "down"
	case KeyRight:
		return "right"
	case KeyLeft:
		return "left"
	case KeyHome:
		return "home"
	case KeyEnd:
		return "end"
	case KeyInsert:
		return "insert"
	case KeyDelete:
		return "delete"
	case KeyPageUp:
		return "page-up"
	case KeyPageDown:
		return "page-down"
	case KeyFunction:
		return fmt.Sprintf("f%d", key.Function)
	default:
		return "unknown"
	}
}

func keyActionText(action KeyAction) string {
	switch action {
	case KeyPress:
		return "press"
	case KeyRepeat:
		return "repeat"
	case KeyRelease:
		return "release"
	default:
		return "unknown"
	}
}

func keyProtocolText(protocol KeyProtocol) string {
	if protocol == KeyProtocolLegacy {
		return "legacy"
	}
	return "unknown"
}

func mouseKindText(kind MouseKind) string {
	switch kind {
	case MousePress:
		return "press"
	case MouseRelease:
		return "release"
	case MouseMove:
		return "move"
	default:
		return "scroll"
	}
}

func mouseButtonText(mouse MouseEvent) string {
	switch mouse.Button {
	case MouseLeft:
		return "left"
	case MouseMiddle:
		return "middle"
	case MouseRight:
		return "right"
	case MouseWheelUp:
		return "wheel-up"
	case MouseWheelDown:
		return "wheel-down"
	case MouseWheelLeft:
		return "wheel-left"
	case MouseWheelRight:
		return "wheel-right"
	case MouseOther:
		return fmt.Sprintf("other-%d", mouse.OtherButton)
	default:
		return "none"
	}
}

func modifierText(modifiers Modifiers) string {
	var names []string
	if modifiers.Shift {
		names = append(names, "shift")
	}
	if modifiers.Alt {
		names = append(names, "alt")
	}
	if modifiers.Control {
		names = append(names, "control")
	}
	if modifiers.Meta {
		names = append(names, "meta")
	}
	if len(names) == 0 {
		return "-"
	}
	return strings.Join(names, "+")
}

func scalarText(text string) string {
	if text == "" {
		return "-"
	}
	var values []string
	for _, character := range text {
		values = append(values, fmt.Sprintf("U+%04X", character))
	}
	return strings.Join(values, "+")
}

func fixtureOperations(value string) []TerminalOp {
	rawOperations := strings.Split(value, ";")
	operations := make([]TerminalOp, len(rawOperations))
	for index, operation := range rawOperations {
		fields := strings.Split(operation, ",")
		switch fields[0] {
		case "move-to":
			operations[index] = MoveTo(fixtureUnsigned(fields[1]), fixtureUnsigned(fields[2]))
		case "move-relative":
			operations[index] = MoveRelative(fixtureSigned(fields[1]), fixtureSigned(fields[2]))
		case "set-style":
			operations[index] = SetStyle(fixtureStyle(fields[1]))
		case "reset-style":
			operations[index] = ResetStyle()
		case "write":
			operations[index] = WriteText(fixtureScalarText(fields[1]))
		case "erase-line":
			operations[index] = EraseLine(fixtureEraseMode(fields[1]))
		case "erase-display":
			operations[index] = EraseDisplay(fixtureEraseMode(fields[1]))
		case "show-cursor":
			operations[index] = ShowCursor()
		case "hide-cursor":
			operations[index] = HideCursor()
		case "cursor-shape":
			operations[index] = SetCursorShape(fixtureCursorShape(fields[1]))
		case "enter-alternate":
			operations[index] = EnterAlternateScreen()
		case "leave-alternate":
			operations[index] = LeaveAlternateScreen()
		case "enable-paste":
			operations[index] = EnableBracketedPaste()
		case "disable-paste":
			operations[index] = DisableBracketedPaste()
		case "enable-mouse":
			operations[index] = EnableMouse(fixtureMouseTracking(fields[1]))
		case "disable-mouse":
			operations[index] = DisableMouse()
		case "enable-focus":
			operations[index] = EnableFocus()
		case "disable-focus":
			operations[index] = DisableFocus()
		case "set-clipboard":
			if fields[1] == "-" {
				operations[index] = SetClipboard("")
			} else {
				operations[index] = SetClipboard(fixtureScalarText(fields[1]))
			}
		case "begin-sync":
			operations[index] = BeginSynchronizedUpdate()
		case "end-sync":
			operations[index] = EndSynchronizedUpdate()
		default:
			panic("invalid terminal operation " + operation)
		}
	}
	return operations
}

func fixtureStyle(value string) SgrStyle {
	var style SgrStyle
	if value == "-" {
		return style
	}
	for _, token := range strings.Split(value, "+") {
		switch {
		case strings.HasPrefix(token, "fg-"):
			style.Foreground = fixtureColor(strings.TrimPrefix(token, "fg-"))
		case strings.HasPrefix(token, "bg-"):
			style.Background = fixtureColor(strings.TrimPrefix(token, "bg-"))
		case strings.HasPrefix(token, "underline-color-"):
			style.UnderlineColor = SomeSgrColor(fixtureColor(strings.TrimPrefix(token, "underline-color-")))
		default:
			switch token {
			case "bold":
				style.Bold = true
			case "dim":
				style.Dim = true
			case "italic":
				style.Italic = true
			case "underline":
				style.Underline = true
			case "blink":
				style.Blink = true
			case "reverse":
				style.Reverse = true
			case "hidden":
				style.Hidden = true
			case "strikethrough":
				style.Strikethrough = true
			default:
				panic("unknown fixture style " + token)
			}
		}
	}
	return style
}

func fixtureColor(value string) SgrColor {
	if value == "default" {
		return SgrColor{}
	}
	if index, ok := strings.CutPrefix(value, "indexed-"); ok {
		number, err := strconv.ParseUint(index, 10, 8)
		if err != nil {
			panic(fmt.Sprintf("invalid color index %s: %v", index, err))
		}
		return IndexedSgrColor(uint8(number))
	}
	if rgb, ok := strings.CutPrefix(value, "rgb-"); ok {
		if len(rgb) != 6 {
			panic("invalid RGB color " + rgb)
		}
		return RGBSgrColor(uint8(fixtureHex(rgb[:2])), uint8(fixtureHex(rgb[2:4])), uint8(fixtureHex(rgb[4:])))
	}
	panic("unknown fixture color " + value)
}

func fixtureScalarText(value string) string {
	var output strings.Builder
	for _, scalar := range strings.Split(value, "+") {
		character := rune(fixtureHex(scalar))
		if !utf8.ValidRune(character) {
			panic("invalid Unicode scalar " + scalar)
		}
		output.WriteRune(character)
	}
	return output.String()
}

func fixtureEraseMode(value string) EraseMode {
	switch value {
	case "after":
		return EraseAfter
	case "before":
		return EraseBefore
	case "all":
		return EraseAll
	default:
		panic("unknown erase mode " + value)
	}
}

func fixtureCursorShape(value string) CursorShape {
	switch value {
	case "default":
		return CursorDefault
	case "blinking-block":
		return CursorBlinkingBlock
	case "steady-block":
		return CursorSteadyBlock
	case "blinking-underline":
		return CursorBlinkingUnderline
	case "steady-underline":
		return CursorSteadyUnderline
	case "blinking-bar":
		return CursorBlinkingBar
	case "steady-bar":
		return CursorSteadyBar
	default:
		panic("unknown cursor shape " + value)
	}
}

func fixtureMouseTracking(value string) MouseTracking {
	switch value {
	case "press":
		return MouseTrackingPress
	case "button":
		return MouseTrackingButton
	case "any":
		return MouseTrackingAny
	default:
		panic("unknown mouse tracking " + value)
	}
}

func fixtureUnsigned(value string) uint32 {
	number, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		panic(fmt.Sprintf("invalid unsigned integer %s: %v", value, err))
	}
	return uint32(number)
}

func fixtureSigned(value string) int32 {
	number, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		panic(fmt.Sprintf("invalid signed integer %s: %v", value, err))
	}
	return int32(number)
}

func fixtureHex(value string) uint32 {
	number, err := strconv.ParseUint(value, 16, 32)
	if err != nil {
		panic(fmt.Sprintf("invalid hexadecimal integer %s: %v", value, err))
	}
	return uint32(number)
}
