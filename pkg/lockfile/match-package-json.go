package lockfile

import (
	"encoding/json"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	jsonUtils "github.com/google/osv-scanner/internal/json"
)

const (
	typeDependencies = iota
	typeDevDependencies
	typeOptionalDependencies
)

type PackageJSONMatcher struct {
	// Used to store the patterns for workspaces in a given root package.json
	WorkspacePatterns []string
}

type WorkspacePackageJSON struct {
	Workspaces []string `json:"workspaces"`
}

/*
packageJSONDependencyMap is here to have access to all MatcherDependencyMap methods and at the same time having
a different type to have a clear UnmarshallJSON method for the json decoder and avoid overlaps with other matchers.
*/
type packageJSONDependencyMap struct {
	MatcherDependencyMap
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

	for _, pkg := range depMap.Packages {
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
		var depGroup string
		switch depMap.RootType {
		case typeDependencies:
			depGroup = "prod"
		case typeDevDependencies:
			depGroup = "dev"
		case typeOptionalDependencies:
			depGroup = "optional"
		}

		if (depMap.RootType == typeDevDependencies || depMap.RootType == typeOptionalDependencies) && pkg.BlockLocation.Line.Start != 0 {
			// If it is a dev or optional dependency definition and we already found a package location,
			// we skip it to prioritize non-dev dependencies
			pkgIndexes = []int{}
		}
		depMap.UpdatePackageDetails(pkg, content, pkgIndexes, depGroup)
	}

	return nil
}

func globWorkspacePackageJsons(workspacePatterns []string, basePath string) []string {
	var results []string
	// Create a filesystem rooted at the directory containing basePath
	baseDir := filepath.Dir(basePath)
	fsys := os.DirFS(baseDir)

	for _, pattern := range workspacePatterns {
		// Convert npm workspace pattern to package.json file pattern
		// Important: When we pass the pattern to doublestar.Glob, we need to ensure
		// it uses forward slashes as path separators, regardless of OS
		searchPattern := path.Join(pattern, "package.json")

		// Use the new function signature with the filesystem
		matches, err := doublestar.Glob(fsys, searchPattern)
		if err != nil {
			continue
		}

		results = append(results, matches...)
	}

	return results
}

/*
Match works by leveraging the json decoder to only parse json sections of interest (e.g dependencies)
Whenever the json decoder try to deserialize a file, it will look at json sections it needs to deserialize
and then call the proper UnmarshallJSON method of the type. As the JSON decoder expect us to only deserialize it,
not trying to find the exact location in the file of the content, it does not provide us buffer information (offset, file path, etc...)

To work around this limitation, we are pre-filling the structure with all the field we will need during the deserialization :
  - The root type to know which json section we are deserializing
  - The file path to be able to fill properly location fields of PackageDetails
  - The line offset to be able to compute the line of any found dependencies in the file
  - And a list of pointer to the original PackageDetails extracted by the parser to be able to modify them with the json section content
*/
func (m PackageJSONMatcher) Match(sourcefile DepFile, packages []PackageDetails) error {
	var wpj struct {
		Workspaces []string `json:"workspaces"`
	}

	content, err := io.ReadAll(sourcefile)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(content, &wpj); err != nil {
		err = m.matchFile(sourcefile, packages, content)
		if err != nil {
			return err
		}

		return nil
	}

	// Match in root package.json
	err = m.matchFile(sourcefile, packages, content)
	if err != nil {
		return err
	}

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
			if err := m.matchFile(workspacePkg, packages, workspaceContent); err != nil {
				continue
			}
		}
	}

	return nil
}

func (m PackageJSONMatcher) matchFile(file DepFile, packages []PackageDetails, content []byte) error {
	contentStr := string(content)
	dependenciesLineOffset := jsonUtils.GetSectionOffset("dependencies", contentStr)
	devDependenciesLineOffset := jsonUtils.GetSectionOffset("devDependencies", contentStr)
	optionalDepenenciesLineOffset := jsonUtils.GetSectionOffset("optionalDependencies", contentStr)

	jsonFile := packageJSONFile{
		Dependencies: packageJSONDependencyMap{
			MatcherDependencyMap: MatcherDependencyMap{
				RootType:   typeDependencies,
				FilePath:   file.Path(),
				LineOffset: dependenciesLineOffset,
			},
		},
		DevDependencies: packageJSONDependencyMap{
			MatcherDependencyMap: MatcherDependencyMap{
				RootType:   typeDevDependencies,
				FilePath:   file.Path(),
				LineOffset: devDependenciesLineOffset,
			},
		},
		OptionalDependencies: packageJSONDependencyMap{
			MatcherDependencyMap: MatcherDependencyMap{
				RootType:   typeOptionalDependencies,
				FilePath:   file.Path(),
				LineOffset: optionalDepenenciesLineOffset,
			},
		},
	}
	packagesPtr := make([]*PackageDetails, len(packages))
	for index := range packages {
		packagesPtr[index] = &packages[index]
	}
	jsonFile.Dependencies.Packages = packagesPtr
	jsonFile.DevDependencies.Packages = packagesPtr
	jsonFile.OptionalDependencies.Packages = packagesPtr

	return json.Unmarshal(content, &jsonFile)
}

var _ Matcher = PackageJSONMatcher{}
