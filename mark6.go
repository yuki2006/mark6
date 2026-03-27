package mark6

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"strings"

	"golang.org/x/net/html"
)

var (
	ERASE       = errors.New("erase")
	PARSE_ERROR = errors.New("parse error")
)

// EraseDetail は削除された要素の詳細を表す
type EraseDetail struct {
	Reason string // "disallowed_attr", "dangerous_href", "empty_anchor"
	Tag    string // タグ名
	Attr   string // 属性名（該当する場合）
	Value  string // 属性値（該当する場合）
}

func (d EraseDetail) String() string {
	tag := "&lt;" + d.Tag + "&gt;"
	switch d.Reason {
	case "disallowed_tag":
		return tag + "タグは許可されていません"
	case "disallowed_attr":
		return tag + "タグの属性「" + d.Attr + "」は許可されていません"
	case "dangerous_href":
		return tag + "タグのhref「" + d.Value + "」は許可されていません（httpで始まるURLのみ使用可）"
	case "empty_anchor":
		return tag + "タグにhref属性がないため削除されました"
	default:
		return tag + "タグの一部が削除されました"
	}
}

// EraseError は削除が発生した場合のエラー。詳細情報を含む。
type EraseError struct {
	Details []EraseDetail
}

func (e *EraseError) Error() string {
	return "erase"
}

// Is は errors.Is でERASEと一致させるため
func (e *EraseError) Is(target error) bool {
	return target == ERASE
}

// FormatDetails は削除された内容を人間が読みやすい文字列にする
func (e *EraseError) FormatDetails() string {
	if len(e.Details) == 0 {
		return ""
	}
	seen := make(map[string]bool)
	var lines []string
	for _, d := range e.Details {
		s := d.String()
		if !seen[s] {
			seen[s] = true
			lines = append(lines, "・"+s)
		}
	}
	return strings.Join(lines, "<br>")
}

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

func traversal(node *html.Node, allowTags AllowTags, callBack map[string]func(node html.Node) *string, eraseErr *EraseError) (res string, err error) {

	res = ""

	switch node.Type {
	case html.TextNode:
		return template.HTMLEscapeString(node.Data), nil
	case html.ElementNode:
		tagName := strings.ToLower(node.Data)
		allowMap, found := allowTags[tagName]

		if !found {
			err = ERASE
			eraseErr.Details = append(eraseErr.Details, EraseDetail{
				Reason: "disallowed_tag", Tag: tagName,
			})
		}
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
							eraseErr.Details = append(eraseErr.Details, EraseDetail{
								Reason: "dangerous_href", Tag: tagName, Attr: "href", Value: attr.Val,
							})
							continue
						}
					}
					t := fmt.Sprintf(`%s="%s"`, attr.Key, template.HTMLEscapeString(attr.Val))
					attrs = append(attrs, t)
				} else {
					err = ERASE
					eraseErr.Details = append(eraseErr.Details, EraseDetail{
						Reason: "disallowed_attr", Tag: tagName, Attr: attr.Key, Value: attr.Val,
					})
				}
			}
			if f, ok := callBack[tagName+"."+className]; ok {
				if s := f(*node); s != nil {
					return *s, nil
				}
			}
			// elseではない
			if f, ok := callBack[tagName+"#"+id]; ok {
				if s := f(*node); s != nil {
					return *s, nil
				}
			}
			if f, ok := callBack[tagName]; ok {
				if s := f(*node); s != nil {
					return *s, nil
				}
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
							r, e := traversal(c, allowTags, callBack, eraseErr)
							if e != nil {
								err = e
							}
							res += r
						}
						// 属性なしで a タグの場合タグ自体削除
						err = ERASE
						eraseErr.Details = append(eraseErr.Details, EraseDetail{
							Reason: "empty_anchor", Tag: tagName,
						})
						return
					}
					// それ以外のタグは属性がなくても追加 （そういうタグがあるのか？）
					res += fmt.Sprintf("<%s>", tagName)
				}

				for c := node.FirstChild; c != nil; c = c.NextSibling {
					r, e := traversal(c, allowTags, callBack, eraseErr)
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
			r, e := traversal(c, allowTags, callBack, eraseErr)
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
	return ParseReader(strings.NewReader(src), allowTags)
}

func ParseCallBack(src string, allowTags AllowTags, callBack map[string]func(node html.Node) *string) (template.HTML, error) {
	return ParseCallBackReader(strings.NewReader(src), allowTags, callBack)
}

func ParseReader(r io.Reader, allowTags AllowTags) (template.HTML, error) {
	return ParseCallBackReader(r, allowTags, map[string]func(node html.Node) *string{})
}

func ParseCallBackReader(r io.Reader, allowTags AllowTags, callBack map[string]func(node html.Node) *string) (template.HTML, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", err
	}

	body := getFirstElementByTagName(doc, "body")
	if body == nil {
		return "", PARSE_ERROR
	}

	eraseErr := &EraseError{}
	res := ""
	for c := body.FirstChild; c != nil; c = c.NextSibling {
		r, e := traversal(c, allowTags, callBack, eraseErr)
		if e != nil {
			err = e
		}
		res += r
	}

	// ERASEが発生した場合、詳細付きのEraseErrorを返す
	if errors.Is(err, ERASE) && len(eraseErr.Details) > 0 {
		err = eraseErr
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
