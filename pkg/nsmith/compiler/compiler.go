package compiler

import (
	"context"
	"fmt"
)

// Language identifies a Neo smart contract source language.
type Language string

const (
	LangGo     Language = "go"
	LangPython Language = "python"
	LangJava   Language = "java"
	LangCSharp Language = "csharp"
)

// CompileRequest holds inputs for a compile invocation.
type CompileRequest struct {
	Path       string
	Lang       Language // empty = auto-detect
	OutPrefix  string
	Name       string
	ConfigPath string
	Debug      bool
	WorkDir    string
}

// CompileResult holds paths to generated artifacts.
type CompileResult struct {
	Language Language
	NEF      string
	Manifest string
	Extras   []string
}

// Backend compiles contracts for one language.
type Backend interface {
	Language() Language
	Detect(path string) (score int, err error)
	Ensure(ctx context.Context) error
	Compile(ctx context.Context, req CompileRequest) (CompileResult, error)
}

var backends = []Backend{
	&GoBackend{},
	&PythonBackend{},
	&CSharpBackend{},
	&JavaBackend{},
}

// Backends returns registered compiler backends.
func Backends() []Backend {
	out := make([]Backend, len(backends))
	copy(out, backends)
	return out
}

// BackendFor returns the backend for a language.
func BackendFor(lang Language) (Backend, error) {
	for _, b := range backends {
		if b.Language() == lang {
			return b, nil
		}
	}
	return nil, fmt.Errorf("unsupported language %q", lang)
}

// Compile runs the appropriate backend after optional auto-detection.
func Compile(ctx context.Context, req CompileRequest) (CompileResult, error) {
	lang := req.Lang
	target := req.Path
	if lang == "" {
		var err error
		lang, target, err = DetectLanguage(req.Path)
		if err != nil {
			return CompileResult{}, err
		}
		req.Path = target
		req.Lang = lang
	}
	b, err := BackendFor(lang)
	if err != nil {
		return CompileResult{}, err
	}
	if err := b.Ensure(ctx); err != nil {
		return CompileResult{}, err
	}
	res, err := b.Compile(ctx, req)
	if err != nil {
		return CompileResult{}, err
	}
	res.Language = lang
	return res, nil
}
