package lockfile

import (
	"encoding/json"
	"golang.org/x/exp/maps"
	"io"
	"regexp"
	"strings"

	"github.com/datadog/osv-scanner/internal/cachedregexp"

	"github.com/datadog/osv-scanner/pkg/models"
)

const (
	typeDependencies = iota
	typeDevDependencies
	typeOptionalDependencies
)

type PackageJSONMatcher struct{}

type packageJsonDependencyMap struct {
	rootType   int
	filePath   string
	lineOffset int
	packages   []*PackageDetails
}

type packageJsonFile struct {
	Dependencies         packageJsonDependencyMap `json:"dependencies"`
	DevDependencies      packageJsonDependencyMap `json:"devDependencies"`
	OptionalDependencies packageJsonDependencyMap `json:"optionalDependencies"`
}

func (m PackageJSONMatcher) GetSourceFile(lockfile DepFile) (DepFile, error) {
	return lockfile.Open("package.json")
}

func (depMap *packageJsonDependencyMap) UnmarshalJSON(data []byte) error {
	content := string(data)

	for _, pkg := range depMap.packages {
		var pkgIndexes []int

		for _, targetedVersion := range pkg.TargetVersions {
			pkgIndexes = depMap.extractPackageIndexes(pkg.Name, targetedVersion, content)
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

/*
This method find where a package is defined in a json source file. It returns the block indexes along with
name and version. It assumes the package won't be declared twice in the same block.

You can see the regex in action here: https://regex101.com/r/zzrEAh/1

The expected result of FindAllStringSubmatchIndex is a [6]int, with the following structure :
- index 0/1 represents block start/end
- index 2/3 represents name start/end
- index 4/5 represents version start/end
*/
// TODO : Unify in a util class with the composer one
func (depMap *packageJsonDependencyMap) extractPackageIndexes(pkgName, targetedVersion, content string) []int {
	pkgMatcher := cachedregexp.MustCompile(`"(?P<pkgName>` + pkgName + `)\"\s*:\s*\"(?P<version>` + regexp.QuoteMeta(targetedVersion) + `)"`)
	result := pkgMatcher.FindAllStringSubmatchIndex(content, -1)

	if len(result) == 0 || len(result[0]) < 6 {
		return []int{}
	}

	return result[0]
}

// TODO : Unify it in a util class with the composer one
func (depMap *packageJsonDependencyMap) updatePackageDetails(pkg *PackageDetails, content string, indexes []int) {
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

/*
*
This method computes the start line of any section in the file.
To see the regex in action, check out https://regex101.com/r/3EHqB8/1 (it uses the dependencies section as an example)
*/
// TODO : Unify it with composer
func getSectionOffset(sectionName string, content string) int {
	sectionMatcher := cachedregexp.MustCompile(`(?m)^\s*"` + sectionName + `":\s*{\s*$`)
	sectionIndex := sectionMatcher.FindStringIndex(content)
	if len(sectionIndex) < 2 {
		return -1
	}
	return strings.Count(content[:sectionIndex[1]], "\n")
}

func (m PackageJSONMatcher) Match(sourcefile DepFile, packages []PackageDetails) error {
	content, err := io.ReadAll(sourcefile)
	if err != nil {
		return err
	}
	contentStr := string(content)
	dependenciesLineOffset := getSectionOffset("dependencies", contentStr)
	devDependenciesLineOffset := getSectionOffset("devDependencies", contentStr)
	optionalDepenenciesLineOffset := getSectionOffset("optionalDependencies", contentStr)

	jsonFile := packageJsonFile{
		Dependencies: packageJsonDependencyMap{
			rootType:   typeDependencies,
			filePath:   sourcefile.Path(),
			lineOffset: dependenciesLineOffset,
		},
		DevDependencies: packageJsonDependencyMap{
			rootType:   typeDevDependencies,
			filePath:   sourcefile.Path(),
			lineOffset: devDependenciesLineOffset,
		},
		OptionalDependencies: packageJsonDependencyMap{
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
