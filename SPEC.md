# cooked

A rendering proxy that takes a URL to a raw document (markdown, etc.) and serves it as styled HTML. The opposite of "raw" — you give it a raw file URL, it gives you the cooked version.

**The binary must be fully self-contained.** All CSS, JavaScript (including mermaid.js), fonts, favicon, and templates must be embedded into the Go binary via `go:embed`. No CDN requests, no external resources, no network calls except to the upstream URL. This thing runs in air-gapped environments.

## Overview

```
Browser request:
  https://cooked.example.com/https://cgit.internal/github.com/cilium/cilium/plain/README.md

cooked does:
  1. Fetches https://cgit.internal/github.com/cilium/cilium/plain/README.md
  2. Detects file type from extension (.md)
  3. Renders markdown → HTML with styling
  4. Returns styled, self-contained HTML page to browser
```

The user sends people links like:

```
https://cooked.example.com/https://raw.example.com/path/to/README.md
https://cooked.example.com/https://s3.internal/mirror-bucket/project/docs/INSTALL.md
https://cooked.example.com/https://cgit.internal/github.com/cilium/cilium/plain/README.md
```

The upstream URL is everything after the first `/` in the path. cooked is completely agnostic about the upstream — it just fetches whatever URL you give it and renders the result.

## Language & Dependencies

- **Go** (single static binary, no runtime dependencies)
- **github.com/yuin/goldmark** — CommonMark + GFM markdown parser
- **github.com/yuin/goldmark-highlighting** — syntax highlighting for fenced code blocks (uses chroma)
- **go.abhg.dev/goldmark/mermaid** — mermaid diagram support (client-side rendering mode)
- **github-markdown-css** — GitHub-style CSS (both light and dark variants, embedded)
- **mermaid.js** — vendored/embedded into the binary (download a minified copy at build time)

All static assets (CSS, JS, SVG favicon, HTML templates) must be embedded via `go:embed`. The rendered HTML pages must never make external network requests.

## URL Scheme

```
https://cooked.example.com/{upstream_url}
```

The `{upstream_url}` is a full URL including scheme:

```
/https://cgit.internal/repo/plain/README.md
/https://s3.internal/bucket/path/file.md
/http://some-internal-server/docs/guide.md
```

cooked extracts the upstream URL from the request path by stripping the leading `/` and using everything that follows.

### Query strings

Query strings on the cooked URL are passed through to the upstream. This is important for S3 presigned URLs and similar auth mechanisms:

```
/https://s3.internal/bucket/file.md?X-Amz-Signature=abc123
```

cooked's own parameters (like theme selection) use a reserved prefix: `_cooked_theme=dark`.

### Fragment / anchor support

URL fragments (`#some-heading`) are client-side and just work — the browser handles them after the page loads. Auto-generated heading IDs ensure anchors work.

### Special paths

- `GET /` — simple landing page explaining what cooked is, with a text input field where you can paste a raw URL
- `GET /healthz` — returns 200 OK (for k8s probes)
- `GET /_cooked/` — reserved namespace for cooked's own embedded assets (mermaid.min.js, etc.)

## HTML Testability & Inspectability

The rendered HTML must be easy to inspect, query, and assert against using browser DevTools, Puppeteer, Playwright, or any DOM inspection tool. This is critical for automated testing and CI pipelines.

### Semantic structure

- Use proper semantic HTML5 elements: `<header>`, `<nav>`, `<main>`, `<article>`, `<footer>`, `<section>`
- Every functional region must have a unique `id` attribute
- The DOM tree must be clean and shallow — no unnecessary wrapper `<div>` nesting

### Required IDs

Every cooked page must have these elements queryable by ID:

| ID | Element | Contains |
|----|---------|----------|
| `#cooked-header` | `<header>` | The metadata bar |
| `#cooked-source-link` | `<a>` | The upstream URL link |
| `#cooked-modified` | `<span>` or `<time>` | Last-Modified display |
| `#cooked-size` | `<span>` | File size display |
| `#cooked-type` | `<span>` | File type label |
| `#cooked-theme-toggle` | `<button>` | Theme toggle button |
| `#cooked-toc-toggle` | `<button>` | TOC toggle button (absent if no TOC) |
| `#cooked-toc` | `<nav>` | Table of contents (absent if no TOC) |
| `#cooked-content` | `<article>` | The rendered document content |
| `#cooked-error` | `<div>` | Error message (only on error pages) |

