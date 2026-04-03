## ADDED Requirements

### Requirement: Image src attributes are proxied through cooked
Relative `src` attributes on `<img>` tags in rendered HTML SHALL be rewritten to route through `/_cooked/raw/` instead of pointing directly at the upstream origin.

#### Scenario: Relative image in markdown
- **WHEN** a markdown file at `https://cgit.internal/repo/plain/README.md` contains `![arch](./docs/arch.png)`
- **THEN** the rendered HTML contains `src="/_cooked/raw/https://cgit.internal/repo/plain/docs/arch.png"`

#### Scenario: Relative image without dot-slash prefix
- **WHEN** a markdown file at `https://cgit.internal/repo/plain/README.md` contains `![logo](images/logo.svg)`
- **THEN** the rendered HTML contains `src="/_cooked/raw/https://cgit.internal/repo/plain/images/logo.svg"`

#### Scenario: Absolute image URL is not rewritten
- **WHEN** a markdown file contains `![badge](https://img.shields.io/badge/build-passing-green.svg)`
- **THEN** the rendered HTML contains `src="https://img.shields.io/badge/build-passing-green.svg"` unchanged

#### Scenario: Data URI image is not rewritten
- **WHEN** a markdown file contains `![](data:image/png;base64,abc)`
- **THEN** the rendered HTML preserves the data URI unchanged

### Requirement: Non-image src attributes are proxied through cooked
Relative `src` attributes on `<video>`, `<audio>`, `<source>`, and `<embed>` tags SHALL also be rewritten through `/_cooked/raw/`.

#### Scenario: Video source proxied
- **WHEN** rendered HTML contains `<video src="demo.mp4">`
- **THEN** the `src` is rewritten to `/_cooked/raw/https://upstream/path/demo.mp4`

### Requirement: Download href attributes point at upstream
Relative `href` attributes on `<a>` tags linking to non-renderable files SHALL continue to point directly at the upstream origin (not proxied through `/_cooked/raw/`).

#### Scenario: Download link points at upstream
- **WHEN** a markdown file contains `[download](archive.zip)`
- **THEN** the rendered HTML contains `href="https://cgit.internal/repo/plain/archive.zip"` (direct upstream)

### Requirement: Raw handler preserves upstream Content-Type
The `/_cooked/raw/` handler SHALL forward the upstream server's `Content-Type` response header instead of hardcoding `text/plain`.

#### Scenario: PNG image served with correct Content-Type
- **WHEN** a browser requests `/_cooked/raw/https://cgit.internal/repo/plain/logo.png`
- **AND** the upstream responds with `Content-Type: image/png`
- **THEN** cooked responds with `Content-Type: image/png`

#### Scenario: Upstream without Content-Type header
- **WHEN** the upstream does not include a `Content-Type` header
- **THEN** cooked detects the content type from the response body and sets it accordingly

### Requirement: Raw handler sets cache headers
The `/_cooked/raw/` handler SHALL set `Cache-Control: public, max-age=300` on successful responses.

#### Scenario: Proxied asset is cacheable
- **WHEN** a browser requests `/_cooked/raw/https://cgit.internal/repo/plain/logo.png`
- **THEN** the response includes `Cache-Control: public, max-age=300`

### Requirement: Base URL prefix applied when configured
When `--base-url` is set, proxied asset URLs SHALL include the base URL prefix.

#### Scenario: Base URL prepended to raw proxy path
- **WHEN** cooked runs with `--base-url=https://cooked.example.com`
- **AND** a markdown file contains `![](logo.png)`
- **THEN** the rendered HTML contains `src="https://cooked.example.com/_cooked/raw/https://upstream/path/logo.png"`
