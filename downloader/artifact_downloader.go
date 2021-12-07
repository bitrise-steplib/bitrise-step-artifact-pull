package downloader

import (
	"fmt"
	"io"
	"os"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/bitrise-step-artifact-pull/api"
)

const (
	filePermission               = 0755
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

	semaphore := make(chan struct{}, maxConcurrentDownloadThreads)
	for _, artifact := range ad.Artifacts {
		semaphore <- struct{}{} // acquire
		go func(artifactInfo api.ArtifactResponseItemModel) {
			downloadResultChan <- ad.download(artifactInfo, targetDir)

			<-semaphore // release
		}(artifact)
	}

	var downloadResults []ArtifactDownloadResult
	for { // block until results are filled up
		result, ok := <-downloadResultChan
		if !ok {
			break
		}

		downloadResults = append(downloadResults, result)
		if len(downloadResults) == len(ad.Artifacts) {
			break
		}
	}

	return downloadResults, nil
}

func (ad *ConcurrentArtifactDownloader) download(artifactInfo api.ArtifactResponseItemModel, dir string) ArtifactDownloadResult {
	ad.Logger.Debugf("downloading %s into %s dir", artifactInfo.DownloadPath, dir)

	dataReader, err := ad.Downloader.DownloadFileFromURL(artifactInfo.DownloadPath)
	defer func() {
		if dataReader != nil {
			if err := dataReader.Close(); err != nil {
				ad.Logger.Errorf("failed to close data reader")
			}
		}
	}()

	if err != nil {
		return ArtifactDownloadResult{DownloadError: err, DownloadURL: artifactInfo.DownloadPath}
	}

	ad.Logger.Debugf("%s download completed", artifactInfo.Title)

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

	if _, err := io.Copy(out, dataReader); err != nil {
		return ArtifactDownloadResult{DownloadError: err, DownloadURL: artifactInfo.DownloadPath}
	}

	ad.Logger.Debugf("downloaded %s into %s dir", artifactInfo.Title, fileFullPath)

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
