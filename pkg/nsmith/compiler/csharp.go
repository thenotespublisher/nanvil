package compiler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nspcc-dev/neo-go/pkg/nsmith/detect"
)

// CSharpBackend compiles C# contracts via nccs.
type CSharpBackend struct{}

func (CSharpBackend) Language() Language { return LangCSharp }

func (CSharpBackend) Detect(path string) (int, error) {
	res, err := detect.Detect(path)
	if err != nil {
		if strings.EqualFold(filepath.Ext(path), ".csproj") {
			return 3, nil
		}
		return 0, nil
	}
	if res.Language != detect.LangCSharp {
		return 0, nil
	}
	return res.Score, nil
}

func (CSharpBackend) Ensure(ctx context.Context) error {
	m, err := defaultManager()
	if err != nil {
		return err
	}
	return m.EnsureLanguage(ctx, "csharp")
}

func (CSharpBackend) Compile(ctx context.Context, req CompileRequest) (CompileResult, error) {
	csproj := req.Path
	info, err := os.Stat(csproj)
	if err != nil {
		return CompileResult{}, err
	}
	if info.IsDir() {
		csproj, err = findCSProj(csproj)
		if err != nil {
			return CompileResult{}, err
		}
	} else if !strings.EqualFold(filepath.Ext(csproj), ".csproj") {
		return CompileResult{}, fmt.Errorf("C# compile expects a .csproj file or project directory")
	}
	m, err := defaultManager()
	if err != nil {
		return CompileResult{}, err
	}
	dir := filepath.Dir(csproj)
	csFile, err := findContractCS(dir, csproj)
	if err != nil {
		return CompileResult{}, err
	}
	if err := runCmd(ctx, dir, compileEnv(), m.NCCSPath(), csFile); err != nil {
		return CompileResult{}, fmt.Errorf("nccs: %w", err)
	}
	scDir := filepath.Join(dir, "bin", "sc")
	nef, err := findArtifact(scDir, ".nef")
	if err != nil {
		// nccs may emit next to csproj
		nef, err = findArtifact(dir, ".nef")
		if err != nil {
			return CompileResult{}, err
		}
	}
	manifest := strings.TrimSuffix(nef, ".nef") + ".manifest.json"
	if _, err := os.Stat(manifest); err != nil {
		return CompileResult{}, fmt.Errorf("manifest not found beside %s", nef)
	}
	if req.OutPrefix != "" {
		out := req.OutPrefix
		if !filepath.IsAbs(out) {
			out = filepath.Join(dir, out)
		}
		if err := copyIfDifferent(nef, out+".nef"); err != nil {
			return CompileResult{}, err
		}
		if err := copyIfDifferent(manifest, out+".manifest.json"); err != nil {
			return CompileResult{}, err
		}
		nef, manifest = out+".nef", out+".manifest.json"
	}
	return CompileResult{NEF: nef, Manifest: manifest}, nil
}

func findContractCS(dir, csproj string) (string, error) {
	var csFiles []string
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.EqualFold(filepath.Ext(p), ".cs") {
			return nil
		}
		base := filepath.Base(p)
		if base == "AssemblyInfo.cs" {
			return nil
		}
		if strings.Contains(filepath.ToSlash(p), "/obj/") || strings.Contains(filepath.ToSlash(p), "/bin/") {
			return nil
		}
		csFiles = append(csFiles, p)
		return nil
	})
	if len(csFiles) == 0 {
		return "", fmt.Errorf("no .cs contract source in %s", dir)
	}
	for _, p := range csFiles {
		if strings.EqualFold(filepath.Base(p), "Contract.cs") {
			return p, nil
		}
	}
	if len(csFiles) == 1 {
		return csFiles[0], nil
	}
	return "", fmt.Errorf("multiple .cs files in %s; expected Contract.cs (found %d files)", dir, len(csFiles))
}

func findCSProj(dir string) (string, error) {
	var projects []string
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(p), ".csproj") {
			projects = append(projects, p)
		}
		return nil
	})
	if len(projects) == 0 {
		return "", fmt.Errorf("no .csproj in %s", dir)
	}
	return projects[0], nil
}
