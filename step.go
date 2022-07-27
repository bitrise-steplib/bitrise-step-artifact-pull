package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-steplib/bitrise-step-artifact-pull/api"
	"github.com/bitrise-steplib/bitrise-step-artifact-pull/downloader"
	"github.com/bitrise-steplib/bitrise-step-artifact-pull/export"
	"github.com/bitrise-steplib/bitrise-step-artifact-pull/model"
)

const downloadDirPrefix = "_artifact_pull"

type Input struct {
	ArtifactSources       string          `env:"artifact_sources,required"`
	Verbose               bool            `env:"verbose,opt[true,false]"`
	AppSlug               string          `env:"app_slug,required"`
	FinishedStages        string          `env:"finished_stage,required"`
	BitriseAPIBaseURL     string          `env:"bitrise_api_base_url,required"`
	BitriseAPIAccessToken stepconf.Secret `env:"bitrise_api_access_token,required"`
}

type Config struct {
	ArtifactSources       []string
	AppSlug               string
	FinishedStages        model.FinishedStages
	BitriseAPIBaseURL     string
	BitriseAPIAccessToken string
}

type Result struct {
	IntermediateFiles map[string]string
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
		return Config{}, err
	}

	stepconf.Print(input)
	a.logger.EnableDebugLog(input.Verbose)

	finishedStages := input.FinishedStages
	var finishedStagesModel model.FinishedStages
	if err := json.Unmarshal([]byte(finishedStages), &finishedStagesModel); err != nil {
		return Config{}, fmt.Errorf("invalid finished stages: %w", err)
	}

	return Config{
		ArtifactSources:       strings.Split(input.ArtifactSources, ","),
		AppSlug:               input.AppSlug,
		FinishedStages:        finishedStagesModel,
		BitriseAPIBaseURL:     input.BitriseAPIBaseURL,
		BitriseAPIAccessToken: string(input.BitriseAPIAccessToken),
	}, nil
}

func (a ArtifactPull) Run(cfg Config) (Result, error) {
	buildIdGetter := NewBuildIDGetter(cfg.FinishedStages, cfg.ArtifactSources)
	buildIDs, err := buildIdGetter.GetBuildIDs()
	if err != nil {
		return Result{}, fmt.Errorf("failed to get build IDs: %w", err)
	}

	a.logger.Println()
	a.logger.Debugf("Downloading artifacts for builds %+v", buildIDs)
	a.logger.Printf("Getting the list of artifacts of %d builds", len(buildIDs))

	artifactLister, err := api.NewArtifactLister(cfg.BitriseAPIBaseURL, cfg.BitriseAPIAccessToken, a.logger)
	if err != nil {
		return Result{}, fmt.Errorf("failed to create artifact lister: %w", err)
	}
	artifacts, err := artifactLister.ListBuildArtifactDetails(cfg.AppSlug, buildIDs)
	if err != nil {
		return Result{}, fmt.Errorf("failed to list artifacts: %w", err)
	}

	a.logger.Printf("Downloading %d artifacts", len(artifacts))

	targetDir, err := pathutil.NormalizedOSTempDirPath(downloadDirPrefix)
	if err != nil {
		return Result{}, fmt.Errorf("failed to create artifact download directory: %w", err)
	}

	artifactDownloader := downloader.NewConcurrentArtifactDownloader(artifacts, 5*time.Minute, targetDir, a.logger)
	downloadResults, err := artifactDownloader.DownloadAndSaveArtifacts()
	if err != nil {
		return Result{}, fmt.Errorf("failed to download artifacts: %w", err)
	}

	intermediateFiles := map[string]string{}
	for _, downloadResult := range downloadResults {
		if downloadResult.DownloadError != nil {
			return Result{}, fmt.Errorf("failed to download artifact from %s, error: %s", downloadResult.DownloadURL, downloadResult.DownloadError.Error())
		}

		a.logger.Debugf("Artifact downloaded: %s", downloadResult.DownloadPath)

		intermediateFiles[downloadResult.EnvKey] = downloadResult.DownloadPath
	}

	return Result{IntermediateFiles: intermediateFiles}, nil
}

func (a ArtifactPull) Export(result Result) error {
	exporter := export.OutputExporter{
		Logger:            a.logger,
		EnvRepository:     a.envRepository,
		IntermediateFiles: result.IntermediateFiles,
	}

	return exporter.Export()
}
