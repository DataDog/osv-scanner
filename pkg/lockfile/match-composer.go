package lockfile

import (
	"encoding/json"
	"github.com/google/osv-scanner/internal/cachedregexp"
	"github.com/google/osv-scanner/pkg/models"
	"io"
	"path/filepath"
	"strings"
)

const composerFilename = "composer.json"

type ComposerMatcher struct{}

const (
	typeRequire = iota
	typeRequireDev
)

type dependencyMap struct {
	rootType   int
	filePath   string
	lineOffset int
	packages   []*PackageDetails
}

type composerFile struct {
	Require    dependencyMap `json:"require"`
	RequireDev dependencyMap `json:"require-dev"`
}

func (depMap *dependencyMap) UnmarshalJSON(bytes []byte) error {
	content := string(bytes)

	for _, pkg := range depMap.packages {
		if depMap.rootType == typeRequireDev && pkg.BlockLocation.Line.Start != 0 {
			// If it is dev dependency definition and we already found a package location,
			// we skip it to prioritize non-dev dependencies
			continue
		}
		pkgIndexes := depMap.extractPackageIndexes(pkg.Name, content)
		if pkgIndexes == nil || len(pkgIndexes) == 0 || len(pkgIndexes[0]) < 6 {
			// The matcher haven't found package information, lets skip the package
			continue
		}
		depMap.updatePackageDetails(pkg, content, pkgIndexes)
	}

	return nil
}

func (matcher ComposerMatcher) GetSourceFile(lockfile DepFile) (DepFile, error) {
	lockfileDir := filepath.Dir(lockfile.Path())
	sourceFilePath := filepath.Join(lockfileDir, composerFilename)
	file, err := OpenLocalDepFile(sourceFilePath)

	return file, err
}

func (matcher ComposerMatcher) Match(sourceFile DepFile, packages []PackageDetails) error {
	content, err := io.ReadAll(sourceFile)
	if err != nil {
		return err
	}
	contentStr := string(content)
	requireIndex := cachedregexp.MustCompile("\"require\"\\s*:\\s*{").FindStringIndex(contentStr)
	requireDevIndex := cachedregexp.MustCompile("\"require-dev\"\\s*:\\s*{").FindStringIndex(contentStr)
	requireLineOffset, requireDevLineOffset := 0, 0

	if requireIndex != nil && len(requireIndex) > 1 {
		requireLineOffset = strings.Count(contentStr[:requireIndex[1]], "\n")
	}
	if requireDevIndex != nil && len(requireDevIndex) > 1 {
		requireDevLineOffset = strings.Count(contentStr[:requireDevIndex[1]], "\n")
	}

	jsonFile := composerFile{
		Require: dependencyMap{
			rootType:   typeRequire,
			filePath:   sourceFile.Path(),
			lineOffset: requireLineOffset,
			packages:   make([]*PackageDetails, len(packages)),
		},
		RequireDev: dependencyMap{
			rootType:   typeRequireDev,
			filePath:   sourceFile.Path(),
			lineOffset: requireDevLineOffset,
			packages:   make([]*PackageDetails, len(packages)),
		},
	}

	for index, _ := range packages {
		jsonFile.Require.packages[index] = &packages[index]
		jsonFile.RequireDev.packages[index] = &packages[index]
	}

	return json.Unmarshal(content, &jsonFile)
}

/*
This method find where a package is defined in the composer.json file. It returns the block indexes along with
name and version. The expected result is a [1][6]int, with the only existing row being as follows :
- index 0/1 represents block start/end
- index 2/3 represents name start/end
- index 4/5 represents version start/end
*/
func (depMap *dependencyMap) extractPackageIndexes(pkgName string, content string) [][]int {
	pkgMatcher := cachedregexp.MustCompile(".*\"(?P<pkgName>" + pkgName + ")\"\\s*:\\s*\"(?P<version>.*)\"")

	return pkgMatcher.FindAllStringSubmatchIndex(content, -1)
}

func (depMap *dependencyMap) updatePackageDetails(pkg *PackageDetails, content string, indexes [][]int) {
	lineStart := depMap.lineOffset + strings.Count(content[:indexes[0][0]], "\n")
	lineStartIndex := strings.LastIndex(content[:indexes[0][0]], "\n")
	lineEnd := depMap.lineOffset + strings.Count(content[:indexes[0][1]], "\n")
	lineEndIndex := strings.LastIndex(content[:indexes[0][1]], "\n")

	pkg.BlockLocation = models.FilePosition{
		Filename: depMap.filePath,
		Line: models.Position{
			Start: lineStart + 1,
			End:   lineEnd + 1,
		},
		Column: models.Position{
			Start: indexes[0][0] - lineStartIndex,
			End:   indexes[0][1] - lineEndIndex,
		},
	}

	pkg.NameLocation = &models.FilePosition{
		Filename: depMap.filePath,
		Line: models.Position{
			Start: lineStart + 1,
			End:   lineStart + 1,
		},
		Column: models.Position{
			Start: indexes[0][2] - lineStartIndex,
			End:   indexes[0][3] - lineStartIndex,
		},
	}

	pkg.VersionLocation = &models.FilePosition{
		Filename: depMap.filePath,
		Line: models.Position{
			Start: lineEnd + 1,
			End:   lineEnd + 1,
		},
		Column: models.Position{
			Start: indexes[0][4] - lineEndIndex,
			End:   indexes[0][5] - lineEndIndex,
		},
	}
}
