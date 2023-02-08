package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// Settings for the Plugin.
type Settings struct {
	Endpoint               string
	AccessKey              string
	SecretKey              string
	Bucket                 string
	Region                 string
	Source                 string
	Target                 string
	Delete                 bool
	ACL                    map[string]string
	CacheControl           map[string]string
	ContentType            map[string]string
	ContentEncoding        map[string]string
	Metadata               map[string]map[string]string
	Redirects              map[string]string
	CloudFrontDistribution string
	DryRun                 bool
	PathStyle              bool
	Client                 AWS
	Jobs                   []Job
	MaxConcurrency         int
}

type Job struct {
	local  string
	remote string
	action string
}

type Result struct {
	j   Job
	err error
}

// Validate handles the settings validation of the plugin.
func (p *Plugin) Validate() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error while retrieving working directory: %w", err)
	}

	p.settings.Source = filepath.Join(wd, p.settings.Source)
	p.settings.Target = strings.TrimPrefix(p.settings.Target, "/")

	return nil
}

// Execute provides the implementation of the plugin.
func (p *Plugin) Execute() error {
	p.settings.Jobs = make([]Job, 1)
	p.settings.Client = NewAWS(p)

	if err := p.createSyncJobs(); err != nil {
		return fmt.Errorf("error while creating sync job: %w", err)
	}

	if len(p.settings.CloudFrontDistribution) > 0 {
		p.settings.Jobs = append(p.settings.Jobs, Job{
			local:  "",
			remote: filepath.Join("/", p.settings.Target, "*"),
			action: "invalidateCloudFront",
		})
	}

	if err := p.runJobs(); err != nil {
		return fmt.Errorf("error while creating sync job: %w", err)
	}

	return nil
}

func (p *Plugin) createSyncJobs() error {
	remote, err := p.settings.Client.List(p.settings.Target)
	if err != nil {
		return err
	}

	local := make([]string, 0)

	err = filepath.Walk(p.settings.Source, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		localPath := path
		if p.settings.Source != "." {
			localPath = strings.TrimPrefix(path, p.settings.Source)
			localPath = strings.TrimPrefix(localPath, "/")
		}
		local = append(local, localPath)
		p.settings.Jobs = append(p.settings.Jobs, Job{
			local:  filepath.Join(p.settings.Source, localPath),
			remote: filepath.Join(p.settings.Target, localPath),
			action: "upload",
		})

		return nil
	})
	if err != nil {
		return err
	}

	for path, location := range p.settings.Redirects {
		path = strings.TrimPrefix(path, "/")
		local = append(local, path)
		p.settings.Jobs = append(p.settings.Jobs, Job{
			local:  path,
			remote: location,
			action: "redirect",
		})
	}

	if p.settings.Delete {
		for _, remote := range remote {
			found := false
			remotePath := strings.TrimPrefix(remote, p.settings.Target+"/")

			for _, l := range local {
				if l == remotePath {
					found = true

					break
				}
			}

			if !found {
				p.settings.Jobs = append(p.settings.Jobs, Job{
					local:  "",
					remote: remote,
					action: "delete",
				})
			}
		}
	}

	return nil
}

func (p *Plugin) runJobs() error {
	client := p.settings.Client
	jobChan := make(chan struct{}, p.settings.MaxConcurrency)
	results := make(chan *Result, len(p.settings.Jobs))

	var invalidateJob *Job

	logrus.Infof("Synchronizing with bucket '%s'", p.settings.Bucket)

	for _, job := range p.settings.Jobs {
		jobChan <- struct{}{}

		go func(job Job) {
			var err error

			switch job.action {
			case "upload":
				err = client.Upload(job.local, job.remote)
			case "redirect":
				err = client.Redirect(job.local, job.remote)
			case "delete":
				err = client.Delete(job.remote)
			case "invalidateCloudFront":
				invalidateJob = &job
			default:
				err = nil
			}
			results <- &Result{job, err}

			<-jobChan
		}(job)
	}

	for range p.settings.Jobs {
		r := <-results
		if r.err != nil {
			return fmt.Errorf("failed to %s %s to %s: %w", r.j.action, r.j.local, r.j.remote, r.err)
		}
	}

	if invalidateJob != nil {
		err := client.Invalidate(invalidateJob.remote)
		if err != nil {
			return fmt.Errorf("failed to %s %s to %s: %w", invalidateJob.action, invalidateJob.local, invalidateJob.remote, err)
		}
	}

	return nil
}
