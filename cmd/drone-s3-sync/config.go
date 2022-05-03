package main

import (
	"github.com/thegeeklab/drone-s3-sync/plugin"
	"github.com/urfave/cli/v2"
)

// settingsFlags has the cli.Flags for the plugin.Settings.
func settingsFlags(settings *plugin.Settings) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "endpoint",
			Usage:       "endpoint for the s3 connection",
			EnvVars:     []string{"PLUGIN_ENDPOINT", "S3_SYNC_ENDPOINT", "S3_ENDPOINT"},
			Destination: &settings.Endpoint,
		},
		&cli.StringFlag{
			Name:        "access-key",
			Usage:       "aws access key",
			EnvVars:     []string{"PLUGIN_ACCESS_KEY", "AWS_ACCESS_KEY_ID"},
			Destination: &settings.AccessKey,
		},
		&cli.StringFlag{
			Name:        "secret-key",
			Usage:       "aws secret key",
			EnvVars:     []string{"PLUGIN_SECRET_KEY", "AWS_SECRET_ACCESS_KEY"},
			Destination: &settings.SecretKey,
		},
		&cli.BoolFlag{
			Name:        "path-style",
			Usage:       "use path style for bucket paths",
			EnvVars:     []string{"PLUGIN_PATH_STYLE"},
			Destination: &settings.PathStyle,
		},
		&cli.StringFlag{
			Name:        "bucket",
			Usage:       "name of bucket",
			EnvVars:     []string{"PLUGIN_BUCKET"},
			Destination: &settings.Bucket,
		},
		&cli.StringFlag{
			Name:        "region",
			Usage:       "aws region",
			Value:       "us-east-1",
			EnvVars:     []string{"PLUGIN_REGION"},
			Destination: &settings.Region,
		},
		&cli.StringFlag{
			Name:        "source",
			Usage:       "upload source path",
			Value:       ".",
			EnvVars:     []string{"PLUGIN_SOURCE"},
			Destination: &settings.Source,
		},
		&cli.StringFlag{
			Name:        "target",
			Usage:       "target path",
			Value:       "/",
			EnvVars:     []string{"PLUGIN_TARGET"},
			Destination: &settings.Target,
		},
		&cli.BoolFlag{
			Name:        "delete",
			Usage:       "delete locally removed files from the target",
			EnvVars:     []string{"PLUGIN_DELETE"},
			Destination: &settings.Delete,
		},
		&cli.GenericFlag{
			Name:    "access",
			Usage:   "access control settings",
			EnvVars: []string{"PLUGIN_ACCESS", "PLUGIN_ACL"},
			Value:   &StringMapFlag{},
		},
		&cli.GenericFlag{
			Name:    "content-type",
			Usage:   "content-type settings for uploads",
			EnvVars: []string{"PLUGIN_CONTENT_TYPE"},
			Value:   &StringMapFlag{},
		},
		&cli.GenericFlag{
			Name:    "content-encoding",
			Usage:   "content-encoding settings for uploads",
			EnvVars: []string{"PLUGIN_CONTENT_ENCODING"},
			Value:   &StringMapFlag{},
		},
		&cli.GenericFlag{
			Name:    "cache-control",
			Usage:   "cache-control settings for uploads",
			EnvVars: []string{"PLUGIN_CACHE_CONTROL"},
			Value:   &StringMapFlag{},
		},
		&cli.GenericFlag{
			Name:    "metadata",
			Usage:   "additional metadata for uploads",
			EnvVars: []string{"PLUGIN_METADATA"},
			Value:   &DeepStringMapFlag{},
		},
		&cli.GenericFlag{
			Name:    "redirects",
			Usage:   "redirects to create",
			EnvVars: []string{"PLUGIN_REDIRECTS"},
			Value:   &MapFlag{},
		},
		&cli.StringFlag{
			Name:        "cloudfront-distribution",
			Usage:       "id of cloudfront distribution to invalidate",
			EnvVars:     []string{"PLUGIN_CLOUDFRONT_DISTRIBUTION"},
			Destination: &settings.CloudFrontDistribution,
		},
		&cli.BoolFlag{
			Name:        "dry-run",
			Usage:       "dry run disables api calls",
			EnvVars:     []string{"DRY_RUN", "PLUGIN_DRY_RUN"},
			Destination: &settings.DryRun,
		},
		&cli.StringFlag{
			Name:        "env-file",
			Usage:       "source env file",
			Destination: &settings.EnvFile,
		},
		&cli.IntFlag{
			Name:        "max-concurrency",
			Usage:       "customize number concurrent files to process",
			Value:       100,
			EnvVars:     []string{"PLUGIN_MAX_CONCURRENCY"},
			Destination: &settings.MaxConcurrency,
		},
	}
}
