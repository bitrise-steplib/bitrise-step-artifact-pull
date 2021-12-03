package downloader

import (
	"io"
	"os"
	"strings"
	"sync"

	"github.com/bitrise-io/go-utils/log"
)

const (
	realtiveDownloadPath         = "_tmp"
	filePermission               = 0755
	maxConcurrentDownloadThreads = 10
)

type ArtifactDownloader interface {
	DownloadAndSaveArtifacts() ([]string, []error, error)
}

type ConcurrentArtifactDownloader struct {
	DownloadURLs []string
	Downloader   FileDownloader
	Logger       log.Logger
	valueMutex   sync.Mutex
}

func (ad *ConcurrentArtifactDownloader) DownloadAndSaveArtifacts() ([]string, []error, error) {
	paths := make([]string, len(ad.DownloadURLs))
	errors := make([]error, len(ad.DownloadURLs))

	targetDir, err := getTargetDir(realtiveDownloadPath)
	if err != nil {
		return nil, nil, err
	}

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := os.Mkdir(targetDir, filePermission); err != nil {
			return nil, nil, err
		}
	}

	semaphore := make(chan struct{}, maxConcurrentDownloadThreads)
	wg := &sync.WaitGroup{}
	for i, url := range ad.DownloadURLs {
		wg.Add(1)
		go ad.download(i, url, targetDir, paths, errors, wg, semaphore)
	}
	wg.Wait()

	return paths, errors, nil
}

func (ad *ConcurrentArtifactDownloader) download(
	index int, url string, downloadDir string, paths []string, errors []error, wg *sync.WaitGroup, semaphore chan struct{},
) {
	defer wg.Done()
	semaphore <- struct{}{} // acquire

	ad.Logger.Printf("Downloading %d file from %s URL", index, url)

	ad.valueMutex.Lock()
	defer ad.valueMutex.Unlock()

	dataReader, err := ad.Downloader.DownloadFileFromURL(url)
	if err != nil {
		errors[index] = err
	}

	<-semaphore // release

	fileFullPath := downloadDir + "/" + getFileNameFromURL(url)

	ad.Logger.Printf("Saving %d file from %s URL", index, fileFullPath)

	out, err := os.Create(fileFullPath)
	if err != nil {
		errors[index] = err
	}

	if _, err := io.Copy(out, dataReader); err != nil {
		errors[index] = err
	}

	err = out.Close()
	if err != nil {
		errors[index] = err
	}

	err = dataReader.Close()
	if err != nil {
		errors[index] = err
	}

	paths[index] = fileFullPath
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

	return pwd + "/" + dirName, nil
}

func NewConcurrentArtifactDownloader(downloadURLs []string, downloader FileDownloader, logger log.Logger) ArtifactDownloader {
	return &ConcurrentArtifactDownloader{
		DownloadURLs: downloadURLs,
		Downloader:   downloader,
		Logger:       logger}
}
