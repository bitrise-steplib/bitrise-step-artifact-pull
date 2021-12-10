package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/bitrise-step-artifact-pull/api"
	"github.com/bitrise-steplib/bitrise-step-artifact-pull/downloader"
	"github.com/bitrise-steplib/bitrise-step-artifact-pull/model"
)

type Input struct {
	Verbose               bool   `env:"verbose,required"`
	ArtifactSources       string `env:"artifact_sources"`
	FinishedStages        string `env:"finished_stage"`
	BitriseAPIAccessToken string `env:"bitrise_api_access_token"`
	BitriseAPIBaseURL     string `env:"bitrise_api_base_url"`
}

type Config struct {
	VerboseLogging        bool
	ArtifactSources       []string
	FinishedStages        model.FinishedStages
	BitriseAPIAccessToken string
	BitriseAPIBaseURL     string
	AppSlug               string
}

type Result struct {
	ArtifactLocations []string
}

type ArtifactPull struct {
	inputParser   stepconf.InputParser
	envRepository env.Repository
	cmdFactory    command.Factory
	logger        log.Logger
}

func (a ArtifactPull) ProcessConfig() (Config, error) {
	var input Input
	err := a.inputParser.Parse(&input)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse step inputs: %w", err)
	}

	stepconf.Print(input)

	finishedStages := input.FinishedStages
	var finishedStagesModel model.FinishedStages
	if finishedStages != "" {
		if err := json.Unmarshal([]byte(finishedStages), &finishedStagesModel); err != nil {
			return Config{}, fmt.Errorf("failed to parse step inputs: %w", err)
		}
	}

	appSlug := a.envRepository.Get("BITRISE_APP_SLUG")
	if err != nil {
		return Config{}, fmt.Errorf("app slug (BITRISE_APP_SLUG env var) not found")
	}

	// TODO: validate inputs here and possibly convert from string to a concrete type
	return Config{
		VerboseLogging:        input.Verbose,
		ArtifactSources:       strings.Split(input.ArtifactSources, ","),
		FinishedStages:        finishedStagesModel,
		BitriseAPIAccessToken: input.BitriseAPIAccessToken,
		BitriseAPIBaseURL:     input.BitriseAPIBaseURL,
		AppSlug:               appSlug,
	}, nil
}

func (a ArtifactPull) Run(cfg Config) (Result, error) {
	a.logger.EnableDebugLog(cfg.VerboseLogging)
	buildIdGetter := NewBuildIDGetter(cfg.FinishedStages, cfg.ArtifactSources)
	buildIDs, err := buildIdGetter.GetBuildIDs()
	if err != nil {
		return Result{}, err
	}

	a.logger.Debugf("Downloading artifacts for builds %+v", buildIDs)

	apiClient, err := api.NewDefaultBitriseAPIClient(cfg.BitriseAPIBaseURL, cfg.BitriseAPIAccessToken)
	if err != nil {
		return Result{}, err
	}

	a.logger.Printf("getting the list of artifacts of %d builds", len(buildIDs))

	artifactLister := api.NewArtifactLister(&apiClient, a.logger)
	artifacts, err := artifactLister.ListBuildArtifacts(cfg.AppSlug, buildIDs)

	if err != nil {
		a.logger.Printf("failed", err)
		return Result{}, err
	}

	a.logger.Printf("downloading %d artifacts", len(artifacts))

	fileDownloader := downloader.NewDefaultFileDownloader(a.logger)
	artifactDownloader := downloader.NewConcurrentArtifactDownloader(artifacts, fileDownloader, a.logger)

	downloadResults, err := artifactDownloader.DownloadAndSaveArtifacts()
	if err != nil {
		a.logger.Printf("failed", err)
		return Result{}, err
	}

	var downloadedArtifactLocatins []string
	for _, downloadResult := range downloadResults {
		if downloadResult.DownloadError != nil {
			a.logger.Errorf("failed to download artifact from %s, error: %s", downloadResult.DownloadURL, downloadResult.DownloadError.Error())
		} else {
			a.logger.Printf("artifact downloaded: %s", downloadResult.DownloadPath)
			downloadedArtifactLocatins = append(downloadedArtifactLocatins, downloadResult.DownloadPath)
		}
	}

	return Result{ArtifactLocations: downloadedArtifactLocatins}, nil
}

func (a ArtifactPull) Export(result Result) error {
	if err := a.envRepository.Set("PULLED_ARTIFACT_LOCATIONS", strings.Join(result.ArtifactLocations, ",")); err != nil {
		return fmt.Errorf("failed to export pulled artifact locations, error: %s", err)
	}

	return nil
}
