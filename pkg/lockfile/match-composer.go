package lockfile

import (
	"encoding/json"
	jsonUtils "github.com/google/osv-scanner/internal/json"
	"io"
	"path/filepath"
	"strings"

	"github.com/google/osv-scanner/pkg/models"
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
		pkgIndexes := jsonUtils.ExtractPackageIndexes(pkg.Name, "", content)
		if len(pkgIndexes) == 0 {
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
	requireLineOffset := jsonUtils.GetSectionOffset("require", contentStr)
	requireDevLineOffset := jsonUtils.GetSectionOffset("require-dev", contentStr)

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

	for index := range packages {
		jsonFile.Require.packages[index] = &packages[index]
		jsonFile.RequireDev.packages[index] = &packages[index]
	}

	return json.Unmarshal(content, &jsonFile)
}

func (depMap *dependencyMap) updatePackageDetails(pkg *PackageDetails, content string, indexes []int) {
	lineStart := depMap.lineOffset + strings.Count(content[:indexes[0]], "\n")
	lineStartIndex := strings.LastIndex(content[:indexes[0]], "\n")
	lineEnd := depMap.lineOffset + strings.Count(content[:indexes[1]], "\n")
	lineEndIndex := strings.LastIndex(content[:indexes[1]], "\n")

	pkg.IsDirect = true

	pkg.BlockLocation = models.FilePosition{
		Filename: depMap.filePath,
		Line: models.Position{
			Start: lineStart + 1,
			End:   lineEnd + 1,
		},
		Column: models.Position{
			Start: indexes[0] - lineStartIndex,
			End:   indexes[1] - lineEndIndex,
		},
	}

	pkg.NameLocation = &models.FilePosition{
		Filename: depMap.filePath,
		Line: models.Position{
			Start: lineStart + 1,
			End:   lineStart + 1,
		},
		Column: models.Position{
			Start: indexes[2] - lineStartIndex,
			End:   indexes[3] - lineStartIndex,
		},
	}

	pkg.VersionLocation = &models.FilePosition{
		Filename: depMap.filePath,
		Line: models.Position{
			Start: lineEnd + 1,
			End:   lineEnd + 1,
		},
		Column: models.Position{
			Start: indexes[4] - lineEndIndex,
			End:   indexes[5] - lineEndIndex,
		},
	}
}
