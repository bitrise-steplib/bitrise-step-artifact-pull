package export

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/log"
)

type ArtifactLocation struct {
	Path   string
	EnvKey string
}

type OutputExporter struct {
	Logger        log.Logger
	EnvRepository env.Repository

	ArtifactLocations []ArtifactLocation
}

func (oe OutputExporter) Export() error {
	if len(oe.ArtifactLocations) == 0 {
		return nil
	}

	oe.Logger.Println()
	oe.Logger.Printf("The following outputs are exported as environment variables:")

	for _, artifact := range oe.ArtifactLocations {
		if artifact.EnvKey != "" {
			if err := oe.exportOutputVariable(artifact.EnvKey, artifact.Path); err != nil {
				return err
			}
		}
	}
	return oe.simpleOutputExport()
}

func (oe OutputExporter) simpleOutputExport() error {
	var paths []string
	for _, artifact := range oe.ArtifactLocations {
		paths = append(paths, artifact.Path)
	}
	exportValues := strings.Join(paths, "|")
	err := oe.exportOutputVariable("BITRISE_ARTIFACT_PATHS", exportValues)
	if err != nil {
		return err
	}

	oe.Logger.Donef("$BITRISE_ARTIFACT_PATHS = %s", exportValues)

	return nil
}

func (oe OutputExporter) exportOutputVariable(key string, value string) error {
	if err := oe.EnvRepository.Set(key, value); err != nil {
		return fmt.Errorf("failed to export pulled artifact location (%s=%s), error: %s", key, value, err)
	}

	oe.Logger.Donef("$%s = %s", key, value)

	return nil
}
