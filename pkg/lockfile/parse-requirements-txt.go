package lockfile

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/osv-scanner/pkg/models"

	"github.com/google/osv-scanner/internal/cachedregexp"
	"github.com/google/osv-scanner/internal/utility/fileposition"
	"golang.org/x/exp/maps"
)

const PipEcosystem Ecosystem = "PyPI"

// https://regex101.com/r/ppD7Uj/1
var wheelURLPattern = cachedregexp.MustCompile(
	`^.*?\/(?P<distribution>[^-/]+)-(?P<version>[^-/]+)(-(?P<buildtag>[^-/]+))?-(?P<pythontag>[^-/]+)-(?P<abitag>[^-/]+)-(?P<platformtag>[^-/]+)\.whl\s*$`)

// normalizedName ensures that the package name is normalized per PEP-0503
// and then removing "added support" syntax if present.
//
// This is done to ensure we don't miss any advisories, as while the OSV
// specification says that the normalized name should be used for advisories,
// that's not the case currently in our databases, _and_ Pip itself supports
// non-normalized names in the requirements.txt, so we need to normalize
// on both sides to ensure we don't have false negatives.
//
// It's possible that this will cause some false positives, but that is better
// than false negatives, and can be dealt with when/if it actually happens.
func normalizedRequirementName(name string) string {
	// per https://www.python.org/dev/peps/pep-0503/#normalized-names
	name = cachedregexp.MustCompile(`[-_.]+`).ReplaceAllString(name, "-")
	name = strings.ToLower(name)
	name, _, _ = strings.Cut(name, "[")

	return name
}

func isNotRequirementLine(line string) bool {
	return line == "" ||
		// flags are not supported
		strings.HasPrefix(line, "-") ||
		// file urls
		strings.HasPrefix(line, "https://") ||
		strings.HasPrefix(line, "http://") ||
		// file paths are not supported (relative or absolute)
		strings.HasPrefix(line, ".") ||
		strings.HasPrefix(line, "/")
}

func isLineContinuation(line string) bool {
	// checks that the line ends with an odd number of back slashes,
	// meaning the last one isn't escaped
	var re = cachedregexp.MustCompile(`([^\\]|^)(\\{2})*\\$`)

	return re.MatchString(line)
}

// Please note the whl filename has been standardized here :
// https://packaging.python.org/en/latest/specifications/binary-distribution-format/#file-name-convention
func extractVersionFromWheelURL(wheelURL string) string {
	matches := wheelURLPattern.FindStringSubmatch(wheelURL)

	if len(matches) == 0 {
		return ""
	}

	if version := matches[wheelURLPattern.SubexpIndex("version")]; version != "" {
		return version
	}

	return ""
}

type RequirementsTxtExtractor struct{}

func (e RequirementsTxtExtractor) ShouldExtract(path string) bool {
	baseFilepath := filepath.Base(path)
	return strings.Contains(baseFilepath, "requirements") && strings.HasSuffix(baseFilepath, ".txt")
}

func (e RequirementsTxtExtractor) Extract(f DepFile) ([]PackageDetails, error) {
	return parseRequirementsTxt(f, map[string]struct{}{})
}

func parseRequirementsTxt(f DepFile, requiredAlready map[string]struct{}) ([]PackageDetails, error) {
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
	var lineNumber, lineOffset, columnStart, columnEnd int

	for scanner.Scan() {
		lineNumber += lineOffset + 1
		lineOffset = 0

		line := scanner.Text()
		lastLine := line
		columnStart = fileposition.GetFirstNonEmptyCharacterIndexInLine(line)

		for isLineContinuation(line) {
			line = strings.TrimSuffix(line, "\\")

			if scanner.Scan() {
				lineOffset++
				newLine := scanner.Text()
				line += "\n" + newLine
				lastLine = newLine
			}
		}

		columnEnd = fileposition.GetLastNonEmptyCharacterIndexInLine(lastLine)

		cleanLine := commentsRegexp.ReplaceAllLiteralString(strings.TrimSpace(line), "")
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

				details, err := parseRequirementsTxt(af, requiredAlready)

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

		detail, err := parseRequirementLine(f.Path(), models.Requirements, line, cleanLine, lineNumber, lineOffset, columnStart, columnEnd)
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
		return []PackageDetails{}, fmt.Errorf("error while scanning %s: %w", f.Path(), err)
	}

	return maps.Values(packages), nil
}

var _ Extractor = RequirementsTxtExtractor{}

//nolint:gochecknoinits
func init() {
	registerExtractor("requirements.txt", RequirementsTxtExtractor{})
}

func ParseRequirementsTxt(pathToLockfile string) ([]PackageDetails, error) {
	return extractFromFile(pathToLockfile, RequirementsTxtExtractor{})
}
