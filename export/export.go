package export

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/log"
)

type OutputExporter struct {
	ExportPattern map[string]string
	ExportValues  string
	Logger        log.Logger
	EnvRepository env.Repository
}

func ProcessRawExportMap(rawMap string) map[string]string {
	res := make(map[string]string)
	rawExportMapArray := strings.Split(strings.TrimSpace(rawMap), "\n")
	for _, line := range rawExportMapArray {
		if len(strings.TrimSpace(line)) > 0 {
			parsedLine := strings.Split(line, ":")

			variableName := strings.TrimSpace(parsedLine[0])
			patternValue := strings.TrimSpace(parsedLine[1])
			if len(parsedLine) != 2 ||
				len(variableName) == 0 ||
				len(patternValue) == 0 {
				continue
			}

			res[variableName] = patternValue
		}
	}

	return res
}

func (oe OutputExporter) Export() error {
	if len(oe.ExportPattern) == 0 {
		return oe.simpleOutputExport()
	}
	return oe.patternBasedOutputExport()
}

func (oe OutputExporter) simpleOutputExport() error {
	err := oe.exportOutputVariable("BITRISE_ARTIFACT_PATHS", oe.ExportValues)
	if err != nil {
		return err
	}

	oe.Logger.Println()
	oe.Logger.Printf("The following outputs are exported as environment variables:")
	oe.Logger.Printf("$BITRISE_ARTIFACT_PATHS = %s", oe.ExportValues)

	return nil
}

func (oe OutputExporter) patternBasedOutputExport() error {
	filePaths := strings.Split(oe.ExportValues, "|")

	exportMap := make(map[string][]string)

	for k, v := range oe.ExportPattern {
		for _, filePath := range filePaths {
			valueExpressions := strings.Split(v, ",")

			for _, expression := range valueExpressions {
				matched, err := filepath.Match(expression, filepath.Base(filePath))
				if err != nil {
					return err
				}

				if matched {
					if el := exportMap[k]; el != nil {
						el = append(el, filePath)
						exportMap[k] = el
					} else {
						exportMap[k] = []string{filePath}
					}
				}
			}
		}
	}

	oe.Logger.Println()
	oe.Logger.Printf("\nThe following outputs are exported as environment variables:")

	for k, v := range exportMap {
		exportVal := strings.Join(v, "|")

		err := oe.exportOutputVariable(k, exportVal)
		if err != nil {
			return err
		}

		oe.Logger.Printf("$%s = %s", k, exportVal)
	}

	return nil
}

func (oe OutputExporter) exportOutputVariable(key string, value string) error {
	if err := oe.EnvRepository.Set(key, value); err != nil {
		return fmt.Errorf("failed to export pulled artifact locations, error: %s", err)
	}
	return nil
}
