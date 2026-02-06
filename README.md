# cooked

A rendering proxy that takes a URL to a raw document (markdown, etc.) and serves it as styled HTML. The opposite of "raw" — you give it a raw file URL, it gives you the cooked version.

Designed for air-gapped environments. The binary is fully self-contained — all CSS, JavaScript, fonts, and templates are embedded. No CDN requests, no external resources.

## How it works

```
https://cooked.example.com/https://cgit.internal/repo/plain/README.md
```

cooked fetches the upstream URL, detects the file type, renders it to styled HTML, and returns it. That's it.

## Quick start

```bash
make deps    # Download embedded assets (mermaid.js, github-markdown-css)
make build   # Build the binary
./cooked     # Start on :8080
```

## Development

```bash
make help    # Show all available targets
make test    # Run tests
make lint    # Run gitleaks
```

## License

[MIT](LICENSE)