### Data attributes for test hooks

Use `data-*` attributes to expose state and metadata to automated tools:

```html
<html data-theme="auto|light|dark"
      data-cooked-version="1.0.0"
      data-upstream-url="https://..."
      data-content-type="markdown|mdx|code|plaintext|unsupported"
      data-cache-status="hit|miss|revalidated|expired">

<header id="cooked-header"
        data-upstream-status="200"
        data-file-size="14832"
        data-last-modified="2025-02-06T12:00:00Z">

<article id="cooked-content"
         data-has-mermaid="true|false"
         data-has-toc="true|false"
         data-heading-count="12"
         data-code-block-count="5">
```

These attributes make it trivial to write test assertions like:

```javascript
// Playwright / Puppeteer
expect(await page.getAttribute('html', 'data-content-type')).toBe('markdown');
expect(await page.getAttribute('#cooked-content', 'data-has-toc')).toBe('true');
expect(await page.getAttribute('html', 'data-cache-status')).toBe('miss');
```

### Code blocks

Every rendered code block must have:

```html
<div class="cooked-code-block" data-language="python" data-line-count="42">
  <div class="cooked-code-header">
    <span class="cooked-code-language">python</span>
    <button class="cooked-copy-btn" data-state="idle|copied">Copy</button>
  </div>
  <pre><code class="language-python">...</code></pre>
</div>
```

The `data-language` and `data-line-count` attributes make it easy to find and validate specific code blocks. The `data-state` on the copy button tracks whether it was just clicked.

### TOC entries

```html
<nav id="cooked-toc">
  <ul>
    <li data-level="1"><a href="#installation">Installation</a></li>
    <li data-level="2"><a href="#prerequisites">Prerequisites</a></li>
    <li data-level="2"><a href="#steps">Steps</a></li>
  </ul>
</nav>
```

### Error pages

Error pages must use the same template and be equally inspectable:

```html
<html data-content-type="error"
      data-upstream-url="https://..."
      data-error-type="upstream-error|timeout|too-large|blocked|unreachable">
<div id="cooked-error"
     data-status-code="404"
     data-error-message="Not Found">
```

### CSS class naming

All cooked-specific CSS classes use the `cooked-` prefix to avoid collisions with github-markdown-css or upstream HTML content. The content inside `#cooked-content` uses github-markdown-css classes (`markdown-body`), everything outside uses `cooked-*` classes.

### Clean DOM guarantee

- No inline styles on elements (all styling via classes and the embedded `<style>` block) except where chroma syntax highlighting requires it
- No minified/mangled class names — everything is human-readable
- Proper indentation in the HTML output (not minified) — readability over bytes
- Comments in the HTML marking the major sections:
  ```html
  <!-- cooked: header -->
  <!-- cooked: table of contents -->
  <!-- cooked: content -->
  <!-- cooked: scripts -->
  ```

## Rendering

### Markdown (.md, .markdown, .mdown, .mkd, .mdx)

Use goldmark with the following extensions enabled:

- **GFM** (tables, strikethrough, autolinks, task lists)
- **Footnotes**
- **Definition lists**
- **Syntax highlighting** via goldmark-highlighting/chroma (inline CSS classes, styles embedded in page)
- **Mermaid diagrams** via goldmark-mermaid in client-side mode (mermaid.js served from `/_cooked/mermaid.min.js`, embedded in binary)
- **Typographer** (smart quotes, em-dashes, etc.)
- **Auto heading IDs** (so `## Foo Bar` gets `id="foo-bar"` for anchor links)
- **Unsafe HTML** enabled (many READMEs contain inline HTML)

#### MDX preprocessing (.mdx)

