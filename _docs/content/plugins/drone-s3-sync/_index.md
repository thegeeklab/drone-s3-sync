---
title: drone-s3-sync
---

[![Build Status](https://img.shields.io/drone/build/thegeeklab/drone-s3-sync?logo=drone&server=https%3A%2F%2Fdrone.thegeeklab.de)](https://drone.thegeeklab.de/thegeeklab/drone-s3-sync)
[![Docker Hub](https://img.shields.io/badge/dockerhub-latest-blue.svg?logo=docker&logoColor=white)](https://hub.docker.com/r/thegeeklab/drone-s3-sync)
[![Quay.io](https://img.shields.io/badge/quay-latest-blue.svg?logo=docker&logoColor=white)](https://quay.io/repository/thegeeklab/drone-s3-sync)
[![GitHub contributors](https://img.shields.io/github/contributors/thegeeklab/drone-s3-sync)](https://github.com/thegeeklab/drone-s3-sync/graphs/contributors)
[![Source: GitHub](https://img.shields.io/badge/source-github-blue.svg?logo=github&logoColor=white)](https://github.com/thegeeklab/drone-s3-sync)
[![License: MIT](https://img.shields.io/github/license/thegeeklab/drone-s3-sync)](https://github.com/thegeeklab/drone-s3-sync/blob/main/LICENSE)

Drone plugin to synchronize a directory with an S3 bucket.

<!-- prettier-ignore-start -->
<!-- spellchecker-disable -->
{{< toc >}}
<!-- spellchecker-enable -->
<!-- prettier-ignore-end -->

## Build

Build the binary with the following command:

```Shell
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0
export GO111MODULE=on

make build
```

Build the Docker image with the following command:

```Shell
docker build --file docker/Dockerfile.amd64 --tag thegeeklab/drone-s3-sync .
```

## Usage

```Shell
docker run --rm \
  -e PLUGIN_BUCKET=my_bucket \
  -e AWS_ACCESS_KEY_ID=abc123 \
  -e AWS_SECRET_ACCESS_KEY=xyc789 \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  thegeeklab/drone-s3-sync
```

### Parameters

endpoint
: endpoint for the s3 connection

access-key
: s3 access key

secret-key
: s3 secret key

path-style
: use path style for bucket paths

bucket
: name of the bucket

region
: s3 region (default `us-east-1`)

source
: upload source path (default `.`)

target
: target path (default `/`)

delete
: delete locally removed files from the target

access
: access control settings

content-type
: content-type settings for uploads

content-encoding
: content-encoding settings for uploads

cache_control
: cache-control settings for uploads

metadata
: additional metadata for uploads

redirects
: redirects to create

cloudfront-distribution
: id of cloudfront distribution to invalidate

dry_run
: dry run disables api calls

max_concurrency
: customize number concurrent files to process (default `100`)
