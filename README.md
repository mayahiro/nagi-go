# Nagi shared libraries for Go

[日本語](README_ja.md)

`github.com/mayahiro/nagi-go` provides independently reusable Text and VT
foundations for Nagi terminal applications. It deliberately has no package at
the module root

## Requirements

- Go 1.25 or newer

## Packages

| Package | Responsibility |
| --- | --- |
| [`text`](text) | Unicode 17 grapheme segmentation, terminal Cell width, wrapping, truncation, and byte/Cell positions |
| [`vt`](vt) | Pure typed VT input decoding and output encoding, terminal capabilities, Color, Attributes, and Style |

These packages own no application runtime, Surface, Widget, terminal session,
or CLI command lifecycle

## Installation

```sh
go get github.com/mayahiro/nagi-go@v0.1.0
```

Import only the packages an application uses

## Example

The [Text and VT example](examples/text-and-vt/README.md) uses both packages in
one executable program

```sh
go run ./examples/text-and-vt
```

## License

Source code is available under the MIT License. Generated Unicode data is
distributed under the [Unicode License v3](UNICODE-LICENSE)
