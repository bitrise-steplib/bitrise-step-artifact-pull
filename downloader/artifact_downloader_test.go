package downloader

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-steplib/bitrise-step-artifact-pull/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const relativeDownloadPath = "_tmp"

type MockDownloader struct {
	mock.Mock
}

func (m *MockDownloader) DownloadFileFromURL(_ string) ([]byte, error) {
	args := m.Called()
	// as it is used concurrently we need to give back new return value every time
	return []byte("shiny!"), args.Error(1)
}

func getDownloadDir(dirName string) (string, error) {
	tempPath, err := pathutil.NormalizedOSTempDirPath(dirName)
	if err != nil {
		return "", err
	}

	return tempPath, nil
}

func Test_DownloadAndSaveArtifacts(t *testing.T) {
	mockDownloader := &MockDownloader{}
	mockDownloader.
		On("DownloadFileFromURL", mock.AnythingOfTypeArgument("string")).
		Return(ioutil.NopCloser(bytes.NewReader([]byte("asd"))), nil)

	targetDir, err := getDownloadDir(relativeDownloadPath)
	assert.NoError(t, err)

	var artifacts []api.ArtifactResponseItemModel
	var expectedDownloadResults []ArtifactDownloadResult
	for i := 1; i <= 11; i++ {
		downloadURL := fmt.Sprintf("https://nice-file.hu/%d.txt", i)
		artifacts = append(artifacts, api.ArtifactResponseItemModel{DownloadPath: downloadURL, Title: fmt.Sprintf("%d.txt", i)})
		expectedDownloadResults = append(expectedDownloadResults, ArtifactDownloadResult{
			DownloadPath: targetDir + fmt.Sprintf("/%d.txt", i),
			DownloadURL:  downloadURL,
		})
	}

	artifactDownloader := NewConcurrentArtifactDownloader(artifacts, mockDownloader, targetDir, log.NewLogger())

	downloadResults, err := artifactDownloader.DownloadAndSaveArtifacts()

	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedDownloadResults, downloadResults)

	_ = os.RemoveAll(targetDir)
}
