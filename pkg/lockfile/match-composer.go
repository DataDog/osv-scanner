package lockfile

import (
	"encoding/json"
	"io"
	"path/filepath"

	jsonUtils "github.com/datadog/osv-scanner/internal/json"
)

const composerFilename = "composer.json"

type ComposerMatcher struct{}

const (
	typeRequire = iota
	typeRequireDev
)

type ComposerMatcherDependencyMap struct {
	MatcherDependencyMap
}

type composerFile struct {
	Require    ComposerMatcherDependencyMap `json:"require"`
	RequireDev ComposerMatcherDependencyMap `json:"require-dev"`
}

func (depMap *ComposerMatcherDependencyMap) UnmarshalJSON(bytes []byte) error {
	content := string(bytes)

	for _, pkg := range depMap.Packages {
		if depMap.RootType == typeRequireDev && pkg.BlockLocation.Line.Start != 0 {
			// If it is dev dependency definition and we already found a package location,
			// we skip it to prioritize non-dev dependencies
			continue
		}
		pkgIndexes := jsonUtils.ExtractPackageIndexes(pkg.Name, "", content)
		if len(pkgIndexes) == 0 {
			// The matcher haven't found package information, lets skip the package
			continue
		}
		depMap.UpdatePackageDetails(pkg, content, pkgIndexes, "")
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
		Require: ComposerMatcherDependencyMap{
			MatcherDependencyMap: MatcherDependencyMap{
				RootType:   typeRequire,
				FilePath:   sourceFile.Path(),
				LineOffset: requireLineOffset,
				Packages:   make([]*PackageDetails, len(packages)),
			},
		},
		RequireDev: ComposerMatcherDependencyMap{
			MatcherDependencyMap: MatcherDependencyMap{
				RootType:   typeRequireDev,
				FilePath:   sourceFile.Path(),
				LineOffset: requireDevLineOffset,
				Packages:   make([]*PackageDetails, len(packages)),
			},
		},
	}

	for index := range packages {
		jsonFile.Require.Packages[index] = &packages[index]
		jsonFile.RequireDev.Packages[index] = &packages[index]
	}

	return json.Unmarshal(content, &jsonFile)
}
