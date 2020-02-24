package mark6

import (
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"html/template"
	"strings"
)

var (
	ERASE       = errors.New("erase")
	PARSE_ERROR = errors.New("parse error")
)

func allowAttrs(attrs ...string) map[string]bool {
	mp := make(map[string]bool)
	for _, attr := range attrs {
		mp[attr] = true
	}
	return mp
}

var allowTags = map[string]map[string]bool{
	"a":          allowAttrs("href", "target"),
	"b":          allowAttrs(),
	"i":          allowAttrs("class"),
	"p":          allowAttrs(),
	"br":         allowAttrs(),
	"h1":         allowAttrs("class"),
	"h2":         allowAttrs("class"),
	"h3":         allowAttrs("class"),
	"h4":         allowAttrs("class"),
	"h5":         allowAttrs("class"),
	"h6":         allowAttrs("class"),
	"span":       allowAttrs(),
	"div":        allowAttrs("class"),
	"pre":        allowAttrs(),
	"img":        allowAttrs("src", "alt"),
	"ul":         allowAttrs(),
	"ol":         allowAttrs(),
	"li":         allowAttrs(),
	"table":      allowAttrs("class", "border"),
	"thead":      allowAttrs(),
	"tr":         allowAttrs(),
	"th":         allowAttrs("data-defaultsort"),
	"tbody":      allowAttrs(),
	"td":         allowAttrs("class"),
	"strong":     allowAttrs(),
	"em":         allowAttrs(),
	"code":       allowAttrs(),
	"dl":         allowAttrs(),
	"dt":         allowAttrs(),
	"dd":         allowAttrs(),
	"del":        allowAttrs(),
	"sup":        allowAttrs(),
	"sub":        allowAttrs(),
	"u":          allowAttrs(),
	"blockquote": allowAttrs(),
	"s":          allowAttrs(),
	"marquee":    allowAttrs(),
}

func traversal(node *html.Node, callBack map[string]func(node html.Node)) (res string, err error) {

	res = ""

	switch node.Type {
	case html.TextNode:
		return template.HTMLEscapeString(node.Data), nil
	case html.ElementNode:
		tagName := strings.ToLower(node.Data)
		allowMap, found := allowTags[tagName]

		if found {
			attrs := make([]string, 0, 5)
			className := ""
			for _, attr := range node.Attr {
				if attr.Key == "class" {
					className = attr.Val
				}
				if strings.HasPrefix(attr.Key, "data-") || allowMap[attr.Key] {
					if tagName == "a" && attr.Key == "href" {
						if strings.Contains(attr.Val, ":") && !strings.HasPrefix(attr.Val, "http") {
							err = ERASE
							continue
						}
					}
					t := fmt.Sprintf(`%s="%s"`, attr.Key, template.HTMLEscapeString(attr.Val))
					attrs = append(attrs, t)
				} else {
					err = ERASE
				}
			}
			if f, ok := callBack[tagName+"."+className]; ok {
				f(*node)
			}
			attr := strings.Join(attrs, " ")

			switch tagName {
			case "br", "img":
				if len(attr) > 0 {
					return fmt.Sprintf("<%s %s />", tagName, attr), nil
				} else {
					return fmt.Sprintf("<%s />", tagName), nil
				}
			default:
				if len(attr) > 0 {
					res += fmt.Sprintf("<%s %s>", tagName, attr)
				} else {
					if tagName == "a" {
						for c := node.FirstChild; c != nil; c = c.NextSibling {
							r, e := traversal(c, callBack)
							if e != nil {
								err = e
							}
							res += r
						}
						// 属性なしで a タグの場合タグ自体削除
						err = ERASE
						return
					}
					// それ以外のタグは属性がなくても追加 （そういうタグがあるのか？）
					res += fmt.Sprintf("<%s>", tagName)
				}

				for c := node.FirstChild; c != nil; c = c.NextSibling {
					r, e := traversal(c, callBack)
					if e != nil {
						err = e
					}
					res += r
				}

				res += fmt.Sprintf("</%s>", tagName)
			}
		}
	case html.DocumentNode:
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			r, e := traversal(c, callBack)
			if e != nil {
				err = e
			}
			res += r
		}
	}

	return
}

func getFirstElementByTagName(node *html.Node, tagName string) *html.Node {
	if node.Type == html.ElementNode && node.Data == tagName {
		return node
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		res := getFirstElementByTagName(c, tagName)
		if res != nil {
			return res
		}
	}
	return nil
}

func Parse(src string) (template.HTML, error) {
	return ParseCallBack(src, map[string]func(node html.Node){})
}
func ParseCallBack(src string, callBack map[string]func(node html.Node)) (template.HTML, error) {
	doc, err := html.Parse(strings.NewReader(src))
	if err != nil {
		return "", err
	}

	body := getFirstElementByTagName(doc, "body")
	if body == nil {
		return "", PARSE_ERROR
	}

	res := ""
	for c := body.FirstChild; c != nil; c = c.NextSibling {
		r, e := traversal(c, callBack)
		if e != nil {
			err = e
		}
		res += r
	}

	return template.HTML(res), err
}

func GetFirstElementByTag(src string, tag string) (*html.Node, error) {
	doc, err := html.Parse(strings.NewReader(src))
	if err != nil {
		return nil, err
	}
	element := getFirstElementByTagName(doc, tag)
	return element, nil
}
