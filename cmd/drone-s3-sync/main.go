package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/thegeeklab/drone-plugin-lib/v2/urfave"
	"github.com/thegeeklab/drone-s3-sync/plugin"
	"github.com/urfave/cli/v2"
)

//nolint:gochecknoglobals
var (
	BuildVersion = "devel"
	BuildDate    = "00000000"
)

var ErrTypeAssertionFailed = errors.New("type assertion failed")

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
		Flags:   append(settingsFlags(settings, urfave.FlagsPluginCategory), urfave.Flags()...),
		Action:  run(settings),
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(settings *plugin.Settings) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		urfave.LoggingFromContext(ctx)

		acl, ok := ctx.Generic("acl").(*StringMapFlag)
		if !ok {
			return fmt.Errorf("%w: failed to read acl input", ErrTypeAssertionFailed)
		}

		cacheControl, ok := ctx.Generic("cache-control").(*StringMapFlag)
		if !ok {
			return fmt.Errorf("%w: failed to read cache-control input", ErrTypeAssertionFailed)
		}

		contentType, ok := ctx.Generic("content-type").(*StringMapFlag)
		if !ok {
			return fmt.Errorf("%w: failed to read content-type input", ErrTypeAssertionFailed)
		}

		contentEncoding, ok := ctx.Generic("content-encoding").(*StringMapFlag)
		if !ok {
			return fmt.Errorf("%w: failed to read content-encoding input", ErrTypeAssertionFailed)
		}

		metadata, ok := ctx.Generic("metadata").(*DeepStringMapFlag)
		if !ok {
			return fmt.Errorf("%w: failed to read metadata input", ErrTypeAssertionFailed)
		}

		redirects, ok := ctx.Generic("redirects").(*MapFlag)
		if !ok {
			return fmt.Errorf("%w: failed to read redirects input", ErrTypeAssertionFailed)
		}

		settings.ACL = acl.Get()
		settings.CacheControl = cacheControl.Get()
		settings.ContentType = contentType.Get()
		settings.ContentEncoding = contentEncoding.Get()
		settings.Metadata = metadata.Get()
		settings.Redirects = redirects.Get()

		plugin := plugin.New(
			*settings,
			urfave.PipelineFromContext(ctx),
			urfave.NetworkFromContext(ctx),
		)

		if err := plugin.Validate(); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		if err := plugin.Execute(); err != nil {
			return fmt.Errorf("execution failed: %w", err)
		}

		return nil
	}
}
