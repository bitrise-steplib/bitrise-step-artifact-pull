package downloader

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/bitrise-step-artifact-pull/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDownloader struct {
	mock.Mock
}

func (m *MockDownloader) DownloadFileFromURL(_ string) (io.ReadCloser, error) {
	args := m.Called()
	// as it is used concurrently we need to give back new return value every time
	return io.NopCloser(strings.NewReader("shiny!")), args.Error(1)
}

func getDownloadDir(dirName string) string {
	pwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	return pwd + "/" + dirName
}

func Test_DownloadAndSaveArtifacts(t *testing.T) {
	mockDownloader := &MockDownloader{}
	mockDownloader.
		On("DownloadFileFromURL", mock.AnythingOfTypeArgument("string")).
		Return(ioutil.NopCloser(bytes.NewReader([]byte("asd"))), nil)
  
	var downloadURLs []string
	var expectedDownloadResults []ArtifactDownloadResult
	for i := 1; i <= 11; i++ {
		downloadURL := fmt.Sprintf("https://nice-file.hu/%d.txt", i)
		downloadURLs = append(downloadURLs, downloadURL)
		expectedDownloadResults = append(expectedDownloadResults, ArtifactDownloadResult{
			DownloadPath: getDownloadDir(relativeDownloadPath) + fmt.Sprintf("/%d.txt", i),
			DownloadURL:  downloadURL,
		})
	}

	artifactDownloader := NewConcurrentArtifactDownloader(downloadURLs, mockDownloader, log.NewLogger())

	downloadResults, err := artifactDownloader.DownloadAndSaveArtifacts()

	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedDownloadResults, downloadResults)

	_ = os.RemoveAll(getDownloadDir(relativeDownloadPath))
}
