# cooked

A rendering proxy that takes a URL to a raw document (markdown, etc.) and serves it as styled HTML. The opposite of "raw" — you give it a raw file URL, it gives you the cooked version.

Designed for air-gapped environments. The binary is fully self-contained — all CSS, JavaScript, fonts, and templates are embedded. No CDN requests, no external resources.

## How it works

```
https://cooked.example.com/https://cgit.internal/repo/plain/README.md
```

cooked fetches the upstream URL, detects the file type, renders it to styled HTML, and returns it. That's it.

## Supported formats

- **Markdown** — `.md`, `.markdown`, `.mdown`, `.mkd`
- **MDX** — `.mdx` (JSX imports/exports and component tags are stripped before rendering)
- **Code** — `.py`, `.go`, `.js`, `.ts`, `.rs`, `.c`, `.h`, `.cpp`, `.hpp`, `.java`, `.rb`, `.lua`, `.pl`, `.sh`, `.bash`, `.zsh`, `.fish`, `.yaml`, `.yml`, `.json`, `.toml`, `.xml`, `.csv`, `.sql`, `.graphql`, `.tf`, `.hcl`, `.dockerfile`, `.diff`, `.patch` (plus `Dockerfile`, `Makefile`, `Jenkinsfile` by filename)
- **Plaintext** — `.txt`, `.text`, `.log`, `.conf`, `.cfg`, `.ini`, `.env`

## Quick start

```bash
make deps    # Download embedded assets (mermaid.js, github-markdown-css)
make build   # Build the binary
./cooked     # Start on :8080
```

## Configuration

| Flag | Env Var | Default | Description |
|------|---------|---------|-------------|
| `--listen` | `COOKED_LISTEN` | `127.0.0.1:8080` | Listen address (loopback only by default; Docker overrides to `0.0.0.0:8080`) |
| `--cache-ttl` | `COOKED_CACHE_TTL` | `5m` | Cache TTL duration |
| `--cache-max-size` | `COOKED_CACHE_MAX_SIZE` | `100MB` | Max cache size (e.g. 100MB) |
| `--fetch-timeout` | `COOKED_FETCH_TIMEOUT` | `30s` | Upstream fetch timeout |
| `--max-file-size` | `COOKED_MAX_FILE_SIZE` | `5MB` | Max file size to render (e.g. 5MB) |
| `--allowed-upstreams` | `COOKED_ALLOWED_UPSTREAMS` | *(empty)* | Comma-separated allowed upstreams: hostnames, `*.wildcard`, or CIDR ranges |
| `--base-url` | `COOKED_BASE_URL` | *(auto-detect)* | Public base URL of cooked |
| `--default-theme` | `COOKED_DEFAULT_THEME` | `auto` | Default theme: auto, light, or dark |
| `--tls-skip-verify` | `COOKED_TLS_SKIP_VERIFY` | `false` | Disable TLS certificate verification for upstream fetches |

All flags have environment variable equivalents prefixed with `COOKED_`.

## Security

### Allowed upstreams

The `--allowed-upstreams` flag restricts which upstream hosts cooked will fetch from. It accepts a comma-separated list of three entry types:

- **Hostnames** — exact match plus subdomain matching (e.g. `cgit.internal` also allows `sub.cgit.internal`)
- **Wildcards** — `*.internal` matches any host ending in `.internal` (e.g. `foo.internal`, `a.b.internal`)
- **CIDR ranges** — `10.0.0.0/8` matches IP-literal URLs in that range (e.g. `http://10.0.1.50/file.md`)

Redirect targets are also validated against the allowlist.

```bash
# Only allow fetching from two internal hosts
./cooked --allowed-upstreams="cgit.internal,gitea.corp.example.com"

# CIDR ranges for IP-based upstreams
./cooked --allowed-upstreams="10.0.0.0/8,172.16.0.0/12"

# Wildcard DNS patterns
./cooked --allowed-upstreams="*.internal,*.corp.example.com"

# Mixed: all three types together
./cooked --allowed-upstreams="*.internal,10.0.0.0/8,gitea.specific.host"
```

### Private IP protection (SSRF)

When `--allowed-upstreams` is **not set**, cooked blocks requests to private and loopback IP ranges to prevent server-side request forgery (SSRF). The blocked ranges include:

- IPv4/IPv6 loopback, private (RFC 1918), link-local, multicast, unspecified
- CGNAT (`100.64.0.0/10`)

When `--allowed-upstreams` **is set**, the allowlist becomes the trust boundary and private-IP blocking is disabled. This is required for air-gapped environments where upstreams are on private networks (10.x, 172.16.x, 192.168.x).

Redirects are capped at 5 hops and validated against the allowlist when set.

### HTML sanitization

Rendered markdown/MDX output is sanitized: `<script>`, `<iframe>`, `<object>`, `<embed>`, `<form>`, `<input>` tags and all `on*` event handler attributes are stripped. Additionally, `javascript:`, `vbscript:`, and `data:text/html` URIs in `href`/`src` attributes are removed.

### TLS verification

By default, cooked verifies TLS certificates when fetching upstream URLs. For internal CAs, add your CA certificate to the system trust store (see [Docker with internal CAs](#internal-ca-certificates)). Use `--tls-skip-verify` only as a last resort.

### No credential forwarding

cooked does not forward cookies, authorization headers, or other credentials to upstream servers.

## Docker

### Basic usage

```bash
docker run -p 8080:8080 cooked
```

### Air-gapped with allowed upstreams

```bash
docker run -p 8080:8080 cooked \
  --allowed-upstreams="*.internal,10.0.0.0/8,gitea.corp.example.com"
```

### Internal CA certificates

If your upstreams use certificates signed by an internal CA, add the CA cert to the image:

```dockerfile
FROM ghcr.io/air-gapped/cooked:latest
COPY my-internal-ca.crt /usr/local/share/ca-certificates/
RUN update-ca-certificates
```

### Docker Compose

```yaml
services:
  cooked:
    image: ghcr.io/air-gapped/cooked:latest
    ports:
      - "8080:8080"
    command: ["--allowed-upstreams=*.internal,10.0.0.0/8,gitea.corp.example.com"]
```

## Response headers

cooked sets response headers for monitoring and debugging:

| Header | Description |
|--------|-------------|
| `X-Cooked-Version` | Application version |
| `X-Cooked-Upstream` | Upstream URL that was fetched |
| `X-Cooked-Upstream-Status` | HTTP status code from upstream |
| `X-Cooked-Cache` | Cache status (hit/miss/revalidated/stale) |
| `X-Cooked-Content-Type` | Detected file type (markdown/code/plaintext) |
| `X-Cooked-Render-Ms` | Time spent rendering HTML (milliseconds) |
| `X-Cooked-Upstream-Ms` | Time spent fetching from upstream (milliseconds) |

## Health check

`GET /healthz` returns `200 OK`. Use this for load balancer and container health checks.

## Development

```bash
make help    # Show all available targets
make deps    # Download embedded assets
make build   # Build the binary
make test    # Run tests
make lint    # Run golangci-lint + gitleaks
make docker  # Build Docker image
```

## License

[MIT](LICENSE)
