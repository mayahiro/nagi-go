package vt

// Modifiers is keyboard or mouse modifier state supplied by a terminal
// protocol
type Modifiers struct {
	// Shift indicates the Shift modifier
	Shift bool
	// Alt indicates the Alt modifier
	Alt bool
	// Control indicates the Control modifier
	Control bool
	// Meta indicates Meta when distinguished from Alt
	Meta bool
	// Super indicates the Windows, Linux, or Command modifier
	Super bool
	// Hyper indicates the Hyper modifier
	Hyper bool
	// CapsLock is the Caps Lock state reported by an extended protocol
	CapsLock bool
	// NumLock is the Num Lock state reported by an extended protocol
	NumLock bool
}

// WithoutLocks returns shortcut modifiers without lock-key state
func (m Modifiers) WithoutLocks() Modifiers {
	m.CapsLock = false
	m.NumLock = false
	return m
}

// KeyCode identifies a normalized logical key
type KeyCode uint8

const (
	// KeyUnknown is a control key without a more precise identity
	KeyUnknown KeyCode = iota
	// KeyCharacter uses KeyEvent.Character
	KeyCharacter
	// KeyEnter is Enter or return
	KeyEnter
	// KeyTab is horizontal tab
	KeyTab
	// KeyBackspace is Backspace
	KeyBackspace
	// KeyEscape is Escape
	KeyEscape
	// KeyUp is the up arrow
	KeyUp
	// KeyDown is the down arrow
	KeyDown
	// KeyRight is the right arrow
	KeyRight
	// KeyLeft is the left arrow
	KeyLeft
	// KeyHome is Home
	KeyHome
	// KeyEnd is End
	KeyEnd
	// KeyInsert is Insert
	KeyInsert
	// KeyDelete is Delete
	KeyDelete
	// KeyPageUp is Page Up
	KeyPageUp
	// KeyPageDown is Page Down
	KeyPageDown
	// KeyFunction uses KeyEvent.Function
	KeyFunction
	// KeyFunctional uses KeyEvent.Functional for a protocol-defined functional key
	KeyFunctional
)

// KeyAction is keyboard action information supplied by the input protocol
type KeyAction uint8

const (
	// KeyActionUnknown means the legacy protocol supplied no action
	KeyActionUnknown KeyAction = iota
	// KeyPress is a key press
	KeyPress
	// KeyRepeat is an automatic repeat
	KeyRepeat
	// KeyRelease is a key release
	KeyRelease
)

// KeyProtocol identifies the keyboard protocol that supplied an event
type KeyProtocol uint8

const (
	// KeyProtocolUnknown means the protocol was not identifiable
	KeyProtocolUnknown KeyProtocol = iota
	// KeyProtocolLegacy is traditional C0, CSI, or SS3 input
	KeyProtocolLegacy
	// KeyProtocolKitty is Kitty keyboard protocol CSI input
	KeyProtocolKitty
)

// KeyEvent is a normalized keyboard event
type KeyEvent struct {
	// Code is the logical key category
	Code KeyCode
	// Character is set when Code is KeyCharacter
	Character rune
	// Function is set when Code is KeyFunction
	Function uint8
	// Functional is set when Code is KeyFunctional
	Functional uint32
	// Modifiers contains supplied modifiers
	Modifiers Modifiers
	// Action contains supplied press, repeat, or release information
	Action KeyAction
	// Text is associated text when HasText is true
	Text string
	// HasText indicates that the protocol supplied associated text
	HasText bool
	// Protocol is the source keyboard protocol
	Protocol KeyProtocol
}

// MouseKind identifies a normalized mouse event category
type MouseKind uint8

const (
	// MousePress is a button press
	MousePress MouseKind = iota
	// MouseRelease is a button release
	MouseRelease
	// MouseMove is pointer movement
	MouseMove
	// MouseScroll is a wheel or scrolling action
	MouseScroll
)

// MouseButton identifies a normalized mouse button
type MouseButton uint8

const (
	// MouseNone means no button is held
	MouseNone MouseButton = iota
	// MouseLeft is the left button
	MouseLeft
	// MouseMiddle is the middle button
	MouseMiddle
	// MouseRight is the right button
	MouseRight
	// MouseWheelUp is wheel up
	MouseWheelUp
	// MouseWheelDown is wheel down
	MouseWheelDown
	// MouseWheelLeft is wheel left
	MouseWheelLeft
	// MouseWheelRight is wheel right
	MouseWheelRight
	// MouseOther is a protocol button without a standard mapping
	MouseOther
)

// MouseEvent is a zero-based SGR mouse event
type MouseEvent struct {
	// Kind is the event category
	Kind MouseKind
	// Button is the button or wheel direction
	Button MouseButton
	// OtherButton is set when Button is MouseOther
	OtherButton uint16
	// X is the zero-based horizontal cell coordinate
	X uint32
	// Y is the zero-based vertical cell coordinate
	Y uint32
	// Modifiers contains supplied modifiers
	Modifiers Modifiers
}

// EventKind identifies a normalized VT input event variant
type EventKind uint8

const (
	// EventKey uses Event.Key
	EventKey EventKind = iota
	// EventText uses Event.Text and contains one Unicode scalar
	EventText
	// EventPaste uses Event.Text and contains one bracketed paste payload
	EventPaste
	// EventMouse uses Event.Mouse
	EventMouse
	// EventFocusIn indicates terminal focus gain
	EventFocusIn
	// EventFocusOut indicates terminal focus loss
	EventFocusOut
	// EventTerminalResponse uses Event.Bytes
	EventTerminalResponse
	// EventUnknownSequence uses Event.Bytes
	EventUnknownSequence
)

// Event is one normalized VT input event
//
// The Kind field selects which of Key, Text, Mouse, or Bytes is meaningful.
type Event struct {
	// Kind selects the event variant
	Kind EventKind
	// Key contains EventKey data
	Key KeyEvent
	// Text contains EventText or EventPaste data
	Text string
	// Mouse contains EventMouse data
	Mouse MouseEvent
	// Bytes contains original EventTerminalResponse or EventUnknownSequence data
	Bytes []byte
}

func legacyKey(code KeyCode, character rune, function uint8, modifiers Modifiers) Event {
	return Event{Kind: EventKey, Key: KeyEvent{
		Code: code, Character: character, Function: function, Modifiers: modifiers,
		Action: KeyActionUnknown, Protocol: KeyProtocolLegacy,
	}}
}

func altCharacter(character rune) Event {
	event := legacyKey(KeyCharacter, character, 0, Modifiers{Alt: true})
	event.Key.Text = string(character)
	event.Key.HasText = true
	return event
}
