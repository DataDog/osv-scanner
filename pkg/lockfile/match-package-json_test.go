package lockfile_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/osv-scanner/pkg/lockfile"
	"github.com/google/osv-scanner/pkg/models"
	"github.com/stretchr/testify/assert"
)

var packageJSONMatcher = lockfile.PackageJSONMatcher{}

func TestPackageJSONMatcher_GetSourceFile_FileDoesNotExist(t *testing.T) {
	t.Parallel()

	lockFile, err := lockfile.OpenLocalDepFile("fixtures/package-json/does-not-exist/npm-v1.json")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	sourceFile, err := packageJSONMatcher.GetSourceFile(lockFile)
	expectErrIs(t, err, fs.ErrNotExist)
	assert.Equal(t, "", sourceFile.Path())
}

func TestPackageJSONMatcher_GetSourceFile(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	basePath := "fixtures/package-json/one-package/"
	sourcefilePath := filepath.FromSlash(filepath.Join(dir, basePath+"package.json"))

	lockFile, err := lockfile.OpenLocalDepFile(basePath + "npm-v1.json")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	sourceFile, err := packageJSONMatcher.GetSourceFile(lockFile)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	assert.Equal(t, sourcefilePath, sourceFile.Path())
}

func TestPackageJSONMatcher_Match_OnePackage(t *testing.T) {
	t.Parallel()

	sourceFile, err := lockfile.OpenLocalDepFile("fixtures/package-json/one-package/package.json")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	packages := []lockfile.PackageDetails{
		{
			Name:           "lodash",
			TargetVersions: []string{"^4.0.0"},
		},
	}
	err = packageJSONMatcher.Match(sourceFile, packages)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:           "lodash",
			TargetVersions: []string{"^4.0.0"},
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 4, End: 4},
				Column:   models.Position{Start: 5, End: 23},
				Filename: sourceFile.Path(),
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 4, End: 4},
				Column:   models.Position{Start: 6, End: 12},
				Filename: sourceFile.Path(),
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 4, End: 4},
				Column:   models.Position{Start: 16, End: 22},
				Filename: sourceFile.Path(),
			},
		},
	})
}

func TestPackageJSONMatcher_Match_TransitiveDependencies(t *testing.T) {
	t.Parallel()

	sourceFile, err := lockfile.OpenLocalDepFile("fixtures/package-json/transitive/package.json")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	packages := []lockfile.PackageDetails{
		{
			Name:           "commander",
			TargetVersions: []string{"~2.0.0"},
		},
		{
			Name:           "debug",
			TargetVersions: []string{"^0.7", "~0.7.2"},
		},
		{
			Name:           "jear",
			TargetVersions: []string{"^0.1.4"},
		},
		{
			Name:           "shelljs",
			TargetVersions: []string{"~0.1.4"},
		},
		{
			Name:           "velocityjs",
			TargetVersions: []string{"~0.3.15"},
		},
	}
	err = packageJSONMatcher.Match(sourceFile, packages)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:           "commander",
			TargetVersions: []string{"~2.0.0"},
		},
		{
			Name:           "debug",
			TargetVersions: []string{"^0.7", "~0.7.2"},
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 5, End: 5},
				Column:   models.Position{Start: 5, End: 20},
				Filename: sourceFile.Path(),
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 5, End: 5},
				Column:   models.Position{Start: 6, End: 11},
				Filename: sourceFile.Path(),
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 5, End: 5},
				Column:   models.Position{Start: 15, End: 19},
				Filename: sourceFile.Path(),
			},
		},
		{
			Name:           "jear",
			TargetVersions: []string{"^0.1.4"},
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 4, End: 4},
				Column:   models.Position{Start: 5, End: 21},
				Filename: sourceFile.Path(),
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 4, End: 4},
				Column:   models.Position{Start: 6, End: 10},
				Filename: sourceFile.Path(),
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 4, End: 4},
				Column:   models.Position{Start: 14, End: 20},
				Filename: sourceFile.Path(),
			},
		},
		{
			Name:           "shelljs",
			TargetVersions: []string{"~0.1.4"},
		},
		{
			Name:           "velocityjs",
			TargetVersions: []string{"~0.3.15"},
		},
	})
}

