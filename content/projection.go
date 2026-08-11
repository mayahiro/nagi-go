package content

import "strings"

// AnnotationRange is one application annotation projected onto a semantic UTF-8 byte range
type AnnotationRange struct {
	annotation AnnotationID
	start      int
	end        int
}

// Annotation returns the application-resolved annotation identity
func (r AnnotationRange) Annotation() AnnotationID {
	return r.annotation
}

// Start returns the inclusive UTF-8 byte offset
func (r AnnotationRange) Start() int {
	return r.start
}

// End returns the exclusive UTF-8 byte offset
func (r AnnotationRange) End() int {
	return r.end
}

// SemanticText is owned semantic UTF-8 text and annotation ranges
type SemanticText struct {
	text        string
	annotations []AnnotationRange
}

// Text returns the projected semantic UTF-8 text
func (t SemanticText) Text() string {
	return t.text
}

// Annotations returns a copy of annotation ranges in element preorder
func (t SemanticText) Annotations() []AnnotationRange {
	return append([]AnnotationRange(nil), t.annotations...)
}

type projectionEventKind uint8

const (
	projectionNode projectionEventKind = iota
	projectionBoundary
	projectionCloseAnnotation
)

type projectionEvent struct {
	kind       projectionEventKind
	content    Content
	boundary   SemanticBoundary
	rangeIndex int
}

// ProjectSemanticText projects source-neutral content into semantic text and annotation ranges
func ProjectSemanticText(content Content) SemanticText {
	var builder strings.Builder
	var annotations []AnnotationRange
	stack := []projectionEvent{{kind: projectionNode, content: content}}
	for len(stack) > 0 {
		last := len(stack) - 1
		event := stack[last]
		stack = stack[:last]
		switch event.kind {
		case projectionNode:
			switch event.content.Kind() {
			case ContentText:
				value, _ := event.content.Text()
				builder.WriteString(value)
			case ContentHardBreak:
				builder.WriteByte('\n')
			case ContentElement:
				element, _ := event.content.Element()
				if annotation, ok := element.Annotation(); ok {
					rangeIndex := len(annotations)
					annotations = append(annotations, AnnotationRange{
						annotation: annotation, start: builder.Len(), end: builder.Len(),
					})
					stack = append(stack, projectionEvent{
						kind: projectionCloseAnnotation, rangeIndex: rangeIndex,
					})
				}
				children := element.children()
				for index := len(children) - 1; index >= 0; index-- {
					stack = append(stack, projectionEvent{kind: projectionNode, content: children[index]})
					if index > 0 {
						stack = append(stack, projectionEvent{
							kind: projectionBoundary, boundary: element.Boundary(),
						})
					}
				}
			}
		case projectionBoundary:
			builder.WriteString(event.boundary.Text())
		case projectionCloseAnnotation:
			annotations[event.rangeIndex].end = builder.Len()
		}
	}
	return SemanticText{text: builder.String(), annotations: annotations}
}
