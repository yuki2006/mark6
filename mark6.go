package mark6

import (
	"errors"
	"fmt"
	"html/template"
	"strings"

	"golang.org/x/net/html"
)

var (
	ERASE       = errors.New("erase")
	PARSE_ERROR = errors.New("parse error")
)

// AllowTags は許可するHTMLタグとその属性のホワイトリストを表す型
type AllowTags map[string]map[string]bool

// AllowAttrs は許可する属性のセットを生成するヘルパー関数
func AllowAttrs(attrs ...string) map[string]bool {
	mp := make(map[string]bool)
	for _, attr := range attrs {
		mp[attr] = true
	}
	return mp
}

func traversal(node *html.Node, allowTags AllowTags, callBack map[string]func(node html.Node)) (res string, err error) {

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
			id := ""
			for _, attr := range node.Attr {
				if attr.Key == "class" {
					className = attr.Val
				} else if attr.Key == "id" {
					id = attr.Val
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
			// elseではない
			if f, ok := callBack[tagName+"#"+id]; ok {
				f(*node)
			}
			attr := strings.Join(attrs, " ")

			switch tagName {
			case "br", "hr", "img":
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
							r, e := traversal(c, allowTags, callBack)
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
					r, e := traversal(c, allowTags, callBack)
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
			r, e := traversal(c, allowTags, callBack)
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

func Parse(src string, allowTags AllowTags) (template.HTML, error) {
	return ParseCallBack(src, allowTags, map[string]func(node html.Node){})
}
func ParseCallBack(src string, allowTags AllowTags, callBack map[string]func(node html.Node)) (template.HTML, error) {
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
		r, e := traversal(c, allowTags, callBack)
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
