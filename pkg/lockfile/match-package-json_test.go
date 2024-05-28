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

var matcher = lockfile.PackageJSONMatcher{}

func TestPackageJSONMatcher_GetSourceFile_FileDoesNotExist(t *testing.T) {
	t.Parallel()

	lockFile, err := lockfile.OpenLocalDepFile("fixtures/package-json/does-not-exist/npm-v1.json")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	sourceFile, err := matcher.GetSourceFile(lockFile)
	expectErrIs(t, err, fs.ErrNotExist)
	assert.Equal(t, "", sourceFile.Path())
}

func TestPackageJSONMatcher_GetSourceFile(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/package-json/one-package"))
	lockFile, err := lockfile.OpenLocalDepFile(path + "/npm-v1.json")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	sourceFile, err := matcher.GetSourceFile(lockFile)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	assert.Equal(t, path+"/package.json", sourceFile.Path())
}

func TestPackageJSONMatcher_Match_OnePackage(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/package-json/one-package"))
	sourceFile, err := lockfile.OpenLocalDepFile(path + "/package.json")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	packages := []lockfile.PackageDetails{
		{
			Name:           "lodash",
			TargetVersions: []string{"^4.0.0"},
		},
	}
	err = matcher.Match(sourceFile, packages)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:           "lodash",
			TargetVersions: []string{"^4.0.0"},
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 11, End: 11},
				Column:   models.Position{Start: 5, End: 23},
				Filename: sourceFile.Path(),
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 11, End: 11},
				Column:   models.Position{Start: 6, End: 12},
				Filename: sourceFile.Path(),
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 11, End: 11},
				Column:   models.Position{Start: 16, End: 22},
				Filename: sourceFile.Path(),
			},
		},
	})
}

func TestPackageJSONMatcher_Match_TransitiveDependencies(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/package-json/transitive-dependencies"))
	sourceFile, err := lockfile.OpenLocalDepFile(path + "/package.json")
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
	err = matcher.Match(sourceFile, packages)
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
				Line:     models.Position{Start: 4, End: 4},
				Column:   models.Position{Start: 5, End: 20},
				Filename: sourceFile.Path(),
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 4, End: 4},
				Column:   models.Position{Start: 6, End: 11},
				Filename: sourceFile.Path(),
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 4, End: 4},
				Column:   models.Position{Start: 15, End: 19},
				Filename: sourceFile.Path(),
			},
		},
		{
			Name:           "jear",
			TargetVersions: []string{"^0.1.4"},
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 5, End: 5},
				Column:   models.Position{Start: 5, End: 21},
				Filename: sourceFile.Path(),
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 5, End: 5},
				Column:   models.Position{Start: 6, End: 10},
				Filename: sourceFile.Path(),
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 5, End: 5},
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
