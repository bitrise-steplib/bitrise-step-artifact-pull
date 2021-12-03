package downloader

import (
	"io"
	"os"
	"strings"
	"sync"

	"github.com/bitrise-io/go-utils/log"
)

const DOWNLOAD_PATH = "_tmp" // under current work dir
const FILE_PERMISSION = 0755
const MAX_CONCURRENT_DOWNLOAD_THREAD = 10

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

	targetDir, err := getTargetDir(DOWNLOAD_PATH)
	if err != nil {
		return nil, nil, err
	}

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		err = os.Mkdir(targetDir, FILE_PERMISSION)
		if err != nil {
			return nil, nil, err
		}
	}

	semaphore := make(chan struct{}, MAX_CONCURRENT_DOWNLOAD_THREAD)
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
	defer ad.Downloader.CloseResponseWithErrorLogging(dataReader)

	<-semaphore // release

	fileFullPath := fmt.Sprintf("%s/%s", downloadDir, getFileNameFromURL(url))

	ad.Logger.Printf("Saving %d file from %s URL", index, fileFullPath)

	out, err := os.Create(fileFullPath)
	if err != nil {
		errors[index] = err
	}
	defer out.Close()

	_, err = io.Copy(out, dataReader)
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

	return fmt.Sprintf("%s/%s", pwd, dirName), nil
}

func NewConcurrentArtifactDownloader(downloadURLs []string, downloader FileDownloader, logger log.Logger) ArtifactDownloader {
	return &ConcurrentArtifactDownloader{
		DownloadURLs: downloadURLs,
		Downloader:   downloader,
		Logger:       logger}
}
