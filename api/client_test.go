package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ShowBuildArtifact_returnsArtifactModel(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{"data":{"title":"artifact-slug","expiring_download_url":"https://some.path.com/path","slug":"artifact1"}}`
		_, err := w.Write([]byte(response))
		assert.NoError(t, err)
	}))
	defer svr.Close()

	client, err := NewDefaultBitriseAPIClient(svr.URL, "token")
	assert.NoError(t, err)

	artifact, showErr := client.ShowBuildArtifact("app-slug", "build-slug", "artifact-slug")

	assert.NoError(t, showErr)
	expectedArtifact := ArtifactResponseItemModel{
		Title:       "artifact-slug",
		DownloadURL: "https://some.path.com/path",
		Slug:        "artifact1",
	}
	assert.Equal(t, expectedArtifact, artifact)
}

func Test_ShowBuildArtifact_unauthorizedError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client, err := NewDefaultBitriseAPIClient(svr.URL, "token")
	assert.NoError(t, err)

	artifact, showErr := client.ShowBuildArtifact("app-slug", "build-slug", "artifact-slug")

	assert.Equal(t, ArtifactResponseItemModel{}, artifact)
	assert.EqualError(t, showErr, fmt.Sprintf("request to %s/v0.2/apps/app-slug/builds/build-slug/artifacts/artifact-slug failed - status code should be 2XX (401)", svr.URL))
}

func Test_ListBuildArtifacts_no_paging_returnsListOfArtifacts(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{"data":[{"title":"artifact1","slug":"slug1"},{"title":"artifact2","slug":"slug2"},{"title":"artifact3","slug":"slug3"}],"paging":{"total_item_count":3,"page_item_limit":10}}`
		_, err := w.Write([]byte(response))
		assert.NoError(t, err)
	}))
	defer svr.Close()

	client, err := NewDefaultBitriseAPIClient(svr.URL, "token")
	assert.NoError(t, err)

	artifactList, showErr := client.ListBuildArtifacts("app-slug", "build-slug")

	assert.NoError(t, showErr)
	expectedArtifactList := []ArtifactListElementResponseModel{
		{"artifact1", "slug1"}, {"artifact2", "slug2"}, {"artifact3", "slug3"},
	}
	assert.Equal(t, expectedArtifactList, artifactList)
}

func Test_ListBuildArtifacts_paging_returnsListOfArtifacts(t *testing.T) {
	expectedArtifactList := []ArtifactListElementResponseModel{
		{"artifact1", "slug1"}, {"artifact2", "slug2"}, {"artifact3", "slug3"}, {"artifact4", "slug4"},
	}
	var nextPageIndex int64
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queryValues := r.URL.Query()
		nextQueryValue, ok := queryValues["next"]

		if nextPageIndex == 0 {
			assert.False(t, ok)
		} else {
			assert.Equal(t, fmt.Sprintf("%d", nextPageIndex), nextQueryValue[0])
		}

		nextPageIndex++

		var nextPage string
		if nextPageIndex < int64(len(expectedArtifactList)) {
			nextPage = fmt.Sprintf("%d", nextPageIndex)
		}

		response := ListBuildArtifactsResponse{
			Data:   []ArtifactListElementResponseModel{expectedArtifactList[nextPageIndex-1]},
			Paging: PagingModel{TotalItemCount: int64(len(expectedArtifactList)), PageItemLimit: 1, Next: nextPage},
		}
		result, err := json.Marshal(response)
		assert.NoError(t, err)
		_, err = w.Write(result)
		assert.NoError(t, err)
	}))
	defer svr.Close()

	client, err := NewDefaultBitriseAPIClient(svr.URL, "token")
	assert.NoError(t, err)

	artifactList, showErr := client.ListBuildArtifacts("app-slug", "build-slug")

	assert.NoError(t, showErr)
	assert.Equal(t, expectedArtifactList, artifactList)
}

func Test_ListBuildArtifacts_unauthorizedError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client, err := NewDefaultBitriseAPIClient(svr.URL, "token")
	assert.NoError(t, err)

	artifactList, showErr := client.ListBuildArtifacts("app-slug", "build-slug")

	assert.Nil(t, artifactList)
	assert.EqualError(t, showErr, fmt.Sprintf("request to %s/v0.2/apps/app-slug/builds/build-slug/artifacts failed - status code should be 2XX (401)", svr.URL))
}
