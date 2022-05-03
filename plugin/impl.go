package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
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
	Access                 map[string]string
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
	EnvFile                string
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

var MissingAwsValuesMessage = "Must set 'bucket'"

// Validate handles the settings validation of the plugin.
func (p *Plugin) Validate() error {
	if len(p.settings.Bucket) == 0 {
		return fmt.Errorf("no bucket name provided")
	}

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
	if p.settings.EnvFile != "" {
		_ = godotenv.Load(p.settings.EnvFile)
	}

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

	local := make([]string, 1)

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
		for _, r := range remote {
			found := false
			rPath := strings.TrimPrefix(r, p.settings.Target+"/")
			for _, l := range local {
				if l == rPath {
					found = true
					break
				}
			}

			if !found {
				p.settings.Jobs = append(p.settings.Jobs, Job{
					local:  "",
					remote: r,
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
	for _, j := range p.settings.Jobs {
		jobChan <- struct{}{}
		go func(j Job) {
			var err error
			switch j.action {
			case "upload":
				err = client.Upload(j.local, j.remote)
			case "redirect":
				err = client.Redirect(j.local, j.remote)
			case "delete":
				err = client.Delete(j.remote)
			case "invalidateCloudFront":
				invalidateJob = &j
			default:
				err = nil
			}
			results <- &Result{j, err}
			<-jobChan
		}(j)
	}

	for range p.settings.Jobs {
		r := <-results
		if r.err != nil {
			return fmt.Errorf("failed to %s %s to %s: %+v", r.j.action, r.j.local, r.j.remote, r.err)
		}
	}

	if invalidateJob != nil {
		err := client.Invalidate(invalidateJob.remote)
		if err != nil {
			return fmt.Errorf("failed to %s %s to %s: %+v", invalidateJob.action, invalidateJob.local, invalidateJob.remote, err)
		}
	}

	return nil
}
