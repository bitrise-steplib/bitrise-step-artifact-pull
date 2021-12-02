package downloader

import (
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
	return resp.Body, nil
}

func NewDefaultFileDownloader() FileDownloader {
	return DefaultFileDownloader{}
}