MDX files are Markdown + JSX, widely used in documentation sites (Docusaurus, Next.js, Nextra, Gatsby). cooked cannot execute React components, but most of the content in MDX files is standard markdown. Before passing to goldmark, preprocess MDX files:

1. **Strip `import` statements**: Remove lines matching `import ... from '...'` or `import "..."`.
2. **Strip `export` statements**: Remove `export default`, `export const`, etc. (often used for metadata or layout wrappers).
3. **Preserve YAML frontmatter**: Pass through the `---` frontmatter block (strip it from rendered output but optionally use `title` for the page `<title>`).
4. **Handle JSX component blocks**: JSX tags like `<Tabs>`, `<TabItem>`, `<Callout>`, `<CodeBlock>`, etc. cannot be rendered. Strategy:
   - **Self-closing tags** (`<Component />`): remove silently.
   - **Tags wrapping plain text/markdown** (`<TabItem label="npm">content</TabItem>`): strip the tags, keep the inner content. If there is a `label`, `title`, or `value` attribute, render it as a small bold heading before the content.
   - **Tags wrapping only other JSX** with no renderable content: remove the entire block silently.
5. **JSX expressions**: Strip curly-brace expressions like `{props.foo}` or `{<SomeComponent />}`.

The goal is graceful degradation — the reader gets all the prose, code blocks, and standard markdown, with JSX interactive widgets stripped cleanly rather than rendered as broken HTML. This covers 80-90% of the useful content in typical MDX documentation files.

### Plain text and code files

For known code extensions, render inside a `<pre><code>` block with syntax highlighting via chroma. Detect language from file extension.

Common extensions to handle:
`.txt`, `.text`, `.log`, `.conf`, `.cfg`, `.ini`, `.env`,
`.yaml`, `.yml`, `.json`, `.toml`, `.xml`, `.csv`,
`.sh`, `.bash`, `.zsh`, `.fish`,
`.py`, `.go`, `.js`, `.ts`, `.rs`, `.c`, `.h`, `.cpp`, `.hpp`, `.java`, `.rb`, `.lua`, `.pl`,
`.sql`, `.graphql`,
`.dockerfile`, `Dockerfile`, `Makefile`, `Jenkinsfile`,
`.tf`, `.hcl`,
`.diff`, `.patch`

If extension is unknown or ambiguous, fall back to plain monospace text rendering.

### Unsupported / binary files

Return a simple HTML page saying "This file type is not supported for rendering" with a direct link to the upstream URL so the user can download/view it directly.

## HTML Template & Theming

### Theme modes

cooked supports three theme modes: **auto** (follows system preference), **light**, and **dark**.

- Default is **auto**, which uses CSS `prefers-color-scheme` media query
- A small theme toggle button in the header cycles through: auto → light → dark → auto
- Theme preference is stored in a cookie (`_cooked_theme`) so it persists across page loads
- Can also be set via query parameter `_cooked_theme=dark|light|auto`
- Both light and dark variants of github-markdown-css must be embedded
- Chroma syntax highlighting needs both a light and dark theme too

### Document metadata header

At the top of every rendered page, show a subtle metadata header bar containing:

- **Source URL**: clickable link to the upstream raw file
- **Last Modified**: from the upstream's `Last-Modified` HTTP response header (if provided). Format as human-readable relative time ("2 hours ago") with full ISO timestamp on hover via title attribute. If no Last-Modified header, omit this field.
- **File size**: from `Content-Length` or measured from response body. Human-readable (e.g. "14.2 KB").
- **File type**: e.g. "Markdown", "Python", "YAML"
- **Theme toggle**: auto/light/dark switcher (sun/moon/auto icon)
- **TOC toggle**: hamburger/list icon (only shown if document has 3+ headings)

The header should be visually subtle — thin bar, small font, slightly different background from the document body. It should be sticky (stays at top on scroll). It should not distract from the document content.

In print: the header is hidden.

### Table of contents

For markdown documents, generate a table of contents from headings:

