package api

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/go-utils/log"
)

type BitriseAPIClient interface {
	// ListBuildArtifacts lists all build artifacts that have been generated for an app’s build - https://api-docs.bitrise.io/#/build-artifact/artifact-list
	ListBuildArtifacts(appSlug, buildSlug string) ([]ArtifactListElementResponseModel, error)
	// ShowBuildArtifact retrieves data of a specific build artifact - https://api-docs.bitrise.io/#/build-artifact/artifact-show
	ShowBuildArtifact(appSlug, buildSlug, artifactSlug string) (ArtifactResponseItemModel, error)
}

type ArtifactLister struct {
	apiClient                         BitriseAPIClient
	logger                            log.Logger
	maxConcurrentListArtifactAPICalls int
	maxConcurrentShowArtifactAPICalls int
}

func NewArtifactLister(client BitriseAPIClient, logger log.Logger) ArtifactLister {
	return ArtifactLister{
		apiClient:                         client,
		logger:                            logger,
		maxConcurrentListArtifactAPICalls: 3,
		maxConcurrentShowArtifactAPICalls: 3, // per list artifact call (so the total number of max API calls is maxConcurrentListArtifactAPICalls * maxConcurrentShowArtifactAPICalls)
	}
}

func (lister ArtifactLister) ListBuildArtifacts(appSlug string, buildSlugs []string) ([]ArtifactResponseItemModel, error) {
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
		lister.logger.Debugf("listing artifacts for build %v", buildSlug)
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
				} else {
					artifacts = append(artifacts, res.artifact)
				}
			}

			results <- listArtifactsResult{buildSlug: buildSlug, artifacts: artifacts}
		}
	}
}

func (lister ArtifactLister) showArtifactWorker(appSlug, buildSlug string, artifactSlugs chan string, results chan showArtifactResult) {
	for artifactSlug := range artifactSlugs {
		lister.logger.Debugf("getting artifact details for artifact %v", artifactSlug)

		artifact, err := lister.apiClient.ShowBuildArtifact(appSlug, buildSlug, artifactSlug)
		if err != nil {
			results <- showArtifactResult{buildSlug: buildSlug, err: err}
		} else {
			results <- showArtifactResult{buildSlug: buildSlug, artifact: artifact}
		}
	}
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
