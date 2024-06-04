package lockfile_test

import (
	"github.com/google/osv-scanner/pkg/models"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/osv-scanner/pkg/lockfile"
)

func TestParsePnpmLock_v9_NoPackages(t *testing.T) {
	t.Parallel()

	packages, err := lockfile.ParsePnpmLock("fixtures/pnpm/no-packages.v9.yaml")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{})
}

func TestParsePnpmLock_v9_OnePackage(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/pnpm/one-package.v9.yaml"))
	packages, err := lockfile.ParsePnpmLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:           "acorn",
			Version:        "8.11.3",
			TargetVersions: []string{"^8.11.3"},
			Ecosystem:      lockfile.PnpmEcosystem,
			CompareAs:      lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 17, End: 20},
					Column:   models.Position{Start: 3, End: 17},
					Filename: path,
				},
			},
		},
	})
}

func TestParsePnpmLock_v9_OnePackageDev(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/pnpm/one-package-dev.v9.yaml"))
	packages, err := lockfile.ParsePnpmLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:           "acorn",
			Version:        "8.11.3",
			TargetVersions: []string{"^8.11.3"},
			Ecosystem:      lockfile.PnpmEcosystem,
			CompareAs:      lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 17, End: 20},
					Column:   models.Position{Start: 3, End: 17},
					Filename: path,
				},
			},
		},
	})
}

func TestParsePnpmLock_v9_ScopedPackages(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/pnpm/scoped-packages.v9.yaml"))
	packages, err := lockfile.ParsePnpmLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:           "@typescript-eslint/types",
			Version:        "5.62.0",
			TargetVersions: []string{"^5.0.0"},
			Ecosystem:      lockfile.PnpmEcosystem,
			CompareAs:      lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 17, End: 19},
					Column:   models.Position{Start: 3, End: 54},
					Filename: path,
				},
			},
		},
	})
}

func TestParsePnpmLock_v9_PeerDependencies(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/pnpm/peer-dependencies.v9.yaml"))
	packages, err := lockfile.ParsePnpmLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:           "acorn-jsx",
			Version:        "5.3.2",
			TargetVersions: []string{"^5.3.2"},
			Ecosystem:      lockfile.PnpmEcosystem,
			CompareAs:      lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 17, End: 20},
					Column:   models.Position{Start: 3, End: 40},
					Filename: path,
				},
			},
		},
		{
			Name:      "acorn",
			Version:   "8.11.3",
			Ecosystem: lockfile.PnpmEcosystem,
			CompareAs: lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 22, End: 25},
					Column:   models.Position{Start: 3, End: 17},
					Filename: path,
				},
			},
		},
	})
}

func TestParsePnpmLock_v9_PeerDependenciesAdvanced(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/pnpm/peer-dependencies-advanced.v9.yaml"))
	packages, err := lockfile.ParsePnpmLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:      "@eslint-community/eslint-utils",
			Version:   "4.4.0",
			Ecosystem: lockfile.PnpmEcosystem,
			CompareAs: lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 26, End: 30},
					Column:   models.Position{Start: 3, End: 42},
					Filename: path,
				},
			},
		},
		{
			Name:      "@eslint/eslintrc",
			Version:   "2.1.4",
			Ecosystem: lockfile.PnpmEcosystem,
			CompareAs: lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 32, End: 34},
					Column:   models.Position{Start: 3, End: 54},
					Filename: path,
				},
			},
		},
		{
			Name:           "@typescript-eslint/eslint-plugin",
			Version:        "5.62.0",
			TargetVersions: []string{"^5.12.0"},
			Ecosystem:      lockfile.PnpmEcosystem,
			CompareAs:      lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 36, End: 45},
					Column:   models.Position{Start: 3, End: 23},
					Filename: path,
				},
			},
		},
		{
			Name:           "@typescript-eslint/parser",
			Version:        "5.62.0",
			TargetVersions: []string{"^5.12.0"},
			Ecosystem:      lockfile.PnpmEcosystem,
			CompareAs:      lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 47, End: 55},
					Column:   models.Position{Start: 3, End: 23},
					Filename: path,
				},
			},
		},
		{
			Name:      "@typescript-eslint/type-utils",
			Version:   "5.62.0",
			Ecosystem: lockfile.PnpmEcosystem,
			CompareAs: lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 57, End: 65},
					Column:   models.Position{Start: 3, End: 23},
					Filename: path,
				},
			},
		},
		{
			Name:      "@typescript-eslint/typescript-estree",
			Version:   "5.62.0",
			Ecosystem: lockfile.PnpmEcosystem,
			CompareAs: lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 67, End: 74},
					Column:   models.Position{Start: 3, End: 23},
					Filename: path,
				},
			},
		},
		{
			Name:      "@typescript-eslint/utils",
			Version:   "5.62.0",
			Ecosystem: lockfile.PnpmEcosystem,
			CompareAs: lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 76, End: 80},
					Column:   models.Position{Start: 3, End: 41},
					Filename: path,
				},
			},
		},
		{
			Name:      "debug",
			Version:   "4.3.4",
			Ecosystem: lockfile.PnpmEcosystem,
			CompareAs: lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 82, End: 89},
					Column:   models.Position{Start: 3, End: 23},
					Filename: path,
				},
			},
		},
		{
			Name:           "eslint",
			Version:        "8.57.0",
			TargetVersions: []string{"^8.0.0"},
			Ecosystem:      lockfile.PnpmEcosystem,
			CompareAs:      lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 91, End: 94},
					Column:   models.Position{Start: 3, End: 17},
					Filename: path,
				},
			},
		},
		{
			Name:      "has-flag",
			Version:   "4.0.0",
			Ecosystem: lockfile.PnpmEcosystem,
			CompareAs: lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 96, End: 98},
					Column:   models.Position{Start: 3, End: 27},
					Filename: path,
				},
			},
		},
		{
			Name:      "supports-color",
			Version:   "7.2.0",
			Ecosystem: lockfile.PnpmEcosystem,
			CompareAs: lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 100, End: 102},
					Column:   models.Position{Start: 3, End: 27},
					Filename: path,
				},
			},
		},
		{
			Name:      "tsutils",
			Version:   "3.21.0",
			Ecosystem: lockfile.PnpmEcosystem,
			CompareAs: lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 104, End: 108},
					Column:   models.Position{Start: 3, End: 158},
					Filename: path,
				},
			},
		},
		{
			Name:           "typescript",
			Version:        "4.9.5",
			TargetVersions: []string{"^4.0.0"},
			Ecosystem:      lockfile.PnpmEcosystem,
			CompareAs:      lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 110, End: 113},
					Column:   models.Position{Start: 3, End: 17},
					Filename: path,
				},
			},
		},
	})
}

