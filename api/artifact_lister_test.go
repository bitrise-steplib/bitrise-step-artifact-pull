package api

import (
	"sync"
	"testing"

	"github.com/gobuffalo/nulls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBitriseAPIClient struct {
	mock.Mock
}

func (m *MockBitriseAPIClient) ListBuildArtifacts(appSlug, buildSlug string) ([]ArtifactListElementResponseModel, error) {
	args := m.Called(appSlug, buildSlug)

	r0, ok := args.Get(0).([]ArtifactListElementResponseModel)
	if !ok {
		panic("Type mismatch")
	}
	r1 := args.Error(1)

	return r0, r1
}

func (m *MockBitriseAPIClient) ShowBuildArtifact(appSlug, buildSlug, artifactSlug string) (ArtifactResponseItemModel, error) {
	args := m.Called(appSlug, buildSlug, artifactSlug)

	r0, ok := args.Get(0).(ArtifactResponseItemModel)
	if !ok {
		panic("Type mismatch")
	}
	r1 := args.Error(1)

	return r0, r1
}

func Test_showArtifact_returnsArtifact(t *testing.T) {
	mockDownloadPath := "http://download.com"
	mockArtifact := ArtifactResponseItemModel{Title: nulls.NewString("artifact"), DownloadPath: &mockDownloadPath}
	mockClient := &MockBitriseAPIClient{}
	mockClient.
		On("ShowBuildArtifact", mock.AnythingOfTypeArgument("string"), mock.AnythingOfTypeArgument("string"), mock.AnythingOfTypeArgument("string")).
		Return(mockArtifact, nil)

	lister := NewDefaultArtifactLister(mockClient)

	showResultChan := make(chan showArtifactResult, 1)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	lister.showArtifact("app-slug", "build-slug", "artifact-slug", showResultChan, wg)
	wg.Wait()

	res := <-showResultChan

	assert.NoError(t, res.err)
	assert.Equal(t, mockArtifact, res.artifact)
}

func Test_listArtifactsOfBuild_returnsArtifactList(t *testing.T) {
	mockArtifactList := []ArtifactListElementResponseModel{
		ArtifactListElementResponseModel{Slug: "artifact1"},
		ArtifactListElementResponseModel{Slug: "artifact2"},
		ArtifactListElementResponseModel{Slug: "artifact3"},
	}
	mockDownloadPath := "http://download.com"
	mockArtifact := ArtifactResponseItemModel{Title: nulls.NewString("artifact"), DownloadPath: &mockDownloadPath}
	mockClient := &MockBitriseAPIClient{}
	mockClient.
		On("ListBuildArtifacts", mock.AnythingOfTypeArgument("string"), mock.AnythingOfTypeArgument("string")).
		Return(mockArtifactList, nil)
	mockClient.
		On("ShowBuildArtifact", mock.AnythingOfTypeArgument("string"), mock.AnythingOfTypeArgument("string"), mock.AnythingOfTypeArgument("string")).
		Return(mockArtifact, nil)

	lister := NewDefaultArtifactLister(mockClient)

	resultsChan := make(chan listArtifactsOfBuildResult, 1)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	lister.listArtifactsOfBuild("app-slug", "build-slug", resultsChan, wg)
	wg.Wait()

	res := <-resultsChan

	assert.NoError(t, res.err)
	assert.Equal(t, mockArtifactList, res.artifacts[0])
}
