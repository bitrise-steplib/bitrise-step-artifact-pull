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
	expectedArtifact := ArtifactResponseItemModel{
		Title:        "artifact-slug",
		DownloadPath: "https://some.path.com/path",
		Slug:         "artifact1",
	}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ShowBuildArtifactResponse{
			Data: expectedArtifact,
		}
		result, err := json.Marshal(response)
		assert.NoError(t, err)
		w.Write(result)
	}))
	defer svr.Close()

	client, err := NewDefaultBitriseAPIClient(svr.URL, "token")
	assert.NoError(t, err)

	artifact, showErr := client.ShowBuildArtifact("app-slug", "build-slug", "artifact-slug")

	assert.NoError(t, showErr)
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
	expectedArtifactList := []ArtifactListElementResponseModel{
		{"artifact1", "slug1"}, {"artifact2", "slug2"}, {"artifact3", "slug3"},
	}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ListBuildArtifactsResponse{
			Data:   expectedArtifactList,
			Paging: PagingModel{TotalItemCount: 3, PageItemLimit: 10},
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

func Test_ListBuildArtifacts_paging_returnsListOfArtifacts(t *testing.T) {
	expectedArtifactList := []ArtifactListElementResponseModel{
		{"artifact1", "slug1"}, {"artifact2", "slug2"}, {"artifact3", "slug3"}, {"artifact4", "slug4"},
	}
	var nextPageIndex int64
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextPageIndex++

		nextPage := "some page index"
		if nextPageIndex == int64(len(expectedArtifactList)) {
			nextPage = ""
		}

		response := ListBuildArtifactsResponse{
			Data:   []ArtifactListElementResponseModel{expectedArtifactList[nextPageIndex-1]},
			Paging: PagingModel{TotalItemCount: int64(len(expectedArtifactList)), PageItemLimit: 1, Next: nextPage},
		}
		result, err := json.Marshal(response)
		assert.NoError(t, err)
		w.Write(result)
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
