package content

import (
	"fmt"

	celltext "github.com/mayahiro/nagi-go/text"
)

// ContentKind is the closed kind of one content node
type ContentKind uint8

const (
	// ContentText contains normalized UTF-8 text
	ContentText ContentKind = iota
	// ContentHardBreak represents one explicit semantic U+000A
	ContentHardBreak
	// ContentElement contains an ordered element with metadata and children
	ContentElement
)

// ElementKind is the mechanical structure of a content element
type ElementKind uint8

const (
	// ElementInline contains inline children forming one semantic run
	ElementInline ElementKind = iota
	// ElementFlow contains ordered block children
	ElementFlow
	// ElementParagraph contains inline children forming one paragraph
	ElementParagraph
	// ElementSequence contains ordered fields or cells
	ElementSequence
)

// DefaultBoundary returns the default semantic boundary for this element kind
func (k ElementKind) DefaultBoundary() SemanticBoundary {
	switch k {
	case ElementFlow:
		return BoundaryLine
	case ElementSequence:
		return BoundaryTab
	default:
		return BoundaryNone
	}
}

// SemanticBoundary is text inserted between adjacent direct children during semantic projection
type SemanticBoundary uint8

const (
	// BoundaryNone inserts no text
	BoundaryNone SemanticBoundary = iota
	// BoundarySpace inserts one U+0020
	BoundarySpace
	// BoundaryLine inserts one U+000A
	BoundaryLine
	// BoundaryTab inserts one U+0009
	BoundaryTab
)

// Text returns the UTF-8 text represented by this boundary
func (b SemanticBoundary) Text() string {
	switch b {
	case BoundarySpace:
		return " "
	case BoundaryLine:
		return "\n"
	case BoundaryTab:
		return "\t"
	default:
		return ""
	}
}

func (b SemanticBoundary) valid() bool {
	return b <= BoundaryTab
}

// InvalidBoundaryError reports a semantic boundary outside the closed set
type InvalidBoundaryError struct {
	// Boundary is the rejected value
	Boundary SemanticBoundary
}

// Error returns the invalid-boundary diagnostic
func (e *InvalidBoundaryError) Error() string {
	return fmt.Sprintf("invalid semantic boundary %d", e.Boundary)
}

// Content is one immutable source-neutral content node
//
// Its zero value is an empty text node
type Content struct {
	_     noCompare
	inner *contentData
}

type noCompare [0]func()

type contentData struct {
	kind    ContentKind
	text    string
	element *elementData
}

var hardBreakData = contentData{kind: ContentHardBreak}
var emptyElementData = elementData{kind: ElementInline, boundary: BoundaryNone}

// NewText creates a text node and normalizes invalid UTF-8 runs
func NewText(value string) Content {
	return Content{inner: &contentData{kind: ContentText, text: celltext.NormalizeUTF8(value)}}
}

// NewTextBytes creates a text node and replaces each invalid UTF-8 run with U+FFFD
func NewTextBytes(value []byte) Content {
	return NewText(string(value))
}

// NewHardBreak creates one explicit semantic hard break
func NewHardBreak() Content {
	return Content{inner: &hardBreakData}
}

// Kind returns this node's closed kind
func (c Content) Kind() ContentKind {
	if c.inner == nil {
		return ContentText
	}
	return c.inner.kind
}

// Text returns this node's text and reports whether this is a text node
func (c Content) Text() (string, bool) {
	if c.inner == nil {
		return "", true
	}
	if c.inner.kind != ContentText {
		return "", false
	}
	return c.inner.text, true
}

// Element returns this node's element and reports whether this is an element node
func (c Content) Element() (Element, bool) {
	if c.inner == nil || c.inner.kind != ContentElement {
		return Element{}, false
	}
	return Element{inner: c.inner.element}, true
}

// Element is immutable ordered content and source-neutral metadata
//
// Its zero value is an empty inline element
type Element struct {
	_     noCompare
	inner *elementData
}

type elementData struct {
	kind          ElementKind
	id            ElementID
	hasID         bool
	revision      uint64
	roles         []Role
	classes       []Class
	annotation    AnnotationID
	hasAnnotation bool
	boundary      SemanticBoundary
	children      []Content
}

// NewInline creates an inline element with no semantic child boundary
func NewInline(children []Content) Element {
	return newElement(ElementInline, children)
}

// NewFlow creates an ordered block element with a line child boundary
func NewFlow(children []Content) Element {
	return newElement(ElementFlow, children)
}

// NewParagraph creates a paragraph element with no semantic child boundary
func NewParagraph(children []Content) Element {
	return newElement(ElementParagraph, children)
}

// NewSequence creates an ordered field element with a tab child boundary
func NewSequence(children []Content) Element {
	return newElement(ElementSequence, children)
}

func newElement(kind ElementKind, children []Content) Element {
	return Element{inner: &elementData{
		kind: kind, boundary: kind.DefaultBoundary(),
		children: append([]Content(nil), children...),
	}}
}

func (e Element) data() *elementData {
	if e.inner != nil {
		return e.inner
	}
	return &emptyElementData
}

func (e Element) cloneData() elementData {
	return *e.data()
}

