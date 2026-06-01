# PR Buddy

PR Buddy generates a single HTML review guide for a pull request comparison
using local git context, CI metadata, and an LLM provider. The CLI is designed
to run locally or inside GitHub Actions, and it writes the HTML document to
stdout so another command or workflow step can decide what to do with it.

## Install

```bash
go install github.com/pangobit/prbuddy/cmd/prbuddy@v0.1.0
```

During local development, run the command directly from the repository:

```bash
GEMINI_API_KEY=... go run ./cmd/prbuddy render --provider gemini --base origin/main --head HEAD
```

## Usage

```bash
prbuddy render --provider gemini --base origin/main --head HEAD > pr-review.html
prbuddy render --provider grok --base origin/main --head HEAD > pr-review.html
```

To test Gemini through a trusted local proxy, override the provider URL. This
path allows an empty Gemini key because the proxy owns authentication:

```bash
prbuddy render --provider gemini --provider-url https://ai --base origin/main --head HEAD > pr-review.html
```

The command writes:

- stdout: one complete HTML document
- stderr: diagnostics and errors

PR Buddy does not save, serve, open, or publish the HTML file. Redirect stdout
when a file is useful.

Provider credentials are read from environment variables:

- `--provider gemini` requires `GEMINI_API_KEY`
- `--provider grok` requires `XAI_API_KEY`

`--provider gemini --provider-url <url>` allows `GEMINI_API_KEY` to be empty for
trusted proxy workflows.

Provider models can be overridden with `--model`, `PRBUDDY_MODEL`,
`PRBUDDY_GEMINI_MODEL`, or `PRBUDDY_GROK_MODEL`.

The provider returns Markdown. PR Buddy converts that Markdown into the final
HTML document, supports tables and Mermaid code blocks, and sanitizes rendered
Markdown before writing the document.

When running in GitHub Actions, PR Buddy reads available workflow metadata from
the default `GITHUB_*` environment variables and `$GITHUB_EVENT_PATH`. It does
not call the GitHub API, use `gh`, or fetch `diff_url`.

## GitHub Actions

```yaml
steps:
  - uses: actions/checkout@v4
    with:
      fetch-depth: 0

  - uses: actions/setup-go@v5
    with:
      go-version: "1.25.x"

  - run: go install github.com/pangobit/prbuddy/cmd/prbuddy@v0.1.0

  - name: Render PR review HTML
    id: prbuddy
    env:
      GEMINI_API_KEY: ${{ secrets.GEMINI_API_KEY }}
    run: |
      prbuddy render --provider gemini > pr-review.html
      echo "html_path=pr-review.html" >> "$GITHUB_OUTPUT"

  - name: Use rendered HTML
    run: some-other-tool "${{ steps.prbuddy.outputs.html_path }}"
```

For handoff to another job, upload the generated file as an artifact:

```yaml
- uses: actions/upload-artifact@v4
  with:
    name: pr-review-html
    path: pr-review.html
```
