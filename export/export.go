package export

import (
	"fmt"

	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/log"
)

type OutputExporter struct {
	Logger        log.Logger
	EnvRepository env.Repository
}

func NewOutputExporter(logger log.Logger, envRepository env.Repository) OutputExporter {
	return OutputExporter{
		Logger:        logger,
		EnvRepository: envRepository,
	}
}

func (oe OutputExporter) Export(intermediateFiles map[string]string) error {
	if len(intermediateFiles) == 0 {
		return nil
	}

	oe.Logger.Println()
	oe.Logger.Printf("The following outputs are exported as environment variables:")

	for envKey, path := range intermediateFiles {
		if err := oe.exportOutputVariable(envKey, path); err != nil {
			return err
		}
	}
	return nil
}

func (oe OutputExporter) exportOutputVariable(key string, value string) error {
	if err := oe.EnvRepository.Set(key, value); err != nil {
		return fmt.Errorf("failed to export pulled artifact location (%s=%s), error: %s", key, value, err)
	}

	oe.Logger.Donef("$%s = %s", key, value)

	return nil
}
