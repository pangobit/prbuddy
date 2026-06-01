package app

import (
	"bytes"
	"embed"
	"strings"
	"text/template"

	"github.com/pangobit/prbuddy/internal/llm"
	"github.com/pangobit/prbuddy/internal/review"
)

//go:embed templates/*.tmpl
var promptTemplateFS embed.FS

var reviewPromptTemplates = template.Must(template.New("review-prompts").Funcs(template.FuncMap{
	"emptyFallback": emptyFallback,
	"shortHash":     shortHash,
}).ParseFS(promptTemplateFS, "templates/*.tmpl"))

// BuildPrompt creates the review prompt for an LLM review guide.
func BuildPrompt(ctx review.Context) llm.Prompt {
	return llm.Prompt{
		System: renderPromptTemplate("review_system.md.tmpl", ctx),
		User:   renderPromptTemplate("review_user.md.tmpl", ctx),
	}
}

func renderPromptTemplate(name string, ctx review.Context) string {
	var out bytes.Buffer
	if err := reviewPromptTemplates.ExecuteTemplate(&out, name, ctx); err != nil {
		panic(err)
	}

	return strings.TrimSpace(out.String())
}

func emptyFallback(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}

func shortHash(hash string) string {
	if len(hash) <= 7 {
		return hash
	}

	return hash[:7]
}
