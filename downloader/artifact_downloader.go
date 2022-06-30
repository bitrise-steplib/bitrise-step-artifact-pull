package downloader

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/bitrise-io/go-utils/filedownloader"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-steplib/bitrise-step-artifact-pull/api"
)

const (
	filePermission               = 0o655
	maxConcurrentDownloadThreads = 10
)

type ConcurrentArtifactDownloader struct {
	Artifacts []api.ArtifactResponseItemModel
	Logger    log.Logger
	TargetDir string
	Timeout   time.Duration
}

type ArtifactDownloadResult struct {
	DownloadError error
	DownloadPath  string
	DownloadURL   string
	EnvKey        string
}

type downloadJob struct {
	ResponseModel api.ArtifactResponseItemModel
	TargetDir     string
}

func (ad *ConcurrentArtifactDownloader) DownloadAndSaveArtifacts() ([]ArtifactDownloadResult, error) {
	if _, err := os.Stat(ad.TargetDir); os.IsNotExist(err) {
		if err := os.Mkdir(ad.TargetDir, filePermission); err != nil {
			return nil, err
		}
	}

	return ad.downloadParallel(ad.TargetDir)
}

func (ad *ConcurrentArtifactDownloader) downloadParallel(targetDir string) ([]ArtifactDownloadResult, error) {
	var downloadResults []ArtifactDownloadResult

	jobs := make(chan downloadJob, len(ad.Artifacts))
	results := make(chan ArtifactDownloadResult, len(ad.Artifacts))

	for i := 0; i < maxConcurrentDownloadThreads; i++ {
		go ad.download(jobs, results)
	}

	for _, artifact := range ad.Artifacts {
		jobs <- downloadJob{
			ResponseModel: artifact,
			TargetDir:     targetDir,
		}
	}
	close(jobs)

	for i := 0; i < len(ad.Artifacts); i++ {
		res := <-results
		downloadResults = append(downloadResults, res)
	}

	return downloadResults, nil
}

func (ad *ConcurrentArtifactDownloader) download(jobs <-chan downloadJob, results chan<- ArtifactDownloadResult) {
	for j := range jobs {
		fileFullPath := filepath.Join(j.TargetDir, j.ResponseModel.Title)

		ctx, cancel := context.WithTimeout(context.Background(), ad.Timeout)

		downloader := filedownloader.NewWithContext(ctx, retry.NewHTTPClient().StandardClient())
		err := downloader.Get(fileFullPath, j.ResponseModel.DownloadURL)

		cancel()

		if err != nil {
			results <- ArtifactDownloadResult{DownloadError: err, DownloadURL: j.ResponseModel.DownloadURL}
			return
		}

		results <- ArtifactDownloadResult{DownloadPath: fileFullPath, DownloadURL: j.ResponseModel.DownloadURL, EnvKey: j.ResponseModel.EnvKey}
	}
}

func NewConcurrentArtifactDownloader(artifacts []api.ArtifactResponseItemModel, timeout time.Duration, targetDir string, logger log.Logger) *ConcurrentArtifactDownloader {
	return &ConcurrentArtifactDownloader{
		Artifacts: artifacts,
		Timeout:   timeout,
		Logger:    logger,
		TargetDir: targetDir,
	}
}
