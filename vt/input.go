package vt

import (
	"bytes"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	escapeByte = 0x1B
	// MaxSequenceBytes is the maximum buffered CSI, SS3, or control-string size
	MaxSequenceBytes = 4_096
	// MaxPasteBytes is the maximum buffered bracketed paste payload size
	MaxPasteBytes = 1_048_576
)

var (
	pasteStart = []byte("\x1B[200~")
	pasteEnd   = []byte("\x1B[201~")
)

type decoderState uint8

const (
	stateGround decoderState = iota
	stateEscape
	stateCSI
	stateSS3
	stateAltUTF8
	stateControlString
	statePaste
	stateDiscardPaste
)

type controlKind uint8

const (
	controlOther controlKind = iota
	controlOSC
)

// Decoder is a pure streaming VT input decoder
//
// Ordinary text produces one EventText per Unicode scalar so event boundaries
// do not depend on input byte chunks. Call FlushPending when an external ESC
// deadline expires or when no more input will arrive.
type Decoder struct {
	state          decoderState
	sequence       []byte
	controlKind    controlKind
	sawEscape      bool
	utf8Pending    []byte
	invalidTextRun bool
}

// NewDecoder returns a decoder in the ground state
func NewDecoder() *Decoder {
	return &Decoder{utf8Pending: make([]byte, 0, utf8.UTFMax)}
}

// Feed consumes one arbitrary byte chunk and returns all completed events
func (d *Decoder) Feed(input []byte) []Event {
	var events []Event
	for _, value := range input {
		d.process(value, &events)
	}
	return events
}

// HasPending reports whether incomplete UTF-8 or VT input is buffered
func (d *Decoder) HasPending() bool {
	return d.state != stateGround || len(d.utf8Pending) != 0 || d.invalidTextRun
}

// HasPendingEscape reports whether the only VT prefix is a lone ESC byte
func (d *Decoder) HasPendingEscape() bool {
	return d.state == stateEscape
}

// FlushPending resolves all currently incomplete input without waiting
//
// A lone ESC becomes an Escape key, incomplete UTF-8 becomes one U+FFFD, and
// any other incomplete VT sequence becomes EventUnknownSequence.
func (d *Decoder) FlushPending() []Event {
	var events []Event
	d.finishText(&events)
	switch d.state {
	case stateGround:
	case stateEscape:
		events = append(events, escapeEvent())
	case statePaste, stateDiscardPaste:
		events = append(events, unknownEvent(pasteStart))
	default:
		events = append(events, unknownEvent(d.sequence))
	}
	d.state = stateGround
	d.sequence = nil
	d.sawEscape = false
	return events
}

func (d *Decoder) process(value byte, events *[]Event) {
	switch d.state {
	case stateGround:
		d.processGround(value, events)
	case stateEscape:
		d.processEscape(value, events)
	case stateCSI:
		d.sequence = append(d.sequence, value)
		switch {
		case len(d.sequence) > MaxSequenceBytes:
			*events = append(*events, unknownEvent(d.sequence))
			d.resetState()
		case value >= 0x40 && value <= 0x7E:
			if bytes.Equal(d.sequence, pasteStart) {
				d.state = statePaste
				d.sequence = nil
			} else {
				*events = append(*events, parseCSI(d.sequence))
				d.resetState()
			}
		case value < 0x20 || value > 0x3F:
			*events = append(*events, unknownEvent(d.sequence))
			d.resetState()
		}
	case stateSS3:
		d.sequence = append(d.sequence, value)
		switch {
		case value >= 0x40 && value <= 0x7E:
			*events = append(*events, parseSS3(d.sequence))
			d.resetState()
		case len(d.sequence) > MaxSequenceBytes || value < 0x20 || value > 0x3F:
			*events = append(*events, unknownEvent(d.sequence))
			d.resetState()
		}
	case stateAltUTF8:
		d.sequence = append(d.sequence, value)
		payload := d.sequence[1:]
		expected := utf8Expected(payload[0])
		switch {
		case expected == 0 || !validUTF8Prefix(payload):
			*events = append(*events, unknownEvent(d.sequence))
			d.resetState()
		case len(payload) == expected:
			character, _ := utf8.DecodeRune(payload)
			*events = append(*events, altCharacter(character))
			d.resetState()
		}
	case stateControlString:
		d.sequence = append(d.sequence, value)
		terminated := (d.controlKind == controlOSC && value == 0x07) || (d.sawEscape && value == '\\')
		switch {
		case terminated:
			*events = append(*events, responseEvent(d.sequence))
			d.resetState()
		case len(d.sequence) > MaxSequenceBytes:
			*events = append(*events, unknownEvent(d.sequence))
			d.resetState()
		default:
			d.sawEscape = value == escapeByte
		}
	case statePaste:
		d.sequence = append(d.sequence, value)
		switch {
		case bytes.HasSuffix(d.sequence, pasteEnd):
			payload := d.sequence[:len(d.sequence)-len(pasteEnd)]
			if len(payload) > MaxPasteBytes {
				*events = append(*events, unknownEvent(pasteStart))
			} else {
				*events = append(*events, Event{Kind: EventPaste, Text: normalizeUTF8(payload)})
			}
			d.resetState()
		case len(d.sequence) > MaxPasteBytes+len(pasteEnd)-1:
			keep := len(pasteEnd) - 1
			tail := append([]byte(nil), d.sequence[len(d.sequence)-keep:]...)
			*events = append(*events, unknownEvent(pasteStart))
			d.state = stateDiscardPaste
			d.sequence = tail
		}
	case stateDiscardPaste:
		d.sequence = append(d.sequence, value)
		if bytes.HasSuffix(d.sequence, pasteEnd) {
			d.resetState()
		} else if len(d.sequence) >= len(pasteEnd) {
			d.sequence = append(d.sequence[:0], d.sequence[1:]...)
		}
	}
}

