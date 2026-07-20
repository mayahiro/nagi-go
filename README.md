# Nagi shared libraries for Go

[日本語](README_ja.md)

`github.com/mayahiro/nagi-go` contains the Go foundations shared by Nagi TUI
and Nagi CLI

It deliberately has no package at the module root

## Packages

| Package | Responsibility |
| --- | --- |
| [`text`](text) | Unicode 17 grapheme segmentation, terminal cell width, wrapping, truncation, and byte/cell positions |
| [`vt`](vt) | Pure typed VT input decoding and output encoding, terminal capabilities, colors, attributes, and styles |

The packages own no application runtime, Surface, Widget, terminal session, or
CLI command lifecycle

## Installation after the first release

After v0.1.0 is published, import only the package an application needs

```sh
go get github.com/mayahiro/nagi-go/text@v0.1.0
go get github.com/mayahiro/nagi-go/vt@v0.1.0
```

Both packages are released together under the module version

## Example

```go
package main

import (
	"fmt"

	"github.com/mayahiro/nagi-go/text"
	"github.com/mayahiro/nagi-go/vt"
)

func main() {
	width := text.Width("Nagi", text.ModernWidth())
	style := vt.Style{Foreground: vt.IndexedColor(4), Bold: true}
	fmt.Println(width, style.Attributes())
}
```

## Development

```sh
make check
```

Shared fixture tests use `NAGI_FIXTURES` when it names the fixture directory in
the Nagi coordination repository

## License

Source code is available under the MIT License. Generated Unicode data is
distributed under the [Unicode License v3](UNICODE-LICENSE)
