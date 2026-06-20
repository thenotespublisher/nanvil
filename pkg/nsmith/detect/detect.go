package detect

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Language identifies a Neo smart contract source language.
type Language string

const (
	LangGo     Language = "go"
	LangPython Language = "python"
	LangJava   Language = "java"
	LangCSharp Language = "csharp"
)

// Result holds language detection output.
type Result struct {
	Language Language
	Path     string
	Score    int
}

// Detect picks the best matching language for path (file or directory).
func Detect(path string) (Result, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return Result{}, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return Result{}, err
	}
	dir := abs
	if !info.IsDir() {
		dir = filepath.Dir(abs)
	}
	scores := map[Language]int{
		LangGo:     scoreGo(dir, abs, info.IsDir()),
		LangPython: scorePython(dir, abs, info.IsDir()),
		LangCSharp: scoreCSharp(dir, abs, info.IsDir()),
		LangJava:   scoreJava(dir, abs, info.IsDir()),
	}
	var best Result
	for lang, score := range scores {
		if score > best.Score {
			best = Result{Language: lang, Path: abs, Score: score}
		}
	}
	if best.Score == 0 {
		return Result{}, fmt.Errorf("could not detect contract language for %s", path)
	}
	if !info.IsDir() {
		best.Path = abs
	}
	return best, nil
}

func scoreGo(dir, path string, isDir bool) int {
	if !isDir && strings.EqualFold(filepath.Ext(path), ".go") {
		if hasNeoGoImport(path) {
			return 10
		}
		return 1
	}
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
	if score > 0 {
		return score + 2
	}
	return 0
}

func hasNeoGoImport(path string) bool {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	s := string(raw)
	return strings.Contains(s, "github.com/nspcc-dev/neo-go/pkg/interop")
}

func scorePython(dir, path string, isDir bool) int {
	if !isDir && strings.EqualFold(filepath.Ext(path), ".py") {
		if hasNeoBoaMarkers(path) {
			return 10
		}
		return 2
	}
	if hasRequirementsNeoBoa(dir) {
		return 8
	}
	score := 0
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(p), ".py") {
			score++
			if hasNeoBoaMarkers(p) {
				score += 5
			}
		}
		return nil
	})
	return score
}

func hasNeoBoaMarkers(path string) bool {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	s := string(raw)
	return strings.Contains(s, "neo3.boa") || strings.Contains(s, "@n3") ||
		strings.Contains(s, "from boa3")
}

func hasRequirementsNeoBoa(dir string) bool {
	raw, err := os.ReadFile(filepath.Join(dir, "requirements.txt"))
	if err != nil {
		return false
	}
	return strings.Contains(string(raw), "neo3-boa")
}

func scoreCSharp(dir, path string, isDir bool) int {
	if !isDir && strings.EqualFold(filepath.Ext(path), ".csproj") {
		if csprojReferencesNeo(path) {
			return 12
		}
		return 3
	}
	score := 0
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(p), ".csproj") && csprojReferencesNeo(p) {
			score += 10
		}
		return nil
	})
	return score
}

func csprojReferencesNeo(path string) bool {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(raw), "Neo.SmartContract.Framework")
}

func scoreJava(dir, _ string, _ bool) int {
	for _, name := range []string{"build.gradle", "build.gradle.kts"} {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		if strings.Contains(string(raw), "neow3j") || strings.Contains(string(raw), "io.neow3j") {
			return 12
		}
	}
	score := 0
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(p, "build.gradle") || strings.HasSuffix(p, "build.gradle.kts") {
			raw, err := os.ReadFile(p)
			if err == nil && (strings.Contains(string(raw), "neow3j") || strings.Contains(string(raw), "io.neow3j")) {
				score += 10
			}
		}
		return nil
	})
	return score
}
