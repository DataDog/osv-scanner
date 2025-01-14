package scan

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/datadog/osv-scanner/pkg/lockfile"

	"github.com/datadog/osv-scanner/pkg/osvscanner"
	"github.com/datadog/osv-scanner/pkg/reporter"
	"github.com/urfave/cli/v2"
)

func Command(stdout, stderr io.Writer, r *reporter.Reporter) *cli.Command {
	return &cli.Command{
		Name:        "scan",
		Usage:       "scans various mediums for dependencies",
		Description: "scans various mediums for dependencies",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "sets the output format; value can be: " + strings.Join(reporter.Format(), ", "),
				Value:   "cyclonedx-1-5",
				Action: func(context *cli.Context, s string) error {
					if slices.Contains(reporter.Format(), s) {
						return nil
					}

					return fmt.Errorf("unsupported output format \"%s\" - must be one of: %s", s, strings.Join(reporter.Format(), ", "))
				},
			},
			&cli.StringFlag{
				Name:      "output",
				Usage:     "saves the result to the given file path",
				TakesFile: true,
			},
			&cli.BoolFlag{
				Name:    "recursive",
				Aliases: []string{"r"},
				Usage:   "check subdirectories",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:  "skip-git",
				Usage: "DEPRECATED: do nothing, will be removed in a later release",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "no-ignore",
				Usage: "also scan files that would be ignored by .gitignore",
				Value: false,
			},
			&cli.StringFlag{
				Name:  "verbosity",
				Usage: "specify the level of information that should be provided during runtime; value can be: " + strings.Join(reporter.VerbosityLevels(), ", "),
				Value: "info",
			},
			&cli.BoolFlag{
				Name:  "experimental-only-packages",
				Usage: "only collects packages, does not scan for vulnerabilities",
			},
			&cli.BoolFlag{
				Name:  "consider-scan-path-as-root",
				Usage: "Transform package path root to be the scanning path, thus removing any information about the host",
			},
			&cli.BoolFlag{
				Name:  "paths-relative-to-scan-dir",
				Usage: "Same than --consider-scan-path-as-root but reports a path relative to the scan dir (removing the leading path separator)",
			},
			&cli.StringSliceFlag{
				Name:  "enable-parsers",
				Usage: fmt.Sprintf("Explicitly define which lockfile to parse. If set, any non-set parsers will be ignored. (Available parsers: %v)", lockfile.ListExtractors()),
			},
		},
		ArgsUsage: "[directory1 directory2...]",
		Action: func(c *cli.Context) error {
			var err error
			*r, err = action(c, stdout, stderr)

			return err
		},
	}
}

func action(context *cli.Context, stdout, stderr io.Writer) (reporter.Reporter, error) {
	format := context.String("format")
	outputPath := context.String("output")
	var err error

	if outputPath != "" { // Output is definitely a file
		stdout, err = os.Create(outputPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create output file: %w", err)
		}
	}

	verbosityLevel, err := reporter.ParseVerbosityLevel(context.String("verbosity"))
	if err != nil {
		return nil, err
	}
	r, err := reporter.New(format, stdout, stderr, verbosityLevel)
	if err != nil {
		return r, err
	}

	vulnResult, err := osvscanner.DoScan(osvscanner.ScannerActions{
		Recursive:              context.Bool("recursive"),
		NoIgnore:               context.Bool("no-ignore"),
		DirectoryPaths:         context.Args().Slice(),
		ConsiderScanPathAsRoot: context.Bool("consider-scan-path-as-root"),
		PathRelativeToScanDir:  context.Bool("paths-relative-to-scan-dir"),
		EnableParsers:          context.StringSlice("enable-parsers"),
		ExperimentalScannerActions: osvscanner.ExperimentalScannerActions{
			OnlyPackages: context.Bool("experimental-only-packages"),
		},
	}, r)

	if err != nil && !errors.Is(err, osvscanner.NoPackagesFoundErr) && !errors.Is(err, osvscanner.VulnerabilitiesFoundErr) {
		return r, err
	}

	if errPrint := r.PrintResult(&vulnResult); errPrint != nil {
		return r, fmt.Errorf("failed to write output: %w", errPrint)
	}

	// This may be nil.
	return r, err
}
