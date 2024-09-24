package lockfile

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"github.com/google/osv-scanner/internal/utility/fileposition"
	"github.com/google/osv-scanner/pkg/models"

	"golang.org/x/exp/maps"
)

// Spec: https://setuptools.pypa.io/en/latest/userguide/declarative_config.html

type SetupCfgExtractor struct{}

func (e SetupCfgExtractor) ShouldExtract(path string) bool {
	return filepath.Base(path) == "setup.cfg"
}

func (e SetupCfgExtractor) Extract(f DepFile) ([]PackageDetails, error) {
	return parseSetupCfg(f, map[string]struct{}{})
}

func parseSetupCfg(f DepFile, requiredAlready map[string]struct{}) ([]PackageDetails, error) {
	packages := map[string]PackageDetails{}

	group := strings.TrimSuffix(filepath.Base(f.Path()), filepath.Ext(f.Path()))
	hasGroup := func(groups []string) bool {
		for _, g := range groups {
			if g == group {
				return true
			}
		}

		return false
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	scanner.Buffer([]byte{}, math.MaxInt)
	var lineNumber, lineOffset, columnStart, columnEnd int

	var inOptions, inInstallRequires bool
	for scanner.Scan() {
		lineNumber += lineOffset + 1
		lineOffset = 0

		line := scanner.Text()

		columnStart = fileposition.GetFirstNonEmptyCharacterIndexInLine(line)
		columnEnd = fileposition.GetLastNonEmptyCharacterIndexInLine(line)

		// Empty line
		if columnStart == -1 && columnEnd == -1 {
			continue
		}

		cleanLine := strings.TrimSpace(line)
		cleanLine = commentsRegexp.ReplaceAllLiteralString(cleanLine, "")
		if cleanLine == "" {
			continue
		}

		if !inOptions {
			if strings.HasPrefix(cleanLine, "[options]") {
				inOptions = true
			}

			continue
		} else if !inInstallRequires {
			matches := installRequiresRegexp.FindStringSubmatch(cleanLine)
			if len(matches) == 0 {
				continue
			}

			inInstallRequires = true

			requirementsText := matches[installRequiresRegexp.SubexpIndex("requirements")]
			if len(requirementsText) == 0 {
				continue
			}

			requirements := strings.Split(requirementsText, ";")

			for _, requirement := range requirements {
				detail, err := ParseRequirementLine(f.Path(), models.SetupTools, line, requirement, lineNumber, lineOffset, columnStart, columnEnd)
				if err != nil {
					return nil, err
				}
				key := detail.Name + "@" + detail.Version
				if _, ok := packages[key]; !ok {
					packages[key] = *detail
				}
				d := packages[key]
				if !hasGroup(d.DepGroups) {
					d.DepGroups = append(d.DepGroups, group)
					packages[key] = d
				}
			}

			break
		}

		if columnStart == 1 {
			break
		}

		if ar := strings.TrimPrefix(cleanLine, "-r "); ar != cleanLine {
			if strings.HasPrefix(ar, "http://") || strings.HasPrefix(ar, "https://") {
				// If the linked requirement file is not locally stored, we skip it
				continue
			}
			err := func() error {
				af, err := f.Open(ar)

				if err != nil {
					return fmt.Errorf("failed to include %s: %w", line, err)
				}

				defer af.Close()

				if _, ok := requiredAlready[af.Path()]; ok {
					return nil
				}

				requiredAlready[af.Path()] = struct{}{}

				details, err := parseSetupCfg(af, requiredAlready)

				if err != nil {
					return fmt.Errorf("failed to include %s: %w", line, err)
				}

				for _, detail := range details {
					packages[detail.Name+"@"+detail.Version] = detail
				}

				return nil
			}()

			if err != nil {
				return []PackageDetails{}, err
			}

			continue
		}

		if isNotRequirementLine(cleanLine) {
			continue
		}

		detail, err := ParseRequirementLine(f.Path(), models.SetupTools, line, cleanLine, lineNumber, lineOffset, columnStart, columnEnd)
		if err != nil {
			return nil, err
		}
		key := detail.Name + "@" + detail.Version
		if _, ok := packages[key]; !ok {
			packages[key] = *detail
		}
		d := packages[key]
		if !hasGroup(d.DepGroups) {
			d.DepGroups = append(d.DepGroups, group)
			packages[key] = d
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error while scanning %s: %w", f.Path(), err)
	}

	if !inOptions || !inInstallRequires {
		return nil, errors.New("could not find options.install_requires")
	}

	return maps.Values(packages), nil
}

var _ Extractor = SetupCfgExtractor{}

//nolint:gochecknoinits
func init() {
	registerExtractor("setup.cfg", SetupCfgExtractor{})
}

func ParseSetupCfg(pathToLockfile string) ([]PackageDetails, error) {
	return extractFromFile(pathToLockfile, SetupCfgExtractor{})
}
