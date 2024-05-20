package lockfile_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/osv-scanner/pkg/models"

	"github.com/google/osv-scanner/pkg/lockfile"
)

func TestPoetryLockExtractor_ShouldExtract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "",
			path: "",
			want: false,
		},
		{
			name: "",
			path: "poetry.lock",
			want: true,
		},
		{
			name: "",
			path: "path/to/my/poetry.lock",
			want: true,
		},
		{
			name: "",
			path: "path/to/my/poetry.lock/file",
			want: false,
		},
		{
			name: "",
			path: "path/to/my/poetry.lock.file",
			want: false,
		},
		{
			name: "",
			path: "path.to.my.poetry.lock",
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := lockfile.PoetryLockExtractor{}
			got := e.ShouldExtract(tt.path)
			if got != tt.want {
				t.Errorf("Extract() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePoetryLock_FileDoesNotExist(t *testing.T) {
	t.Parallel()

	packages, err := lockfile.ParsePoetryLock("fixtures/poetry/does-not-exist")

	expectErrIs(t, err, fs.ErrNotExist)
	expectPackages(t, packages, []lockfile.PackageDetails{})
}

func TestParsePoetryLock_InvalidToml(t *testing.T) {
	t.Parallel()

	packages, err := lockfile.ParsePoetryLock("fixtures/poetry/not-toml.txt")

	expectErrContaining(t, err, "could not decode toml from")
	expectPackages(t, packages, []lockfile.PackageDetails{})
}

func TestParsePoetryLock_NoPackages(t *testing.T) {
	t.Parallel()

	packages, err := lockfile.ParsePoetryLock("fixtures/poetry/empty.lock")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{})
}

func TestParsePoetryLock_OnePackage(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/poetry/one-package.lock"))
	packages, err := lockfile.ParsePoetryLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:    "numpy",
			Version: "1.23.3",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 1, End: 7},
				Column:   models.Position{Start: 1, End: 26},
				Filename: path,
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 2, End: 2},
				Column:   models.Position{Start: 9, End: 14},
				Filename: path,
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 3, End: 3},
				Column:   models.Position{Start: 12, End: 18},
				Filename: path,
			},
			Ecosystem: lockfile.PoetryEcosystem,
			CompareAs: lockfile.PoetryEcosystem,
		},
	})
}

func TestParsePoetryLock_TwoPackages(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/poetry/two-packages.lock"))
	packages, err := lockfile.ParsePoetryLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:    "proto-plus",
			Version: "1.22.0",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 1, End: 13},
				Column:   models.Position{Start: 1, End: 47},
				Filename: path,
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 2, End: 2},
				Column:   models.Position{Start: 9, End: 19},
				Filename: path,
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 3, End: 3},
				Column:   models.Position{Start: 12, End: 18},
				Filename: path,
			},
			Ecosystem: lockfile.PoetryEcosystem,
			CompareAs: lockfile.PoetryEcosystem,
		},
		{
			Name:    "protobuf",
			Version: "4.21.5",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 15, End: 21},
				Column:   models.Position{Start: 1, End: 26},
				Filename: path,
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 16, End: 16},
				Column:   models.Position{Start: 9, End: 17},
				Filename: path,
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 17, End: 17},
				Column:   models.Position{Start: 12, End: 18},
				Filename: path,
			},
			Ecosystem: lockfile.PoetryEcosystem,
			CompareAs: lockfile.PoetryEcosystem,
		},
	})
}

func TestParsePoetryLock_PackageWithMetadata(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/poetry/one-package-with-metadata.lock"))
	packages, err := lockfile.ParsePoetryLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:    "emoji",
			Version: "2.0.0",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 1, End: 10},
				Column:   models.Position{Start: 1, End: 42},
				Filename: path,
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 2, End: 2},
				Column:   models.Position{Start: 9, End: 14},
				Filename: path,
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 3, End: 3},
				Column:   models.Position{Start: 12, End: 17},
				Filename: path,
			},
			Ecosystem: lockfile.PoetryEcosystem,
			CompareAs: lockfile.PoetryEcosystem,
		},
	})
}

func TestParsePoetryLock_PackageWithGitSource(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/poetry/source-git.lock"))
	packages, err := lockfile.ParsePoetryLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:    "ike",
			Version: "0.2.0",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 1, End: 14},
				Column:   models.Position{Start: 1, End: 64},
				Filename: path,
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 2, End: 2},
				Column:   models.Position{Start: 9, End: 12},
				Filename: path,
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 3, End: 3},
				Column:   models.Position{Start: 12, End: 17},
				Filename: path,
			},
			Ecosystem: lockfile.PoetryEcosystem,
			CompareAs: lockfile.PoetryEcosystem,
			Commit:    "cd66602cd29f61a2d2e7fb995fef1e61708c034d",
		},
	})
}

func TestParsePoetryLock_PackageWithLegacySource(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/poetry/source-legacy.lock"))
	packages, err := lockfile.ParsePoetryLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:    "appdirs",
			Version: "1.4.4",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 1, End: 12},
				Column:   models.Position{Start: 1, End: 23},
				Filename: path,
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 2, End: 2},
				Column:   models.Position{Start: 9, End: 16},
				Filename: path,
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 3, End: 3},
				Column:   models.Position{Start: 12, End: 17},
				Filename: path,
			},
			Ecosystem: lockfile.PoetryEcosystem,
			CompareAs: lockfile.PoetryEcosystem,
			Commit:    "",
		},
	})
}

func TestParsePoetryLock_OptionalPackage(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/poetry/optional-package.lock"))
	packages, err := lockfile.ParsePoetryLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:    "numpy",
			Version: "1.23.3",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 1, End: 7},
				Column:   models.Position{Start: 1, End: 26},
				Filename: path,
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 2, End: 2},
				Column:   models.Position{Start: 9, End: 14},
				Filename: path,
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 3, End: 3},
				Column:   models.Position{Start: 12, End: 18},
				Filename: path,
			},
			Ecosystem: lockfile.PoetryEcosystem,
			CompareAs: lockfile.PoetryEcosystem,
			DepGroups: []string{"optional"},
		},
	})
}
