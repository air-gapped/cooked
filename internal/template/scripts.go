package template

import (
	"bytes"
)

func writeScripts(buf *bytes.Buffer) {
	buf.WriteString(`  <script>
    // Theme toggle: auto → light → dark → auto
    (function() {
      var toggle = document.getElementById('cooked-theme-toggle');
      if (!toggle) return;

      function getTheme() {
        var cookie = document.cookie.match(/_cooked_theme=(\w+)/);
        if (cookie) return cookie[1];
        return document.documentElement.getAttribute('data-theme') || 'auto';
      }

      function setTheme(theme) {
        document.documentElement.setAttribute('data-theme', theme);
        document.cookie = '_cooked_theme=' + theme + ';path=/;max-age=31536000;SameSite=Lax';
      }

      // Check URL param override
      var params = new URLSearchParams(window.location.search);
      var paramTheme = params.get('_cooked_theme');
      if (paramTheme && ['auto','light','dark'].indexOf(paramTheme) !== -1) {
        setTheme(paramTheme);
      } else {
        var saved = getTheme();
        if (saved !== document.documentElement.getAttribute('data-theme')) {
          setTheme(saved);
        }
      }

      toggle.addEventListener('click', function() {
        var current = document.documentElement.getAttribute('data-theme');
        var next = current === 'auto' ? 'light' : current === 'light' ? 'dark' : 'auto';
        setTheme(next);
      });
    })();

    // TOC toggle
    (function() {
      var toggle = document.getElementById('cooked-toc-toggle');
      var toc = document.getElementById('cooked-toc');
      if (!toggle || !toc) return;

      toggle.addEventListener('click', function() {
        toc.hidden = !toc.hidden;
      });

      // Close TOC on mobile when clicking a link
      toc.addEventListener('click', function(e) {
        if (e.target.tagName === 'A' && window.innerWidth <= 768) {
          toc.hidden = true;
        }
      });
    })();

    // Copy buttons on code blocks
    (function() {
      document.querySelectorAll('.cooked-copy-btn').forEach(function(btn) {
        btn.addEventListener('click', function() {
          var block = btn.closest('.cooked-code-block');
          if (!block) return;
          var code = block.querySelector('pre code, pre');
          if (!code) return;
          navigator.clipboard.writeText(code.textContent).then(function() {
            btn.textContent = 'Copied!';
            btn.setAttribute('data-state', 'copied');
            setTimeout(function() {
              btn.textContent = 'Copy';
              btn.setAttribute('data-state', 'idle');
            }, 2000);
          });
        });
      });
    })();
  </script>
`)
}
