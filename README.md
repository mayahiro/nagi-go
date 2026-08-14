# Nagi shared libraries for Go

[日本語](README_ja.md)

`github.com/mayahiro/nagi-go` provides independently reusable Content, Text,
and VT foundations for Nagi terminal applications. It deliberately has no
package at the module root

## Requirements

- Go 1.25 or newer

## Packages

| Package | Responsibility |
| --- | --- |
| [`content`](content) | Immutable source-neutral text structure, semantic projection, annotations, and resource validation |
| [`text`](text) | Unicode 17 grapheme segmentation, terminal Cell width, wrapping, truncation, and byte/Cell positions |
| [`vt`](vt) | Pure typed VT input decoding and output encoding, terminal capabilities, Color, Attributes, and Style |

These packages own no application runtime, Surface, Widget, terminal session,
or CLI command lifecycle

## Installation

```sh
go get github.com/mayahiro/nagi-go@latest
```

Import only the packages an application uses

## Examples

- [Source-neutral Content](examples/content/README.md)
- [Text and VT](examples/text-and-vt/README.md)

```sh
go run ./examples/content
go run ./examples/text-and-vt
```

## License

Source code is available under the MIT License. Generated Unicode data is
distributed under the [Unicode License v3](UNICODE-LICENSE)
