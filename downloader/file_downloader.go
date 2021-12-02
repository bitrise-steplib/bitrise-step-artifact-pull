package downloader

import (
	"fmt"
	"io"
	"net/http"

	"github.com/bitrise-io/go-utils/log"
)

type FileDownloader interface {
	DownloadFileFromURL(url string) (io.ReadCloser, error)
	CloseResponseWithErrorLogging(resp io.ReadCloser)
}

type DefaultFileDownloader struct {
	Logger log.Logger
}

func (fd DefaultFileDownloader) DownloadFileFromURL(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file from %s, status code: %d", url, resp.StatusCode)
	}

	return resp.Body, nil
}

func (fd DefaultFileDownloader) CloseResponseWithErrorLogging(resp io.ReadCloser) {
	err := resp.Close()

	if err != nil {
		fd.Logger.Errorf("failed to close response reader")
	}
}

func NewDefaultFileDownloader(logger log.Logger) FileDownloader {
	return DefaultFileDownloader{Logger: logger}
}
