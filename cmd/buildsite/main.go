// Static site generator for GitHub Pages (landing + docs).
package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/russross/blackfriday/v2"
)

var docNav = []struct {
	Slug  string
	Title string
}{
	{"index", "Overview"},
	{"getting-started", "Getting started"},
	{"examples", "Examples"},
	{"cli-reference", "CLI reference"},
	{"nsmith", "nsmith compiler"},
	{"rpc-reference", "RPC reference"},
	{"explorer", "Block explorer"},
	{"forking", "Forking"},
	{"fork-troubleshooting", "Fork troubleshooting"},
	{"impersonation", "Impersonation"},
	{"state-management", "State management"},
	{"tracing", "Tracing"},
	{"architecture", "Architecture"},
	{"anvil-comparison", "Anvil comparison"},
	{"development", "Development"},
	{"upstream-sync", "Upstream sync"},
}

var slugPattern = regexp.MustCompile(`^[a-z0-9-]+$`)
var mdLinkPattern = regexp.MustCompile(`href="([^"]+)\.md([^"]*)"`)

const docsPageTmpl = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{.Title}} · Nanvil Docs</title>
  <script>(function(){try{var k='nanvil-theme',t=localStorage.getItem(k);if(t!=='light'&&t!=='dark')t=matchMedia('(prefers-color-scheme:light)').matches?'light':'dark';document.documentElement.setAttribute('data-theme',t);}catch(e){document.documentElement.setAttribute('data-theme','dark');}})();</script>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=DM+Sans:ital,opsz,wght@0,9..40,400;0,9..40,500;0,9..40,600;0,9..40,700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
  <link rel="icon" type="image/png" href="../favicon.png">
  <link rel="stylesheet" href="../theme.css">
  <link rel="stylesheet" href="../docs.css">
</head>
<body>
  <div class="bg-mesh" aria-hidden="true"></div>
  <header class="docs-topbar">
    <a href="../index.html" class="docs-brand">
      <img src="../nanvil-logo.png" alt="" class="logo-img" width="36" height="36">
      <span>
        <strong>Nanvil</strong>
        <small>Documentation</small>
      </span>
    </a>
    <nav class="docs-topnav">
      <a href="../index.html">Home</a>
      <a href="index.html" class="active">Docs</a>
      <a href="https://github.com/merl111/nanvil" target="_blank" rel="noopener noreferrer">GitHub</a>
      <div class="theme-switcher" role="group" aria-label="Theme">
        <button type="button" class="theme-option" data-theme="light" aria-pressed="false" title="Light theme">
          <span class="theme-option-icon" aria-hidden="true">☀</span>
          <span class="theme-option-label">Light</span>
        </button>
        <button type="button" class="theme-option" data-theme="dark" aria-pressed="false" title="Dark theme">
          <span class="theme-option-icon" aria-hidden="true">☾</span>
          <span class="theme-option-label">Dark</span>
        </button>
      </div>
    </nav>
  </header>
  <div class="docs-layout">
    <aside class="docs-sidebar">
      <p class="docs-sidebar-label">Contents</p>
      <ul>
        {{range .Nav}}
        <li><a href="{{.Slug}}.html"{{if eq .Slug $.Current}} class="active"{{end}}>{{.Title}}</a></li>
        {{end}}
      </ul>
    </aside>
    <main class="docs-content">
      <article class="markdown-body">
        {{.Body}}
      </article>
    </main>
  </div>
  <footer class="docs-footer">
    <span>From <code>docs/</code> in the Nanvil repository</span>
    <span>·</span>
    <a href="../index.html">Back to home</a>
  </footer>
  <script src="../theme.js"></script>
</body>
</html>`

type docsPageData struct {
	Title   string
	Body    template.HTML
	Nav     []docNavItem
	Current string
}

type docNavItem struct {
	Slug  string
	Title string
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		fatal(err)
	}

	outDir := filepath.Join(root, "website", "dist")
	docsSrc := filepath.Join(root, "docs")
	staticSrc := filepath.Join(root, "pkg", "nanvil", "explorer", "static")
	websiteSrc := filepath.Join(root, "website")

	if err := os.RemoveAll(outDir); err != nil {
		fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(outDir, "docs"), 0o755); err != nil {
		fatal(err)
	}

	copyFiles := map[string]string{
		filepath.Join(staticSrc, "theme.css"):      filepath.Join(outDir, "theme.css"),
		filepath.Join(staticSrc, "theme.js"):       filepath.Join(outDir, "theme.js"),
		filepath.Join(staticSrc, "docs.css"):       filepath.Join(outDir, "docs.css"),
		filepath.Join(staticSrc, "nanvil-logo.png"): filepath.Join(outDir, "nanvil-logo.png"),
		filepath.Join(staticSrc, "favicon.png"):    filepath.Join(outDir, "favicon.png"),
		filepath.Join(websiteSrc, "index.html"):    filepath.Join(outDir, "index.html"),
		filepath.Join(websiteSrc, "landing.css"):   filepath.Join(outDir, "landing.css"),
		filepath.Join(websiteSrc, "releases.js"):  filepath.Join(outDir, "releases.js"),
	}
	for src, dst := range copyFiles {
		if err := copyFile(src, dst); err != nil {
			fatal(fmt.Errorf("copy %s: %w", src, err))
		}
	}

	tmpl, err := template.New("docs").Parse(docsPageTmpl)
	if err != nil {
		fatal(err)
	}

	nav := make([]docNavItem, len(docNav))
	for i, item := range docNav {
		nav[i] = docNavItem{Slug: item.Slug, Title: item.Title}
	}

	for _, item := range docNav {
		if err := renderDocPage(tmpl, docsSrc, outDir, nav, item.Slug); err != nil {
			fatal(fmt.Errorf("render %s: %w", item.Slug, err))
		}
	}

	fmt.Printf("=> Built static site at %s\n", outDir)
}

func renderDocPage(tmpl *template.Template, docsSrc, outDir string, nav []docNavItem, slug string) error {
	if !slugPattern.MatchString(slug) {
		return fmt.Errorf("invalid slug %q", slug)
	}

	raw, err := os.ReadFile(filepath.Join(docsSrc, slug+".md"))
	if err != nil {
		return err
	}

	title := pageTitleFromMarkdown(string(raw), slug)
	body := blackfriday.Run(raw, blackfriday.WithExtensions(
		blackfriday.CommonExtensions|blackfriday.AutoHeadingIDs|blackfriday.HardLineBreak,
	))
	html := string(body)
	html = mdLinkPattern.ReplaceAllString(html, `href="$1.html$2"`)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, docsPageData{
		Title:   title,
		Body:    template.HTML(html),
		Nav:     nav,
		Current: slug,
	}); err != nil {
		return err
	}

	outPath := filepath.Join(outDir, "docs", slug+".html")
	return os.WriteFile(outPath, buf.Bytes(), 0o644)
}

func pageTitleFromMarkdown(md, slug string) string {
	for _, line := range strings.Split(md, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	for _, item := range docNav {
		if item.Slug == slug {
			return item.Title
		}
	}
	return slug
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(path.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "buildsite: %v\n", err)
	os.Exit(1)
}
