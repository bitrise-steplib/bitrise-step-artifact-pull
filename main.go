package main

import (
	"os"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-steputils/stepenv"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/log"
)

func main() {
	logger := log.NewLogger()
	if err := run(logger); err != nil {
		logger.Errorf(err.Error())
		os.Exit(1)
	}
}

func run(logger log.Logger) error {
	envRepository := stepenv.NewRepository(env.NewRepository())
	cmdFactory := command.NewFactory(envRepository)
	inputParser := stepconf.NewInputParser(envRepository)

	artifactPull := ArtifactPull{
		inputParser:   inputParser,
		envRepository: envRepository,
		cmdFactory:    cmdFactory,
		logger:        logger,
	}

	config, err := artifactPull.ProcessConfig()
	if err != nil {
		return err
	}

	result, err := artifactPull.Run(config)
	if err != nil {
		return err
	}

	if err := artifactPull.Export(result, config.ExportMap); err != nil {
		return err
	}

	return nil
}
