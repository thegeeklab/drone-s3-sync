package plugin

import (
	//nolint:gosec
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ryanuber/go-glob"
	"github.com/sirupsen/logrus"
)

type AWS struct {
	client   *s3.S3
	cfClient *cloudfront.CloudFront
	remote   []string
	local    []string
	plugin   *Plugin
}

func NewAWS(plugin *Plugin) AWS {
	sessCfg := &aws.Config{
		S3ForcePathStyle: aws.Bool(plugin.settings.PathStyle),
		Region:           aws.String(plugin.settings.Region),
	}

	if plugin.settings.Endpoint != "" {
		sessCfg.Endpoint = &plugin.settings.Endpoint
		sessCfg.DisableSSL = aws.Bool(strings.HasPrefix(plugin.settings.Endpoint, "http://"))
	}

	// allowing to use the instance role or provide a key and secret
	if plugin.settings.AccessKey != "" && plugin.settings.SecretKey != "" {
		sessCfg.Credentials = credentials.NewStaticCredentials(plugin.settings.AccessKey, plugin.settings.SecretKey, "")
	}

	sess, _ := session.NewSession(sessCfg)

	c := s3.New(sess)
	cf := cloudfront.New(sess)
	r := make([]string, 1)
	l := make([]string, 1)

	return AWS{c, cf, r, l, plugin}
}

//nolint:gocognit,gocyclo,maintidx
func (a *AWS) Upload(local, remote string) error {
	plugin := a.plugin

	if local == "" {
		return nil
	}

	file, err := os.Open(local)
	if err != nil {
		return err
	}

	defer file.Close()

	var acl string

	for pattern := range plugin.settings.ACL {
		if match := glob.Glob(pattern, local); match {
			acl = plugin.settings.ACL[pattern]

			break
		}
	}

	if acl == "" {
		acl = "private"
	}

	fileExt := filepath.Ext(local)

	var contentType string

	for patternExt := range plugin.settings.ContentType {
		if patternExt == fileExt {
			contentType = plugin.settings.ContentType[patternExt]

			break
		}
	}

	if contentType == "" {
		contentType = mime.TypeByExtension(fileExt)
	}

	var contentEncoding string

	for patternExt := range plugin.settings.ContentEncoding {
		if patternExt == fileExt {
			contentEncoding = plugin.settings.ContentEncoding[patternExt]

			break
		}
	}

	var cacheControl string

	for pattern := range plugin.settings.CacheControl {
		if match := glob.Glob(pattern, local); match {
			cacheControl = plugin.settings.CacheControl[pattern]

			break
		}
	}

	metadata := map[string]*string{}

	for pattern := range plugin.settings.Metadata {
		if match := glob.Glob(pattern, local); match {
			for k, v := range plugin.settings.Metadata[pattern] {
				metadata[k] = aws.String(v)
			}

			break
		}
	}

	var AWSErr awserr.Error

	head, err := a.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(plugin.settings.Bucket),
		Key:    aws.String(remote),
	})
	if err != nil && errors.As(err, &AWSErr) {
		//nolint:errorlint,forcetypeassert
		if err.(awserr.Error).Code() == "404" {
			return err
		}

		logrus.Debugf(
			"'%s' not found in bucket, uploading with content-type '%s' and permissions '%s'",
			local,
			contentType,
			acl,
		)

		putObject := &s3.PutObjectInput{
			Bucket:      aws.String(plugin.settings.Bucket),
			Key:         aws.String(remote),
			Body:        file,
			ContentType: aws.String(contentType),
			ACL:         aws.String(acl),
			Metadata:    metadata,
		}

		if len(cacheControl) > 0 {
			putObject.CacheControl = aws.String(cacheControl)
		}

		if len(contentEncoding) > 0 {
			putObject.ContentEncoding = aws.String(contentEncoding)
		}

		// skip upload during dry run
		if a.plugin.settings.DryRun {
			return nil
		}

		_, err = a.client.PutObject(putObject)

		return err
	}

	//nolint:gosec
	hash := md5.New()
	_, _ = io.Copy(hash, file)
	sum := fmt.Sprintf("'%x'", hash.Sum(nil))

	//nolint:nestif
	if sum == *head.ETag {
		shouldCopy := false

		if head.ContentType == nil && contentType != "" {
			logrus.Debugf("content-type has changed from unset to %s", contentType)

			shouldCopy = true
		}

		if !shouldCopy && head.ContentType != nil && contentType != *head.ContentType {
			logrus.Debugf("content-type has changed from %s to %s", *head.ContentType, contentType)

			shouldCopy = true
		}

		if !shouldCopy && head.ContentEncoding == nil && contentEncoding != "" {
			logrus.Debugf("Content-Encoding has changed from unset to %s", contentEncoding)

			shouldCopy = true
		}

		if !shouldCopy && head.ContentEncoding != nil && contentEncoding != *head.ContentEncoding {
			logrus.Debugf("Content-Encoding has changed from %s to %s", *head.ContentEncoding, contentEncoding)

			shouldCopy = true
		}

		if !shouldCopy && head.CacheControl == nil && cacheControl != "" {
			logrus.Debugf("cache-control has changed from unset to %s", cacheControl)

			shouldCopy = true
		}

		if !shouldCopy && head.CacheControl != nil && cacheControl != *head.CacheControl {
			logrus.Debugf("cache-control has changed from %s to %s", *head.CacheControl, cacheControl)

			shouldCopy = true
		}

		if !shouldCopy && len(head.Metadata) != len(metadata) {
			logrus.Debugf("count of metadata values has changed for %s", local)

			shouldCopy = true
		}

		if !shouldCopy && len(metadata) > 0 {
			for k, v := range metadata {
				if hv, ok := head.Metadata[k]; ok {
					if *v != *hv {
						logrus.Debugf("metadata values have changed for %s", local)

						shouldCopy = true

						break
					}
				}
			}
		}

		if !shouldCopy {
			grant, err := a.client.GetObjectAcl(&s3.GetObjectAclInput{
				Bucket: aws.String(plugin.settings.Bucket),
				Key:    aws.String(remote),
			})
			if err != nil {
				return err
			}

			previousACL := "private"

			for _, grant := range grant.Grants {
				grantee := *grant.Grantee
				if grantee.URI != nil {
					if *grantee.URI == "http://acs.amazonaws.com/groups/global/AllUsers" {
						if *grant.Permission == "READ" {
							previousACL = "public-read"
						} else if *grant.Permission == "WRITE" {
							previousACL = "public-read-write"
						}
					}

					if *grantee.URI == "http://acs.amazonaws.com/groups/global/AuthenticatedUsers" {
						if *grant.Permission == "READ" {
							previousACL = "authenticated-read"
						}
					}
				}
			}

			if previousACL != acl {
				logrus.Debugf("permissions for '%s' have changed from '%s' to '%s'", remote, previousACL, acl)

				shouldCopy = true
			}
		}

		if !shouldCopy {
			logrus.Debugf("skipping '%s' because hashes and metadata match", local)

			return nil
		}

		logrus.Debugf("updating metadata for '%s' content-type: '%s', ACL: '%s'", local, contentType, acl)

		copyObject := &s3.CopyObjectInput{
			Bucket:            aws.String(plugin.settings.Bucket),
			Key:               aws.String(remote),
			CopySource:        aws.String(fmt.Sprintf("%s/%s", plugin.settings.Bucket, remote)),
			ACL:               aws.String(acl),
			ContentType:       aws.String(contentType),
			Metadata:          metadata,
			MetadataDirective: aws.String("REPLACE"),
		}

		if len(cacheControl) > 0 {
			copyObject.CacheControl = aws.String(cacheControl)
		}

		if len(contentEncoding) > 0 {
			copyObject.ContentEncoding = aws.String(contentEncoding)
		}

		// skip update if dry run
		if a.plugin.settings.DryRun {
			return nil
		}

		_, err = a.client.CopyObject(copyObject)

		return err
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	logrus.Debugf("uploading '%s' with content-type '%s' and permissions '%s'", local, contentType, acl)

	putObject := &s3.PutObjectInput{
		Bucket:      aws.String(plugin.settings.Bucket),
		Key:         aws.String(remote),
		Body:        file,
		ContentType: aws.String(contentType),
		ACL:         aws.String(acl),
		Metadata:    metadata,
	}

	if len(cacheControl) > 0 {
		putObject.CacheControl = aws.String(cacheControl)
	}

	if len(contentEncoding) > 0 {
		putObject.ContentEncoding = aws.String(contentEncoding)
	}

	// skip upload if dry run
	if a.plugin.settings.DryRun {
		return nil
	}

	_, err = a.client.PutObject(putObject)

	return err
}