// Kind returns the element's mechanical kind
func (e Element) Kind() ElementKind {
	return e.data().kind
}

// ID returns the optional stable identity
func (e Element) ID() (ElementID, bool) {
	data := e.data()
	return data.id, data.hasID
}

// Revision returns the caller-owned opaque revision
func (e Element) Revision() uint64 {
	return e.data().revision
}

// Roles returns a copy of the ordered unique semantic roles
func (e Element) Roles() []Role {
	return append([]Role(nil), e.data().roles...)
}

// HasRole reports whether this element contains role in its ordered semantic
// role set without allocating a copy of that set.
func (e Element) HasRole(role Role) bool {
	for _, candidate := range e.data().roles {
		if candidate == role {
			return true
		}
	}
	return false
}

// Classes returns a copy of the ordered unique presentation classes
func (e Element) Classes() []Class {
	return append([]Class(nil), e.data().classes...)
}

// HasClass reports whether this element contains class in its ordered
// presentation class set without allocating a copy of that set.
func (e Element) HasClass(class Class) bool {
	for _, candidate := range e.data().classes {
		if candidate == class {
			return true
		}
	}
	return false
}

// Annotation returns the optional application-resolved annotation identity
func (e Element) Annotation() (AnnotationID, bool) {
	data := e.data()
	return data.annotation, data.hasAnnotation
}

// Boundary returns the semantic boundary inserted between adjacent children
func (e Element) Boundary() SemanticBoundary {
	return e.data().boundary
}

// Children returns a copy of the immutable ordered children
func (e Element) Children() []Content {
	return append([]Content(nil), e.data().children...)
}

// ChildCount returns the number of immutable ordered children.
func (e Element) ChildCount() int {
	return len(e.data().children)
}

// Child returns one immutable child by zero-based index without allocating a
// copy of the complete child sequence.
func (e Element) Child(index int) (Content, bool) {
	children := e.data().children
	if index < 0 || index >= len(children) {
		return Content{}, false
	}
	return children[index], true
}

// WithID returns this element with a stable identity
func (e Element) WithID(id ElementID) (Element, error) {
	if id.value == "" {
		return Element{}, identifierError(IdentifierEmpty)
	}
	data := e.cloneData()
	data.id = id
	data.hasID = true
	return Element{inner: &data}, nil
}

// WithRevision returns this element with a caller-owned opaque revision
func (e Element) WithRevision(revision uint64) Element {
	data := e.cloneData()
	data.revision = revision
	return Element{inner: &data}
}

// WithRoles returns this element with a validated ordered role set
func (e Element) WithRoles(roles []Role) (Element, error) {
	owned := append([]Role(nil), roles...)
	seen := make(map[string]struct{}, len(owned))
	for _, role := range owned {
		if err := validatePortableToken(role.value); err != nil {
			return Element{}, err
		}
		if _, exists := seen[role.value]; exists {
			return Element{}, &DuplicateRoleError{Role: role}
		}
		seen[role.value] = struct{}{}
	}
	data := e.cloneData()
	data.roles = owned
	return Element{inner: &data}, nil
}

// WithClasses returns this element with a validated ordered class set
func (e Element) WithClasses(classes []Class) (Element, error) {
	owned := append([]Class(nil), classes...)
	seen := make(map[string]struct{}, len(owned))
	for _, class := range owned {
		if err := validatePortableToken(class.value); err != nil {
			return Element{}, err
		}
		if _, exists := seen[class.value]; exists {
			return Element{}, &DuplicateClassError{Class: class}
		}
		seen[class.value] = struct{}{}
	}
	data := e.cloneData()
	data.classes = owned
	return Element{inner: &data}, nil
}

// WithAnnotation returns this element with an application-resolved annotation
func (e Element) WithAnnotation(annotation AnnotationID) (Element, error) {
	if err := validatePortableToken(annotation.value); err != nil {
		return Element{}, err
	}
	data := e.cloneData()
	data.annotation = annotation
	data.hasAnnotation = true
	return Element{inner: &data}, nil
}

// WithBoundary returns this element with an explicit semantic child boundary
func (e Element) WithBoundary(boundary SemanticBoundary) (Element, error) {
	if !boundary.valid() {
		return Element{}, &InvalidBoundaryError{Boundary: boundary}
	}
	data := e.cloneData()
	data.boundary = boundary
	return Element{inner: &data}, nil
}

// Content converts this configured element into a content node
func (e Element) Content() Content {
	return Content{inner: &contentData{kind: ContentElement, element: e.data()}}
}

func (e Element) roles() []Role {
	return e.data().roles
}

func (e Element) classes() []Class {
	return e.data().classes
}

func (e Element) children() []Content {
	return e.data().children
}

// DuplicateRoleError reports one role repeated on an element
type DuplicateRoleError struct {
	// Role is the repeated role
	Role Role
}

// Error returns the duplicate-role diagnostic
func (e *DuplicateRoleError) Error() string {
	return "duplicate content role " + e.Role.String()
}

// DuplicateClassError reports one class repeated on an element
type DuplicateClassError struct {
	// Class is the repeated class
	Class Class
}

// Error returns the duplicate-class diagnostic
func (e *DuplicateClassError) Error() string {
	return "duplicate content class " + e.Class.String()
}
