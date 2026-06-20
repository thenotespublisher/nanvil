package detect_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/nsmith/detect"
	"github.com/stretchr/testify/require"
)

func TestDetectGoContract(t *testing.T) {
	src := filepath.Join("..", "..", "..", "integration", "testcontracts", "explorer_deploy", "main.go")
	require.FileExists(t, src)
	res, err := detect.Detect(src)
	require.NoError(t, err)
	require.Equal(t, detect.LangGo, res.Language)
}

func TestDetectPythonRequirements(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("neo3-boa\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "contract.py"), []byte("from boa3 import public\n"), 0o644))
	res, err := detect.Detect(dir)
	require.NoError(t, err)
	require.Equal(t, detect.LangPython, res.Language)
}

func TestDetectCSharpCSProj(t *testing.T) {
	dir := t.TempDir()
	csproj := `<Project Sdk="Microsoft.NET.Sdk">
  <ItemGroup>
    <PackageReference Include="Neo.SmartContract.Framework" Version="3.9.1" />
  </ItemGroup>
</Project>`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Contract.csproj"), []byte(csproj), 0o644))
	res, err := detect.Detect(dir)
	require.NoError(t, err)
	require.Equal(t, detect.LangCSharp, res.Language)
}

func TestDetectJavaGradle(t *testing.T) {
	dir := t.TempDir()
	gradle := `plugins { id 'io.neow3j.gradle-plugin' version '3.24.0' }`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "build.gradle"), []byte(gradle), 0o644))
	res, err := detect.Detect(dir)
	require.NoError(t, err)
	require.Equal(t, detect.LangJava, res.Language)
}
