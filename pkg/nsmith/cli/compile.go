package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	nsmithcompiler "github.com/nspcc-dev/neo-go/pkg/nsmith/compiler"
	"github.com/urfave/cli/v2"
)

func compileCmd() *cli.Command {
	return &cli.Command{
		Name:      "compile",
		Usage:     "Compile a Neo smart contract to NEF and manifest",
		ArgsUsage: "[path]",
		Flags:     compileFlags(),
		Action: func(c *cli.Context) error {
			path := "."
			if c.NArg() > 0 {
				path = c.Args().First()
			}
			req := compileRequestFromFlags(c, path)
			if req.ConfigPath == "" && req.Lang != nsmithcompiler.LangPython {
				tryConfig := filepath.Join(path, "contract.yml")
				if info, err := os.Stat(tryConfig); err == nil && !info.IsDir() {
					req.ConfigPath = tryConfig
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			res, err := nsmithcompiler.Compile(ctx, req)
			if err != nil {
				return err
			}
			return printCompileResult(c, res)
		},
	}
}

func parseLangs(c *cli.Context) []string {
	if c.Bool("all") {
		return []string{"go", "python", "java", "csharp"}
	}
	lang := c.String("lang")
	if lang == "" {
		return []string{"go", "python", "java", "csharp"}
	}
	parts := strings.Split(lang, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func langFlag() cli.Flag {
	return &cli.StringFlag{
		Name:  "lang",
		Usage: "Language(s): go, python, java, csharp (comma-separated) or use --all",
	}
}

func allFlag() cli.Flag {
	return &cli.BoolFlag{Name: "all", Usage: "All supported languages"}
}
