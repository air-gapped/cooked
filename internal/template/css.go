package template

import (
	"bytes"
	"fmt"
	"strings"
)

// faviconSVG is an inline SVG favicon — a simple cooking pot icon.
const faviconSVG = `%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'%3E%3Ctext y='.9em' font-size='90'%3E🍳%3C/text%3E%3C/svg%3E`

// chromaLight is CSS for chroma syntax highlighting (light theme).
const chromaLight = `
/* chroma light */
.chroma { background-color: #f8f8f8; }
.chroma .ln { color: #7f7f7f; margin-right: 0.4em; padding: 0 0.4em; }
.chroma .lnt { color: #7f7f7f; margin-right: 0.4em; padding: 0 0.4em; }
.chroma .hl { background-color: #ffffcc; display: block; }
.chroma .lnlinks { outline: none; text-decoration: none; color: inherit; }
.chroma .err { color: #a61717; background-color: #e3d2d2; }
.chroma .k { color: #000; font-weight: bold; }
.chroma .kc { color: #000; font-weight: bold; }
.chroma .kd { color: #000; font-weight: bold; }
.chroma .kn { color: #000; font-weight: bold; }
.chroma .kp { color: #000; font-weight: bold; }
.chroma .kr { color: #000; font-weight: bold; }
.chroma .kt { color: #458; font-weight: bold; }
.chroma .na { color: #008080; }
.chroma .nb { color: #0086b3; }
.chroma .nc { color: #458; font-weight: bold; }
.chroma .no { color: #008080; }
.chroma .nd { color: #3c5d5d; font-weight: bold; }
.chroma .ni { color: #800080; }
.chroma .ne { color: #900; font-weight: bold; }
.chroma .nf { color: #900; font-weight: bold; }
.chroma .nn { color: #555; }
.chroma .nt { color: #000080; }
.chroma .nv { color: #008080; }
.chroma .s { color: #d14; }
.chroma .sa { color: #d14; }
.chroma .sb { color: #d14; }
.chroma .sc { color: #d14; }
.chroma .dl { color: #d14; }
.chroma .sd { color: #d14; }
.chroma .s2 { color: #d14; }
.chroma .se { color: #d14; }
.chroma .sh { color: #d14; }
.chroma .si { color: #d14; }
.chroma .sx { color: #d14; }
.chroma .sr { color: #009926; }
.chroma .s1 { color: #d14; }
.chroma .ss { color: #990073; }
.chroma .c { color: #998; font-style: italic; }
.chroma .ch { color: #998; font-style: italic; }
.chroma .cm { color: #998; font-style: italic; }
.chroma .c1 { color: #998; font-style: italic; }
.chroma .cs { color: #999; font-weight: bold; font-style: italic; }
.chroma .cp { color: #999; font-weight: bold; font-style: italic; }
.chroma .cpf { color: #999; font-weight: bold; font-style: italic; }
.chroma .m { color: #099; }
.chroma .mb { color: #099; }
.chroma .mf { color: #099; }
.chroma .mh { color: #099; }
.chroma .mi { color: #099; }
.chroma .il { color: #099; }
.chroma .mo { color: #099; }
.chroma .o { color: #000; font-weight: bold; }
.chroma .ow { color: #000; font-weight: bold; }
.chroma .p { color: #000; }
.chroma .w { color: #bbb; }
`

