package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	nsmithcompiler "github.com/nspcc-dev/neo-go/pkg/nsmith/compiler"
	"github.com/nspcc-dev/neo-go/pkg/nsmith/initscaffold"
	"github.com/nspcc-dev/neo-go/pkg/nsmith/toolchain"
	"github.com/urfave/cli/v2"
)

func initCmd() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     "Scaffold a minimal contract project",
		ArgsUsage: "<name>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "lang", Required: true, Usage: "go, python, java, or csharp"},
			&cli.StringFlag{Name: "dir", Usage: "Output directory (default: name)"},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("usage: nsmith init <name> --lang go|python|java|csharp")
			}
			dir, err := initscaffold.Create(initscaffold.Options{
				Name: c.Args().First(),
				Lang: c.String("lang"),
				Dir:  c.String("dir"),
			})
			if err != nil {
				return err
			}
			fmt.Printf("created %s project in %s\n", c.String("lang"), dir)
			return nil
		},
	}
}

func installCmd() *cli.Command {
	return &cli.Command{
		Name:  "install",
		Usage: "Download and pin compiler toolchains",
		Flags: []cli.Flag{langFlag(), allFlag()},
		Action: func(c *cli.Context) error {
			return runToolchain(c, false)
		},
	}
}

func updateCmd() *cli.Command {
	return &cli.Command{
		Name:  "update",
		Usage: "Update pinned compiler toolchains to latest versions",
		Flags: []cli.Flag{langFlag(), allFlag()},
		Action: func(c *cli.Context) error {
			return runToolchain(c, true)
		},
	}
}

func runToolchain(c *cli.Context, update bool) error {
	m, err := toolchain.NewManager()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	return m.Install(ctx, toolchain.InstallOptions{
		Languages: parseLangs(c),
		Update:    update,
	})
}

func toolchainListCmd() *cli.Command {
	return &cli.Command{
		Name:  "toolchain",
		Usage: "Toolchain management",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "Show installed and latest available toolchain versions",
				Action: func(c *cli.Context) error {
					m, err := toolchain.NewManager()
					if err != nil {
						return err
					}
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					installed, latest, err := m.ListInstalled(ctx)
					if err != nil {
						return err
					}
					fmt.Printf("cache: %s\n\n", m.Root)
					for _, lang := range []string{"go", "python", "java", "csharp"} {
						ins := installed[lang]
						if ins == "" {
							ins = "(not installed)"
						}
						fmt.Printf("  %-8s installed: %-16s latest: %s\n", lang, ins, latest[lang])
					}
					return nil
				},
			},
		},
	}
}

func doctorCmd() *cli.Command {
	return &cli.Command{
		Name:  "doctor",
		Usage: "Check host prerequisites and toolchain installs",
		Flags: []cli.Flag{langFlag(), allFlag()},
		Action: func(c *cli.Context) error {
			m, err := toolchain.NewManager()
			if err != nil {
				return err
			}
			checks := m.Doctor(parseLangs(c))
			ok := true
			for _, ch := range checks {
				status := "FAIL"
				if ch.OK {
					status = "OK"
				} else {
					ok = false
				}
				fmt.Printf("[%s] %-8s %s\n", status, ch.Language, ch.Detail)
			}
			if !ok {
				return fmt.Errorf("doctor found issues")
			}
			return nil
		},
	}
}

func buildCmd() *cli.Command {
	return &cli.Command{
		Name:      "build",
		Usage:     "Compile and optionally deploy via ncast",
		ArgsUsage: "[path]",
		Flags: append(compileFlags(), []cli.Flag{
			&cli.BoolFlag{Name: "deploy", Usage: "Deploy after compile using ncast"},
			&cli.StringFlag{Name: "wif", Usage: "Signer WIF for deploy", EnvVars: []string{"NCAST_WIF", "NANVIL_WIF"}},
			&cli.StringFlag{Name: "rpc", Usage: "RPC URL for deploy", EnvVars: []string{"NCAST_RPC", "NANVIL_RPC"}, Value: "http://127.0.0.1:8545"},
		}...),
		Action: func(c *cli.Context) error {
			path := "."
			if c.NArg() > 0 {
				path = c.Args().First()
			}
			req := compileRequestFromFlags(c, path)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			res, err := nsmithcompiler.Compile(ctx, req)
			if err != nil {
				return err
			}
			if !c.Bool("deploy") {
				return printCompileResult(c, res)
			}
			wif := c.String("wif")
			if wif == "" {
				return fmt.Errorf("--deploy requires --wif or NCAST_WIF")
			}
			ncast, err := exec.LookPath("ncast")
			if err != nil {
				bin := "./bin/ncast"
				if _, statErr := os.Stat(bin); statErr != nil {
					return fmt.Errorf("ncast not found in PATH or ./bin/ncast")
				}
				ncast = bin
			}
			cmd := exec.CommandContext(ctx, ncast,
				"--rpc", c.String("rpc"),
				"deploy", "--wif", wif,
				"--nef", res.NEF,
				"--manifest", res.Manifest,
			)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		},
	}
}

func compileFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "lang", Usage: "Source language: go, python, java, csharp"},
		&cli.StringFlag{Name: "out", Usage: "Output file prefix (without extension)"},
		&cli.StringFlag{Name: "name", Usage: "Contract manifest name"},
		&cli.StringFlag{Name: "config", Usage: "Go contract YAML config path"},
		&cli.BoolFlag{Name: "debug", Usage: "Emit debug artifacts where supported"},
	}
}

func compileRequestFromFlags(c *cli.Context, path string) nsmithcompiler.CompileRequest {
	req := nsmithcompiler.CompileRequest{
		Path:       path,
		Lang:       nsmithcompiler.Language(c.String("lang")),
		OutPrefix:  c.String("out"),
		Name:       c.String("name"),
		ConfigPath: c.String("config"),
		Debug:      c.Bool("debug"),
	}
	return req
}

func printCompileResult(c *cli.Context, res nsmithcompiler.CompileResult) error {
	if jsonOut(c) {
		return encodeJSON(c, res)
	}
	w := c.App.Writer
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintf(w, "language: %s\n", res.Language)
	fmt.Fprintf(w, "nef:      %s\n", res.NEF)
	fmt.Fprintf(w, "manifest: %s\n", res.Manifest)
	for _, extra := range res.Extras {
		fmt.Fprintf(w, "extra:    %s\n", extra)
	}
	return nil
}

func encodeJSON(c *cli.Context, v any) error {
	w := c.App.Writer
	if w == nil {
		w = os.Stdout
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
