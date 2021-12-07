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
	Verbose         bool   `env:"verbose,required"`
	ArtifactSources string `env:"artifact_sources"`
}

type Config struct {
	Verbose         bool
	ArtifactSources []string
	FinishedStages  model.FinishedStages
}

type Result struct {
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

	finishedStages := a.envRepository.Get("BITRISEIO_FINISHED_STAGES")

	var finishedStagesModel model.FinishedStages
	if finishedStages != "" {
		if err := json.Unmarshal([]byte(finishedStages), &finishedStagesModel); err != nil {
			return Config{}, fmt.Errorf("failed to parse step inputs: %w", err)
		}
	}

	// TODO: validate inputs here and possibly convert from string to a concrete type
	return Config{
		Verbose:         input.Verbose,
		ArtifactSources: strings.Split(input.ArtifactSources, ","),
		FinishedStages:  finishedStagesModel,
	}, nil
}

func (a ArtifactPull) Run(cfg Config) (Result, error) {
	a.logger.EnableDebugLog(cfg.Verbose)
	buildIdGetter := NewDefaultBuildIDGetter(cfg.FinishedStages, cfg.ArtifactSources)
	buildIDs, err := buildIdGetter.GetBuildIDs()
	if err != nil {
		return Result{}, err
	}

	// TODO: Do not print these build IDs, remove this line. It is just for the developer who will implement thed artifact download
	a.logger.Printf("%+v", buildIDs)

	authToken := a.envRepository.Get("BITRISEIO_ARTIFACT_PULL_TOKEN")
	if authToken == "" {
		return Result{}, fmt.Errorf("missing access token (BITRISEIO_ARTIFACT_PULL_TOKEN env var is not set)")
	}

	apiClient, err := api.NewDefaultBitriseAPIClient("https://api-staging.bitrise.io", authToken)
	if err != nil {
		return Result{}, err
	}

	a.logger.Printf("Getting artifact info")

	appSlug := a.envRepository.Get("BITRISE_APP_SLUG")
	if appSlug == "" {
		return Result{}, fmt.Errorf("missing app slug (BITRISE_APP_SLUG env var is not set)")
	}

	artifactLister := api.NewDefaultArtifactLister(&apiClient)
	artifacts, err := artifactLister.ListBuildArtifacts(appSlug, buildIDs) // TODO: how to get app slug?
	if err != nil {
		a.logger.Printf("failed", err)
		return Result{}, err
	}

	a.logger.Printf("Got artifact infos")
	for _, artifact := range artifacts {
		a.logger.Printf("%v - %v", artifact.Title, *artifact.DownloadPath)
	}

	a.logger.Printf("Downloading artifacts")
	fileDownloader := downloader.NewDefaultFileDownloader(a.logger)
	artifactDownloader := downloader.NewConcurrentArtifactDownloader(artifacts, fileDownloader, a.logger)

	downloadResults, err := artifactDownloader.DownloadAndSaveArtifacts()
	if err != nil {
		a.logger.Printf("failed", err)
		return Result{}, err
	}

	a.logger.Printf("Download result")
	for _, downloadResult := range downloadResults {
		if downloadResult.DownloadError != nil {
			a.logger.Errorf("failed to download artifact from %s, error: %s", downloadResult.DownloadURL, downloadResult.DownloadError.Error())
		} else {
			a.logger.Printf("artifact downloaded: %s", downloadResult.DownloadPath)
		}
	}

	// TODO
	return Result{}, nil
}

func (a ArtifactPull) Export(result Result) error {
	// TODO
	return nil
}
