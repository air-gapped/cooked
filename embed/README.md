# embed/

This directory holds static assets that are embedded into the cooked binary via `go:embed`.

Run `make deps` to download pinned versions of:

- **mermaid.min.js** — client-side Mermaid diagram rendering
- **github-markdown-light.css** — GitHub-style light theme CSS
- **github-markdown-dark.css** — GitHub-style dark theme CSS

These files are `.gitignore`d because they are downloaded artifacts. The `LICENSES.md` file (tracking third-party licenses for bundled assets) is committed.
