# Nagi Go共有library

[English](README.md)

`github.com/mayahiro/nagi-go`はNagi terminal application向けに独立して再利用できるContent、Text、VTの基盤を提供します。Module rootには意図的にpackageを置きません

## 要件

- Go 1.25以降

## Package

| Package | 責務 |
| --- | --- |
| [`content`](content) | Immutableなsource-neutral text構造、semantic projection、annotation、resource validation |
| [`text`](text) | Unicode 17 grapheme segmentation、terminal Cell幅、wrap、truncate、byte／Cell位置変換 |
| [`vt`](vt) | 純粋なtyped VT input decoderとoutput encoder、terminal capability、Color、Attributes、Style |

Application runtime、Surface、Widget、terminal session、CLI command lifecycleはこのmoduleの責務に含めません

## 導入

```sh
go get github.com/mayahiro/nagi-go@v0.1.0
```

Applicationが使用するpackageだけをimportします

## Example

- [Source-neutral Content](examples/content/README.md)
- [TextとVT](examples/text-and-vt/README.md)

```sh
go run ./examples/content
go run ./examples/text-and-vt
```

## License

Source codeはMIT Licenseで提供します。生成済みUnicode dataは[Unicode License v3](UNICODE-LICENSE)で配布します