// chromaDark is CSS for chroma syntax highlighting (dark theme).
const chromaDark = `
/* chroma dark */
.chroma { background-color: #1e1e1e; color: #d4d4d4; }
.chroma .ln { color: #6e7681; margin-right: 0.4em; padding: 0 0.4em; }
.chroma .lnt { color: #6e7681; margin-right: 0.4em; padding: 0 0.4em; }
.chroma .hl { background-color: #2a2a2a; display: block; }
.chroma .err { color: #f44747; }
.chroma .k { color: #569cd6; }
.chroma .kc { color: #569cd6; }
.chroma .kd { color: #569cd6; }
.chroma .kn { color: #c586c0; }
.chroma .kp { color: #569cd6; }
.chroma .kr { color: #569cd6; }
.chroma .kt { color: #4ec9b0; }
.chroma .na { color: #9cdcfe; }
.chroma .nb { color: #4ec9b0; }
.chroma .nc { color: #4ec9b0; }
.chroma .no { color: #4fc1ff; }
.chroma .nd { color: #dcdcaa; }
.chroma .ni { color: #d7ba7d; }
.chroma .ne { color: #4ec9b0; }
.chroma .nf { color: #dcdcaa; }
.chroma .nn { color: #4ec9b0; }
.chroma .nt { color: #569cd6; }
.chroma .nv { color: #9cdcfe; }
.chroma .s { color: #ce9178; }
.chroma .sa { color: #ce9178; }
.chroma .sb { color: #ce9178; }
.chroma .sc { color: #ce9178; }
.chroma .dl { color: #ce9178; }
.chroma .sd { color: #ce9178; }
.chroma .s2 { color: #ce9178; }
.chroma .se { color: #d7ba7d; }
.chroma .sh { color: #ce9178; }
.chroma .si { color: #ce9178; }
.chroma .sx { color: #ce9178; }
.chroma .sr { color: #d16969; }
.chroma .s1 { color: #ce9178; }
.chroma .ss { color: #ce9178; }
.chroma .c { color: #6a9955; font-style: italic; }
.chroma .ch { color: #6a9955; font-style: italic; }
.chroma .cm { color: #6a9955; font-style: italic; }
.chroma .c1 { color: #6a9955; font-style: italic; }
.chroma .cs { color: #6a9955; font-style: italic; }
.chroma .cp { color: #c586c0; }
.chroma .cpf { color: #ce9178; }
.chroma .m { color: #b5cea8; }
.chroma .mb { color: #b5cea8; }
.chroma .mf { color: #b5cea8; }
.chroma .mh { color: #b5cea8; }
.chroma .mi { color: #b5cea8; }
.chroma .il { color: #b5cea8; }
.chroma .mo { color: #b5cea8; }
.chroma .o { color: #d4d4d4; }
.chroma .ow { color: #c586c0; }
.chroma .p { color: #d4d4d4; }
.chroma .w { color: #d4d4d4; }
`

func writeThemeCSS(buf *bytes.Buffer, lightCSS, darkCSS, chromaLightCSS, chromaDarkCSS string) {
	// Light theme
	fmt.Fprintf(buf, "    /* Theme: light */\n")
	fmt.Fprintf(buf, "    [data-theme=\"light\"] { color-scheme: light; }\n")
	if lightCSS != "" {
		buf.WriteString(prefixThemeCSS(lightCSS, `[data-theme="light"]`))
	}
	fmt.Fprintf(buf, "    [data-theme=\"light\"] %s\n", chromaLightCSS)

	// Dark theme
	fmt.Fprintf(buf, "    /* Theme: dark */\n")
	fmt.Fprintf(buf, "    [data-theme=\"dark\"] { color-scheme: dark; }\n")
	if darkCSS != "" {
		buf.WriteString(prefixThemeCSS(darkCSS, `[data-theme="dark"]`))
	}
	fmt.Fprintf(buf, "    [data-theme=\"dark\"] %s\n", chromaDarkCSS)

	// Auto theme (system preference)
	fmt.Fprintf(buf, "    /* Theme: auto (system preference) */\n")
	fmt.Fprintf(buf, "    [data-theme=\"auto\"] { color-scheme: light dark; }\n")
	if lightCSS != "" {
		buf.WriteString(prefixThemeCSS(lightCSS, `[data-theme="auto"]`))
	}
	fmt.Fprintf(buf, "    [data-theme=\"auto\"] %s\n", chromaLightCSS)

	fmt.Fprintf(buf, "    @media (prefers-color-scheme: dark) {\n")
	fmt.Fprintf(buf, "      [data-theme=\"auto\"] { color-scheme: dark; }\n")
	if darkCSS != "" {
		buf.WriteString(prefixThemeCSS(darkCSS, `[data-theme="auto"]`))
	}
	fmt.Fprintf(buf, "      [data-theme=\"auto\"] %s\n", chromaDarkCSS)
	fmt.Fprintf(buf, "    }\n")
}

