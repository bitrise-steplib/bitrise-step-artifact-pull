package downloader

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitrise-io/go-utils/log"
	"github.com/stretchr/testify/assert"
)

func Test_DownloadFileFromURL(t *testing.T) {
	expected := "dummy data"
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expected)
	}))
	defer svr.Close()
	c := NewDefaultFileDownloader(log.NewLogger())
	res, err := c.DownloadFileFromURL(svr.URL)
	assert.NoError(t, err)

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(res)
	assert.NoError(t, err)

	data := buf.String()
	err = res.Close()

	assert.NoError(t, err)
	assert.Equal(t, expected, string(data))
}

func Test_DownloadFileFromURL_UnauthorizedError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()
	c := NewDefaultFileDownloader(log.NewLogger())
	res, err := c.DownloadFileFromURL(svr.URL)

	assert.Equal(t, http.NoBody, res)
	assert.EqualError(t, err, fmt.Sprintf("failed to download file from %s, status code: 401", svr.URL))
}

func Test_DownloadFileFromURL_ServerNotFound(t *testing.T) {
	c := NewDefaultFileDownloader(log.NewLogger())
	_, err := c.DownloadFileFromURL("http://dummy-url")

	assert.EqualError(t, err, `Get "http://dummy-url": dial tcp: lookup dummy-url: no such host`)
}
