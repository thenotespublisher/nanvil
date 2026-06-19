package explorer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadDocPage(t *testing.T) {
	html, title, err := loadDocPage("getting-started")
	require.NoError(t, err)
	require.Equal(t, "Getting started", title)
	require.Contains(t, html, "<h1")
	require.Contains(t, html, `href="/docs/forking"`)

	_, _, err = loadDocPage("not-a-real-page")
	require.Error(t, err)
}

func TestDocSlugValidation(t *testing.T) {
	require.True(t, slugPattern.MatchString("getting-started"))
	require.False(t, slugPattern.MatchString("../etc/passwd"))
	require.False(t, slugPattern.MatchString("foo/bar"))
}

func TestPageTitleFromMarkdown(t *testing.T) {
	md := "# Hello World\n\nbody"
	require.Equal(t, "Hello World", pageTitleFromMarkdown(md, "index"))
	require.Equal(t, "Overview", pageTitleFromMarkdown("no heading", "index"))
}

func TestMarkdownLinkRewrite(t *testing.T) {
	_, _, err := loadDocPage("index")
	require.NoError(t, err)
	raw, err := docsFS.ReadFile("embedded-docs/index.md")
	require.NoError(t, err)
	require.True(t, strings.Contains(string(raw), "getting-started.md"))
	html, _, err := loadDocPage("index")
	require.NoError(t, err)
	require.Contains(t, html, `href="/docs/getting-started"`)
}
