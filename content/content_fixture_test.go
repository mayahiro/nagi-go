package content

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/mayahiro/nagi-go/internal/conformance"
)

func TestIdentifierFixtures(t *testing.T) {
	records := contentRecords(
		t, "content/identifiers.txt", "content-identifiers", "kind", "value", "expected",
	)
	for _, record := range records {
		err := constructFixtureIdentifier(record.Field("kind"), record.Bytes("value"))
		actual := "ok"
		if err != nil {
			var identifierError *IdentifierError
			if !errors.As(err, &identifierError) {
				t.Fatalf("case %s: unexpected error %v", record.ID, err)
			}
			actual = identifierError.Kind().String()
		}
		if want := record.Field("expected"); actual != want {
			t.Errorf("case %s: identifier result = %q, want %q", record.ID, actual, want)
		}
	}
}

func TestElementFixtures(t *testing.T) {
	records := contentRecords(
		t, "content/elements.txt", "content-elements",
		"kind", "boundary", "id", "revision", "roles", "classes", "annotation", "children",
	)
	for _, record := range records {
		element := elementFixtureCase(record.ID)
		if got, want := elementKindName(element.Kind()), record.Field("kind"); got != want {
			t.Errorf("case %s kind: got %q, want %q", record.ID, got, want)
		}
		if got, want := boundaryName(element.Boundary()), record.Field("boundary"); got != want {
			t.Errorf("case %s boundary: got %q, want %q", record.ID, got, want)
		}
		id := "none"
		if value, ok := element.ID(); ok {
			id = value.String()
		}
		if want := record.Field("id"); id != want {
			t.Errorf("case %s id: got %q, want %q", record.ID, id, want)
		}
		if got, want := element.Revision(), fixtureNumber(record.Field("revision")); got != want {
			t.Errorf("case %s revision: got %d, want %d", record.ID, got, want)
		}
		roles := element.Roles()
		roleValues := make([]string, len(roles))
		for index, role := range roles {
			roleValues[index] = role.String()
		}
		if got, want := fixtureTokens(roleValues), record.Field("roles"); got != want {
			t.Errorf("case %s roles: got %q, want %q", record.ID, got, want)
		}
		classes := element.Classes()
		classValues := make([]string, len(classes))
		for index, class := range classes {
			classValues[index] = class.String()
		}
		if got, want := fixtureTokens(classValues), record.Field("classes"); got != want {
			t.Errorf("case %s classes: got %q, want %q", record.ID, got, want)
		}
		annotation := "none"
		if value, ok := element.Annotation(); ok {
			annotation = value.String()
		}
		if want := record.Field("annotation"); annotation != want {
			t.Errorf("case %s annotation: got %q, want %q", record.ID, annotation, want)
		}
		if got, want := uint64(len(element.Children())), fixtureNumber(record.Field("children")); got != want {
			t.Errorf("case %s children: got %d, want %d", record.ID, got, want)
		}
	}
}

func TestSemanticFixtures(t *testing.T) {
	records := contentRecords(
		t, "content/semantic.txt", "content-semantic",
		"expected", "annotations", "nodes", "depth", "semantic-bytes", "metadata-bytes",
		"annotation-count", "identified-count", "max-tokens",
	)
	for _, record := range records {
		content := semanticFixtureCase(record.ID)
		projected := ProjectSemanticText(content)
		if got, want := projected.Text(), record.Text("expected"); got != want {
			t.Errorf("case %s text: got %q, want %q", record.ID, got, want)
		}

		annotations := projected.Annotations()
		annotationText := "none"
		if len(annotations) > 0 {
			parts := make([]string, len(annotations))
			for index, annotation := range annotations {
				parts[index] = fmt.Sprintf(
					"%s:%d-%d", annotation.Annotation(), annotation.Start(), annotation.End(),
				)
			}
			annotationText = strings.Join(parts, ",")
		}
		if want := record.Field("annotations"); annotationText != want {
			t.Errorf("case %s annotations: got %q, want %q", record.ID, annotationText, want)
		}

		got, err := Validate(content, UnlimitedLimits())
		if err != nil {
			t.Fatalf("case %s: Validate() error = %v", record.ID, err)
		}
		want := Stats{
			Nodes:               fixtureNumber(record.Field("nodes")),
			MaxDepth:            fixtureNumber(record.Field("depth")),
			SemanticBytes:       fixtureNumber(record.Field("semantic-bytes")),
			MetadataBytes:       fixtureNumber(record.Field("metadata-bytes")),
			Annotations:         fixtureNumber(record.Field("annotation-count")),
			IdentifiedElements:  fixtureNumber(record.Field("identified-count")),
			MaxTokensPerElement: fixtureNumber(record.Field("max-tokens")),
		}
		if got != want {
			t.Errorf("case %s stats: got %+v, want %+v", record.ID, got, want)
		}
	}
}

