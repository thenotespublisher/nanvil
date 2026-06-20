package cli

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/urfave/cli/v2"
)

const version = "0.1.0"

// NewApp creates the nsmith CLI application.
func NewApp() *cli.App {
	app := &cli.App{
		Name:  "nsmith",
		Usage: "Multi-language Neo smart contract compiler (Nanvil forge)",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Machine-readable JSON output where supported",
			},
		},
		Commands: []*cli.Command{
			compileCmd(),
			buildCmd(),
			initCmd(),
			installCmd(),
			updateCmd(),
			toolchainListCmd(),
			doctorCmd(),
			versionCmd(),
		},
	}
	return app
}

func jsonOut(ctx *cli.Context) bool {
	for _, c := range ctx.Lineage() {
		if c.Bool("json") {
			return true
		}
	}
	return false
}

func versionCmd() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Print nsmith version",
		Action: func(c *cli.Context) error {
			goVer := "unknown"
			if info, ok := debug.ReadBuildInfo(); ok {
				goVer = info.Main.Version
				if goVer == "" || goVer == "(devel)" {
					goVer = version
				}
			}
			fmt.Printf("nsmith %s (embedded go compiler via %s)\n", version, goVer)
			return nil
		},
	}
}

var _ = os.Stderr