func writeLayoutCSS(buf *bytes.Buffer) {
	buf.WriteString(layoutCSS)
}

// prefixThemeCSS prepends a theme selector before each .markdown-body selector
// in the embedded github-markdown CSS. The CSS files use .markdown-body as the
// root selector for every rule, so we prefix each occurrence to scope rules to
// the active theme (e.g. [data-theme="light"] .markdown-body hr { ... }).
//
// This avoids wrapping the CSS in a block like [data-theme] .markdown-body { ... }
// which would cause CSS nesting and produce broken double selectors like
// [data-theme] .markdown-body .markdown-body hr that never match the DOM.
func prefixThemeCSS(css, themeSelector string) string {
	return strings.ReplaceAll(css, ".markdown-body", themeSelector+" .markdown-body")
}

const layoutCSS = `
    /* cooked layout */
    * { box-sizing: border-box; }
    body { margin: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif; }

    #cooked-header {
      position: sticky; top: 0; z-index: 100;
      display: flex; align-items: center; justify-content: space-between;
      padding: 4px 16px; font-size: 12px;
      border-bottom: 1px solid rgba(128,128,128,0.2);
      background: rgba(246,248,250,0.95); color: #656d76;
    }
    [data-theme="dark"] #cooked-header {
      background: rgba(22,27,34,0.95); color: #8b949e;
    }
    @media (prefers-color-scheme: dark) {
      [data-theme="auto"] #cooked-header {
        background: rgba(22,27,34,0.95); color: #8b949e;
      }
    }

    .cooked-meta { display: flex; align-items: center; gap: 8px; overflow: hidden; flex: 1; }
    .cooked-meta a { color: inherit; text-decoration: none; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .cooked-meta a:hover { text-decoration: underline; }
    .cooked-meta span, .cooked-meta time { white-space: nowrap; }
    .cooked-divider { width: 1px; height: 12px; background: rgba(128,128,128,0.4); flex-shrink: 0; }
    .cooked-copy-url {
      background: none; border: none; cursor: pointer; padding: 0 2px;
      color: inherit; opacity: 0.6; font-size: 12px; flex-shrink: 0; line-height: 1;
    }
    .cooked-copy-url:hover { opacity: 1; }
    .cooked-controls { display: flex; gap: 4px; }
    .cooked-controls button {
      background: none; border: 1px solid rgba(128,128,128,0.3); border-radius: 4px;
      cursor: pointer; padding: 2px 6px; font-size: 14px; color: inherit;
    }
    .cooked-controls button:hover { background: rgba(128,128,128,0.1); }
    .cooked-copy-md {
      font-size: 11px !important; padding: 2px 8px !important;
      display: flex; align-items: center; gap: 4px;
    }

    main { max-width: 1012px; margin: 0 auto; padding: 32px 16px; }
    .markdown-body {
      font-size: 16px; line-height: 1.5; word-wrap: break-word;
      border: 1px solid #d0d7de; border-radius: 6px;
      padding: 32px 40px;
    }
    [data-theme="dark"] .markdown-body { border-color: #30363d; }
    @media (prefers-color-scheme: dark) {
      [data-theme="auto"] .markdown-body { border-color: #30363d; }
    }
    #cooked-toc {
      position: fixed; top: 32px; left: 0; bottom: 0; width: 280px;
      overflow-y: auto; padding: 16px; font-size: 13px;
      background: rgba(246,248,250,0.98); border-right: 1px solid rgba(128,128,128,0.2);
      z-index: 50;
    }
    [data-theme="dark"] #cooked-toc {
      background: rgba(22,27,34,0.98); border-color: rgba(128,128,128,0.2);
    }
    @media (prefers-color-scheme: dark) {
      [data-theme="auto"] #cooked-toc {
        background: rgba(22,27,34,0.98);
      }
    }
    #cooked-toc ul { list-style: none; padding: 0; margin: 0; }
    #cooked-toc li { padding: 2px 0; }
    #cooked-toc li[data-level="2"] { padding-left: 12px; }
    #cooked-toc li[data-level="3"] { padding-left: 24px; }
    #cooked-toc li[data-level="4"] { padding-left: 36px; }
    #cooked-toc li[data-level="5"] { padding-left: 48px; }
    #cooked-toc li[data-level="6"] { padding-left: 60px; }
    #cooked-toc a { color: inherit; text-decoration: none; }
    #cooked-toc a:hover { text-decoration: underline; }
    #cooked-toc li.active > a { color: #0969da; font-weight: 600; }
    [data-theme="dark"] #cooked-toc li.active > a { color: #58a6ff; }
    @media (prefers-color-scheme: dark) {
      [data-theme="auto"] #cooked-toc li.active > a { color: #58a6ff; }
    }

    .cooked-code-block {
      position: relative; margin: 16px 0;
      border: 1px solid #d0d7de; border-radius: 6px; overflow: hidden;
    }
    .cooked-code-block pre { margin: 0 !important; border-radius: 0 !important; border: none !important; }
    .cooked-code-block pre code { padding: 16px !important; display: block; overflow-x: auto; }
    .cooked-code-header {
      display: flex; justify-content: flex-end; align-items: center; gap: 8px;
      padding: 4px 12px; font-size: 12px; color: #656d76;
      background: #f6f8fa; border-bottom: 1px solid #d0d7de;
    }
    .cooked-copy-btn {
      background: none; border: 1px solid rgba(128,128,128,0.3); border-radius: 4px;
      cursor: pointer; padding: 2px 8px; font-size: 11px; color: inherit;
    }
    .cooked-copy-btn:hover { background: rgba(128,128,128,0.15); }
    [data-theme="dark"] .cooked-code-block { border-color: #30363d; }
    [data-theme="dark"] .cooked-code-header { background: #161b22; border-color: #30363d; color: #8b949e; }
    @media (prefers-color-scheme: dark) {
      [data-theme="auto"] .cooked-code-block { border-color: #30363d; }
      [data-theme="auto"] .cooked-code-header { background: #161b22; border-color: #30363d; color: #8b949e; }
    }

    #cooked-error {
      text-align: center; padding: 80px 16px;
    }
    #cooked-error h1 { font-size: 48px; margin: 0 0 16px; color: #656d76; }
    #cooked-error p { color: #656d76; font-size: 16px; }
    #cooked-error a { color: #0969da; }

    @media print {
      /* Hide interactive UI elements */
      #cooked-header, #cooked-toc { display: none !important; }
      .cooked-code-header { display: none !important; }
      .cooked-copy-btn, .cooked-copy-url, .cooked-copy-md { display: none !important; }

      /* Remove screen layout constraints */
      main { max-width: 100%; padding: 0; }
      .markdown-body {
        max-width: 100%; padding: 0;
        border: none; border-radius: 0;
        font-size: 12px;
      }

      /* Scale down headings */
      .markdown-body h1 { font-size: 1.6em; }
      .markdown-body h2 { font-size: 1.3em; }
      .markdown-body h3 { font-size: 1.15em; }

      /* Code blocks: wrap long lines, shrink font, remove decorative borders */
      .cooked-code-block { border: 1px solid #ccc; border-radius: 0; }
      .cooked-code-block pre code {
        white-space: pre-wrap !important;
        word-break: break-all;
        font-size: 10px !important;
        padding: 8px !important;
      }

      /* Constrain images */
      .markdown-body img { max-width: 100%; max-height: 300px; object-fit: contain; }

      /* Force light colors for print (save ink) */
      html { color: #000 !important; background: #fff !important; }
      .markdown-body { color: #000 !important; background: #fff !important; }

      /* Avoid wasteful page breaks */
      .cooked-code-block { break-inside: avoid; }
      h1, h2, h3, h4, h5, h6 { break-after: avoid; }
    }

    @media (max-width: 768px) {
      #cooked-toc { width: 100%; }
      main { padding: 16px 8px; }
      .markdown-body { padding: 16px; border-radius: 0; border-left: 0; border-right: 0; }
    }
`
