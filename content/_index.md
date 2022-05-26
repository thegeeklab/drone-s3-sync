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

## Usage

```YAML
kind: pipeline
name: default

steps:
  - name: sync
    image: thegeeklab/drone-s3-sync
    settings:
      access_key: a50d28f4dd477bc184fbd10b376de753
      secret_key: bc5785d3ece6a9cdefa42eb99b58986f9095ff1c
      region: us-east-1
      bucket: my-bucket.s3-website-us-east-1.amazonaws.com
      source: folder/to/archive
      target: /target/location
```

### Parameters

<!-- prettier-ignore-start -->
<!-- spellchecker-disable -->
{{< propertylist name=drone-s3-sync.data >}}
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

## Test

```Shell
docker run --rm \
  -e PLUGIN_BUCKET=my_bucket \
  -e AWS_ACCESS_KEY_ID=abc123 \
  -e AWS_SECRET_ACCESS_KEY=xyc789 \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  thegeeklab/drone-s3-sync
```
