package emailclient

import (
	"bytes"
	"context"
	"strings"

	jobtemporal "webauthn-server/internal/jobs/temporal"

	"golang.org/x/net/html"
)

type EmailClient interface {
	FetchRecentMetadata(ctx context.Context, email, bankName, cardLast4 string) ([]jobtemporal.EmailMetadata, error)
	FetchFullBodyByUID(ctx context.Context, email string, uid uint32) (string, error)
}

func extractTextFromHTML(htmlBytes []byte) string {
	doc, err := html.Parse(bytes.NewReader(htmlBytes))
	if err != nil {
		return cleanText(string(htmlBytes))
	}

	var buf strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && (node.Data == "style" || node.Data == "script" || node.Data == "head") {
			return
		}
		if node.Type == html.TextNode {
			if text := cleanText(node.Data); text != "" {
				buf.WriteString(text)
				buf.WriteString("\n")
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return buf.String()
}

func cleanText(input string) string {
	input = strings.ReplaceAll(input, "\ufffd", " ")
	input = strings.ReplaceAll(input, "\u00a0", " ")
	return strings.TrimSpace(input)
}

func snippet(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) > 100 {
		return value[:100]
	}
	return value
}