func (d *Decoder) processGround(value byte, events *[]Event) {
	switch {
	case value == escapeByte:
		d.finishText(events)
		d.state = stateEscape
	case value < 0x20 || value == 0x7F:
		d.finishText(events)
		*events = append(*events, controlEvent(value))
	default:
		d.processTextByte(value, events)
	}
}

func (d *Decoder) processEscape(value byte, events *[]Event) {
	switch value {
	case '[':
		d.state = stateCSI
		d.sequence = []byte{escapeByte, '['}
	case 'O':
		d.state = stateSS3
		d.sequence = []byte{escapeByte, 'O'}
	case ']':
		d.state = stateControlString
		d.controlKind = controlOSC
		d.sequence = []byte{escapeByte, ']'}
	case 'P', '^', '_':
		d.state = stateControlString
		d.controlKind = controlOther
		d.sequence = []byte{escapeByte, value}
	case escapeByte:
		*events = append(*events, escapeEvent())
		d.state = stateEscape
	case 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
		0x18, 0x19, 0x1A, 0x1C, 0x1D, 0x1E, 0x1F, 0x7F:
		*events = append(*events, escapeEvent())
		d.resetState()
		d.processGround(value, events)
	default:
		switch {
		case value >= 0x20 && value <= 0x7E:
			*events = append(*events, altCharacter(rune(value)))
			d.resetState()
		case utf8Expected(value) != 0:
			d.state = stateAltUTF8
			d.sequence = []byte{escapeByte, value}
		default:
			*events = append(*events, unknownEvent([]byte{escapeByte, value}))
			d.resetState()
		}
	}
}

func (d *Decoder) processTextByte(value byte, events *[]Event) {
	if value < utf8.RuneSelf {
		if len(d.utf8Pending) != 0 {
			d.invalidTextRun = true
			d.utf8Pending = d.utf8Pending[:0]
		}
		d.flushInvalidText(events)
		*events = append(*events, Event{Kind: EventText, Text: string(rune(value))})
		return
	}
	if len(d.utf8Pending) == 0 {
		if utf8Expected(value) != 0 {
			d.utf8Pending = append(d.utf8Pending, value)
		} else {
			d.invalidTextRun = true
		}
		return
	}
	d.utf8Pending = append(d.utf8Pending, value)
	if !validUTF8Prefix(d.utf8Pending) {
		d.invalidTextRun = true
		d.utf8Pending = d.utf8Pending[:0]
		d.processTextByte(value, events)
		return
	}
	expected := utf8Expected(d.utf8Pending[0])
	if len(d.utf8Pending) != expected {
		return
	}
	character, _ := utf8.DecodeRune(d.utf8Pending)
	d.utf8Pending = d.utf8Pending[:0]
	d.flushInvalidText(events)
	*events = append(*events, Event{Kind: EventText, Text: string(character)})
}

func (d *Decoder) finishText(events *[]Event) {
	if len(d.utf8Pending) != 0 {
		d.invalidTextRun = true
		d.utf8Pending = d.utf8Pending[:0]
	}
	d.flushInvalidText(events)
}

