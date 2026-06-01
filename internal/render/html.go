// Package render turns review context into output artifacts.
package render

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"

	"github.com/pangobit/prbuddy/internal/review"
)

const htmlTemplate = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>PR Buddy Review Guide</title>
  <style>
    body { color: #1f2937; font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; line-height: 1.5; margin: 2rem auto; max-width: 960px; padding: 0 1rem; }
    h1, h2 { color: #111827; line-height: 1.2; }
    code, pre { background: #f3f4f6; border-radius: 6px; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
    code { padding: 0.1rem 0.3rem; }
    pre { overflow-x: auto; padding: 1rem; }
    pre.mermaid { background: #ffffff; border: 1px solid #e5e7eb; }
    table { border-collapse: collapse; width: 100%; }
    th, td { border-bottom: 1px solid #e5e7eb; padding: 0.5rem; text-align: left; vertical-align: top; }
    .muted { color: #6b7280; }
    .review { margin-top: 2rem; }
  </style>
</head>
<body>
  <main>
    <h1>PR Buddy Review Guide</h1>
    <section>
      <h2>Repository</h2>
      <p><strong>{{ .Repository.Name }}</strong></p>
      <p class="muted">{{ .Repository.Root }}</p>
      {{ if .Repository.RemoteURL }}<p><code>{{ .Repository.RemoteURL }}</code></p>{{ end }}
    </section>
    <section>
      <h2>Comparison</h2>
      <p><code>{{ .Comparison.BaseRef }}</code> to <code>{{ .Comparison.HeadRef }}</code></p>
      <p class="muted">Merge base: <code>{{ .Comparison.MergeBase }}</code></p>
    </section>
    {{ if .PullRequest }}
    <section>
      <h2>Pull Request</h2>
      <p><strong>#{{ .PullRequest.Number }} {{ .PullRequest.Title }}</strong></p>
      {{ if .PullRequest.URL }}<p><code>{{ .PullRequest.URL }}</code></p>{{ end }}
      {{ if .PullRequest.Author }}<p class="muted">Author: {{ .PullRequest.Author }}</p>{{ end }}
    </section>
    {{ end }}
    <section class="review">
      {{ .ReviewHTML }}
    </section>
    <section>
      <h2>Raw Context</h2>
      <p class="muted">{{ len .Commits }} commits, {{ len .Files }} changed files.</p>
      {{ if .Diff }}<details><summary>Diff</summary><pre><code>{{ .Diff }}</code></pre></details>{{ end }}
    </section>
  </main>
  {{ if .HasMermaid }}
  <script type="module">
    import mermaid from "https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs";
    document.querySelectorAll("pre > code.language-mermaid").forEach((code) => {
      const container = document.createElement("pre");
      container.className = "mermaid";
      container.textContent = code.textContent;
      code.parentElement.replaceWith(container);
    });
    mermaid.initialize({ startOnLoad: true });
  </script>
  {{ end }}
</body>
</html>
`

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
		template: template.Must(template.New("review").Parse(htmlTemplate)),
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

	return r.template.Execute(out, htmlTemplateData{
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
