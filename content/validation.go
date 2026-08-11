package content

import "fmt"

const maximumUint64 = ^uint64(0)

// Limits contains inclusive resource limits for explicit content validation
type Limits struct {
	// MaxDepth is the maximum root-relative node depth, where the root has depth one
	MaxDepth uint64
	// MaxNodes is the maximum number of node occurrences
	MaxNodes uint64
	// MaxSemanticBytes is the maximum projected semantic UTF-8 byte count
	MaxSemanticBytes uint64
	// MaxMetadataBytes is the maximum identifier and token UTF-8 byte count
	MaxMetadataBytes uint64
	// MaxTokensPerElement is the maximum combined role and class count on one element
	MaxTokensPerElement uint64
}

// UnlimitedLimits returns limits that accept every representable resource count
func UnlimitedLimits() Limits {
	return Limits{
		MaxDepth: maximumUint64, MaxNodes: maximumUint64,
		MaxSemanticBytes: maximumUint64, MaxMetadataBytes: maximumUint64,
		MaxTokensPerElement: maximumUint64,
	}
}

// Stats contains complete resource statistics for one validated content tree
type Stats struct {
	// Nodes is the number of node occurrences
	Nodes uint64
	// MaxDepth is the greatest root-relative node depth
	MaxDepth uint64
	// SemanticBytes is the projected semantic UTF-8 byte count
	SemanticBytes uint64
	// MetadataBytes is the identifier and token UTF-8 byte count
	MetadataBytes uint64
	// Annotations is the number of annotated element occurrences
	Annotations uint64
	// IdentifiedElements is the number of identified element occurrences
	IdentifiedElements uint64
	// MaxTokensPerElement is the greatest combined role and class count on one element
	MaxTokensPerElement uint64
}

// ValidationErrorKind is a stable reason that content validation stopped
type ValidationErrorKind uint8

const (
	// ValidationNodeLimit means that the node occurrence limit was exceeded
	ValidationNodeLimit ValidationErrorKind = iota
	// ValidationDepthLimit means that the root-relative depth limit was exceeded
	ValidationDepthLimit
	// ValidationSemanticByteLimit means that the semantic UTF-8 byte limit was exceeded
	ValidationSemanticByteLimit
	// ValidationMetadataByteLimit means that the metadata UTF-8 byte limit was exceeded
	ValidationMetadataByteLimit
	// ValidationTokenLimit means that one element's role-plus-class limit was exceeded
	ValidationTokenLimit
	// ValidationDuplicateElementID means that a stable element ID occurred more than once
	ValidationDuplicateElementID
)

// String returns the stable specification name for this failure kind
func (k ValidationErrorKind) String() string {
	switch k {
	case ValidationNodeLimit:
		return "node-limit"
	case ValidationDepthLimit:
		return "depth-limit"
	case ValidationSemanticByteLimit:
		return "semantic-byte-limit"
	case ValidationMetadataByteLimit:
		return "metadata-byte-limit"
	case ValidationTokenLimit:
		return "token-limit"
	case ValidationDuplicateElementID:
		return "duplicate-element-id"
	default:
		return "unknown"
	}
}

// ValidationError is a structured content validation failure
type ValidationError struct {
	kind         ValidationErrorKind
	observed     uint64
	limit        uint64
	hasLimit     bool
	elementID    ElementID
	hasElementID bool
}

// Kind returns the stable validation failure kind
func (e *ValidationError) Kind() ValidationErrorKind {
	return e.kind
}

// Observed returns the first count that exceeded a limit
func (e *ValidationError) Observed() (uint64, bool) {
	return e.observed, e.hasLimit
}

// Limit returns the configured limit for a resource failure
func (e *ValidationError) Limit() (uint64, bool) {
	return e.limit, e.hasLimit
}

// ElementID returns the repeated element ID for a duplicate failure
func (e *ValidationError) ElementID() (ElementID, bool) {
	return e.elementID, e.hasElementID
}

// Error returns the validation diagnostic
func (e *ValidationError) Error() string {
	if e.hasLimit {
		return fmt.Sprintf("content %s: observed %d, limit %d", e.kind, e.observed, e.limit)
	}
	if e.hasElementID {
		return "duplicate content element ID " + e.elementID.String()
	}
	return e.kind.String()
}

type validationEventKind uint8

const (
	validationNode validationEventKind = iota
	validationBoundary
)

