package downloader

import (
	"fmt"
	"github.com/bitrise-io/go-utils/log"
	"io"
	"os"
	"strings"
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
	DownloadURLs []string
	Downloader   FileDownloader
	Logger       log.Logger
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
	downloadResultChan := make(chan ArtifactDownloadResult)
	defer close(downloadResultChan)

	semaphore := make(chan struct{}, maxConcurrentDownloadThreads)
	for _, url := range ad.DownloadURLs {
		semaphore <- struct{}{} // acquire
		go func(url string) {
			downloadResultChan <- ad.download(url, targetDir)
			<-semaphore // release
		}(url)
	}

	var downloadResults []ArtifactDownloadResult
	for { // block until results are filled up
		result, ok := <-downloadResultChan
		if !ok {
			break
		}

		downloadResults = append(downloadResults, result)
		if len(downloadResults) == len(ad.DownloadURLs) {
			break
		}
	}

	return downloadResults, nil
}

func (ad *ConcurrentArtifactDownloader) download(url, dir string) ArtifactDownloadResult {
	dataReader, err := ad.Downloader.DownloadFileFromURL(url)
	if err != nil {
		return ArtifactDownloadResult{DownloadError: err, DownloadURL: url}
	}
	defer func() {
		if err := dataReader.Close(); err != nil {
			ad.Logger.Errorf("failed to close data reader")
		}
	}()

	fileFullPath := fmt.Sprintf("%s/%s", dir, getFileNameFromURL(url))

	out, err := os.Create(fileFullPath)
	if err != nil {
		return ArtifactDownloadResult{DownloadError: err, DownloadURL: url}
	}
	defer func() {
		if err := out.Close(); err != nil {
			ad.Logger.Errorf("couldn't close file: %s", out.Name())
		}
	}()

	if _, err := io.Copy(out, dataReader); err != nil {
		return ArtifactDownloadResult{DownloadError: err, DownloadURL: url}
	}

	return ArtifactDownloadResult{DownloadPath: fileFullPath, DownloadURL: url}
}

func getFileNameFromURL(url string) string {
	parts := strings.Split(url, "/")

	return parts[len(parts)-1]
}

func getTargetDir(dirName string) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", pwd, dirName), nil
}

func NewConcurrentArtifactDownloader(downloadURLs []string, downloader FileDownloader, logger log.Logger) ArtifactDownloader {
	return &ConcurrentArtifactDownloader{
		DownloadURLs: downloadURLs,
		Downloader:   downloader,
		Logger:       logger,
	}
}