func (a *AWS) Redirect(path, location string) error {
	plugin := a.plugin

	logrus.Debugf("adding redirect from '%s' to '%s'", path, location)

	if a.plugin.settings.DryRun {
		return nil
	}

	_, err := a.client.PutObject(&s3.PutObjectInput{
		Bucket:                  aws.String(plugin.settings.Bucket),
		Key:                     aws.String(path),
		ACL:                     aws.String("public-read"),
		WebsiteRedirectLocation: aws.String(location),
	})

	return err
}

func (a *AWS) Delete(remote string) error {
	plugin := a.plugin

	logrus.Debugf("removing remote file '%s'", remote)

	if a.plugin.settings.DryRun {
		return nil
	}

	_, err := a.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(plugin.settings.Bucket),
		Key:    aws.String(remote),
	})

	return err
}

func (a *AWS) List(path string) ([]string, error) {
	plugin := a.plugin

	remote := make([]string, 0)

	resp, err := a.client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(plugin.settings.Bucket),
		Prefix: aws.String(path),
	})
	if err != nil {
		return remote, err
	}

	for _, item := range resp.Contents {
		remote = append(remote, *item.Key)
	}

	for *resp.IsTruncated {
		resp, err = a.client.ListObjects(&s3.ListObjectsInput{
			Bucket: aws.String(plugin.settings.Bucket),
			Prefix: aws.String(path),
			Marker: aws.String(remote[len(remote)-1]),
		})

		if err != nil {
			return remote, err
		}

		for _, item := range resp.Contents {
			remote = append(remote, *item.Key)
		}
	}

	return remote, nil
}

func (a *AWS) Invalidate(invalidatePath string) error {
	p := a.plugin

	logrus.Debugf("invalidating '%s'", invalidatePath)

	_, err := a.cfClient.CreateInvalidation(&cloudfront.CreateInvalidationInput{
		DistributionId: aws.String(p.settings.CloudFrontDistribution),
		InvalidationBatch: &cloudfront.InvalidationBatch{
			CallerReference: aws.String(time.Now().Format(time.RFC3339Nano)),
			Paths: &cloudfront.Paths{
				Quantity: aws.Int64(1),
				Items: []*string{
					aws.String(invalidatePath),
				},
			},
		},
	})

	return err
}
