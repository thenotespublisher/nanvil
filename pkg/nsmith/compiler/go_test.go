package compiler_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/nsmith/compiler"
	"github.com/stretchr/testify/require"
)

func TestGoBackendCompile(t *testing.T) {
	src := filepath.Join("..", "..", "..", "integration", "testcontracts", "explorer_deploy", "main.go")
	require.FileExists(t, src)
	dir := t.TempDir()
	out := filepath.Join(dir, "contract")
	res, err := compiler.Compile(context.Background(), compiler.CompileRequest{
		Path:      src,
		Lang:      compiler.LangGo,
		OutPrefix: out,
		Name:      "GoCompileTest",
	})
	require.NoError(t, err)
	require.Equal(t, out+".nef", res.NEF)
	require.Equal(t, out+".manifest.json", res.Manifest)
	require.FileExists(t, res.NEF)
	require.FileExists(t, res.Manifest)
}

func TestGoBackendCompileAutoDetect(t *testing.T) {
	src := filepath.Join("..", "..", "..", "integration", "testcontracts", "explorer_deploy", "main.go")
	dir := t.TempDir()
	out := filepath.Join(dir, "auto")
	res, err := compiler.Compile(context.Background(), compiler.CompileRequest{
		Path:      src,
		OutPrefix: out,
		Name:      "AutoDetectTest",
	})
	require.NoError(t, err)
	require.Equal(t, compiler.LangGo, res.Language)
	require.FileExists(t, res.NEF)
}
