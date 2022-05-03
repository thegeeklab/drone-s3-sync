package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/thegeeklab/drone-plugin-lib/errors"
	"github.com/thegeeklab/drone-plugin-lib/urfave"
	"github.com/thegeeklab/drone-s3-sync/plugin"
	"github.com/urfave/cli/v2"
)

var (
	BuildVersion = "devel"
	BuildDate    = "00000000"
)

func main() {
	settings := &plugin.Settings{}

	if _, err := os.Stat("/run/drone/env"); err == nil {
		_ = godotenv.Overload("/run/drone/env")
	}

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("%s version=%s date=%s\n", c.App.Name, c.App.Version, BuildDate)
	}

	app := &cli.App{
		Name:    "drone-s3-sync",
		Usage:   "synchronize a directory with an S3 bucket",
		Version: BuildVersion,
		Flags:   append(settingsFlags(settings), urfave.Flags()...),
		Action:  run(settings),
	}

	if err := app.Run(os.Args); err != nil {
		errors.HandleExit(err)
	}
}

func run(settings *plugin.Settings) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		urfave.LoggingFromContext(ctx)

		settings.Access = ctx.Generic("access").(*StringMapFlag).Get()
		settings.CacheControl = ctx.Generic("cache-control").(*StringMapFlag).Get()
		settings.ContentType = ctx.Generic("content-type").(*StringMapFlag).Get()
		settings.ContentEncoding = ctx.Generic("content-encoding").(*StringMapFlag).Get()
		settings.Metadata = ctx.Generic("metadata").(*DeepStringMapFlag).Get()
		settings.Redirects = ctx.Generic("redirects").(*MapFlag).Get()

		plugin := plugin.New(
			*settings,
			urfave.PipelineFromContext(ctx),
			urfave.NetworkFromContext(ctx),
		)

		if err := plugin.Validate(); err != nil {
			if e, ok := err.(errors.ExitCoder); ok {
				return e
			}

			return errors.ExitMessagef("validation failed: %w", err)
		}

		if err := plugin.Execute(); err != nil {

			if e, ok := err.(errors.ExitCoder); ok {
				return e
			}

			return errors.ExitMessagef("execution failed: %w", err)
		}

		return nil
	}
}
