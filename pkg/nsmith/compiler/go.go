package compiler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gocompiler "github.com/nspcc-dev/neo-go/pkg/compiler"
)

// GoBackend compiles Go Neo contracts via embedded pkg/compiler.
type GoBackend struct{}

func (GoBackend) Language() Language { return LangGo }

func (GoBackend) Detect(path string) (int, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	if info.IsDir() {
		return scoreGoDir(path)
	}
	if strings.EqualFold(filepath.Ext(path), ".go") {
		if hasNeoGoImport(path) {
			return 10, nil
		}
		return 1, nil
	}
	return 0, nil
}

func scoreGoDir(dir string) (int, error) {
	score := 0
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.EqualFold(filepath.Ext(p), ".go") {
			return nil
		}
		score++
		if hasNeoGoImport(p) {
			score += 5
		}
		return nil
	})
	if score == 0 {
		return 0, nil
	}
	return score + 2, nil
}

func hasNeoGoImport(path string) bool {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	s := string(raw)
	return strings.Contains(s, "github.com/nspcc-dev/neo-go/pkg/interop") ||
		strings.Contains(s, `"github.com/nspcc-dev/neo-go/pkg/interop`)
}

func (GoBackend) Ensure(context.Context) error { return nil }

func (GoBackend) Compile(_ context.Context, req CompileRequest) (CompileResult, error) {
	src := req.Path
	if src == "" {
		return CompileResult{}, fmt.Errorf("go compile: missing source path")
	}
	info, err := os.Stat(src)
	if err != nil {
		return CompileResult{}, err
	}
	if info.IsDir() {
		src, err = findMainGo(src)
		if err != nil {
			return CompileResult{}, err
		}
	}

	out := req.OutPrefix
	if out == "" {
		out = strings.TrimSuffix(filepath.Base(src), filepath.Ext(src))
	}
	workDir := req.WorkDir
	if workDir == "" {
		workDir = filepath.Dir(src)
	}
	outPath := out
	if !filepath.IsAbs(outPath) {
		outPath = filepath.Join(workDir, out)
	}

	opts := &gocompiler.Options{
		Name:         req.Name,
		Outfile:      outPath,
		Ext:          "nef",
		ManifestFile: outPath + ".manifest.json",
	}
	if req.Debug {
		opts.DebugInfo = outPath + ".nefdbgnfo"
	}
	if req.ConfigPath != "" {
		conf, err := gocompiler.ParseProjectConfig(req.ConfigPath)
		if err != nil {
			return CompileResult{}, err
		}
		if opts.Name == "" {
			opts.Name = conf.Name
		}
		opts.ContractEvents = conf.Events
		opts.DeclaredNamedTypes = conf.NamedTypes
		opts.ContractSupportedStandards = conf.SupportedStandards
		opts.Permissions = conf.PermissionsManifest()
		opts.SafeMethods = conf.SafeMethods
		opts.Overloads = conf.Overloads
		opts.SourceURL = conf.SourceURL
	}
	if opts.Name == "" {
		opts.Name = strings.TrimSuffix(filepath.Base(src), filepath.Ext(src))
	}

	if _, err := gocompiler.CompileAndSave(src, opts); err != nil {
		return CompileResult{}, err
	}
	res := CompileResult{
		NEF:      outPath + ".nef",
		Manifest: outPath + ".manifest.json",
	}
	if req.Debug && opts.DebugInfo != "" {
		res.Extras = append(res.Extras, opts.DebugInfo)
	}
	return res, nil
}

func findMainGo(dir string) (string, error) {
	var candidates []string
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(p), ".go") && !strings.HasSuffix(p, "_test.go") {
			candidates = append(candidates, p)
		}
		return nil
	})
	if len(candidates) == 0 {
		return "", fmt.Errorf("no .go contract source in %s", dir)
	}
	for _, c := range candidates {
		if filepath.Base(c) == "main.go" {
			return c, nil
		}
	}
	return candidates[0], nil
}
