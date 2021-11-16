package main

import (
	"fmt"
	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/log"
)

type Input struct {
	Verbose bool `env:"verbose,required"`
}

type Config struct {
	Verbose bool
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

	// TODO: validate inputs here and possibly convert from string to a concrete type
	//nolint:gosimple
	return Config{
		Verbose: input.Verbose,
	}, nil
}

func (a ArtifactPull) Run(cfg Config) (Result, error) {
	a.logger.EnableDebugLog(cfg.Verbose)
	// TODO
	return Result{}, nil
}

func (a ArtifactPull) Export(result Result) error {
	// TODO
	return nil
}
