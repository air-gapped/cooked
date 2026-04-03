## Context

cooked rewrites relative URLs in rendered HTML via `internal/rewrite/rewrite.go`. Renderable files (`.md`, `.adoc`, `.org`) are routed through cooked. Non-renderable files (images, fonts, archives) are rewritten to absolute upstream URLs, expecting the browser to fetch them directly.

The `/_cooked/raw/` endpoint already exists in `internal/server/server.go:185` and proxies raw upstream content. However, it hardcodes `Content-Type: text/plain` and is not used by the URL rewriter.

The `RelativeURLs` function takes `(html, upstreamURL, baseURL)` — `baseURL` is the cooked instance's public URL (e.g., `https://cooked.example.com`), used for rendering links. There is currently no parameter for a raw proxy prefix.

## Goals / Non-Goals

**Goals:**
1. Images and other embedded assets load when cooked is on a different origin than the upstream
2. No configuration required — asset proxying is always on
3. Upstream Content-Type is preserved (images serve as `image/png`, not `text/plain`)
4. Allowlist and SSRF protections apply to proxied assets (already enforced by `handleRaw`)

**Non-Goals:**
- Caching proxied assets (the existing in-memory cache is for rendered pages; raw assets use HTTP cache headers)
- Rewriting absolute URLs (only relative `src`/`href` are affected)
- Streaming large binary files (existing `maxFileSize` limit applies)
- Rewriting CSS `url()` references (only HTML `src`/`href` attributes)

## Decisions

**1. Rewrite non-renderable `src` through `/_cooked/raw/`, leave `href` pointing at upstream**

Images (`<img src>`), videos (`<video src>`), audio (`<audio src>`) need to load in the browser — these must go through the proxy. Download links (`<a href="archive.zip">`) should still point at the upstream so the user downloads directly without cooked buffering the entire file.

Alternative: proxy everything through `/_cooked/raw/`. Rejected because large file downloads would be buffered through cooked's memory, hitting `maxFileSize` limits unnecessarily.

**2. Pass upstream Content-Type through in `handleRaw`**

Currently hardcoded to `text/plain`. Change to forward the upstream's `Content-Type` response header. If the upstream doesn't provide one, fall back to Go's `http.DetectContentType` on the first 512 bytes.

**3. Add cache headers to raw-proxied assets**

Set `Cache-Control: public, max-age=300` (same as rendered pages) on raw-proxied responses. The browser will cache images locally, reducing repeat fetches.

**4. Extend `RelativeURLs` signature to accept raw proxy prefix**

Add a `rawProxyPrefix` parameter. When non-empty, non-renderable `src` attributes are rewritten through it. When empty (backward compat for tests), behavior is unchanged.

## Risks / Trade-offs

**[Risk] Increased bandwidth through cooked** — All images now flow through cooked instead of being fetched directly from upstream. Mitigated by browser caching (`max-age=300`) and the fact that this only affects deployments where direct upstream access already fails.

**[Risk] maxFileSize blocks large images** — A 10MB high-res image would be rejected by the default 5MB limit. This is acceptable — the same limit already applies to documents. Users can increase `--max-file-size` if needed.

**[Trade-off] Only `src` attributes are proxied, not `href`** — Download links still point at upstream. This is intentional (see Decision 1) but means `<a href="screenshot.png">` won't load as a page through cooked. This matches user expectation — clicking a link to a .png should download it, not render it.
