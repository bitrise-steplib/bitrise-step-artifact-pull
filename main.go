package main

import (
	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/log"
	"os"
)

func main() {
	os.Exit(run())
}

func run() int {
	logger := log.NewLogger()
	envRepository := env.NewRepository()
	cmdFactory := command.NewFactory(envRepository)
	inputParser := stepconf.NewInputParser(envRepository)

	artifactPull := ArtifactPull{
		inputParser:   inputParser,
		envRepository: envRepository,
		cmdFactory: cmdFactory,
		logger:        logger,
	}

	config, err := artifactPull.ProcessConfig()
	if err != nil {
		logger.Errorf(err.Error())
		return 1
	}

	result, err := artifactPull.Run(config)
	if err != nil {
		logger.Errorf(err.Error())
		return 1
	}

	if err := artifactPull.Export(result); err != nil {
		logger.Errorf(err.Error())
		return 1
	}

	return 0
}