- Show as a toggleable sidebar or dropdown, activated by the TOC button in the header
- Only show the TOC button if the document has 3+ headings
- TOC entries are clickable anchor links that scroll to the heading
- Indent based on heading level
- Clicking a TOC entry closes the TOC on mobile

### Code block features

- **Copy button**: Small clipboard icon in the top-right corner of every `<pre>` code block. On click, copies the code text content to clipboard and shows brief "Copied!" feedback (change icon or show tooltip for ~2 seconds). Pure inline JS, no external dependencies.
- **Language label**: Show the detected language name as a small label in the top-right corner of fenced code blocks (e.g. "python", "yaml"), next to the copy button.

### Page template structure

```html
<!DOCTYPE html>
<html lang="en"
      data-theme="auto"
      data-cooked-version="{version}"
      data-upstream-url="{upstream_url}"
      data-content-type="{content_type}"
      data-cache-status="{cache_status}">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{filename} — cooked</title>
  <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,{inline_svg_favicon}">
  <style>
    /* Theme: light */
    [data-theme="light"] { ... }
    /* Theme: dark */
    [data-theme="dark"] { ... }
    /* Theme: auto (system preference) */
    @media (prefers-color-scheme: dark) {
      [data-theme="auto"] { ... }
    }

    {github-markdown-css-light-and-dark}
    {chroma-syntax-theme-light-and-dark}
    {cooked-layout-styles}

    @media print {
      .cooked-header, .cooked-toc { display: none !important; }
      .markdown-body { max-width: 100%; padding: 0; }
    }
  </style>
</head>
<body>
  <!-- cooked: header -->
  <header id="cooked-header"
          data-upstream-status="{upstream_status}"
          data-file-size="{file_size_bytes}"
          data-last-modified="{iso_timestamp}">
    <div class="cooked-meta">
      <a id="cooked-source-link" href="{upstream_url}" title="{upstream_url}">{upstream_url_truncated}</a>
      <time id="cooked-modified" datetime="{iso_timestamp}" title="{iso_timestamp}">Modified {relative_time}</time>
      <span id="cooked-size">{file_size_human}</span>
      <span id="cooked-type">{file_type_label}</span>
    </div>
    <div class="cooked-controls">
      <button id="cooked-toc-toggle" title="Table of contents">☰</button>
      <button id="cooked-theme-toggle" title="Toggle theme">◑</button>
    </div>
  </header>

  <!-- cooked: table of contents -->
  <nav id="cooked-toc" hidden>
    {table_of_contents_html}
  </nav>

  <!-- cooked: content -->
  <main>
    <article id="cooked-content"
             class="markdown-body"
             data-has-mermaid="{has_mermaid}"
             data-has-toc="{has_toc}"
             data-heading-count="{heading_count}"
             data-code-block-count="{code_block_count}">
      {rendered_content}
    </article>
  </main>

  <!-- cooked: scripts -->
  <script>
    // Inline JS for:
    // - Theme toggle (cycle auto→light→dark→auto, set cookie, update data-theme attr)
    // - TOC toggle (show/hide sidebar)
    // - Copy buttons on code blocks (find all pre>code, inject button, clipboard API)
    // - Relative time display (optional: update "2 hours ago" periodically)
  </script>
  <!-- Only included if the markdown contains ```mermaid blocks -->
  <script src="/_cooked/mermaid.min.js"></script>
  <script>mermaid.initialize({startOnLoad: true, theme: 'default'});</script>
</body>
</html>
```

### Layout

- Center the `.markdown-body` with `max-width: 980px` and appropriate padding (same as GitHub)
- Responsive: readable on mobile, header wraps gracefully
- The upstream URL in the header should be truncated with ellipsis if too long, full URL on hover
- Favicon: embed a minimal SVG — a simple cooking pot icon or stylized "C"

## Relative URL Rewriting

This is critical for usability. Markdown files often contain relative links and images:

```markdown
![architecture](./docs/arch.png)
See [CONTRIBUTING](CONTRIBUTING.md) for details.
```

cooked must rewrite these relative URLs so they resolve correctly. There are two strategies depending on the link target:

1. **Markdown files** (`.md`, `.markdown`, `.mdx`, etc.) — rewrite to go through cooked so they also get rendered:
   ```
   CONTRIBUTING.md → /https://cgit.internal/repo/plain/CONTRIBUTING.md
   ```

2. **Non-markdown files** (images, etc.) — rewrite to point directly at the upstream server:
   ```
   ./docs/arch.png → https://cgit.internal/repo/plain/docs/arch.png
   ```

The base URL for resolving relative links is derived from the upstream URL by stripping the filename component.

### Absolute URLs in markdown

Absolute URLs (`https://...`) in markdown content are left untouched.

