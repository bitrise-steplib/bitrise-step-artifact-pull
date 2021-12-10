package downloader

import (
	"context"
	"fmt"
	"time"

	"github.com/bitrise-io/go-steputils/input"
	"github.com/bitrise-io/go-utils/filedownloader"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/retry"
)

type FileDownloader interface {
	DownloadFileFromURL(url string) ([]byte, error)
}

type DefaultFileDownloader struct {
	Logger  log.Logger
	Timeout time.Duration
}

func (fd *DefaultFileDownloader) DownloadFileFromURL(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fd.Timeout)
	defer cancel()

	downloader := filedownloader.NewWithContext(ctx, retry.NewHTTPClient().StandardClient())
	fileProvider := input.NewFileProvider(downloader)

	contents, err := fileProvider.Contents(url)

	if err != nil {
		return nil, err
	} else if contents == nil {
		return nil, fmt.Errorf("failed to download file from %s", url)
	}

	return contents, nil
}

func NewDefaultFileDownloader(logger log.Logger, timeout time.Duration) FileDownloader {
	return &DefaultFileDownloader{Logger: logger, Timeout: timeout}
}
