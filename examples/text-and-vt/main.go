// Command text-and-vt demonstrates the shared Nagi Go foundations
package main

import (
	"fmt"

	"github.com/mayahiro/nagi-go/text"
	"github.com/mayahiro/nagi-go/vt"
)

func main() {
	width := text.Width("Nagi", text.ModernWidth())
	style := vt.Style{Foreground: vt.IndexedColor(4), Bold: true}
	fmt.Printf("width=%d bold=%t\n", width, style.Attributes().Bold)
}
