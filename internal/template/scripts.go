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

      var icons = { auto: '\u25D1', light: '\u2600', dark: '\u263E' };
      var labels = { auto: 'Auto', light: 'Light', dark: 'Dark' };

      function updateButton() {
        var theme = document.documentElement.getAttribute('data-theme') || 'auto';
        toggle.textContent = icons[theme] || icons.auto;
        toggle.title = 'Theme: ' + (labels[theme] || 'Auto');
      }
      updateButton();

      toggle.addEventListener('click', function() {
        var current = document.documentElement.getAttribute('data-theme');
        var next = current === 'auto' ? 'light' : current === 'light' ? 'dark' : 'auto';
        setTheme(next);
        updateButton();
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

    // TOC scroll sync
    (function() {
      var toc = document.getElementById('cooked-toc');
      if (!toc) return;
      var content = document.getElementById('cooked-content');
      if (!content) return;
      var headings = content.querySelectorAll('h1[id], h2[id], h3[id], h4[id], h5[id], h6[id]');
      if (!headings.length) return;

      var tocLinks = {};
      toc.querySelectorAll('a[href^="#"]').forEach(function(a) {
        tocLinks[a.getAttribute('href').slice(1)] = a;
      });

      var activeLi = null;
      function setActive(id) {
        if (activeLi) activeLi.classList.remove('active');
        var a = tocLinks[id];
        if (a) {
          activeLi = a.parentElement;
          activeLi.classList.add('active');
          activeLi.scrollIntoView({ block: 'nearest' });
        }
      }

      var observer = new IntersectionObserver(function(entries) {
        entries.forEach(function(entry) {
          if (entry.isIntersecting) {
            setActive(entry.target.id);
          }
        });
      }, { rootMargin: '0px 0px -80% 0px' });

      headings.forEach(function(h) { observer.observe(h); });
    })();

    // Copy URL button
    (function() {
      var btn = document.getElementById('cooked-copy-url');
      if (!btn) return;
      var link = document.getElementById('cooked-source-link');
      if (!link) return;
      btn.addEventListener('click', function() {
        navigator.clipboard.writeText(link.href).then(function() {
          btn.textContent = '\u2713';
          setTimeout(function() { btn.innerHTML = '\u2398'; }, 1500);
        });
      });
    })();

    // Copy as Markdown / Copy Source button
    (function() {
      var btn = document.getElementById('cooked-copy-md');
      if (!btn) return;
      var upstreamURL = document.documentElement.getAttribute('data-upstream-url');
      if (!upstreamURL) return;
      var originalHTML = btn.innerHTML;

      btn.addEventListener('click', function() {
        btn.disabled = true;
        btn.textContent = 'Fetching\u2026';
        fetch(upstreamURL).then(function(r) {
          if (!r.ok) throw new Error(r.status);
          return r.text();
        }).then(function(text) {
          return navigator.clipboard.writeText(text);
        }).then(function() {
          btn.innerHTML = '\u2713 Copied!';
          setTimeout(function() {
            btn.innerHTML = originalHTML;
            btn.disabled = false;
          }, 2000);
        }).catch(function() {
          btn.textContent = 'Failed';
          setTimeout(function() {
            btn.innerHTML = originalHTML;
            btn.disabled = false;
          }, 2000);
        });
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
