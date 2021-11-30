package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/bitrise-step-artifact-pull/model"
)

type Input struct {
	Verbose         bool   `env:"verbose,required"`
	ArtifactSources string `env:"artifact_source"`
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

	finishedStages := a.envRepository.Get("FINISHED_STAGES")

	var finishedStagesModel model.FinishedStages
	if err := json.Unmarshal([]byte(finishedStages), &finishedStagesModel); err != nil {
		return Config{}, fmt.Errorf("failed to parse step inputs: %w", err)
	}

	// TODO: validate inputs here and possibly convert from string to a concrete type
	//nolint:gosimple
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

	a.logger.Printf("%+v", buildIDs)
	// TODO
	return Result{}, nil
}

func (a ArtifactPull) Export(result Result) error {
	// TODO
	return nil
}
