package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/nsmith/cli"
	"github.com/stretchr/testify/require"
)

func TestCompileGoJSON(t *testing.T) {
	src := filepath.Join("..", "..", "..", "integration", "testcontracts", "explorer_deploy", "main.go")
	dir := t.TempDir()
	out := filepath.Join(dir, "out")
	var buf bytes.Buffer
	app := cli.NewApp()
	app.Writer = &buf
	err := app.Run([]string{"nsmith", "--json", "compile", "--out", out, "--name", "JsonTest", src})
	require.NoError(t, err)
	require.Contains(t, buf.String(), `"NEF"`)
	require.FileExists(t, out+".nef")
}

func TestVersion(t *testing.T) {
	app := cli.NewApp()
	err := app.Run([]string{"nsmith", "version"})
	require.NoError(t, err)
}

func TestInitGo(t *testing.T) {
	dir := t.TempDir()
	prev, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(prev) })
	require.NoError(t, os.Chdir(dir))
	app := cli.NewApp()
	err = app.Run([]string{"nsmith", "init", "--lang", "go", "Demo"})
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(dir, "Demo", "main.go"))
}
