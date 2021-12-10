package downloader

import (
	"fmt"
	"os"
	"sync"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/bitrise-step-artifact-pull/api"
)

const (
	filePermission               = 0o755
	maxConcurrentDownloadThreads = 10
	relativeDownloadPath         = "_tmp"
)

type ArtifactDownloader interface {
	DownloadAndSaveArtifacts() ([]ArtifactDownloadResult, error)
}

type ConcurrentArtifactDownloader struct {
	Artifacts  []api.ArtifactResponseItemModel
	Downloader FileDownloader
	Logger     log.Logger
}

type ArtifactDownloadResult struct {
	DownloadError error
	DownloadPath  string
	DownloadURL   string
}

func (ad *ConcurrentArtifactDownloader) DownloadAndSaveArtifacts() ([]ArtifactDownloadResult, error) {
	targetDir, err := getTargetDir(relativeDownloadPath)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := os.Mkdir(targetDir, filePermission); err != nil {
			return nil, err
		}
	}

	return ad.downloadParallel(targetDir)
}

func (ad *ConcurrentArtifactDownloader) downloadParallel(targetDir string) ([]ArtifactDownloadResult, error) {
	downloadResultChan := make(chan ArtifactDownloadResult, len(ad.Artifacts))
	defer close(downloadResultChan)
	var downloadResults []ArtifactDownloadResult

	var wg sync.WaitGroup
	wg.Add(len(ad.Artifacts))
	go func() {
		for { // block until results are filled up
			result := <-downloadResultChan

			downloadResults = append(downloadResults, result)
			wg.Done()
			if len(downloadResults) == len(ad.Artifacts) {
				break
			}
		}
	}()

	semaphore := make(chan struct{}, maxConcurrentDownloadThreads)
	for _, artifact := range ad.Artifacts {
		semaphore <- struct{}{} // acquire
		go func(artifactInfo api.ArtifactResponseItemModel) {
			downloadResultChan <- ad.download(artifactInfo, targetDir)

			<-semaphore // release
		}(artifact)
	}

	wg.Wait()

	return downloadResults, nil
}

func (ad *ConcurrentArtifactDownloader) download(artifactInfo api.ArtifactResponseItemModel, dir string) ArtifactDownloadResult {
	fileContent, err := ad.Downloader.DownloadFileFromURL(artifactInfo.DownloadPath)

	if err != nil {
		return ArtifactDownloadResult{DownloadError: err, DownloadURL: artifactInfo.DownloadPath}
	}

	fileFullPath := fmt.Sprintf("%s/%s", dir, artifactInfo.Title)

	out, err := os.Create(fileFullPath)
	if err != nil {
		return ArtifactDownloadResult{DownloadError: err, DownloadURL: artifactInfo.DownloadPath}
	}
	defer func() {
		if err := out.Close(); err != nil {
			ad.Logger.Errorf("couldn't close file: %s", out.Name())
		}
	}()

	if _, err := out.Write(fileContent); err != nil {
		return ArtifactDownloadResult{DownloadError: err, DownloadURL: artifactInfo.DownloadPath}
	}

	return ArtifactDownloadResult{DownloadPath: fileFullPath, DownloadURL: artifactInfo.DownloadPath}
}

func getTargetDir(dirName string) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", pwd, dirName), nil
}

func NewConcurrentArtifactDownloader(artifacts []api.ArtifactResponseItemModel, downloader FileDownloader, logger log.Logger) ArtifactDownloader {
	return &ConcurrentArtifactDownloader{
		Artifacts:  artifacts,
		Downloader: downloader,
		Logger:     logger,
	}
}
