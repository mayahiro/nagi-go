// Command content demonstrates source-neutral structured content
package main

import (
	"fmt"
	"os"

	"github.com/mayahiro/nagi-go/content"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	headingRole, err := content.NewRole("heading")
	if err != nil {
		return err
	}
	heading, err := content.NewParagraph([]content.Content{content.NewText("Nagi")}).WithRoles(
		[]content.Role{headingRole},
	)
	if err != nil {
		return err
	}

	annotation, err := content.NewAnnotationID("guide.content")
	if err != nil {
		return err
	}
	description, err := content.NewParagraph(
		[]content.Content{content.NewText("Source-neutral content")},
	).WithAnnotation(annotation)
	if err != nil {
		return err
	}

	documentID, err := content.NewElementID("example-document")
	if err != nil {
		return err
	}
	document, err := content.NewFlow(
		[]content.Content{heading.Content(), description.Content()},
	).WithID(documentID)
	if err != nil {
		return err
	}
	document = document.WithRevision(1)

	root := document.Content()
	projected := content.ProjectSemanticText(root)
	stats, err := content.Validate(root, content.UnlimitedLimits())
	if err != nil {
		return err
	}
	fmt.Println(projected.Text())
	fmt.Printf("nodes=%d depth=%d\n", stats.Nodes, stats.MaxDepth)
	return nil
}
