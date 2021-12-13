package downloader

import (
	"fmt"
	"os"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/bitrise-step-artifact-pull/api"
)

const (
	filePermission               = 0o755
	maxConcurrentDownloadThreads = 10
)

type ArtifactDownloader interface {
	DownloadAndSaveArtifacts() ([]ArtifactDownloadResult, error)
}

type ConcurrentArtifactDownloader struct {
	Artifacts  []api.ArtifactResponseItemModel
	Downloader FileDownloader
	Logger     log.Logger
	TargetDir  string
}

type ArtifactDownloadResult struct {
	DownloadError error
	DownloadPath  string
	DownloadURL   string
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
		fileContent, err := ad.Downloader.DownloadFileFromURL(j.ResponseModel.DownloadPath)

		if err != nil {
			results <- ArtifactDownloadResult{DownloadError: err, DownloadURL: j.ResponseModel.DownloadPath}
		}

		fileFullPath := fmt.Sprintf("%s/%s", j.TargetDir, j.ResponseModel.Title)

		out, err := os.Create(fileFullPath)
		if err != nil {
			results <- ArtifactDownloadResult{DownloadError: err, DownloadURL: j.ResponseModel.DownloadPath}
		}

		if _, err := out.Write(fileContent); err != nil {
			results <- ArtifactDownloadResult{DownloadError: err, DownloadURL: j.ResponseModel.DownloadPath}
		}

		if err := out.Close(); err != nil {
			ad.Logger.Errorf("couldn't close file: %s", out.Name())
		}

		results <- ArtifactDownloadResult{DownloadPath: fileFullPath, DownloadURL: j.ResponseModel.DownloadPath}
	}
}

func NewConcurrentArtifactDownloader(artifacts []api.ArtifactResponseItemModel, downloader FileDownloader, targetDir string, logger log.Logger) *ConcurrentArtifactDownloader {
	return &ConcurrentArtifactDownloader{
		Artifacts:  artifacts,
		Downloader: downloader,
		Logger:     logger,
		TargetDir:  targetDir,
	}
}