func (d *Decoder) flushInvalidText(events *[]Event) {
	if d.invalidTextRun {
		*events = append(*events, Event{Kind: EventText, Text: "\uFFFD"})
		d.invalidTextRun = false
	}
}

func (d *Decoder) resetState() {
	d.state = stateGround
	d.sequence = nil
	d.controlKind = controlOther
	d.sawEscape = false
}

func utf8Expected(lead byte) int {
	switch {
	case lead >= 0xC2 && lead <= 0xDF:
		return 2
	case lead >= 0xE0 && lead <= 0xEF:
		return 3
	case lead >= 0xF0 && lead <= 0xF4:
		return 4
	default:
		return 0
	}
}

func validUTF8Prefix(input []byte) bool {
	expected := utf8Expected(input[0])
	if expected == 0 || len(input) > expected {
		return false
	}
	for index := 1; index < len(input); index++ {
		value := input[index]
		if value < 0x80 || value > 0xBF {
			return false
		}
		if index == 1 {
			switch input[0] {
			case 0xE0:
				if value < 0xA0 {
					return false
				}
			case 0xED:
				if value > 0x9F {
					return false
				}
			case 0xF0:
				if value < 0x90 {
					return false
				}
			case 0xF4:
				if value > 0x8F {
					return false
				}
			}
		}
	}
	return true
}

func parseCSI(sequence []byte) Event {
	finalByte := sequence[len(sequence)-1]
	body := sequence[2 : len(sequence)-1]
	if len(body) != 0 && body[0] == '<' && (finalByte == 'M' || finalByte == 'm') {
		if mouse, ok := parseSGRMouse(body[1:], finalByte); ok {
			return Event{Kind: EventMouse, Mouse: mouse}
		}
		return unknownEvent(sequence)
	}
	if len(body) == 0 {
		if code, ok := simpleCSIKey(finalByte); ok {
			return legacyKey(code, 0, 0, Modifiers{})
		}
		switch finalByte {
		case 'Z':
			return legacyKey(KeyTab, 0, 0, Modifiers{Shift: true})
		case 'I':
			return Event{Kind: EventFocusIn}
		case 'O':
			return Event{Kind: EventFocusOut}
		}
	}
	parameters, validParameters := parseParameters(body)
	if code, ok := simpleCSIKey(finalByte); ok && validParameters {
		if modifiers, valid := keyModifiers(parameters); valid {
			return legacyKey(code, 0, 0, modifiers)
		}
	}
	if finalByte == '~' && validParameters && len(parameters) != 0 {
		if code, function, ok := tildeKey(parameters[0]); ok {
			var modifiers Modifiers
			valid := len(parameters) == 1
			if len(parameters) == 2 {
				modifiers, valid = xtermModifiers(parameters[1])
			}
			if valid {
				return legacyKey(code, 0, function, modifiers)
			}
		}
	}
	if finalByte == 'R' && validParameters && len(parameters) == 2 && parameters[0] != 0 && parameters[1] != 0 {
		return responseEvent(sequence)
	}
	if finalByte == 'c' || finalByte == 'n' || finalByte == 't' {
		return responseEvent(sequence)
	}
	return unknownEvent(sequence)
}

func parseSS3(sequence []byte) Event {
	finalByte := sequence[len(sequence)-1]
	if code, ok := simpleCSIKey(finalByte); ok {
		return legacyKey(code, 0, 0, Modifiers{})
	}
	if finalByte >= 'P' && finalByte <= 'S' {
		return legacyKey(KeyFunction, 0, uint8(finalByte-'P'+1), Modifiers{})
	}
	return unknownEvent(sequence)
}

func simpleCSIKey(finalByte byte) (KeyCode, bool) {
	switch finalByte {
	case 'A':
		return KeyUp, true
	case 'B':
		return KeyDown, true
	case 'C':
		return KeyRight, true
	case 'D':
		return KeyLeft, true
	case 'H':
		return KeyHome, true
	case 'F':
		return KeyEnd, true
	default:
		return KeyUnknown, false
	}
}

func parseParameters(body []byte) ([]uint32, bool) {
	if len(body) == 0 {
		return nil, true
	}
	parts := strings.Split(string(body), ";")
	parameters := make([]uint32, len(parts))
	for index, part := range parts {
		if part == "" {
			continue
		}
		value, err := strconv.ParseUint(part, 10, 32)
		if err != nil {
			return nil, false
		}
		parameters[index] = uint32(value)
	}
	return parameters, true
}

func keyModifiers(parameters []uint32) (Modifiers, bool) {
	switch len(parameters) {
	case 0, 1:
		return Modifiers{}, true
	case 2:
		return xtermModifiers(parameters[1])
	default:
		return Modifiers{}, false
	}
}

