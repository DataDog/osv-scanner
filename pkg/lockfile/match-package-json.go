package lockfile

import (
	"encoding/json"
	jsonUtils "github.com/google/osv-scanner/internal/json"
	"io"
	"strings"

	"golang.org/x/exp/maps"

	"github.com/google/osv-scanner/pkg/models"
)

const (
	typeDependencies = iota
	typeDevDependencies
	typeOptionalDependencies
)

type PackageJSONMatcher struct{}

type packageJSONDependencyMap struct {
	rootType   int
	filePath   string
	lineOffset int
	packages   []*PackageDetails
}

type packageJSONFile struct {
	Dependencies         packageJSONDependencyMap `json:"dependencies"`
	DevDependencies      packageJSONDependencyMap `json:"devDependencies"`
	OptionalDependencies packageJSONDependencyMap `json:"optionalDependencies"`
}

func (m PackageJSONMatcher) GetSourceFile(lockfile DepFile) (DepFile, error) {
	return lockfile.Open("package.json")
}

func (depMap *packageJSONDependencyMap) UnmarshalJSON(data []byte) error {
	content := string(data)

	for _, pkg := range depMap.packages {
		var pkgIndexes []int

		for _, targetedVersion := range pkg.TargetVersions {
			pkgIndexes = jsonUtils.ExtractPackageIndexes(pkg.Name, targetedVersion, content)
			if len(pkgIndexes) > 0 {
				break
			}
		}

		if len(pkgIndexes) == 0 {
			// The matcher haven't found package information, lets skip it
			continue
		}

		if depMap.rootType == typeDependencies {
			pkg.DepGroups = append(pkg.DepGroups, "prod")
		} else if depMap.rootType == typeDevDependencies {
			pkg.DepGroups = append(pkg.DepGroups, "dev")
		} else if depMap.rootType == typeOptionalDependencies {
			pkg.DepGroups = append(pkg.DepGroups, "optional")
		}
		propagateDepGroups(pkg)

		if (depMap.rootType == typeDevDependencies || depMap.rootType == typeOptionalDependencies) && pkg.BlockLocation.Line.Start != 0 {
			// If it is a dev or optional dependency definition and we already found a package location,
			// we skip it to prioritize non-dev dependencies
			continue
		}
		depMap.updatePackageDetails(pkg, content, pkgIndexes)
	}

	return nil
}

func propagateDepGroups(root *PackageDetails) {
	newDepGroups := make(map[string]bool)
	for _, group := range root.DepGroups {
		newDepGroups[group] = true
	}

	for _, deps := range root.Dependencies {
		for _, group := range deps.DepGroups {
			newDepGroups[group] = true
		}
		deps.DepGroups = maps.Keys(newDepGroups)
		propagateDepGroups(deps)
	}
}

// TODO : Unify it in a util class with the composer one
func (depMap *packageJSONDependencyMap) updatePackageDetails(pkg *PackageDetails, content string, indexes []int) {
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

func (m PackageJSONMatcher) Match(sourcefile DepFile, packages []PackageDetails) error {
	content, err := io.ReadAll(sourcefile)
	if err != nil {
		return err
	}
	contentStr := string(content)
	dependenciesLineOffset := jsonUtils.GetSectionOffset("dependencies", contentStr)
	devDependenciesLineOffset := jsonUtils.GetSectionOffset("devDependencies", contentStr)
	optionalDepenenciesLineOffset := jsonUtils.GetSectionOffset("optionalDependencies", contentStr)

	jsonFile := packageJSONFile{
		Dependencies: packageJSONDependencyMap{
			rootType:   typeDependencies,
			filePath:   sourcefile.Path(),
			lineOffset: dependenciesLineOffset,
		},
		DevDependencies: packageJSONDependencyMap{
			rootType:   typeDevDependencies,
			filePath:   sourcefile.Path(),
			lineOffset: devDependenciesLineOffset,
		},
		OptionalDependencies: packageJSONDependencyMap{
			rootType:   typeOptionalDependencies,
			filePath:   sourcefile.Path(),
			lineOffset: optionalDepenenciesLineOffset,
		},
	}
	packagesPtr := make([]*PackageDetails, len(packages))
	for index := range packages {
		packagesPtr[index] = &packages[index]
	}
	jsonFile.Dependencies.packages = packagesPtr
	jsonFile.DevDependencies.packages = packagesPtr
	jsonFile.OptionalDependencies.packages = packagesPtr

	return json.Unmarshal(content, &jsonFile)
}

var _ Matcher = PackageJSONMatcher{}
