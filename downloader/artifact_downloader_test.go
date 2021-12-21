package downloader

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-steplib/bitrise-step-artifact-pull/api"
	"github.com/stretchr/testify/assert"
)

const relativeDownloadPath = "_tmp"

func getDownloadDir(dirName string) (string, error) {
	tempPath, err := pathutil.NormalizedOSTempDirPath(dirName)
	if err != nil {
		return "", err
	}

	return tempPath, nil
}

func Test_DownloadAndSaveArtifacts(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "dummy data")
	}))
	defer svr.Close()

	targetDir, err := getDownloadDir(relativeDownloadPath)
	assert.NoError(t, err)

	var artifacts []api.ArtifactResponseItemModel
	var expectedDownloadResults []ArtifactDownloadResult
	for i := 1; i <= 11; i++ {
		downloadURL := fmt.Sprintf(svr.URL+"/%d.txt", i)
		artifacts = append(artifacts, api.ArtifactResponseItemModel{DownloadURL: downloadURL, Title: fmt.Sprintf("%d.txt", i)})
		expectedDownloadResults = append(expectedDownloadResults, ArtifactDownloadResult{
			DownloadPath: targetDir + fmt.Sprintf("/%d.txt", i),
			DownloadURL:  downloadURL,
		})
	}

	artifactDownloader := NewConcurrentArtifactDownloader(artifacts, 5*time.Minute, targetDir, log.NewLogger())

	downloadResults, err := artifactDownloader.DownloadAndSaveArtifacts()

	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedDownloadResults, downloadResults)

	_ = os.RemoveAll(targetDir)
}

func Test_DownloadAndSaveArtifacts_DownloadFails(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	targetDir, err := getDownloadDir(relativeDownloadPath)
	assert.NoError(t, err)

	var artifacts []api.ArtifactResponseItemModel
	artifacts = append(artifacts,
		api.ArtifactResponseItemModel{DownloadURL: svr.URL + "/1.txt", Title: "1.txt"})

	artifactDownloader := NewConcurrentArtifactDownloader(artifacts, 5*time.Minute, targetDir, log.NewLogger())

	result, err := artifactDownloader.DownloadAndSaveArtifacts()

	assert.EqualError(t, result[0].DownloadError, fmt.Sprintf("unable to download file from: %s/1.txt. Status code: 401", svr.URL))
	assert.NoError(t, err)

	_ = os.RemoveAll(targetDir)
}
