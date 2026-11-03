package llms

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	markdownRemove = []atom.Atom{
		atom.Script,
		atom.Style,
		atom.Noscript,
		atom.Template,
		atom.Slot,
		atom.Nav,
		atom.Form,
		atom.Iframe,
		atom.Svg,
		atom.Button,
	}
)

type markdownPage struct {
	Title       string
	Description string
	Markdown    []byte
}

func htmlToMarkdown(data []byte, selector, domain string) (*markdownPage, error) {
	doc, err := html.Parse(bytes.NewReader(data))

	if err != nil {
		return nil, err
	}

	page := new(markdownPage)

	if title := findNode(doc, func(n *html.Node) bool { return n.DataAtom == atom.Title }); title != nil {
		page.Title = strings.TrimSpace(textContent(title))
	}

	if meta := findNode(doc, func(n *html.Node) bool {
		return n.DataAtom == atom.Meta && strings.EqualFold(attr(n, "name"), "description")
	}); meta != nil {
		page.Description = strings.TrimSpace(attr(meta, "content"))
	}

	node := findNode(doc, selectorMatcher(selector))

	if node == nil {
		node = findNode(doc, func(n *html.Node) bool { return n.DataAtom == atom.Body })

		// without the selected element, the page header and footer (direct children of the body) are dropped
		if node != nil {
			for c := node.FirstChild; c != nil; {
				next := c.NextSibling

				if c.DataAtom == atom.Header || c.DataAtom == atom.Footer {
					node.RemoveChild(c)
				}

				c = next
			}
		}
	}

	if node == nil {
		node = doc
	}

	removeNodes(node, func(n *html.Node) bool { return slices.Contains(markdownRemove, n.DataAtom) })
	content, err := htmltomarkdown.ConvertNode(node, converter.WithDomain(domain))

	if err != nil {
		return nil, err
	}

	content = bytes.TrimSpace(content)
	var buffer bytes.Buffer

	// the description follows the h1, which is either the page title or the first h1 of the content
	if bytes.HasPrefix(content, []byte("# ")) {
		headline, rest, _ := bytes.Cut(content, []byte("\n"))
		buffer.Write(headline)
		buffer.WriteString("\n\n")
		content = bytes.TrimSpace(rest)
	} else if page.Title != "" {
		buffer.WriteString(fmt.Sprintf("# %s\n\n", oneLine(page.Title)))
	}

	if page.Description != "" {
		buffer.WriteString(fmt.Sprintf("> %s\n\n", oneLine(page.Description)))
	}

	buffer.Write(content)
	page.Markdown = append(bytes.TrimSpace(buffer.Bytes()), '\n')
	return page, nil
}

func selectorMatcher(selector string) func(*html.Node) bool {
	selector = strings.TrimSpace(selector)

	return func(n *html.Node) bool {
		if n.Type != html.ElementNode {
			return false
		}

		if id, ok := strings.CutPrefix(selector, "#"); ok {
			return attr(n, "id") == id
		}

		if class, ok := strings.CutPrefix(selector, "."); ok {
			return slices.Contains(strings.Fields(attr(n, "class")), class)
		}

		return strings.EqualFold(n.Data, selector)
	}
}

func findNode(n *html.Node, match func(*html.Node) bool) *html.Node {
	if n.Type == html.ElementNode && match(n) {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findNode(c, match); found != nil {
			return found
		}
	}

	return nil
}

func removeNodes(n *html.Node, match func(*html.Node) bool) {
	for c := n.FirstChild; c != nil; {
		next := c.NextSibling

		if c.Type == html.ElementNode && match(c) {
			n.RemoveChild(c)
		} else {
			removeNodes(c, match)
		}

		c = next
	}
}

func textContent(n *html.Node) string {
	var sb strings.Builder

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			sb.WriteString(c.Data)
		} else {
			sb.WriteString(textContent(c))
		}
	}

	return sb.String()
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}

	return ""
}

func oneLine(str string) string {
	return strings.Join(strings.Fields(str), " ")
}
