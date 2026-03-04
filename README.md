# mark6

HTMLサニタイザー。許可するタグと属性をホワイトリスト形式で指定し、安全なHTMLを出力します。

## Usage

```go
package main

import (
	"fmt"

	"github.com/yuki2006/mark6"
)

func main() {
	// 許可するタグと属性を定義
	allowTags := mark6.AllowTags{
		"a":          mark6.AllowAttrs("href", "target"),
		"b":          mark6.AllowAttrs(),
		"i":          mark6.AllowAttrs("class"),
		"p":          mark6.AllowAttrs(),
		"br":         mark6.AllowAttrs(),
		"hr":         mark6.AllowAttrs(),
		"h1":         mark6.AllowAttrs("class"),
		"h2":         mark6.AllowAttrs("class"),
		"h3":         mark6.AllowAttrs("class"),
		"h4":         mark6.AllowAttrs("class"),
		"h5":         mark6.AllowAttrs("class"),
		"h6":         mark6.AllowAttrs("class"),
		"span":       mark6.AllowAttrs("class"),
		"details":    mark6.AllowAttrs("class"),
		"div":        mark6.AllowAttrs("class"),
		"font":       mark6.AllowAttrs("size", "color"),
		"pre":        mark6.AllowAttrs(),
		"img":        mark6.AllowAttrs("src", "alt", "width", "height"),
		"ul":         mark6.AllowAttrs(),
		"ol":         mark6.AllowAttrs(),
		"li":         mark6.AllowAttrs(),
		"table":      mark6.AllowAttrs("class", "border"),
		"thead":      mark6.AllowAttrs(),
		"tr":         mark6.AllowAttrs(),
		"th":         mark6.AllowAttrs("data-defaultsort"),
		"tbody":      mark6.AllowAttrs(),
		"td":         mark6.AllowAttrs("class"),
		"strong":     mark6.AllowAttrs(),
		"em":         mark6.AllowAttrs(),
		"code":       mark6.AllowAttrs(),
		"mark":       mark6.AllowAttrs(),
		"dl":         mark6.AllowAttrs(),
		"dt":         mark6.AllowAttrs(),
		"dd":         mark6.AllowAttrs(),
		"del":        mark6.AllowAttrs(),
		"sup":        mark6.AllowAttrs(),
		"sub":        mark6.AllowAttrs(),
		"summary":    mark6.AllowAttrs(),
		"u":          mark6.AllowAttrs(),
		"blockquote": mark6.AllowAttrs(),
		"s":          mark6.AllowAttrs(),
		"marquee":    mark6.AllowAttrs(),
	}

	src := `<p>Hello <strong>World</strong></p><script>alert("xss")</script>`

	result, err := mark6.Parse(src, allowTags)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(result)
	// Output: <p>Hello <strong>World</strong></p>alert(&#34;xss&#34;)
}
```

### ParseCallBack

特定のタグに対してコールバックを実行できます。

```go
callBack := map[string]func(node html.Node){
	"div.highlight": func(node html.Node) {
		// class="highlight" の div が見つかったときの処理
	},
	"p#intro": func(node html.Node) {
		// id="intro" の p が見つかったときの処理
	},
}

result, err := mark6.ParseCallBack(src, allowTags, callBack)
```
