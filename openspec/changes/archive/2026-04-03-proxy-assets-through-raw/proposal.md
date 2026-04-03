## Why

When cooked runs on a different origin than the upstream (e.g., `cooked.example.com` vs `cgit.internal`), images and other embedded assets in rendered documents fail to load. The browser blocks cross-origin requests due to CORS restrictions on the upstream server. This makes cooked unusable as a standalone rendering proxy in any deployment where users can't directly reach the upstream.

## What Changes

- Non-renderable relative URLs (images, fonts, videos, etc.) in rendered HTML will be rewritten through `/_cooked/raw/` instead of pointing directly at the upstream
- The `/_cooked/raw/` handler will pass through the upstream's `Content-Type` instead of hardcoding `text/plain`
- The `RelativeURLs` rewriter will accept a raw proxy prefix and use it for non-renderable assets
- Absolute URLs are left untouched (they already work if the upstream allows cross-origin)

## Capabilities

### New Capabilities
- `asset-proxying`: Proxy non-renderable assets (images, fonts, etc.) through cooked so the browser never needs direct access to the upstream

### Modified Capabilities

## Impact

- `internal/rewrite/rewrite.go` — change non-renderable URL rewriting to route through `/_cooked/raw/`
- `internal/server/server.go` — `handleRaw` must pass through upstream Content-Type, add appropriate cache headers
- `internal/server/server.go` — `handleRender` must pass the raw proxy prefix to `RelativeURLs`
- Existing tests in `internal/rewrite/rewrite_test.go` will need updating for the new URL format
- No new dependencies
