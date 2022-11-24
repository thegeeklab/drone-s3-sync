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
{{< propertylist name=drone-s3-sync.data sort=name >}}
<!-- spellchecker-enable -->
<!-- prettier-ignore-end -->

### Examples

**Customize `acl`, `content_type`, `content_encoding` or `cache_control`:**

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
      acl:
        "public/*": public-read
        "private/*": private
      content_type:
        ".svg": image/svg+xml
      content_encoding:
        ".js": gzip
        ".css": gzip
      cache_control: "public, max-age: 31536000"
```

All `map` parameters can be specified as `map` for a subset of files or as `string` for all files.

- For the `acl` parameter the key must be a glob. Files without a matching rule will default to `private`.
- For the `content_type` parameter, the key must be a file extension (including the leading dot). To apply a configuration to files without extension, the key can be set to an empty string `""`. For files without a matching rule, the content type is determined automatically.
- For the `content_encoding` parameter, the key must be a file extension (including the leading dot). To apply a configuration to files without extension, the key can be set to an empty string `""`. For files without a matching rule, no Content Encoding header is set.
- For the `cache_control` parameter, the key must be a file extension (including the leading dot). If you want to set cache control for files without an extension, set the key to the empty string `""`. For files without a matching rule, no Cache Control header is set.

**Sync to Minio S3:**

To use [Minio S3](https://docs.min.io/) its required to set `path_style: true`.

```YAML
kind: pipeline
name: default

steps:
  - name: sync
    image: thegeeklab/drone-s3-sync
    settings:
      endpoint: https://minio.example.com
      access_key: a50d28f4dd477bc184fbd10b376de753
      secret_key: bc5785d3ece6a9cdefa42eb99b58986f9095ff1c
      bucket: my-bucket
      source: folder/to/archive
      target: /target/location
      path_style: true
```

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
