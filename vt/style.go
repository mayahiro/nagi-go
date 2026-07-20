package vt

// ColorKind identifies a terminal color representation
type ColorKind uint8

const (
	// ColorDefault is the terminal's default color
	ColorDefault ColorKind = iota
	// ColorIndexed is an indexed terminal palette entry
	ColorIndexed
	// ColorRGB is a 24-bit RGB color
	ColorRGB
)

// Color is a terminal color
//
// Its zero value is the terminal default color.
type Color struct {
	kind       ColorKind
	index      uint8
	red, green uint8
	blue       uint8
}

// DefaultColor returns the terminal default color
func DefaultColor() Color {
	return Color{}
}

// IndexedColor returns an indexed terminal palette color
func IndexedColor(index uint8) Color {
	return Color{kind: ColorIndexed, index: index}
}

// RGBColor returns a 24-bit RGB color
func RGBColor(red, green, blue uint8) Color {
	return Color{kind: ColorRGB, red: red, green: green, blue: blue}
}

// Kind returns the color representation
func (c Color) Kind() ColorKind {
	return c.kind
}

// Index returns the palette index when this is an indexed color
func (c Color) Index() (uint8, bool) {
	return c.index, c.kind == ColorIndexed
}

// RGB returns the components when this is an RGB color
func (c Color) RGB() (red, green, blue uint8, ok bool) {
	return c.red, c.green, c.blue, c.kind == ColorRGB
}

// OptionalColor is an optional terminal color
//
// Its zero value has no color.
type OptionalColor struct {
	color Color
	set   bool
}

// Attributes contains the Boolean terminal text attributes
//
// Its zero value disables every attribute.
type Attributes struct {
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

// Merge returns overlay merged over these attributes
func (a Attributes) Merge(overlay Attributes) Attributes {
	return Attributes{
		Bold:          a.Bold || overlay.Bold,
		Dim:           a.Dim || overlay.Dim,
		Italic:        a.Italic || overlay.Italic,
		Underline:     a.Underline || overlay.Underline,
		Blink:         a.Blink || overlay.Blink,
		Reverse:       a.Reverse || overlay.Reverse,
		Hidden:        a.Hidden || overlay.Hidden,
		Strikethrough: a.Strikethrough || overlay.Strikethrough,
	}
}

// Empty reports whether every attribute is disabled
func (a Attributes) Empty() bool {
	return !a.Bold && !a.Dim && !a.Italic && !a.Underline &&
		!a.Blink && !a.Reverse && !a.Hidden && !a.Strikethrough
}

// SomeColor returns an optional color containing color
func SomeColor(color Color) OptionalColor {
	return OptionalColor{color: color, set: true}
}

// Get returns the color and whether it is present
func (c OptionalColor) Get() (Color, bool) {
	return c.color, c.set
}

// Style contains visual attributes attached to a cell
//
// The zero value is the terminal default style. During transparent
// composition, default colors do not replace an existing color, an absent
// underline color is preserved, and Boolean attributes are combined.
type Style struct {
	// Foreground is the text color
	Foreground Color
	// Background is the cell background color
	Background Color
	// UnderlineColor optionally overrides the underline color
	UnderlineColor OptionalColor
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

// Attributes returns the Boolean attributes in this style
func (s Style) Attributes() Attributes {
	return Attributes{
		Bold:          s.Bold,
		Dim:           s.Dim,
		Italic:        s.Italic,
		Underline:     s.Underline,
		Blink:         s.Blink,
		Reverse:       s.Reverse,
		Hidden:        s.Hidden,
		Strikethrough: s.Strikethrough,
	}
}

// WithAttributes returns this style with its Boolean attributes replaced
func (s Style) WithAttributes(attributes Attributes) Style {
	s.Bold = attributes.Bold
	s.Dim = attributes.Dim
	s.Italic = attributes.Italic
	s.Underline = attributes.Underline
	s.Blink = attributes.Blink
	s.Reverse = attributes.Reverse
	s.Hidden = attributes.Hidden
	s.Strikethrough = attributes.Strikethrough
	return s
}

// Merge returns overlay merged over this style for transparent composition
func (s Style) Merge(overlay Style) Style {
	foreground := s.Foreground
	if overlay.Foreground.Kind() != ColorDefault {
		foreground = overlay.Foreground
	}
	background := s.Background
	if overlay.Background.Kind() != ColorDefault {
		background = overlay.Background
	}
	underlineColor := s.UnderlineColor
	if _, ok := overlay.UnderlineColor.Get(); ok {
		underlineColor = overlay.UnderlineColor
	}
	return (Style{
		Foreground:     foreground,
		Background:     background,
		UnderlineColor: underlineColor,
	}).WithAttributes(s.Attributes().Merge(overlay.Attributes()))
}
