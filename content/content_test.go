package content

import (
	"errors"
	"testing"
)

func TestImmutableStorageAndDefensiveSlices(t *testing.T) {
	children := []Content{NewText("original")}
	role, err := NewRole("document")
	if err != nil {
		t.Fatal(err)
	}
	roles := []Role{role}
	element, err := NewInline(children).WithRoles(roles)
	if err != nil {
		t.Fatal(err)
	}
	children[0] = NewText("changed input")
	roles[0] = Role{}

	returned := element.Children()
	returned[0] = NewText("changed result")
	returnedRoles := element.Roles()
	returnedRoles[0] = Role{}
	if got := ProjectSemanticText(element.Content()).Text(); got != "original" {
		t.Fatalf("ProjectSemanticText() = %q, want original", got)
	}
	if got := element.Roles()[0].String(); got != "document" {
		t.Fatalf("Roles()[0] = %q, want document", got)
	}

	content := element.Content()
	cloned := content
	if content.inner != cloned.inner {
		t.Fatal("content copy did not share immutable storage")
	}
}

func TestDuplicateMetadataAndInvalidZeroValues(t *testing.T) {
	role, err := NewRole("heading")
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewInline(nil).WithRoles([]Role{role, role})
	var duplicateRole *DuplicateRoleError
	if !errors.As(err, &duplicateRole) {
		t.Fatalf("WithRoles() error = %v, want DuplicateRoleError", err)
	}

	class, err := NewClass("source.bold")
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewInline(nil).WithClasses([]Class{class, class})
	var duplicateClass *DuplicateClassError
	if !errors.As(err, &duplicateClass) {
		t.Fatalf("WithClasses() error = %v, want DuplicateClassError", err)
	}

	if _, err := NewInline(nil).WithID(ElementID{}); err == nil {
		t.Fatal("WithID() accepted a zero-value ElementID")
	}
	if _, err := NewInline(nil).WithAnnotation(AnnotationID{}); err == nil {
		t.Fatal("WithAnnotation() accepted a zero-value AnnotationID")
	}
	if _, err := NewInline(nil).WithBoundary(SemanticBoundary(255)); err == nil {
		t.Fatal("WithBoundary() accepted an unknown boundary")
	}
}

func TestZeroValuesAreUsable(t *testing.T) {
	var content Content
	if value, ok := content.Text(); !ok || value != "" {
		t.Fatalf("zero Content.Text() = %q, %t", value, ok)
	}
	stats, err := Validate(content, UnlimitedLimits())
	if err != nil {
		t.Fatal(err)
	}
	if stats.Nodes != 1 || stats.MaxDepth != 1 {
		t.Fatalf("zero Content stats = %+v", stats)
	}

	var element Element
	if got := ProjectSemanticText(element.Content()).Text(); got != "" {
		t.Fatalf("zero Element projection = %q", got)
	}
}

func TestValidationErrorDetails(t *testing.T) {
	limits := UnlimitedLimits()
	limits.MaxNodes = 0
	_, err := Validate(NewText("x"), limits)
	var validationError *ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("Validate() error = %v, want ValidationError", err)
	}
	if validationError.Kind() != ValidationNodeLimit {
		t.Fatalf("Kind() = %v, want node limit", validationError.Kind())
	}
	if observed, ok := validationError.Observed(); !ok || observed != 1 {
		t.Fatalf("Observed() = %d, %t, want 1, true", observed, ok)
	}
	if limit, ok := validationError.Limit(); !ok || limit != 0 {
		t.Fatalf("Limit() = %d, %t, want 0, true", limit, ok)
	}
	if _, ok := validationError.ElementID(); ok {
		t.Fatal("ElementID() unexpectedly returned a value")
	}
}

func TestDeepTraversalDoesNotUseTheCallStack(t *testing.T) {
	content := NewText("deep")
	const depth = 4096
	for range depth - 1 {
		content = NewInline([]Content{content}).Content()
	}
	if got := ProjectSemanticText(content).Text(); got != "deep" {
		t.Fatalf("ProjectSemanticText() = %q, want deep", got)
	}
	stats, err := Validate(content, UnlimitedLimits())
	if err != nil {
		t.Fatal(err)
	}
	if stats.MaxDepth != depth {
		t.Fatalf("MaxDepth = %d, want %d", stats.MaxDepth, depth)
	}
}
