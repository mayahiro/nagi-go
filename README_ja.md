# Nagi Go共有library

[English](README.md)

`github.com/mayahiro/nagi-go`はNagi TUIとNagi CLIが共有するGo基盤を提供します

Module rootには意図的にpackageを置きません

## Package

| Package | 責務 |
| --- | --- |
| [`text`](text) | Unicode 17 grapheme segmentation、terminal Cell幅、wrap、truncate、byte／Cell位置変換 |
| [`vt`](vt) | 純粋なtyped VT input decoderとoutput encoder、terminal capability、Color、Attributes、Style |

Application runtime、Surface、Widget、terminal session、CLI command lifecycleはこのmoduleの責務に含めません

## 最初のrelease後のInstallation

v0.1.0の公開後に共有moduleを1回追加し、Applicationでは必要なpackageだけをimportします

```sh
go get github.com/mayahiro/nagi-go@v0.1.0
```

両packageは同じmodule versionでreleaseします

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

## 開発

```sh
make check
```

共有fixture testはNagi調整repositoryのfixture directoryを示す`NAGI_FIXTURES`を使用します

## License

Source codeはMIT Licenseで提供します。生成済みUnicode dataは[Unicode License v3](UNICODE-LICENSE)で配布します
