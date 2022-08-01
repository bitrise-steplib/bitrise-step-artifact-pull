package downloader

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/filedownloader"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-steplib/bitrise-step-artifact-pull/api"
)

const (
	filePermission               = 0o655
	maxConcurrentDownloadThreads = 10
)

type ArtifactDownloadResult struct {
	DownloadError error
	DownloadPath  string
	DownloadURL   string
	EnvKey        string
}

type downloadJob struct {
	ResponseModel api.ArtifactResponseItemModel
	TargetDir     string
}

type ConcurrentArtifactDownloader struct {
	Timeout        time.Duration
	Logger         log.Logger
	CommandFactory command.Factory
}

func NewConcurrentArtifactDownloader(timeout time.Duration, logger log.Logger, commandFactory command.Factory) *ConcurrentArtifactDownloader {
	return &ConcurrentArtifactDownloader{
		Timeout:        timeout,
		Logger:         logger,
		CommandFactory: commandFactory,
	}
}

func (ad *ConcurrentArtifactDownloader) DownloadAndSaveArtifacts(artifacts []api.ArtifactResponseItemModel, targetDir string) ([]ArtifactDownloadResult, error) {
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := os.Mkdir(targetDir, filePermission); err != nil {
			return nil, err
		}
	}

	return ad.downloadParallel(artifacts, targetDir)
}

func (ad *ConcurrentArtifactDownloader) downloadParallel(artifacts []api.ArtifactResponseItemModel, targetDir string) ([]ArtifactDownloadResult, error) {
	var downloadResults []ArtifactDownloadResult

	jobs := make(chan downloadJob, len(artifacts))
	results := make(chan ArtifactDownloadResult, len(artifacts))

	for i := 0; i < maxConcurrentDownloadThreads; i++ {
		go ad.download(jobs, results)
	}

	for _, artifact := range artifacts {
		jobs <- downloadJob{
			ResponseModel: artifact,
			TargetDir:     targetDir,
		}
	}
	close(jobs)

	for i := 0; i < len(artifacts); i++ {
		res := <-results
		downloadResults = append(downloadResults, res)
	}

	return downloadResults, nil
}

func (ad *ConcurrentArtifactDownloader) downloadFile(targetDir, fileName, downloadURL string) (string, error) {
	fileFullPath := filepath.Join(targetDir, fileName)

	ctx, cancel := context.WithTimeout(context.Background(), ad.Timeout)

	downloader := filedownloader.NewWithContext(ctx, retry.NewHTTPClient().StandardClient())
	err := downloader.Get(fileFullPath, downloadURL)
	cancel()
	if err != nil {
		return "", err
	}
	return fileFullPath, nil
}

func (ad *ConcurrentArtifactDownloader) downloadDirectory(targetDir, fileName, downloadURL string) (string, error) {
	client := retry.NewHTTPClient()

	resp, err := client.Get(downloadURL)
	if err != nil {
		return "", err
	}

	fmt.Println("targetDir: ", targetDir)

	dirName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	dirPath := filepath.Join(targetDir, dirName)

	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return "", err
	}

	if err := ad.extractArchive(resp.Body, dirPath); err != nil {
		return "", err
	}

	if err := resp.Body.Close(); err != nil {
		log.Warnf("Failed to close response body: %s", err)
	}

	return dirPath, nil
}

func (ad *ConcurrentArtifactDownloader) download(jobs <-chan downloadJob, results chan<- ArtifactDownloadResult) {
	for j := range jobs {
		var fileFullPath string
		var err error

		if j.ResponseModel.IntermediateFileInfo.IsDir {
			fileFullPath, err = ad.downloadDirectory(j.TargetDir, j.ResponseModel.Title, j.ResponseModel.DownloadURL)
		} else {
			fileFullPath, err = ad.downloadFile(j.TargetDir, j.ResponseModel.Title, j.ResponseModel.DownloadURL)
		}

		if err != nil {
			results <- ArtifactDownloadResult{DownloadError: err, DownloadURL: j.ResponseModel.DownloadURL}
			return
		}

		results <- ArtifactDownloadResult{DownloadPath: fileFullPath, DownloadURL: j.ResponseModel.DownloadURL, EnvKey: j.ResponseModel.IntermediateFileInfo.EnvKey}
	}
}

// extractArchive extracts an archive using the tar CLI tool by piping the archive to the command's input.
func (ad *ConcurrentArtifactDownloader) extractArchive(r io.Reader, targetDir string) error {
	tarArgs := []string{
		"-x",      // -x: extract files from an archive: https://www.gnu.org/software/tar/manual/html_node/extract.html#SEC25
		"-f", "-", // -f "-": reads the archive from standard input: https://www.gnu.org/software/tar/manual/html_node/Device.html#SEC155
	}
	cmd := ad.CommandFactory.Create("tar", tarArgs, &command.Opts{
		Stdin: r,
		Dir:   targetDir,
	})

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %s", cmd.PrintableCommandArgs(), err)
	}

	return nil
}
