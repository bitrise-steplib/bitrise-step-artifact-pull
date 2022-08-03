package api

import (
	"errors"
	"testing"

	"github.com/bitrise-io/go-utils/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockBitriseAPIClient struct {
	mock.Mock
}

func (m *mockBitriseAPIClient) ListBuildArtifacts(appSlug, buildSlug string) ([]ArtifactListElementResponseModel, error) {
	args := m.Called(appSlug, buildSlug)

	r0, ok := args.Get(0).([]ArtifactListElementResponseModel)
	if !ok {
		panic("Type mismatch")
	}
	r1 := args.Error(1)

	return r0, r1
}

func (m *mockBitriseAPIClient) ShowBuildArtifact(appSlug, buildSlug, artifactSlug string) (ArtifactResponseItemModel, error) {
	args := m.Called(appSlug, buildSlug, artifactSlug)

	r0, ok := args.Get(0).(ArtifactResponseItemModel)
	if !ok {
		panic("Type mismatch")
	}
	r1 := args.Error(1)

	return r0, r1
}

type listConcurrencyTestCase struct {
	maxConcurrentListCalls, maxConcurrentShowCalls int
}

func Test_ListBuildArtifactDetails_concurrent_returnsArtifactListForMultipleBuilds(t *testing.T) {
	mockArtifactList := []ArtifactListElementResponseModel{
		{Slug: "artifact1"},
		{Slug: "artifact2"},
		{Slug: "artifact3"},
		{Slug: "artifact4"},
		{Slug: "artifact5"},
		{Slug: "artifact6"},
		{Slug: "artifact7"},
	}

	mockBuildSlugs := []string{"build-slug", "build-slug", "build-slug", "build-slug", "build-slug", "build-slug", "build-slug"}

	mockClient := &mockBitriseAPIClient{}
	mockClient.
		On("ListBuildArtifacts", mock.AnythingOfTypeArgument("string"), mock.AnythingOfTypeArgument("string")).
		Return(mockArtifactList, nil)

	mockClient.On("ShowBuildArtifact", mock.AnythingOfTypeArgument("string"), mock.AnythingOfTypeArgument("string"), mock.AnythingOfTypeArgument("string")).Return(ArtifactResponseItemModel{IntermediateFileInfo: IntermediateFileInfo{EnvKey: "EnvKey"}}, nil)

	testCases := []listConcurrencyTestCase{
		{1, 1}, {3, 3}, {10, 1}, {1, 10}, {10, 10},
	}
	for _, testCase := range testCases {
		lister := newArtifactLister(mockClient, log.NewLogger())
		lister.maxConcurrentListArtifactAPICalls = testCase.maxConcurrentListCalls
		lister.maxConcurrentShowArtifactAPICalls = testCase.maxConcurrentShowCalls

		artifacts, err := lister.ListIntermediateFileDetails("app-slug", mockBuildSlugs)

		assert.NoError(t, err)
		assert.Equal(t, len(mockBuildSlugs)*len(mockArtifactList), len(artifacts))
	}
}

func Test_ListBuildArtifactDetails_returnsErrorWhenApiCallFails(t *testing.T) {
	mockBuildSlugs := []string{"build-slug", "build-slug", "build-slug"}

	mockClient := &mockBitriseAPIClient{}
	mockClient.
		On("ListBuildArtifacts", mock.AnythingOfTypeArgument("string"), mock.AnythingOfTypeArgument("string")).
		Return([]ArtifactListElementResponseModel{}, errors.New("API error"))

	lister := newArtifactLister(mockClient, log.NewLogger())
	_, err := lister.ListIntermediateFileDetails("app-slug", mockBuildSlugs)

	assert.EqualError(t, err, "failed to get artifact download links for build(s): build-slug, build-slug, build-slug")
}
