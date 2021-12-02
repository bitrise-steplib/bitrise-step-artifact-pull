package downloader

import "io"

type FileDownloader interface {
	DownloadFileFromURL(url string) (io.ReadCloser, error)
}

type DefaultFileDownloader struct {
}

func (fd DefaultFileDownloader) DownloadFileFromURL(url string) (io.ReadCloser, error) {
	return nil, nil
}

func NewDefaultFileDownloader() FileDownloader {
	return DefaultFileDownloader{}
}
