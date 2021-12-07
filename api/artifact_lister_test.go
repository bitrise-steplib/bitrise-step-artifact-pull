package api

import (
	"errors"
	"sync"
	"testing"

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
	mockArtifact := ArtifactResponseItemModel{Title: "artifact", DownloadPath: mockDownloadPath}
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
		{Slug: "artifact1"},
		{Slug: "artifact2"},
		{Slug: "artifact3"},
	}

	expectedArtifactList := []ArtifactResponseItemModel{
		{Slug: "artifact1"},
		{Slug: "artifact2"},
		{Slug: "artifact3"},
	}

	mockClient := &MockBitriseAPIClient{}
	mockClient.
		On("ListBuildArtifacts", mock.AnythingOfTypeArgument("string"), mock.AnythingOfTypeArgument("string")).
		Return(mockArtifactList, nil)
	for _, a := range expectedArtifactList {
		mockClient.On("ShowBuildArtifact", mock.AnythingOfTypeArgument("string"), mock.AnythingOfTypeArgument("string"), mock.AnythingOfTypeArgument("string")).
			Return(a, nil).Once()
	}

	lister := NewDefaultArtifactLister(mockClient)

	resultsChan := make(chan listArtifactsOfBuildResult, 1)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	lister.listArtifactsOfBuild("app-slug", "build-slug", resultsChan, wg)
	wg.Wait()

	res := <-resultsChan

	assert.NoError(t, res.err)
	assert.Equal(t, expectedArtifactList, res.artifacts)
}

type listConcurrencyTestCase struct {
	maxConcurrentListCalls, maxConcurrentShowCalls int
}

func Test_ListBuildArtifacts_concurrent_returnsArtifactListForMultipleBuilds(t *testing.T) {
	mockArtifactList := []ArtifactListElementResponseModel{
		{Slug: "artifact1"},
		{Slug: "artifact2"},
		{Slug: "artifact3"},
		{Slug: "artifact4"},
		{Slug: "artifact5"},
	}

	mockBuildSlugs := []string{"build-slug", "build-slug", "build-slug", "build-slug", "build-slug"}

	mockClient := &MockBitriseAPIClient{}
	mockClient.
		On("ListBuildArtifacts", mock.AnythingOfTypeArgument("string"), mock.AnythingOfTypeArgument("string")).
		Return(mockArtifactList, nil)

	mockClient.On("ShowBuildArtifact", mock.AnythingOfTypeArgument("string"), mock.AnythingOfTypeArgument("string"), mock.AnythingOfTypeArgument("string")).Return(ArtifactResponseItemModel{}, nil)

	testCases := []listConcurrencyTestCase{
		{1, 1}, {3, 3}, {10, 1}, {1, 10}, {10, 10},
	}
	for _, testCase := range testCases {
		lister := NewDefaultArtifactLister(mockClient)
		lister.maxConcurrentListArtifactAPICalls = testCase.maxConcurrentListCalls
		lister.maxConcurrentShowArtifactAPICalls = testCase.maxConcurrentShowCalls

		artifacts, err := lister.ListBuildArtifacts("app-slug", mockBuildSlugs)

		assert.NoError(t, err)
		assert.Equal(t, len(mockBuildSlugs)*len(mockArtifactList), len(artifacts))
	}
}

func Test_ListBuildArtifacts_returnsErrorWhenApiCallFails(t *testing.T) {
	mockBuildSlugs := []string{"build-slug", "build-slug", "build-slug"}

	mockClient := &MockBitriseAPIClient{}
	mockClient.
		On("ListBuildArtifacts", mock.AnythingOfTypeArgument("string"), mock.AnythingOfTypeArgument("string")).
		Return([]ArtifactListElementResponseModel{}, errors.New("API error"))

	lister := NewDefaultArtifactLister(mockClient)
	_, err := lister.ListBuildArtifacts("app-slug", mockBuildSlugs)

	assert.EqualError(t, err, "failed to get artifact download links for build(s): build-slug, build-slug, build-slug")
}
