(function () {
  const STORAGE_KEY = 'nanvil-theme';

  function preferredTheme() {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored === 'light' || stored === 'dark') {
        return stored;
      }
    } catch (_) {}
    return window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark';
  }

  function applyTheme(theme) {
    document.documentElement.setAttribute('data-theme', theme);
    try {
      localStorage.setItem(STORAGE_KEY, theme);
    } catch (_) {}
    document.querySelectorAll('.theme-option').forEach((btn) => {
      const active = btn.dataset.theme === theme;
      btn.classList.toggle('active', active);
      btn.setAttribute('aria-pressed', active ? 'true' : 'false');
    });
  }

  function initThemeSwitcher() {
    const theme = preferredTheme();
    applyTheme(theme);
    document.querySelectorAll('.theme-option').forEach((btn) => {
      btn.addEventListener('click', () => {
        if (btn.dataset.theme) {
          applyTheme(btn.dataset.theme);
        }
      });
    });
  }

  window.nanvilTheme = { applyTheme, preferredTheme };

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initThemeSwitcher);
  } else {
    initThemeSwitcher();
  }
})();
