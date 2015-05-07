package mark6

import (
	"code.google.com/p/go.net/html"
	"fmt"
	"html/template"
	"regexp"
	"strings"
)

func allowAttrs(attrs ...string) map[string]bool {
	mp := make(map[string]bool)
	for _, attr := range attrs {
		mp[attr] = true
	}
	return mp
}

var allowTags = map[string]map[string]bool {
	"a" : allowAttrs("href"),
	"b" : allowAttrs(),
	"i" : allowAttrs("class"),
	"p" : allowAttrs(),
	"br" : allowAttrs(),
	"h1" : allowAttrs("class"),
	"h2" : allowAttrs("class"),
	"h3" : allowAttrs("class"),
	"h4" : allowAttrs("class"),
	"h5" : allowAttrs("class"),
	"h6" : allowAttrs("class"),
	"span" : allowAttrs(),
	"div" : allowAttrs("class"),
	"pre" : allowAttrs(),
	"img" : allowAttrs("src", "alt"),
	"ul" : allowAttrs(),
	"ol" : allowAttrs(),
	"li" : allowAttrs(),
	"table" : allowAttrs("class"),
	"thead" : allowAttrs(),
	"tr" : allowAttrs(),
	"th" : allowAttrs("data-defaultsort"),
	"tbody" : allowAttrs(),
	"td" : allowAttrs("class"),
	"strong" : allowAttrs(),
	"em" : allowAttrs(),
	"code" : allowAttrs(),
	"dl" : allowAttrs(),
	"dt" : allowAttrs(),
	"dd" : allowAttrs(),
	"del" : allowAttrs(),
	"sup" : allowAttrs(),
	"sub" : allowAttrs(),
	"u" : allowAttrs(),
	"backquote":allowAttrs(),
	"s":allowAttrs(),
}

func traversal(node *html.Node) string {
	javascriptProtocolChecker := regexp.MustCompile("^\\s*javascript:")
	res := ""

	switch node.Type {
	case html.TextNode :
		return template.HTMLEscapeString(node.Data)
	case html.ElementNode :
		tagName := node.Data
		allowMap, found := allowTags[tagName]

		if found {
			attrs := make([]string, 0, 5)
			for _, attr := range node.Attr {
				if allowMap[attr.Key] {
					if tagName == "a" && attr.Key == "href" {
						if javascriptProtocolChecker.MatchString(attr.Val) {
							continue
						}
					}
					t := fmt.Sprintf(`%s="%s"`, attr.Key, template.HTMLEscapeString(attr.Val))
					attrs = append(attrs, t)
				}
			}
			attr := strings.Join(attrs, " ")

			switch tagName {
			case "br", "img":
				if len(attr) > 0 {
					return fmt.Sprintf("<%s %s />", tagName, attr)
				} else {
					return fmt.Sprintf("<%s />", tagName)
				}
			default:
				if len(attr) > 0 {
					res += fmt.Sprintf("<%s %s>", tagName, attr)
				} else {
					res += fmt.Sprintf("<%s>", tagName)
				}
			}

			for c := node.FirstChild; c != nil; c = c.NextSibling {
				res += traversal(c)
			}

			res += fmt.Sprintf("</%s>", tagName)
		}
	case html.DocumentNode :
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			res += traversal(c)
		}
	}

	return res
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
	doc, err := html.Parse(strings.NewReader(src))
	if err != nil {
		return "", err
	}

	body := getFirstElementByTagName(doc, "body")
	if body == nil {
		return "", fmt.Errorf("parse error")
	}

	res := ""
	for c := body.FirstChild; c != nil; c = c.NextSibling {
		res += traversal(c)
	}

	return template.HTML(res), nil
}

func GetFirstElementByTag(src string, tag string) (*html.Node, error) {
	doc, err := html.Parse(strings.NewReader(src))
	if err != nil {
		return nil, err
	}
	element := getFirstElementByTagName(doc, tag)
	return element, nil
}


