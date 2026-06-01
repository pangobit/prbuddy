// Package render turns review context into output artifacts.
package render

import (
	"bytes"
	"context"
	"embed"
	"html/template"
	"io"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"

	"github.com/pangobit/prbuddy/internal/review"
)

//go:embed templates/*.tmpl
var htmlTemplateFS embed.FS

var reviewHTMLTemplate = template.Must(template.ParseFS(htmlTemplateFS, "templates/*.tmpl"))

type htmlTemplateData struct {
	review.Context
	ReviewHTML template.HTML
	HasMermaid bool
}

// HTMLRenderer renders review markdown as a complete HTML document.
type HTMLRenderer struct {
	template *template.Template
	markdown goldmark.Markdown
	policy   *bluemonday.Policy
}

// NewHTMLRenderer creates an HTML renderer.
func NewHTMLRenderer() *HTMLRenderer {
	return &HTMLRenderer{
		template: reviewHTMLTemplate,
		markdown: goldmark.New(
			goldmark.WithExtensions(extension.GFM),
		),
		policy: markdownPolicy(),
	}
}

// Render writes a complete HTML document to out.
func (r *HTMLRenderer) Render(_ context.Context, out io.Writer, ctx review.Context, markdown string) error {
	reviewHTML, err := r.renderMarkdown(markdown)
	if err != nil {
		return err
	}

	return r.template.ExecuteTemplate(out, "review.html.tmpl", htmlTemplateData{
		Context:    ctx,
		ReviewHTML: reviewHTML,
		HasMermaid: hasMermaid(markdown),
	})
}

func (r *HTMLRenderer) renderMarkdown(markdown string) (template.HTML, error) {
	var out bytes.Buffer
	if err := r.markdown.Convert([]byte(markdown), &out); err != nil {
		return "", err
	}

	return template.HTML(r.policy.Sanitize(out.String())), nil
}

func markdownPolicy() *bluemonday.Policy {
	policy := bluemonday.UGCPolicy()
	policy.AllowAttrs("class").Matching(bluemonday.SpaceSeparatedTokens).OnElements("code")
	policy.AllowElements("table", "thead", "tbody", "tr", "th", "td")
	policy.AllowAttrs("align").OnElements("th", "td")

	return policy
}

func hasMermaid(markdown string) bool {
	return strings.Contains(markdown, "```mermaid")
}
