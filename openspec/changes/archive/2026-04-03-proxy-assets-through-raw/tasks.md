## 1. Raw handler improvements

- [x] 1.1 Pass through upstream Content-Type in `handleRaw` instead of hardcoding `text/plain`; fall back to `http.DetectContentType` when upstream omits the header
- [x] 1.2 Add `Cache-Control: public, max-age=300` to `handleRaw` responses
- [x] 1.3 Add/update tests for `handleRaw` Content-Type passthrough and cache headers

## 2. URL rewriter changes

- [x] 2.1 Add `rawProxyPrefix` parameter to `RelativeURLs` function signature
- [x] 2.2 Change non-renderable `src` attribute rewriting to use `rawProxyPrefix + resolved` instead of bare `resolved`
- [x] 2.3 Keep non-renderable `href` attribute rewriting unchanged (direct upstream)
- [x] 2.4 Update `handleRender` in `server.go` to pass `/_cooked/raw/` (with base URL prefix when configured) to `RelativeURLs`

## 3. Tests

- [x] 3.1 Update rewrite tests: relative image `src` now produces `/_cooked/raw/https://upstream/...` URLs
- [x] 3.2 Add test: absolute `src` URLs are not rewritten
- [x] 3.3 Add test: `data:` URIs in `src` are not rewritten
- [x] 3.4 Add test: `href` on non-renderable files still points at upstream (not proxied)
- [x] 3.5 Add test: base URL prefix is prepended when configured
- [x] 3.6 Run `make test-race` and `make lint` — all pass

## 4. Integration verification

- [x] 4.1 Manual test: render a markdown file with relative images through cooked, verify images load in browser
- [x] 4.2 Update golden files if any are affected by the URL format change
