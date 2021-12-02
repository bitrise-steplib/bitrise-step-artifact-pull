package downloader

import (
	"fmt"
	"io"
	"net/http"
)

type FileDownloader interface {
	DownloadFileFromURL(url string) (io.ReadCloser, error)
}

type DefaultFileDownloader struct {
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

func NewDefaultFileDownloader() FileDownloader {
	return DefaultFileDownloader{}
}