## Caching

Implement a simple in-memory cache:

- Cache key: the full upstream URL (including query string)
- Cache value: the rendered HTML + the ETag/Last-Modified from the upstream response (if provided)
- Default TTL: 5 minutes (configurable via flag/env)
- Maximum cache size: 100 MB (configurable), evict LRU when full
- On cache hit: if the upstream provided ETag or Last-Modified, do a conditional GET (If-None-Match / If-Modified-Since) to validate. If 304 Not Modified, serve from cache and reset TTL.
- On cache miss: fetch, render, cache, serve.
- `Cache-Control` response header: set `public, max-age=300` (matching TTL)

### Response headers

cooked adds custom response headers for programmatic inspection and debugging:

| Header | Example | Description |
|--------|---------|-------------|
| `X-Cooked-Version` | `1.0.0` | cooked version |
| `X-Cooked-Upstream` | `https://cgit.internal/...` | The upstream URL that was fetched |
| `X-Cooked-Upstream-Status` | `200` | HTTP status from upstream |
| `X-Cooked-Cache` | `hit` | Cache status: `hit`, `miss`, `revalidated`, `expired` |
| `X-Cooked-Content-Type` | `markdown` | Detected content type |
| `X-Cooked-Render-Ms` | `12` | Time spent rendering (excluding upstream fetch) |
| `X-Cooked-Upstream-Ms` | `45` | Time spent fetching upstream (0 on cache hit) |

These headers make it trivial to verify behavior from scripts or CI:

```bash
# Quick check from CLI
curl -sI https://cooked.example.com/https://cgit.internal/repo/plain/README.md | grep X-Cooked
```

## Configuration

All configuration via CLI flags with environment variable fallback:

| Flag | Env | Default | Description |
|------|-----|---------|-------------|
| `--listen` | `COOKED_LISTEN` | `:8080` | Listen address |
| `--cache-ttl` | `COOKED_CACHE_TTL` | `5m` | Cache TTL duration |
| `--cache-max-size` | `COOKED_CACHE_MAX_SIZE` | `100MB` | Max cache size |
| `--fetch-timeout` | `COOKED_FETCH_TIMEOUT` | `30s` | Upstream fetch timeout |
| `--max-file-size` | `COOKED_MAX_FILE_SIZE` | `5MB` | Max file size to render |
| `--allowed-upstreams` | `COOKED_ALLOWED_UPSTREAMS` | `` (allow all) | Comma-separated allowed upstream host prefixes (e.g. `cgit.internal,s3.internal`) |
| `--base-url` | `COOKED_BASE_URL` | `` (auto-detect from Host header) | Public base URL of cooked (for generating self-referencing links) |
| `--default-theme` | `COOKED_DEFAULT_THEME` | `auto` | Default theme: `auto`, `light`, or `dark` |
| `--tls-skip-verify` | `COOKED_TLS_SKIP_VERIFY` | `false` | Disable TLS certificate verification for upstream fetches |

## Security

