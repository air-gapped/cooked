package template

import (
	"bytes"
	"fmt"
	"html"
)

// RenderLanding produces the landing page HTML for GET /.
func (r *Renderer) RenderLanding(version, defaultTheme string) []byte {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, `<!DOCTYPE html>
<html lang="en" data-theme="%s" data-cooked-version="%s">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>cooked — rendering proxy</title>
  <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,%s">
  <style>
`,
		html.EscapeString(defaultTheme),
		html.EscapeString(version),
		faviconSVG,
	)

	writeLayoutCSS(&buf)

	buf.WriteString(`
    .cooked-landing {
      max-width: 600px; margin: 80px auto; padding: 0 16px; text-align: center;
    }
    .cooked-landing h1 { font-size: 48px; margin: 0 0 8px; }
    .cooked-landing p { color: #656d76; font-size: 16px; margin: 0 0 32px; }
    .cooked-landing-form { display: flex; gap: 8px; }
    .cooked-landing-form input {
      flex: 1; padding: 10px 14px; font-size: 14px;
      border: 1px solid rgba(128,128,128,0.3); border-radius: 6px;
      background: inherit; color: inherit;
    }
    .cooked-landing-form button {
      padding: 10px 20px; font-size: 14px; font-weight: 600;
      background: #0969da; color: white; border: none; border-radius: 6px;
      cursor: pointer;
    }
    .cooked-landing-form button:hover { background: #0860ca; }
    [data-theme="dark"] .cooked-landing p { color: #8b949e; }
    [data-theme="dark"] .cooked-landing-form input {
      border-color: rgba(128,128,128,0.3); background: #161b22; color: #e6edf3;
    }
    @media (prefers-color-scheme: dark) {
      [data-theme="auto"] .cooked-landing p { color: #8b949e; }
      [data-theme="auto"] .cooked-landing-form input {
        border-color: rgba(128,128,128,0.3); background: #161b22; color: #e6edf3;
      }
    }
  </style>
</head>
<body>
  <div class="cooked-landing">
    <h1>cooked</h1>
    <p>Paste a raw document URL to view it rendered with styling.</p>
    <form class="cooked-landing-form" onsubmit="event.preventDefault(); var u=this.querySelector('input').value.trim(); if(u) window.location.href='/'+u;">
      <input type="url" placeholder="https://example.com/path/to/README.md" autofocus>
      <button type="submit">Cook it</button>
    </form>
    <p style="margin-top:16px;font-size:13px;">
      cooked %s
    </p>
  </div>
  <!-- cooked: scripts -->
`)

	writeScripts(&buf)

	fmt.Fprintf(&buf, "</body>\n</html>\n")

	return buf.Bytes()
}