func TestValidationFixtures(t *testing.T) {
	records := contentRecords(
		t, "content/validation.txt", "content-validation", "limits", "expected",
	)
	for _, record := range records {
		_, err := Validate(validationFixtureCase(record.ID), fixtureLimits(record.Field("limits")))
		actual := "ok"
		if err != nil {
			var validationError *ValidationError
			if !errors.As(err, &validationError) {
				t.Fatalf("case %s: unexpected error %v", record.ID, err)
			}
			actual = validationError.Kind().String()
		}
		if want := record.Field("expected"); actual != want {
			t.Errorf("case %s: validation result = %q, want %q", record.ID, actual, want)
		}
	}
}

func constructFixtureIdentifier(kind string, value []byte) error {
	switch kind {
	case "element":
		_, err := NewElementIDBytes(value)
		return err
	case "role":
		_, err := NewRoleBytes(value)
		return err
	case "class":
		_, err := NewClassBytes(value)
		return err
	case "annotation":
		_, err := NewAnnotationIDBytes(value)
		return err
	default:
		panic("unknown identifier kind " + kind)
	}
}

func elementFixtureCase(id string) Element {
	switch id {
	case "inline-default":
		return NewInline(nil)
	case "flow-default":
		return NewFlow(nil)
	case "paragraph-default":
		return NewParagraph(nil)
	case "sequence-default":
		return NewSequence(nil)
	case "configured":
		return mustBoundary(
			mustAnnotation(
				mustClasses(
					mustRoles(
						mustID(NewSequence([]Content{NewText("left"), NewText("right")}), "document 1"),
						"document", "diagnostic.code",
					),
					"source.bold", "theme.high_contrast",
				),
				"copy.document",
			).WithRevision(42),
			BoundarySpace,
		)
	default:
		panic("unknown element case " + id)
	}
}

func elementKindName(kind ElementKind) string {
	switch kind {
	case ElementInline:
		return "inline"
	case ElementFlow:
		return "flow"
	case ElementParagraph:
		return "paragraph"
	case ElementSequence:
		return "sequence"
	default:
		return "unknown"
	}
}

func boundaryName(boundary SemanticBoundary) string {
	switch boundary {
	case BoundaryNone:
		return "none"
	case BoundarySpace:
		return "space"
	case BoundaryLine:
		return "line"
	case BoundaryTab:
		return "tab"
	default:
		return "unknown"
	}
}

func fixtureTokens(values []string) string {
	if len(values) == 0 {
		return "none"
	}
	return strings.Join(values, ",")
}

func semanticFixtureCase(id string) Content {
	switch id {
	case "empty-text":
		return NewText("")
	case "normalized-text":
		return NewTextBytes([]byte{'A', 0xFF, 0xFE, 'B'})
	case "hard-break":
		return NewHardBreak()
	case "default-boundaries":
		return NewFlow([]Content{
			NewParagraph([]Content{NewText("alpha")}).Content(),
			NewSequence([]Content{NewText("left"), NewText("right")}).Content(),
			NewInline([]Content{NewText("tail"), NewText("!")}).Content(),
		}).Content()
	case "boundary-overrides":
		return NewFlow([]Content{
			mustBoundary(NewParagraph([]Content{NewText("A"), NewText("B")}), BoundarySpace).Content(),
			mustBoundary(NewFlow([]Content{NewText("C"), NewText("D")}), BoundaryNone).Content(),
		}).Content()
	case "nested-annotations":
		return nestedAnnotationsFixture()
	case "empty-annotation":
		return mustAnnotation(NewInline(nil), "marker.empty").Content()
	case "parent-boundary-range":
		return NewSequence([]Content{
			NewText("a"),
			mustAnnotation(NewInline([]Content{NewText("b")}), "cell.middle").Content(),
			NewText("c"),
		}).Content()
	default:
		panic("unknown semantic case " + id)
	}
}

