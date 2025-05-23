package lockfile_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/osv-scanner/internal/testutility"

	"github.com/google/osv-scanner/pkg/lockfile"
	"github.com/google/osv-scanner/pkg/models"
	"github.com/stretchr/testify/assert"
)

var gemfileMatcher = lockfile.GemfileMatcher{}

func TestGemfileMatcher_GetSourceFile_FileDoesNotExist(t *testing.T) {
	t.Parallel()

	lockFile, err := lockfile.OpenLocalDepFile("fixtures/bundler/no-gemfile/Gemfile.lock")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	sourceFile, err := gemfileMatcher.GetSourceFile(lockFile)
	expectErrIs(t, err, fs.ErrNotExist)
	assert.Equal(t, "", sourceFile.Path())
}

func TestGemfileMatcher_GetSourceFile(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	basePath := "fixtures/bundler/one-package/"
	sourcefilePath := filepath.FromSlash(filepath.Join(dir, basePath+"Gemfile"))

	lockFile, err := lockfile.OpenLocalDepFile(basePath + "Gemfile.lock")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	sourceFile, err := gemfileMatcher.GetSourceFile(lockFile)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	assert.Equal(t, sourcefilePath, sourceFile.Path())
}

func TestGemfileMatcher_Match_OnePackage(t *testing.T) {
	t.Parallel()

	sourceFile, err := lockfile.OpenLocalDepFile("fixtures/bundler/one-package/Gemfile")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	packages := []lockfile.PackageDetails{
		{
			Name:           "RedCloth",
			Version:        "4.2.9",
			PackageManager: models.Bundler,
		},
	}
	err = gemfileMatcher.Match(sourceFile, packages)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:           "RedCloth",
			Version:        "4.2.9",
			PackageManager: models.Bundler,
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 3, End: 5},
				Column:   models.Position{Start: 1, End: 24},
				Filename: sourceFile.Path(),
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 3, End: 3},
				Column:   models.Position{Start: 5, End: 15},
				Filename: sourceFile.Path(),
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 4, End: 4},
				Column:   models.Position{Start: 2, End: 12},
				Filename: sourceFile.Path(),
			},
		},
	})
}

func TestGemfileMatcher_Match_Groups(t *testing.T) {
	t.Parallel()

	sourceFile, err := lockfile.OpenLocalDepFile("fixtures/bundler/groups/Gemfile")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	packages := []lockfile.PackageDetails{
		{
			Name:           "base64",
			Version:        "0.2.0",
			PackageManager: models.Bundler,
		},
		{
			Name:           "thor",
			Version:        "1.3.2",
			PackageManager: models.Bundler,
		},
		{
			Name:           "timeout",
			Version:        "0.4.3",
			PackageManager: models.Bundler,
		},
		{
			Name:           "useragent",
			Version:        "0.16.11",
			PackageManager: models.Bundler,
		},
		{
			Name:           "websocket-driver",
			Version:        "0.7.7",
			PackageManager: models.Bundler,
		},
		{
			Name:           "websocket-extensions",
			Version:        "0.1.5",
			PackageManager: models.Bundler,
		},
		{
			Name:           "zeitwerk",
			Version:        "2.7.1",
			PackageManager: models.Bundler,
		},
	}

	err = gemfileMatcher.Match(sourceFile, packages)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	testutility.NewSnapshot().WithJSONNormalization().MatchJSON(t, packages)
}

func TestGemfileMatcher_Filter_Not_In_Lockfile(t *testing.T) {
	t.Parallel()

	sourceFile, err := lockfile.OpenLocalDepFile("fixtures/bundler/lockfile-not-synced/Gemfile")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	packages := []lockfile.PackageDetails{
		{
			Name:           "zeitwerk",
			Version:        "2.6.0",
			PackageManager: models.Bundler,
		},
	}

	err = gemfileMatcher.Match(sourceFile, packages)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	testutility.NewSnapshot().WithJSONNormalization().MatchJSON(t, packages)
}