func xtermModifiers(value uint32) (Modifiers, bool) {
	if value < 1 || value > 16 {
		return Modifiers{}, false
	}
	bits := value - 1
	return Modifiers{
		Shift: bits&1 != 0, Alt: bits&2 != 0, Control: bits&4 != 0, Meta: bits&8 != 0,
	}, true
}

func tildeKey(value uint32) (KeyCode, uint8, bool) {
	switch value {
	case 1, 7:
		return KeyHome, 0, true
	case 2:
		return KeyInsert, 0, true
	case 3:
		return KeyDelete, 0, true
	case 4, 8:
		return KeyEnd, 0, true
	case 5:
		return KeyPageUp, 0, true
	case 6:
		return KeyPageDown, 0, true
	case 11:
		return KeyFunction, 1, true
	case 12:
		return KeyFunction, 2, true
	case 13:
		return KeyFunction, 3, true
	case 14:
		return KeyFunction, 4, true
	case 15:
		return KeyFunction, 5, true
	case 17:
		return KeyFunction, 6, true
	case 18:
		return KeyFunction, 7, true
	case 19:
		return KeyFunction, 8, true
	case 20:
		return KeyFunction, 9, true
	case 21:
		return KeyFunction, 10, true
	case 23:
		return KeyFunction, 11, true
	case 24:
		return KeyFunction, 12, true
	default:
		return KeyUnknown, 0, false
	}
}

func parseSGRMouse(body []byte, finalByte byte) (MouseEvent, bool) {
	parameters, ok := parseParameters(body)
	if !ok || len(parameters) != 3 || parameters[1] == 0 || parameters[2] == 0 {
		return MouseEvent{}, false
	}
	encoded := parameters[0]
	base := encoded & 3
	modifiers := Modifiers{Shift: encoded&4 != 0, Alt: encoded&8 != 0, Control: encoded&16 != 0}
	var kind MouseKind
	var button MouseButton
	switch {
	case encoded&64 != 0:
		kind = MouseScroll
		button = []MouseButton{MouseWheelUp, MouseWheelDown, MouseWheelLeft, MouseWheelRight}[base]
	case encoded&32 != 0:
		kind = MouseMove
		button = baseMouseButton(base)
	case finalByte == 'm':
		kind = MouseRelease
		button = baseMouseButton(base)
	default:
		kind = MousePress
		button = baseMouseButton(base)
	}
	return MouseEvent{
		Kind: kind, Button: button, X: parameters[1] - 1, Y: parameters[2] - 1, Modifiers: modifiers,
	}, true
}

func baseMouseButton(value uint32) MouseButton {
	switch value {
	case 0:
		return MouseLeft
	case 1:
		return MouseMiddle
	case 2:
		return MouseRight
	default:
		return MouseNone
	}
}

func controlEvent(value byte) Event {
	switch value {
	case '\r', '\n':
		return legacyKey(KeyEnter, 0, 0, Modifiers{})
	case '\t':
		return legacyKey(KeyTab, 0, 0, Modifiers{})
	case 0x08, 0x7F:
		return legacyKey(KeyBackspace, 0, 0, Modifiers{})
	case 0:
		return legacyKey(KeyCharacter, ' ', 0, Modifiers{Control: true})
	case 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x0B, 0x0C, 0x0E, 0x0F,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A:
		return legacyKey(KeyCharacter, rune('a'+value-1), 0, Modifiers{Control: true})
	case 0x1C:
		return legacyKey(KeyCharacter, '\\', 0, Modifiers{Control: true})
	case 0x1D:
		return legacyKey(KeyCharacter, ']', 0, Modifiers{Control: true})
	case 0x1E:
		return legacyKey(KeyCharacter, '^', 0, Modifiers{Control: true})
	case 0x1F:
		return legacyKey(KeyCharacter, '_', 0, Modifiers{Control: true})
	default:
		return legacyKey(KeyUnknown, 0, 0, Modifiers{Control: true})
	}
}

func escapeEvent() Event {
	return legacyKey(KeyEscape, 0, 0, Modifiers{})
}

func responseEvent(input []byte) Event {
	return Event{Kind: EventTerminalResponse, Bytes: append([]byte(nil), input...)}
}

func unknownEvent(input []byte) Event {
	return Event{Kind: EventUnknownSequence, Bytes: append([]byte(nil), input...)}
}

func normalizeUTF8(input []byte) string {
	return strings.ToValidUTF8(string(input), "\uFFFD")
}
