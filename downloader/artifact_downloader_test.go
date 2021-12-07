package downloader

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/bitrise-io/go-utils/log"
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

	downloadURLs := []string{
		"https://nice-file.hu/1.txt",
		"https://nice-file.hu/2.txt",
	}
	expectedDownloadResults := []ArtifactDownloadResult{
		{
			DownloadPath: getDownloadDir(relativeDownloadPath) + "/1.txt",
			DownloadURL:  downloadURLs[0],
		},
		{
			DownloadPath: getDownloadDir(relativeDownloadPath) + "/2.txt",
			DownloadURL:  downloadURLs[1],
		},
	}

	artifactDownloader := NewConcurrentArtifactDownloader(downloadURLs, mockDownloader, log.NewLogger())

	downloadResults, err := artifactDownloader.DownloadAndSaveArtifacts()

	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedDownloadResults, downloadResults)

	_ = os.RemoveAll(getDownloadDir(relativeDownloadPath))
}