type validationEvent struct {
	kind     validationEventKind
	content  Content
	depth    uint64
	boundary SemanticBoundary
}

// Validate validates one immutable content tree against explicit resource limits
func Validate(content Content, limits Limits) (Stats, error) {
	var stats Stats
	var ids map[string]struct{}
	stack := []validationEvent{{kind: validationNode, content: content, depth: 1}}
	for len(stack) > 0 {
		last := len(stack) - 1
		event := stack[last]
		stack = stack[:last]
		if event.kind == validationBoundary {
			if err := addSemanticBytes(&stats, uint64(len(event.boundary.Text())), limits.MaxSemanticBytes); err != nil {
				return Stats{}, err
			}
			continue
		}

		stats.Nodes = saturatingAdd(stats.Nodes, 1)
		if err := checkLimit(ValidationNodeLimit, stats.Nodes, limits.MaxNodes); err != nil {
			return Stats{}, err
		}
		if event.depth > stats.MaxDepth {
			stats.MaxDepth = event.depth
		}
		if err := checkLimit(ValidationDepthLimit, event.depth, limits.MaxDepth); err != nil {
			return Stats{}, err
		}

		switch event.content.Kind() {
		case ContentText:
			value, _ := event.content.Text()
			if err := addSemanticBytes(&stats, uint64(len(value)), limits.MaxSemanticBytes); err != nil {
				return Stats{}, err
			}
		case ContentHardBreak:
			if err := addSemanticBytes(&stats, 1, limits.MaxSemanticBytes); err != nil {
				return Stats{}, err
			}
		case ContentElement:
			element, _ := event.content.Element()
			tokens := saturatingAdd(uint64(len(element.roles())), uint64(len(element.classes())))
			if tokens > stats.MaxTokensPerElement {
				stats.MaxTokensPerElement = tokens
			}
			if err := checkLimit(ValidationTokenLimit, tokens, limits.MaxTokensPerElement); err != nil {
				return Stats{}, err
			}

			var metadataBytes uint64
			if id, ok := element.ID(); ok {
				metadataBytes = saturatingAdd(metadataBytes, uint64(len(id.String())))
			}
			for _, role := range element.roles() {
				metadataBytes = saturatingAdd(metadataBytes, uint64(len(role.String())))
			}
			for _, class := range element.classes() {
				metadataBytes = saturatingAdd(metadataBytes, uint64(len(class.String())))
			}
			if annotation, ok := element.Annotation(); ok {
				metadataBytes = saturatingAdd(metadataBytes, uint64(len(annotation.String())))
			}
			stats.MetadataBytes = saturatingAdd(stats.MetadataBytes, metadataBytes)
			if err := checkLimit(ValidationMetadataByteLimit, stats.MetadataBytes, limits.MaxMetadataBytes); err != nil {
				return Stats{}, err
			}

			if id, ok := element.ID(); ok {
				stats.IdentifiedElements = saturatingAdd(stats.IdentifiedElements, 1)
				if _, exists := ids[id.String()]; exists {
					return Stats{}, &ValidationError{
						kind: ValidationDuplicateElementID, elementID: id, hasElementID: true,
					}
				}
				if ids == nil {
					ids = make(map[string]struct{})
				}
				ids[id.String()] = struct{}{}
			}
			if _, ok := element.Annotation(); ok {
				stats.Annotations = saturatingAdd(stats.Annotations, 1)
			}

			children := element.children()
			for index := len(children) - 1; index >= 0; index-- {
				stack = append(stack, validationEvent{
					kind: validationNode, content: children[index], depth: saturatingAdd(event.depth, 1),
				})
				if index > 0 {
					stack = append(stack, validationEvent{
						kind: validationBoundary, boundary: element.Boundary(),
					})
				}
			}
		}
	}
	return stats, nil
}

func addSemanticBytes(stats *Stats, bytes, limit uint64) error {
	stats.SemanticBytes = saturatingAdd(stats.SemanticBytes, bytes)
	return checkLimit(ValidationSemanticByteLimit, stats.SemanticBytes, limit)
}

func checkLimit(kind ValidationErrorKind, observed, limit uint64) error {
	if observed > limit {
		return &ValidationError{
			kind: kind, observed: observed, limit: limit, hasLimit: true,
		}
	}
	return nil
}

func saturatingAdd(left, right uint64) uint64 {
	if maximumUint64-left < right {
		return maximumUint64
	}
	return left + right
}
