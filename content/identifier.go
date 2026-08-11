package content

import "unicode/utf8"

// IdentifierErrorKind is a stable reason that an identifier could not be constructed
type IdentifierErrorKind uint8

const (
	// IdentifierEmpty means that an identifier contained no bytes
	IdentifierEmpty IdentifierErrorKind = iota
	// IdentifierInvalidUTF8 means that an identifier was not valid UTF-8
	IdentifierInvalidUTF8
	// IdentifierInvalidSegmentStart means that a portable token segment did not start with an ASCII lowercase letter
	IdentifierInvalidSegmentStart
	// IdentifierInvalidCharacter means that a portable token contained a byte outside its grammar
	IdentifierInvalidCharacter
)

// String returns the stable specification name for this failure kind
func (k IdentifierErrorKind) String() string {
	switch k {
	case IdentifierEmpty:
		return "empty"
	case IdentifierInvalidUTF8:
		return "invalid-utf8"
	case IdentifierInvalidSegmentStart:
		return "invalid-segment-start"
	case IdentifierInvalidCharacter:
		return "invalid-character"
	default:
		return "unknown"
	}
}

// IdentifierError reports an invalid content identifier
type IdentifierError struct {
	kind IdentifierErrorKind
}

// Kind returns the stable reason for this failure
func (e *IdentifierError) Kind() IdentifierErrorKind {
	return e.kind
}

// Error returns the identifier diagnostic
func (e *IdentifierError) Error() string {
	return "invalid content identifier: " + e.kind.String()
}

// ElementID is opaque stable identity for one element occurrence
type ElementID struct {
	value string
}

// NewElementID creates a non-empty element ID from valid UTF-8
func NewElementID(value string) (ElementID, error) {
	if !utf8.ValidString(value) {
		return ElementID{}, identifierError(IdentifierInvalidUTF8)
	}
	if value == "" {
		return ElementID{}, identifierError(IdentifierEmpty)
	}
	return ElementID{value: value}, nil
}

// NewElementIDBytes creates a non-empty element ID without repairing invalid UTF-8
func NewElementIDBytes(value []byte) (ElementID, error) {
	if !utf8.Valid(value) {
		return ElementID{}, identifierError(IdentifierInvalidUTF8)
	}
	return NewElementID(string(value))
}

// String returns the opaque UTF-8 value
func (i ElementID) String() string {
	return i.value
}

// Role is an open semantic role attached to a content element
type Role struct {
	value string
}

// NewRole creates a role using the portable content-token grammar
func NewRole(value string) (Role, error) {
	if err := validatePortableToken(value); err != nil {
		return Role{}, err
	}
	return Role{value: value}, nil
}

// NewRoleBytes creates a role without repairing invalid UTF-8
func NewRoleBytes(value []byte) (Role, error) {
	if !utf8.Valid(value) {
		return Role{}, identifierError(IdentifierInvalidUTF8)
	}
	return NewRole(string(value))
}

// String returns the portable role token
func (r Role) String() string {
	return r.value
}

// Class is an opaque presentation-rule class attached to a content element
type Class struct {
	value string
}

// NewClass creates a class using the portable content-token grammar
func NewClass(value string) (Class, error) {
	if err := validatePortableToken(value); err != nil {
		return Class{}, err
	}
	return Class{value: value}, nil
}

// NewClassBytes creates a class without repairing invalid UTF-8
func NewClassBytes(value []byte) (Class, error) {
	if !utf8.Valid(value) {
		return Class{}, identifierError(IdentifierInvalidUTF8)
	}
	return NewClass(string(value))
}

// String returns the portable class token
func (c Class) String() string {
	return c.value
}

// AnnotationID is opaque application-resolved annotation identity
type AnnotationID struct {
	value string
}

// NewAnnotationID creates an annotation ID using the portable content-token grammar
func NewAnnotationID(value string) (AnnotationID, error) {
	if err := validatePortableToken(value); err != nil {
		return AnnotationID{}, err
	}
	return AnnotationID{value: value}, nil
}

// NewAnnotationIDBytes creates an annotation ID without repairing invalid UTF-8
func NewAnnotationIDBytes(value []byte) (AnnotationID, error) {
	if !utf8.Valid(value) {
		return AnnotationID{}, identifierError(IdentifierInvalidUTF8)
	}
	return NewAnnotationID(string(value))
}

// String returns the portable annotation token
func (i AnnotationID) String() string {
	return i.value
}

func validatePortableToken(value string) error {
	if !utf8.ValidString(value) {
		return identifierError(IdentifierInvalidUTF8)
	}
	if value == "" {
		return identifierError(IdentifierEmpty)
	}

	segmentStart := true
	for _, value := range []byte(value) {
		if segmentStart {
			if value >= 'a' && value <= 'z' {
				segmentStart = false
				continue
			}
			return identifierError(IdentifierInvalidSegmentStart)
		}
		if value == '.' {
			segmentStart = true
			continue
		}
		if (value < 'a' || value > 'z') &&
			(value < '0' || value > '9') &&
			value != '-' && value != '_' {
			return identifierError(IdentifierInvalidCharacter)
		}
	}
	if segmentStart {
		return identifierError(IdentifierInvalidSegmentStart)
	}
	return nil
}

func identifierError(kind IdentifierErrorKind) error {
	return &IdentifierError{kind: kind}
}
