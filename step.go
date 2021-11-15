package main

import (
	"fmt"
	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/log"
)

type Input struct {
	Example string `env:"example,required"`
}

type Config struct {
	Example string
}

type Result struct {

}

type ArtifactPull struct {
	inputParser stepconf.InputParser
	envRepository env.Repository
	cmdFactory command.Factory
	logger log.Logger
}

func (a ArtifactPull) ProcessConfig() (Config, error) {
	var input Input
	err := a.inputParser.Parse(&input)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse step inputs: %w", err)
	}

	stepconf.Print(input)

	return Config{
		Example: input.Example,
	}, nil
}

func (a ArtifactPull) Run(cfg Config) (Result, error) {
	// TODO
	return Result{}, nil
}

func (a ArtifactPull) Export(result Result) error {
	// TODO
	return nil
}
