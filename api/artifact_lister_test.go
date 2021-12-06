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

func Test_showArtifact(t *testing.T) {
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
