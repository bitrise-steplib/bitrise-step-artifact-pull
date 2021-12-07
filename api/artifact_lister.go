package api

import (
	"fmt"
	"strings"
	"sync"
)

type ArtifactLister interface {
	ListBuildArtifacts(appSlug string, buildSlugs []string) ([]ArtifactResponseItemModel, error)
}

type DefaultArtifactLister struct {
	apiClient                         BitriseAPIClient
	maxConcurrentListArtifactAPICalls int
	maxConcurrentShowArtifactAPICalls int
}

func NewDefaultArtifactLister(client BitriseAPIClient) DefaultArtifactLister {
	return DefaultArtifactLister{
		apiClient:                         client,
		maxConcurrentListArtifactAPICalls: 3,
		maxConcurrentShowArtifactAPICalls: 3, // per list artifact call (so the total number of max API calls is maxConcurrentListArtifactAPICalls * maxConcurrentShowArtifactAPICalls)
	}
}

func (lister DefaultArtifactLister) ListBuildArtifacts(appSlug string, buildSlugs []string) ([]ArtifactResponseItemModel, error) {
	wg := &sync.WaitGroup{}
	resultsChan := make(chan listArtifactsOfBuildResult, len(buildSlugs))

	concurrentCalls := 0
	for _, buildSlug := range buildSlugs {
		wg.Add(1)
		concurrentCalls++
		go lister.listArtifactsOfBuild(appSlug, buildSlug, resultsChan, wg)

		if concurrentCalls >= lister.maxConcurrentListArtifactAPICalls {
			wg.Wait()
			concurrentCalls = 0
		}
	}
	wg.Wait()
	close(resultsChan)

	// process results
	var (
		failedBuildSlugs []string
		artifacts        []ArtifactResponseItemModel
	)
	for res := range resultsChan {
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

// listArtifactsOfBuild gets details of all artifacts of a particular build using the Bitrise API
func (lister DefaultArtifactLister) listArtifactsOfBuild(appSlug, buildSlug string, resultsChan chan listArtifactsOfBuildResult, wg *sync.WaitGroup) {
	defer wg.Done()

	artifactListItems, err := lister.apiClient.ListBuildArtifacts(appSlug, buildSlug)
	if err != nil {
		resultsChan <- listArtifactsOfBuildResult{buildSlug: buildSlug, err: err}
	} else if len(artifactListItems) == 0 {
		resultsChan <- listArtifactsOfBuildResult{buildSlug: buildSlug}
	} else {
		showWg := &sync.WaitGroup{}
		showResultsChan := make(chan showArtifactResult, len(artifactListItems))

		concurrentCalls := 0
		for _, artifactListItem := range artifactListItems {
			showWg.Add(1)
			concurrentCalls++
			go lister.showArtifact(appSlug, buildSlug, artifactListItem.Slug, showResultsChan, showWg)
			if concurrentCalls >= lister.maxConcurrentShowArtifactAPICalls {
				showWg.Wait()
				concurrentCalls = 0
			}
		}

		showWg.Wait()
		close(showResultsChan)

		var artifacts []ArtifactResponseItemModel
		for res := range showResultsChan {
			if res.err != nil {
				resultsChan <- listArtifactsOfBuildResult{buildSlug: buildSlug, err: err}
				return
			} else {
				artifacts = append(artifacts, res.artifact)
			}
		}

		resultsChan <- listArtifactsOfBuildResult{buildSlug: buildSlug, artifacts: artifacts}
	}
}

func (lister DefaultArtifactLister) showArtifact(appSlug, buildSlug, artifactSlug string, resultsChan chan showArtifactResult, wg *sync.WaitGroup) {
	defer wg.Done()

	artifact, err := lister.apiClient.ShowBuildArtifact(appSlug, buildSlug, artifactSlug)
	if err != nil {
		resultsChan <- showArtifactResult{buildSlug: buildSlug, err: err}
	} else {
		resultsChan <- showArtifactResult{buildSlug: buildSlug, artifact: artifact}
	}
}

type listArtifactsOfBuildResult struct {
	buildSlug string
	artifacts []ArtifactResponseItemModel
	err       error
}

type showArtifactResult struct {
	buildSlug string
	artifact  ArtifactResponseItemModel
	err       error
}
