package downloader

import (
	"fmt"
	"io"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/log"
)

// extractCacheArchive invokes tar tool by piping the archive to the command's input.
func extractCacheArchive(r io.Reader, targetDir string, compressed bool) error {
	factory := command.NewFactory(env.NewRepository())
	cmd := factory.Create("tar", []string{processArgs(true, compressed), "-"}, &command.Opts{
		Stdin: r,
		Dir:   targetDir,
	})

	printableCmd := fmt.Sprintf("curl <CACHE_URL> | %s", cmd.PrintableCommandArgs())
	log.Donef(printableCmd)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %s", printableCmd, err)
	}

	return nil
}

func processArgs(relative, compressed bool) string {
	/*
		GNU  tar options
		-f "-" : reads the archive from standard input
		https://www.gnu.org/software/tar/manual/html_node/Device.html#SEC155
		-x : extract files from an archive
		https://www.gnu.org/software/tar/manual/html_node/extract.html#SEC25
		-P : Don't strip an initial `/' from member names
		https://www.gnu.org/software/tar/manual/html_node/absolute.html#SEC120
		-z : tells tar to read or write archives through gzip
		https://www.gnu.org/software/tar/manual/html_node/gzip.html#SEC135
		BSD tar differences
		-z : In	extract	or list	modes, this option is ignored.
		Note that this tar implementation recognizes compress compression automatically when reading archives
		https://www.freebsd.org/cgi/man.cgi?query=bsdtar&sektion=1&manpath=freebsd-release-ports
	*/

	args := "-x"
	if !relative {
		args += "P"
	}
	if compressed {
		args += "z"
	}
	args += "f"
	return args
}
