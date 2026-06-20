package compiler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nspcc-dev/neo-go/pkg/nsmith/detect"
)

// JavaBackend compiles Java contracts via Gradle neow3jCompile.
type JavaBackend struct{}

func (JavaBackend) Language() Language { return LangJava }

func (JavaBackend) Detect(path string) (int, error) {
	res, err := detect.Detect(path)
	if err != nil {
		return 0, nil
	}
	if res.Language != detect.LangJava {
		return 0, nil
	}
	return res.Score, nil
}

func (JavaBackend) Ensure(ctx context.Context) error {
	m, err := defaultManager()
	if err != nil {
		return err
	}
	return m.EnsureLanguage(ctx, "java")
}

func (JavaBackend) Compile(ctx context.Context, req CompileRequest) (CompileResult, error) {
	projectDir := req.Path
	info, err := os.Stat(projectDir)
	if err != nil {
		return CompileResult{}, err
	}
	if !info.IsDir() {
		projectDir = filepath.Dir(projectDir)
	}
	gradleCmd, err := resolveGradle(projectDir)
	if err != nil {
		return CompileResult{}, err
	}
	if err := runCmd(ctx, projectDir, compileEnv(), gradleCmd, "neow3jCompile"); err != nil {
		return CompileResult{}, fmt.Errorf("gradle neow3jCompile: %w", err)
	}
	outDir := filepath.Join(projectDir, "build", "neow3j")
	nef, err := findArtifact(outDir, ".nef")
	if err != nil {
		return CompileResult{}, err
	}
	manifest := strings.TrimSuffix(nef, ".nef") + ".manifest.json"
	if _, err := os.Stat(manifest); err != nil {
		return CompileResult{}, fmt.Errorf("manifest not found beside %s", nef)
	}
	if req.OutPrefix != "" {
		out := req.OutPrefix
		if !filepath.IsAbs(out) {
			out = filepath.Join(projectDir, out)
		}
		if err := copyIfDifferent(nef, out+".nef"); err != nil {
			return CompileResult{}, err
		}
		if err := copyIfDifferent(manifest, out+".manifest.json"); err != nil {
			return CompileResult{}, err
		}
		nef, manifest = out+".nef", out+".manifest.json"
	}
	res := CompileResult{NEF: nef, Manifest: manifest}
	if req.Debug {
		dbg := strings.TrimSuffix(nef, ".nef") + ".nefdbgnfo"
		if _, err := os.Stat(dbg); err == nil {
			res.Extras = append(res.Extras, dbg)
		}
	}
	return res, nil
}

func resolveGradle(dir string) (string, error) {
	gradlew := filepath.Join(dir, "gradlew")
	if _, err := os.Stat(gradlew); err == nil {
		return gradlew, nil
	}
	gradlewBat := filepath.Join(dir, "gradlew.bat")
	if _, err := os.Stat(gradlewBat); err == nil {
		return gradlewBat, nil
	}
	if g, err := exec.LookPath("gradle"); err == nil {
		return g, nil
	}
	return "", fmt.Errorf("gradle not found in %s (add gradlew or install gradle)", dir)
}
