package downloader

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bitrise-io/go-utils/log"
	"github.com/stretchr/testify/assert"
)

func Test_DownloadFileFromURL(t *testing.T) {
	expected := "dummy data"
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expected)
	}))
	defer svr.Close()
	c := NewDefaultFileDownloader(log.NewLogger(), 1*time.Second)
	response, err := c.DownloadFileFromURL(svr.URL)
	assert.NoError(t, err)

	assert.Equal(t, expected, string(response))
}

func Test_DownloadFileFromURL_UnauthorizedError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()
	c := NewDefaultFileDownloader(log.NewLogger(), 1*time.Second)
	response, err := c.DownloadFileFromURL(svr.URL)

	assert.Equal(t, []uint8([]byte(nil)), response)
	assert.EqualError(t, err, fmt.Sprintf("unable to download file from: %s. Status code: 401", svr.URL))
}

func Test_DownloadFileFromURL_Timout(t *testing.T) {
	c := NewDefaultFileDownloader(log.NewLogger(), time.Second/2)
	_, err := c.DownloadFileFromURL("http://dummy-url")

	assert.EqualError(t, err, `Get "http://dummy-url": context deadline exceeded`)
}
