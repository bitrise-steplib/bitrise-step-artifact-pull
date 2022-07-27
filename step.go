package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-steputils/tools"
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
	Verbose               string          `env:"verbose,opt[true,false]"`
	ArtifactSources       string          `env:"artifact_sources"`
	FinishedStages        string          `env:"finished_stage"`
	BitriseAPIAccessToken stepconf.Secret `env:"bitrise_api_access_token"`
	BitriseAPIBaseURL     string          `env:"bitrise_api_base_url"`
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
	ArtifactLocations []export.ArtifactLocation
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
	if appSlug == "" {
		return Config{}, fmt.Errorf("app slug (BITRISE_APP_SLUG env var) not found")
	}

	verboseLoggingValue := false
	if input.Verbose == "true" {
		verboseLoggingValue = true
	}

	return Config{
		VerboseLogging:        verboseLoggingValue,
		ArtifactSources:       strings.Split(input.ArtifactSources, ","),
		FinishedStages:        finishedStagesModel,
		BitriseAPIAccessToken: string(input.BitriseAPIAccessToken),
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

	a.logger.Println()
	a.logger.Debugf("Downloading artifacts for builds %+v", buildIDs)
	a.logger.Printf("Getting the list of artifacts of %d builds", len(buildIDs))

	artifactLister, err := api.NewArtifactLister(cfg.BitriseAPIBaseURL, cfg.BitriseAPIAccessToken, a.logger)
	if err != nil {
		a.logger.Debugf("Failed to create artifact lister", err)
		return Result{}, err
	}
	artifacts, err := artifactLister.ListBuildArtifactDetails(cfg.AppSlug, buildIDs)
	if err != nil {
		a.logger.Debugf("Failed to list artifacts", err)
		return Result{}, err
	}

	a.logger.Printf("Downloading %d artifacts", len(artifacts))

	targetDir, err := dirNamePrefix(downloadDirPrefix)
	if err != nil {
		a.logger.Printf("Failed to determine target artifact download directory", err)
		return Result{}, err
	}

	artifactDownloader := downloader.NewConcurrentArtifactDownloader(artifacts, 5*time.Minute, targetDir, a.logger)

	downloadResults, err := artifactDownloader.DownloadAndSaveArtifacts()
	if err != nil {
		a.logger.Printf("Failed", err)
		return Result{}, err
	}

	var artifactLocations []export.ArtifactLocation
	for _, downloadResult := range downloadResults {
		if downloadResult.DownloadError != nil {
			a.logger.Errorf("Failed to download artifact from %s, error: %s", downloadResult.DownloadURL, downloadResult.DownloadError.Error())

			return Result{}, downloadResult.DownloadError
		} else {
			if cfg.VerboseLogging {
				a.logger.Printf("Artifact downloaded: %s", downloadResult.DownloadPath)
			}

			if downloadResult.EnvKey != "" {
				if err := tools.ExportEnvironmentWithEnvman(downloadResult.EnvKey, downloadResult.DownloadPath); err != nil {
					a.logger.Errorf("Failed to export artifact: %s", err)
				}
			}

			artifactLocations = append(artifactLocations, export.ArtifactLocation{
				Path:   downloadResult.DownloadPath,
				EnvKey: downloadResult.EnvKey,
			})
		}
	}

	return Result{ArtifactLocations: artifactLocations}, nil
}

func (a ArtifactPull) Export(result Result) error {
	exporter := export.OutputExporter{
		Logger:            a.logger,
		EnvRepository:     a.envRepository,
		ArtifactLocations: result.ArtifactLocations,
	}

	return exporter.Export()
}

func dirNamePrefix(dirName string) (string, error) {
	tempPath, err := pathutil.NormalizedOSTempDirPath(dirName)
	if err != nil {
		return "", err
	}

	return tempPath, nil
}
