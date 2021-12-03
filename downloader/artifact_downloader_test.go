package downloader

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/bitrise-io/go-utils/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDownloader struct {
	mock.Mock
}

func (m *MockDownloader) DownloadFileFromURL(url string) (io.ReadCloser, error) {
	args := m.Called()
	return args.Get(0).(io.ReadCloser), args.Error(1)
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
		"http://nice-file.hu/1.txt",
		"http://nice-file.hu/2.txt",
	}
	expectedErrors := []error{nil, nil}
	expectedDownloadPaths := []string{
		getDownloadDir(realtiveDownloadPath) + "/1.txt",
		getDownloadDir(realtiveDownloadPath) + "/2.txt",
	}

	artifactDownloader := NewConcurrentArtifactDownloader(downloadURLs, mockDownloader, log.NewLogger())

	artifactPaths, errors, err := artifactDownloader.DownloadAndSaveArtifacts()

	assert.NoError(t, err)
	assert.Equal(t, expectedErrors, errors)
	assert.Equal(t, expectedDownloadPaths, artifactPaths)

	os.RemoveAll(getDownloadDir(realtiveDownloadPath))
}
