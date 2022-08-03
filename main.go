package main

import (
	"os"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-steputils/stepenv"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/bitrise-step-artifact-pull/step"
)

func main() {
	os.Exit(run())
}

func run() int {
	logger := log.NewLogger()

	downloader := createIntermediateFileDownloader(logger)
	config, err := downloader.ProcessConfig()
	if err != nil {
		logger.Errorf("Process config: %s", err)
		return 1
	}

	result, err := downloader.Run(config)
	if err != nil {
		logger.Errorf("Run: %s", err)
		return 1
	}

	if err := downloader.Export(result); err != nil {
		logger.Errorf("Export outputs: %s", err)
		return 1
	}

	return 0
}

func createIntermediateFileDownloader(logger log.Logger) step.IntermediateFileDownloader {
	envRepository := stepenv.NewRepository(env.NewRepository())
	cmdFactory := command.NewFactory(envRepository)
	inputParser := stepconf.NewInputParser(envRepository)

	return step.NewIntermediateFileDownloader(inputParser, envRepository, cmdFactory, logger)
}
