package api

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/go-utils/log"
)

const (
	maxConcurrentListArtifactAPICalls = 3
	maxConcurrentShowArtifactAPICalls = 3
)

type bitriseAPIClient interface {
	// ListBuildArtifacts lists all build artifacts that have been generated for an app’s build - https://api-docs.bitrise.io/#/build-artifact/artifact-list
	ListBuildArtifacts(appSlug, buildSlug string) ([]ArtifactListElementResponseModel, error)
	// ShowBuildArtifact retrieves data of a specific build artifact - https://api-docs.bitrise.io/#/build-artifact/artifact-show
	ShowBuildArtifact(appSlug, buildSlug, artifactSlug string) (ArtifactResponseItemModel, error)
}

type listArtifactsResult struct {
	buildSlug string
	artifacts []ArtifactResponseItemModel
	err       error
}

type showArtifactResult struct {
	buildSlug string
	artifact  ArtifactResponseItemModel
	err       error
}

type ArtifactLister struct {
	apiClient                         bitriseAPIClient
	logger                            log.Logger
	maxConcurrentListArtifactAPICalls int
	maxConcurrentShowArtifactAPICalls int
}

func NewArtifactLister(apiBaseURL, authToken string, logger log.Logger) (ArtifactLister, error) {
	client, err := NewDefaultBitriseAPIClient(apiBaseURL, authToken)
	if err != nil {
		return ArtifactLister{}, err
	}

	return newArtifactLister(&client, logger), nil
}

func newArtifactLister(client bitriseAPIClient, logger log.Logger) ArtifactLister {
	return ArtifactLister{
		apiClient:                         client,
		logger:                            logger,
		maxConcurrentListArtifactAPICalls: maxConcurrentListArtifactAPICalls,
		maxConcurrentShowArtifactAPICalls: maxConcurrentShowArtifactAPICalls, // per list artifact call (so the total number of max API calls is maxConcurrentListArtifactAPICalls * maxConcurrentShowArtifactAPICalls)
	}
}

func (lister ArtifactLister) ListIntermediateFileDetails(appSlug string, buildSlugs []string) ([]ArtifactResponseItemModel, error) {
	listJobs := make(chan string, len(buildSlugs))
	listResults := make(chan listArtifactsResult, len(buildSlugs))

	for w := 1; w <= lister.maxConcurrentListArtifactAPICalls; w++ {
		go lister.listArtifactsWorker(appSlug, listJobs, listResults)
	}

	for _, buildSlug := range buildSlugs {
		listJobs <- buildSlug
	}
	close(listJobs)

	// process results
	var (
		failedBuildSlugs []string
		artifacts        []ArtifactResponseItemModel
	)
	for i := 1; i <= len(buildSlugs); i++ {
		res := <-listResults
		if res.err != nil {
			failedBuildSlugs = append(failedBuildSlugs, res.buildSlug)
		} else {
			artifacts = append(artifacts, res.artifacts...)
		}
	}

	if failedBuildSlugs != nil {
		return nil, fmt.Errorf("failed to get artifact download links for build(s): %s", strings.Join(failedBuildSlugs, ", "))
	}

	return artifacts, nil
}

// listArtifactsWorker gets details of all artifacts of a particular build using the Bitrise API
func (lister ArtifactLister) listArtifactsWorker(appSlug string, buildSlugs chan string, results chan listArtifactsResult) {
	for buildSlug := range buildSlugs {
		lister.logger.Debugf("Listing artifacts for build: https://app.bitrise.io/build/%v", buildSlug)
		artifactListItems, err := lister.apiClient.ListBuildArtifacts(appSlug, buildSlug)
		if err != nil {
			results <- listArtifactsResult{buildSlug: buildSlug, err: err}
		} else if len(artifactListItems) == 0 {
			results <- listArtifactsResult{buildSlug: buildSlug}
		} else {
			showJobs := make(chan string, len(artifactListItems))
			showResults := make(chan showArtifactResult, len(artifactListItems))

			for w := 1; w <= lister.maxConcurrentShowArtifactAPICalls; w++ {
				go lister.showArtifactWorker(appSlug, buildSlug, showJobs, showResults)
			}

			for _, artifactListItem := range artifactListItems {
				showJobs <- artifactListItem.Slug
			}
			close(showJobs)

			var artifacts []ArtifactResponseItemModel
			for i := 0; i < len(artifactListItems); i++ {
				res := <-showResults
				if res.err != nil {
					results <- listArtifactsResult{buildSlug: buildSlug, err: err}
					return
				}

				// Keep Build Artifacts which are Intermediate Files
				if res.artifact.IntermediateFileInfo.EnvKey != "" {
					artifacts = append(artifacts, res.artifact)
				}
			}

			results <- listArtifactsResult{buildSlug: buildSlug, artifacts: artifacts}
		}
	}
}

func (lister ArtifactLister) showArtifactWorker(appSlug, buildSlug string, artifactSlugs chan string, results chan showArtifactResult) {
	for artifactSlug := range artifactSlugs {
		lister.logger.Debugf("Getting artifact details for artifact %v", artifactSlug)

		artifact, err := lister.apiClient.ShowBuildArtifact(appSlug, buildSlug, artifactSlug)
		if err != nil {
			results <- showArtifactResult{buildSlug: buildSlug, err: err}
		} else {
			results <- showArtifactResult{buildSlug: buildSlug, artifact: artifact}
		}
	}
}