- **Allowed upstreams**: If `--allowed-upstreams` is set, only fetch from those hosts. Reject all others with 403. This prevents cooked from being used as an open proxy.
- **Max file size**: Refuse to fetch/render files larger than `--max-file-size`. Return 413.
- **Fetch timeout**: Hard timeout on upstream fetches.
- **No credentials forwarding**: cooked never forwards cookies, auth headers, or any credentials from the browser request to the upstream.
- **HTML sanitization**: Strip dangerous elements from upstream markdown content before rendering: `<script>`, `<iframe>`, `<object>`, `<embed>`, `<form>`, `<input>`, and all event handler attributes (`onclick`, `onerror`, `onload`, etc.). cooked's own scripts (theme toggle, copy buttons, mermaid) are injected separately after sanitization.
- **Private/loopback protection**: When `--allowed-upstreams` is empty (open mode), refuse to fetch from `127.0.0.0/8`, `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`, `::1`, `fd00::/8`. This prevents SSRF attacks. When `--allowed-upstreams` is explicitly configured, trust the operator and skip this check.
- **TLS verification**: By default, cooked verifies TLS certificates for upstream HTTPS fetches using the system CA store. The preferred approach for internal CAs is to add certificates to the Docker image (see Build & Distribution). The `--tls-skip-verify` flag disables verification entirely as a last resort — log a warning at startup when this is enabled.

## Error Handling

- **Upstream returns non-200**: Show a styled error page with the status code and a link to the upstream URL.
- **Upstream unreachable**: Show "Could not reach upstream server" with the URL.
- **File too large**: Show "File too large to render (X MB, limit is Y MB)" with a direct link.
- **Fetch timeout**: Show "Upstream request timed out after Xs".
- **Blocked upstream**: Show "This upstream is not in the allowed list".

