package lockfile

import (
	"io"
	"strings"

	"github.com/google/osv-scanner/internal/cachedregexp"

	"github.com/google/osv-scanner/internal/utility/fileposition"
	"github.com/google/osv-scanner/pkg/models"
)

type PackageJSONMatcher struct{}

const (
	namePrefix    = "\""
	nameSuffix    = "\":"
	versionPrefix = ":\\s*\""
	versionSuffix = "\",?"
)

func (m PackageJSONMatcher) GetSourceFile(lockfile DepFile) (DepFile, error) {
	return lockfile.Open("package.json")
}

func (m PackageJSONMatcher) Match(sourcefile DepFile, packages []PackageDetails) error {
	content, err := io.ReadAll(sourcefile)
	if err != nil {
		return err
	}

	lines := fileposition.BytesToLines(content)
	for index, line := range lines {
		lineNumber := index + 1
		for key, pkg := range packages {
			// TODO: what to do if version is not in the same line as the name?
			block := []string{line}
			nameRegexp := cachedregexp.QuoteMeta(pkg.Name) + "(@.*)?"
			nameLocation := fileposition.ExtractDelimitedRegexpPositionInBlock(block, nameRegexp, lineNumber, namePrefix, nameSuffix)
			if nameLocation != nil {
				for _, targetVersion := range pkg.TargetVersions {
					var versionLocation *models.FilePosition
					if targetVersion == pkg.Version {
						versionLocation = fileposition.ExtractDelimitedRegexpPositionInBlock(block, targetVersion, lineNumber, versionPrefix, versionSuffix)
					} else {
						versionRegexp := ".*" + cachedregexp.QuoteMeta(targetVersion) + ".*"
						versionLocation = fileposition.ExtractDelimitedRegexpPositionInBlock(block, versionRegexp, lineNumber, versionPrefix, versionSuffix)
					}
					if versionLocation != nil {
						startColumn := fileposition.GetFirstNonEmptyCharacterIndexInLine(line)
						endColumn := fileposition.GetLastNonEmptyCharacterIndexInLine(strings.TrimSuffix(line, ","))
						packages[key].BlockLocation = models.FilePosition{
							Line:     models.Position{Start: lineNumber, End: lineNumber},
							Column:   models.Position{Start: startColumn, End: endColumn},
							Filename: sourcefile.Path(),
						}
						nameLocation.Filename = sourcefile.Path()
						packages[key].NameLocation = nameLocation
						versionLocation.Filename = sourcefile.Path()
						packages[key].VersionLocation = versionLocation
					}
				}
			}
		}
	}

	return nil
}

var _ Matcher = PackageJSONMatcher{}