func TestPackageJSONMatcher_Match_NameConflict(t *testing.T) {
	t.Parallel()

	sourceFile, err := lockfile.OpenLocalDepFile("fixtures/package-json/name-conflict/package.json")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	packages := []lockfile.PackageDetails{
		{
			Name:           "aws-sdk-client-mock",
			TargetVersions: []string{"^2.1.1"},
		},
		{
			Name:           "aws-sdk-client-mock-jest",
			TargetVersions: []string{"^2.1.1"},
		},
	}
	err = packageJSONMatcher.Match(sourceFile, packages)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:           "aws-sdk-client-mock",
			TargetVersions: []string{"^2.1.1"},
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 4, End: 4},
				Column:   models.Position{Start: 5, End: 36},
				Filename: sourceFile.Path(),
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 4, End: 4},
				Column:   models.Position{Start: 6, End: 25},
				Filename: sourceFile.Path(),
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 4, End: 4},
				Column:   models.Position{Start: 29, End: 35},
				Filename: sourceFile.Path(),
			},
		},
		{
			Name:           "aws-sdk-client-mock-jest",
			TargetVersions: []string{"^2.1.1"},
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 5, End: 5},
				Column:   models.Position{Start: 5, End: 41},
				Filename: sourceFile.Path(),
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 5, End: 5},
				Column:   models.Position{Start: 6, End: 30},
				Filename: sourceFile.Path(),
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 5, End: 5},
				Column:   models.Position{Start: 34, End: 40},
				Filename: sourceFile.Path(),
			},
		},
	})
}

func TestPackageJSONMatcher_Match_Resolutions(t *testing.T) {
	t.Parallel()

	sourceFile, err := lockfile.OpenLocalDepFile("fixtures/package-json/resolutions/package.json")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	packages := []lockfile.PackageDetails{
		{
			Name:           "fast-xml-parser",
			Version:        "4.2.5",
			TargetVersions: []string{"4.2.5"},
		},
		{
			Name:           "fast-xml-parser",
			Version:        "4.4.0",
			TargetVersions: []string{"^4.2.5"},
		},
		{
			Name:           "@aws-sdk/core",
			Version:        "3.535.0",
			TargetVersions: []string{"^3.535.0"},
		},
	}
	err = packageJSONMatcher.Match(sourceFile, packages)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:           "fast-xml-parser",
			Version:        "4.2.5",
			TargetVersions: []string{"4.2.5"},
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 4, End: 4},
				Column:   models.Position{Start: 5, End: 41},
				Filename: sourceFile.Path(),
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 4, End: 4},
				Column:   models.Position{Start: 6, End: 31},
				Filename: sourceFile.Path(),
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 4, End: 4},
				Column:   models.Position{Start: 35, End: 40},
				Filename: sourceFile.Path(),
			},
		},
		{
			Name:           "fast-xml-parser",
			Version:        "4.4.0",
			TargetVersions: []string{"^4.2.5"},
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 7, End: 7},
				Column:   models.Position{Start: 5, End: 32},
				Filename: sourceFile.Path(),
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 7, End: 7},
				Column:   models.Position{Start: 6, End: 21},
				Filename: sourceFile.Path(),
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 7, End: 7},
				Column:   models.Position{Start: 25, End: 31},
				Filename: sourceFile.Path(),
			},
		},
		{
			Name:           "@aws-sdk/core",
			Version:        "3.535.0",
			TargetVersions: []string{"^3.535.0"},
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 8, End: 8},
				Column:   models.Position{Start: 5, End: 32},
				Filename: sourceFile.Path(),
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 8, End: 8},
				Column:   models.Position{Start: 6, End: 19},
				Filename: sourceFile.Path(),
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 8, End: 8},
				Column:   models.Position{Start: 23, End: 31},
				Filename: sourceFile.Path(),
			},
		},
	})
}