All error pages use the same HTML template with proper theming (respects the user's theme choice). Include a direct link to the upstream URL.

## Logging

Structured JSON logging to stdout using Go's `slog` package:

```json
{
  "time": "2025-02-06T12:00:00Z",
  "level": "INFO",
  "msg": "request",
  "method": "GET",
  "path": "/https://cgit.internal/repo/plain/README.md",
  "upstream": "https://cgit.internal/repo/plain/README.md",
  "status": 200,
  "cache": "hit",
  "upstream_ms": 0,
  "total_ms": 2,
  "content_type": "markdown",
  "bytes": 14832
}
```

Cache field values: `hit`, `miss`, `revalidated`, `expired`.

Log at WARN level for upstream errors, ERROR for internal failures.

## Graceful Shutdown

Handle `SIGTERM` and `SIGINT`:

1. Stop accepting new connections
2. Wait for in-flight requests to complete (up to 30s timeout)
3. Exit cleanly

Standard for k8s pod lifecycle.

## Build & Distribution

- Clean project structure, split into packages as makes sense
- `go:embed` for all static assets
- **Makefile** with targets:
  - `make deps` — downloads mermaid.min.js (pinned version) and github-markdown-css (light + dark) into an `embed/` directory
  - `make build` — runs `go build` producing a static binary
  - `make docker` — multi-stage Docker build
  - `make test` — runs all tests
- **Dockerfile**: multi-stage build, final image `FROM alpine:latest` with `ca-certificates` and `update-ca-certificates` installed. This allows users to extend the image with internal CA certificates:
  ```dockerfile
  FROM cooked:latest
  COPY internal-ca.crt /usr/local/share/ca-certificates/
  RUN update-ca-certificates
  ```
  Go's `net/http` respects the system CA store, so custom certificates added this way work automatically without any cooked flags. For environments where proper CA distribution isn't feasible, `--tls-skip-verify` disables certificate verification entirely.
- `--version` flag prints version, git commit hash, build date

### Versioning

SemVer with git tags, starting at `v0.1.0`. Pre-1.0 signals no stability guarantee while iterating.

Version, commit SHA, and build date are injected at build time via `-ldflags`:

```go
// cmd/cooked/main.go — set by linker
var (
    version = "dev"
    commit  = "unknown"
    date    = "unknown"
)
```

Untagged or local builds report `version=dev` so you always know if you're running a release or a development build.

The version string appears in:
- `--version` flag output
- `X-Cooked-Version` response header
- `data-cooked-version` HTML attribute

### Docker image tags

Each release produces three Docker tags:

| Tag | Example | Purpose |
|-----|---------|---------|
| Version | `cooked:v0.1.0` | Pinned release |
| Latest | `cooked:latest` | Rolling latest release |
| SHA | `cooked:sha-abc1234` | Exact build from any commit |

The SHA tag uses the short (7-char) git commit hash prefixed with `sha-`.

## Testing

### Unit tests (Go)

- URL parsing: extracting upstream URL from request path
- Relative URL rewriting (markdown links rewrite through cooked, images point to upstream)
- File extension → renderer mapping (including .mdx)
- HTML sanitization (script stripping, event handler removal)
- Query string passthrough
- MDX preprocessing (import stripping, JSX tag handling, content preservation)
- Cache key generation

### Integration tests (Go, using httptest)

- Start server, fetch markdown, verify HTTP 200 and `Content-Type: text/html`
- Verify `--allowed-upstreams` correctly blocks/allows requests
- Verify SSRF protection (loopback/private addresses blocked in open mode)
- Cache behavior: first request is `miss`, second is `hit`, verify via response headers or logs
- Cache revalidation with ETag/Last-Modified (mock upstream returning 304)
- Error pages for: upstream 404, timeout, too large, blocked upstream
- Verify mermaid.js script tag is only present when markdown contains mermaid blocks
- Verify query strings are forwarded to upstream
- Verify `X-Cooked-*` response headers are present and correct on every response (including error pages)

### DOM / HTML structure tests

These can be run with Go (parse HTML with `golang.org/x/net/html` or regex for simple checks) or with a headless browser (Playwright/Puppeteer) in CI:

- **Required IDs exist**: Every rendered page must contain `#cooked-header`, `#cooked-source-link`, `#cooked-size`, `#cooked-type`, `#cooked-theme-toggle`, `#cooked-content`
- **Data attributes are populated**: `html[data-content-type]`, `html[data-upstream-url]`, `html[data-cache-status]` are never empty
- **Content type correctness**: A `.md` file produces `data-content-type="markdown"`, a `.py` file produces `data-content-type="code"`, etc.
- **TOC presence**: Document with 3+ headings has `#cooked-toc` and `#cooked-toc-toggle`; document with fewer does not
- **Code blocks structure**: Every fenced code block is wrapped in `.cooked-code-block` with `data-language` and `data-line-count` attributes, and contains a `.cooked-copy-btn`
- **Theme toggle works**: Clicking `#cooked-theme-toggle` cycles `html[data-theme]` through `auto` → `light` → `dark` → `auto`
- **Error pages**: Error responses contain `#cooked-error` with `data-error-type` and `data-status-code`
- **No external requests**: The HTML must not contain any `src=` or `href=` pointing to external domains (except the upstream source link and links within the rendered markdown content). Verify no CDN references leaked in.
- **Clean DOM**: No inline `style=` attributes on cooked's own elements (chroma output is exempt). All cooked elements use `cooked-*` class prefix or have an `id` starting with `cooked-`.
- **HTML comments**: Verify `<!-- cooked: header -->`, `<!-- cooked: content -->`, `<!-- cooked: scripts -->` section markers are present

## Future / Out of Scope (for now)

These are not needed in v1 but the architecture should not make them hard to add later:

- **reStructuredText (.rst)** rendering — would need external tool (docutils/rst2html)
- **AsciiDoc (.adoc)** rendering — would need external tool (asciidoctor)
- **Org-mode (.org)** rendering — Go library exists (go-org)
- **Directory listing**: If the upstream returns HTML that looks like a directory index, make links clickable through cooked
- **Custom CSS injection**: Flag to specify additional CSS to inject
- **Prometheus metrics endpoint**: `/metrics` with request count, cache hit rate, upstream latency histograms
- **Upstream authentication**: Per-upstream token/header configuration for private S3 buckets, etc.
- **PDF export**: Button to render current page as downloadable PDF
- **Git-aware mode**: Understand branch/tag/commit references in cgit/gitea/forgejo URLs
- **TOC scroll sync**: Highlight current TOC entry based on scroll position
- **Search within page**: Ctrl+F is fine, but a search box in the TOC could be nice for huge docs
- **Multiple file rendering**: Render an entire directory of markdown files as a browsable site
