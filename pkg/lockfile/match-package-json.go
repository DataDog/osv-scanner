package lockfile

import (
	"encoding/json"
	"io"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/google/osv-scanner/internal/cachedregexp"

	"github.com/google/osv-scanner/internal/utility/fileposition"
	"github.com/google/osv-scanner/pkg/models"
)

type PackageJSONMatcher struct {
	// Used to store the patterns for workspaces in a given root package.json
	WorkspacePatterns []string
}

type WorkspacePackageJSON struct {
	Workspaces []string `json:"workspaces"`
}

const (
	namePrefix       = "\""
	nameSuffix       = "\"\\s*:"
	versionPrefix    = ":\\s*\""
	versionSuffix    = "\",?"
	optionalPrefixes = "(file:|link:|portal:)?"
)

func (m PackageJSONMatcher) GetSourceFile(lockfile DepFile) (DepFile, error) {
	return lockfile.Open("package.json")
}

func tryGetNameLocation(name string, line string, lineNumber int) *models.FilePosition {
	nameRegexp := cachedregexp.QuoteMeta(name) + "(@.*)?"

	return fileposition.ExtractDelimitedRegexpPositionInBlock([]string{line}, nameRegexp, lineNumber, namePrefix, nameSuffix)
}

func tryGetVersionLocation(targetVersion string, line string, lineNumber int) *models.FilePosition {
	versionRegexp := optionalPrefixes + cachedregexp.QuoteMeta(targetVersion)

	return fileposition.ExtractDelimitedRegexpPositionInBlock([]string{line}, versionRegexp, lineNumber, versionPrefix, versionSuffix)
}

func updatePackageLocations(pkg *PackageDetails, nameLocation *models.FilePosition, versionLocation *models.FilePosition, line string, lineNumber int, sourcefilePath string) {
	// Update block location
	startColumn := fileposition.GetFirstNonEmptyCharacterIndexInLine(line)
	endColumn := fileposition.GetLastNonEmptyCharacterIndexInLine(strings.TrimSuffix(line, ","))
	pkg.BlockLocation = models.FilePosition{
		Line:     models.Position{Start: lineNumber, End: lineNumber},
		Column:   models.Position{Start: startColumn, End: endColumn},
		Filename: sourcefilePath,
	}
	// Update name location
	nameLocation.Filename = sourcefilePath
	pkg.NameLocation = nameLocation
	// Update version location
	versionLocation.Filename = sourcefilePath
	pkg.VersionLocation = versionLocation
	pkg.IsDirect = true
}

func globWorkspacePackageJsons(workspacePatterns []string, basePath string) []string {
	var results []string

	for _, pattern := range workspacePatterns {
		// Convert npm workspace pattern to package.json file pattern
		// e.g. "packages/*" -> "packages/*/package.json"
		searchPattern := filepath.Join(filepath.Dir(basePath), pattern, "package.json")

		matches, err := doublestar.Glob(searchPattern)
		if err != nil {
			continue
		}

		baseDir := filepath.Dir(basePath)
		for _, match := range matches {
			relPath, err := filepath.Rel(baseDir, match)
			if err != nil {
				continue
			}
			results = append(results, relPath)
		}
	}

	return results
}

func (m PackageJSONMatcher) Match(sourcefile DepFile, packages []PackageDetails) error {
	var wpj struct {
		Workspaces []string `json:"workspaces"`
	}

	content, err := io.ReadAll(sourcefile)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(content, &wpj); err != nil {
		m.matchFile(sourcefile, packages, content)
		return nil
	}

	// Match in root package.json
	m.matchFile(sourcefile, packages, content)

	// For workspaces, find and match in workspace package.json files
	if len(wpj.Workspaces) > 0 {
		matches := globWorkspacePackageJsons(wpj.Workspaces, sourcefile.Path())

		for _, match := range matches {
			workspacePkg, err := sourcefile.Open(match)
			if err != nil {
				continue
			}
			defer workspacePkg.Close()

			workspaceContent, err := io.ReadAll(workspacePkg)
			if err != nil {
				continue
			}

			// Match dependencies in workspace package.json
			m.matchFile(workspacePkg, packages, workspaceContent)
		}
	}

	return nil
}

func (m PackageJSONMatcher) matchFile(file DepFile, packages []PackageDetails, content []byte) {
	lines := fileposition.BytesToLines(content)
	for index, line := range lines {
		lineNumber := index + 1
		for key, pkg := range packages {
			nameLocation := tryGetNameLocation(pkg.Name, line, lineNumber)
			if nameLocation != nil {
				for _, targetVersion := range pkg.TargetVersions {
					versionLocation := tryGetVersionLocation(targetVersion, line, lineNumber)
					if versionLocation != nil {
						updatePackageLocations(&packages[key], nameLocation, versionLocation, line, lineNumber, file.Path())
					}
				}
			}
		}
	}
}

var _ Matcher = PackageJSONMatcher{}