func TestParsePnpmLock_v9_MultipleVersions(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/pnpm/multiple-versions.v9.yaml"))
	packages, err := lockfile.ParsePnpmLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:      "uuid",
			Version:   "8.0.0",
			Ecosystem: lockfile.PnpmEcosystem,
			CompareAs: lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 20, End: 22},
					Column:   models.Position{Start: 3, End: 17},
					Filename: path,
				},
			},
		},
		{
			Name:           "uuid",
			Version:        "8.3.2",
			TargetVersions: []string{"^8.0.0"},
			Ecosystem:      lockfile.PnpmEcosystem,
			CompareAs:      lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 24, End: 26},
					Column:   models.Position{Start: 3, End: 17},
					Filename: path,
				},
			},
		},
		{
			Name:      "xmlbuilder",
			Version:   "11.0.1",
			Ecosystem: lockfile.PnpmEcosystem,
			CompareAs: lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 28, End: 30},
					Column:   models.Position{Start: 3, End: 29},
					Filename: path,
				},
			},
		},
	})
}

func TestParsePnpmLock_v9_Commits(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/pnpm/commits.v9.yaml"))
	packages, err := lockfile.ParsePnpmLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:           "ansi-regex",
			Version:        "6.0.1",
			TargetVersions: []string{"git@github.com/chalk/ansi-regex.git"},
			Ecosystem:      lockfile.PnpmEcosystem,
			CompareAs:      lockfile.PnpmEcosystem,
			Commit:         "02fa893d619d3da85411acc8fd4e2eea0e95a9d9",
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 20, End: 23},
					Column:   models.Position{Start: 3, End: 28},
					Filename: path,
				},
			},
		},
		{
			Name:           "is-number",
			Version:        "7.0.0",
			TargetVersions: []string{"github:jonschlinkert/is-number#master"},
			Ecosystem:      lockfile.PnpmEcosystem,
			CompareAs:      lockfile.PnpmEcosystem,
			Commit:         "98e8ff1da1a89f93d1397a24d7413ed15421c139",
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 25, End: 28},
					Column:   models.Position{Start: 3, End: 32},
					Filename: path,
				},
			},
		},
	})
}

func TestParsePnpmLock_v9_MixedGroups(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/pnpm/mixed-groups.v9.yaml"))
	packages, err := lockfile.ParsePnpmLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:           "ansi-regex",
			Version:        "5.0.1",
			TargetVersions: []string{"^5.0.0"},
			Ecosystem:      lockfile.PnpmEcosystem,
			CompareAs:      lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 25, End: 27},
					Column:   models.Position{Start: 3, End: 27},
					Filename: path,
				},
			},
		},
		{
			Name:           "uuid",
			Version:        "8.3.2",
			TargetVersions: []string{"^8.0.0"},
			Ecosystem:      lockfile.PnpmEcosystem,
			CompareAs:      lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 33, End: 35},
					Column:   models.Position{Start: 3, End: 17},
					Filename: path,
				},
			},
		},
		{
			Name:           "is-number",
			Version:        "7.0.0",
			TargetVersions: []string{"^7.0.0"},
			Ecosystem:      lockfile.PnpmEcosystem,
			CompareAs:      lockfile.PnpmEcosystem,
			LockfileLocations: lockfile.Locations{
				Block: models.FilePosition{
					Line:     models.Position{Start: 29, End: 31},
					Column:   models.Position{Start: 3, End: 32},
					Filename: path,
				},
			},
		},
	})
}
