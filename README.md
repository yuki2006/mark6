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
		"a":      mark6.AllowAttrs("href", "target"),
		"b":      mark6.AllowAttrs(),
		"p":      mark6.AllowAttrs(),
		"br":     mark6.AllowAttrs(),
		"img":    mark6.AllowAttrs("src", "alt", "width", "height"),
		"span":   mark6.AllowAttrs("class"),
		"strong": mark6.AllowAttrs(),
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