func nestedAnnotationsFixture() Content {
	emphasized := mustAnnotation(
		mustClasses(NewInline([]Content{NewText(" now")}), "source.bold"),
		"emphasis.one",
	).Content()
	first := mustAnnotation(
		mustRoles(
			mustID(NewParagraph([]Content{NewText("Go"), emphasized}), "p-1"),
			"paragraph",
		),
		"link.one",
	).Content()
	second := NewParagraph([]Content{NewText("Done")}).Content()
	root := mustAnnotation(
		mustClasses(
			mustRoles(mustID(NewFlow([]Content{first, second}), "doc 1"), "document"),
			"theme.base",
		),
		"copy.all",
	)
	return root.WithRevision(7).Content()
}

func validationFixtureCase(id string) Content {
	switch id {
	case "valid":
		return mustAnnotation(
			mustClasses(
				mustRoles(mustID(NewFlow([]Content{NewText("ok")}), "root"), "document"),
				"theme.base",
			),
			"copy.all",
		).Content()
	case "node-limit":
		return NewFlow([]Content{NewText("a"), NewText("b")}).Content()
	case "depth-limit":
		return NewInline([]Content{
			NewInline([]Content{NewText("deep")}).Content(),
		}).Content()
	case "semantic-byte-limit":
		return NewSequence([]Content{NewText("a"), NewText("b")}).Content()
	case "metadata-byte-limit":
		return mustRoles(mustID(NewInline(nil), "abc"), "heading").Content()
	case "token-limit":
		return mustClasses(mustRoles(NewInline(nil), "a", "b"), "c").Content()
	case "duplicate-element-id":
		return NewFlow([]Content{
			mustID(NewParagraph([]Content{NewText("a")}), "same").Content(),
			mustID(NewParagraph([]Content{NewText("b")}), "same").Content(),
		}).Content()
	case "limit-precedence":
		return NewText("x")
	default:
		panic("unknown validation case " + id)
	}
}

func mustID(element Element, value string) Element {
	id, err := NewElementID(value)
	if err != nil {
		panic(err)
	}
	result, err := element.WithID(id)
	if err != nil {
		panic(err)
	}
	return result
}

func mustRoles(element Element, values ...string) Element {
	roles := make([]Role, len(values))
	for index, value := range values {
		role, err := NewRole(value)
		if err != nil {
			panic(err)
		}
		roles[index] = role
	}
	result, err := element.WithRoles(roles)
	if err != nil {
		panic(err)
	}
	return result
}

func mustClasses(element Element, values ...string) Element {
	classes := make([]Class, len(values))
	for index, value := range values {
		class, err := NewClass(value)
		if err != nil {
			panic(err)
		}
		classes[index] = class
	}
	result, err := element.WithClasses(classes)
	if err != nil {
		panic(err)
	}
	return result
}

func mustAnnotation(element Element, value string) Element {
	annotation, err := NewAnnotationID(value)
	if err != nil {
		panic(err)
	}
	result, err := element.WithAnnotation(annotation)
	if err != nil {
		panic(err)
	}
	return result
}

func mustBoundary(element Element, boundary SemanticBoundary) Element {
	result, err := element.WithBoundary(boundary)
	if err != nil {
		panic(err)
	}
	return result
}

func contentRecords(t *testing.T, path, suite string, fields ...string) []conformance.Record {
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

func fixtureLimits(value string) Limits {
	parts := strings.Split(value, ",")
	if len(parts) != 5 {
		panic("invalid limits " + value)
	}
	return Limits{
		MaxDepth:            fixtureNumber(parts[0]),
		MaxNodes:            fixtureNumber(parts[1]),
		MaxSemanticBytes:    fixtureNumber(parts[2]),
		MaxMetadataBytes:    fixtureNumber(parts[3]),
		MaxTokensPerElement: fixtureNumber(parts[4]),
	}
}

func fixtureNumber(value string) uint64 {
	number, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		panic("invalid fixture number " + value)
	}
	return number
}
