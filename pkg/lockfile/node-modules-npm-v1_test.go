package lockfile_test

import (
	"testing"

	"github.com/datadog/osv-scanner/pkg/models"

	"github.com/datadog/osv-scanner/pkg/lockfile"
)

func TestNodeModulesExtractor_Extract_npm_v1_InvalidJson(t *testing.T) {
	t.Parallel()

	packages, _, err := testParsingNodeModules(t, "fixtures/npm/not-json.txt")

	expectErrContaining(t, err, "could not extract from")
	expectPackages(t, packages, []lockfile.PackageDetails{})
}

func TestNodeModulesExtractor_Extract_npm_v1_NoPackages(t *testing.T) {
	t.Parallel()

	packages, _, err := testParsingNodeModules(t, "fixtures/npm/empty.v1.json")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{})
}

func TestNodeModulesExtractor_Extract_npm_v1_OnePackage(t *testing.T) {
	t.Parallel()

	packages, filePath, err := testParsingNodeModules(t, "fixtures/npm/one-package.v1.json")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:           "wrappy",
			Version:        "1.0.2",
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			PackageManager: models.NPM,
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 5, End: 9},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			IsDirect: true,
		},
	})
}

func TestNodeModulesExtractor_Extract_npm_v1_OnePackageDev(t *testing.T) {
	t.Parallel()

	packages, filePath, err := testParsingNodeModules(t, "fixtures/npm/one-package-dev.v1.json")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:           "wrappy",
			Version:        "1.0.2",
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			PackageManager: models.NPM,
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 5, End: 10},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			DepGroups: []string{"dev"},
			IsDirect:  true,
		},
	})
}

func TestNodeModulesExtractor_Extract_npm_v1_TwoPackages(t *testing.T) {
	t.Parallel()

	packages, filePath, err := testParsingNodeModules(t, "fixtures/npm/two-packages.v1.json")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:           "wrappy",
			Version:        "1.0.2",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 5, End: 9},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			IsDirect: true,
		},
		{
			Name:           "supports-color",
			Version:        "5.5.0",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 10, End: 17},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			IsDirect: true,
		},
	})
}

func TestNodeModulesExtractor_Extract_npm_v1_ScopedPackages(t *testing.T) {
	t.Parallel()

	packages, filePath, err := testParsingNodeModules(t, "fixtures/npm/scoped-packages.v1.json")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:           "wrappy",
			Version:        "1.0.2",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 13, End: 17},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			IsDirect: true,
		},
		{
			Name:           "@babel/code-frame",
			Version:        "7.0.0",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 5, End: 12},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			IsDirect: true,
		},
	})
}

func TestNodeModulesExtractor_Extract_npm_v1_NestedDependencies(t *testing.T) {
	t.Parallel()

	packages, filePath, err := testParsingNodeModules(t, "fixtures/npm/nested-dependencies.v1.json")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:           "postcss",
			Version:        "6.0.23",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 5, End: 14},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			IsDirect: true,
		},
		{
			Name:           "postcss",
			Version:        "7.0.16",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 26, End: 35},
				Column:   models.Position{Start: 9, End: 10},
				Filename: filePath,
			},
			IsDirect: true,
		},
		{
			Name:           "postcss-calc",
			Version:        "7.0.1",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 15, End: 45},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			IsDirect: true,
		},
		{
			Name:           "supports-color",
			Version:        "6.1.0",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 36, End: 43},
				Column:   models.Position{Start: 9, End: 10},
				Filename: filePath,
			},
			IsDirect: true,
		},
		{
			Name:           "supports-color",
			Version:        "5.5.0",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 46, End: 53},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			IsDirect: true,
		},
	})
}

func TestNodeModulesExtractor_Extract_npm_v1_NestedDependenciesDup(t *testing.T) {
	t.Parallel()

	packages, filePath, err := testParsingNodeModules(t, "fixtures/npm/nested-dependencies-dup.v1.json")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	// todo: convert to using expectPackages w/ listing all expected packages
	if len(packages) != 39 {
		t.Errorf("Expected to get 39 packages, but got %d", len(packages))
	}

	expectPackage(t, packages, lockfile.PackageDetails{
		Name:           "supports-color",
		Version:        "6.1.0",
		PackageManager: models.NPM,
		Ecosystem:      lockfile.NpmEcosystem,
		CompareAs:      lockfile.NpmEcosystem,
		BlockLocation: models.FilePosition{
			Line:     models.Position{Start: 749, End: 756},
			Column:   models.Position{Start: 9, End: 10},
			Filename: filePath,
		},
		IsDirect: true,
	})

	expectPackage(t, packages, lockfile.PackageDetails{
		Name:           "supports-color",
		Version:        "5.5.0",
		PackageManager: models.NPM,
		Ecosystem:      lockfile.NpmEcosystem,
		CompareAs:      lockfile.NpmEcosystem,
		BlockLocation: models.FilePosition{
			Line:     models.Position{Start: 759, End: 766},
			Column:   models.Position{Start: 5, End: 6},
			Filename: filePath,
		},
		IsDirect: true,
	})

	expectPackage(t, packages, lockfile.PackageDetails{
		Name:           "supports-color",
		Version:        "2.0.0",
		PackageManager: models.NPM,
		Ecosystem:      lockfile.NpmEcosystem,
		CompareAs:      lockfile.NpmEcosystem,
		BlockLocation: models.FilePosition{
			Line:     models.Position{Start: 186, End: 190},
			Column:   models.Position{Start: 9, End: 10},
			Filename: filePath,
		},
		IsDirect: true,
	})
}

func TestNodeModulesExtractor_Extract_npm_v1_Commits(t *testing.T) {
	t.Parallel()

	packages, filePath, err := testParsingNodeModules(t, "fixtures/npm/commits.v1.json")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:           "@segment/analytics.js-integration-facebook-pixel",
			Version:        "",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			Commit:         "3b1bb80b302c2e552685dc8a029797ec832ea7c9",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 5, End: 18},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			IsDirect: true,
		},
		{
			Name:           "ansi-styles",
			Version:        "1.0.0",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			Commit:         "",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 19, End: 23},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			IsDirect: true,
		},
		{
			Name:           "babel-preset-php",
			Version:        "",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			Commit:         "c5a7ba5e0ad98b8db1cb8ce105403dd4b768cced",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 24, End: 30},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			IsDirect: true,
		},
		{
			Name:           "is-number-1",
			Version:        "",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			Commit:         "af885e2e890b9ef0875edd2b117305119ee5bdc5",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 31, End: 37},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			DepGroups: []string{"dev"},
			IsDirect:  true,
		},
		{
			Name:           "is-number-1",
			Version:        "",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			Commit:         "be5935f8d2595bcd97b05718ef1eeae08d812e10",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 75, End: 81},
				Column:   models.Position{Start: 9, End: 10},
				Filename: filePath,
			},
			DepGroups: []string{"dev"},
			IsDirect:  true,
		},
		{
			Name:           "is-number-2",
			Version:        "",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			Commit:         "d5ac0584ee9ae7bd9288220a39780f155b9ad4c8",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 38, End: 41},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			IsDirect: true,
		},
		{
			Name:           "is-number-2",
			Version:        "",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			Commit:         "82dcc8e914dabd9305ab9ae580709a7825e824f5",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 82, End: 85},
				Column:   models.Position{Start: 9, End: 10},
				Filename: filePath,
			},
			IsDirect: true,
		},
		{
			Name:           "is-number-3",
			Version:        "",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			Commit:         "d5ac0584ee9ae7bd9288220a39780f155b9ad4c8",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 42, End: 46},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			DepGroups: []string{"dev"},
			IsDirect:  true,
		},
		{
			Name:           "is-number-3",
			Version:        "",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			Commit:         "82ae8802978da40d7f1be5ad5943c9e550ab2c89",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 86, End: 90},
				Column:   models.Position{Start: 9, End: 10},
				Filename: filePath,
			},
			DepGroups: []string{"dev"},
			IsDirect:  true,
		},
		{
			Name:           "is-number-4",
			Version:        "",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			Commit:         "af885e2e890b9ef0875edd2b117305119ee5bdc5",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 47, End: 54},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			DepGroups: []string{"dev"},
			IsDirect:  true,
		},
		{
			Name:           "is-number-5",
			Version:        "",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			Commit:         "af885e2e890b9ef0875edd2b117305119ee5bdc5",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 55, End: 62},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			DepGroups: []string{"dev"},
			IsDirect:  true,
		},
		{
			Name:           "is-number-6",
			Version:        "",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			Commit:         "af885e2e890b9ef0875edd2b117305119ee5bdc5",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 63, End: 69},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			DepGroups: []string{"dev"},
			IsDirect:  true,
		},
		{
			Name:           "postcss-calc",
			Version:        "7.0.1",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			Commit:         "",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 70, End: 92},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			IsDirect: true,
		},
		{
			Name:           "raven-js",
			Version:        "",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			Commit:         "c2b377e7a254264fd4a1fe328e4e3cfc9e245570",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 93, End: 96},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			IsDirect: true,
		},
		{
			Name:           "slick-carousel",
			Version:        "",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			Commit:         "280b560161b751ba226d50c7db1e0a14a78c2de0",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 97, End: 101},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			DepGroups: []string{"dev"},
			IsDirect:  true,
		},
	})
}

func TestNodeModulesExtractor_Extract_npm_v1_Files(t *testing.T) {
	t.Parallel()

	packages, filePath, err := testParsingNodeModules(t, "fixtures/npm/files.v1.json")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackagesWithoutLocations(t, packages, []lockfile.PackageDetails{
		{
			Name:           "lodash",
			Version:        "1.3.1",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			Commit:         "",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 5, End: 9},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			IsDirect: true,
		},
		{
			Name:           "other_package",
			Version:        "",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			Commit:         "",
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 10, End: 15},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			IsDirect: true,
		},
	})
}

func TestNodeModulesExtractor_Extract_npm_v1_Alias(t *testing.T) {
	t.Parallel()

	packages, filePath, err := testParsingNodeModules(t, "fixtures/npm/alias.v1.json")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackagesWithoutLocations(t, packages, []lockfile.PackageDetails{
		{
			Name:           "@babel/code-frame",
			Version:        "7.0.0",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 5, End: 12},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			IsDirect: true,
		},
		{
			Name:           "string-width",
			Version:        "4.2.0",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 23, End: 32},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			IsDirect: true,
		},
		{
			Name:           "string-width",
			Version:        "5.1.2",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 13, End: 22},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			IsDirect: true,
		},
	})
}

func TestNodeModulesExtractor_Extract_npm_v1_OptionalPackage(t *testing.T) {
	t.Parallel()

	packages, filePath, err := testParsingNodeModules(t, "fixtures/npm/optional-package.v1.json")
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackagesWithoutLocations(t, packages, []lockfile.PackageDetails{
		{
			Name:           "wrappy",
			Version:        "1.0.2",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 5, End: 11},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			DepGroups: []string{"dev", "optional"},
			IsDirect:  true,
		},
		{
			Name:           "supports-color",
			Version:        "5.5.0",
			PackageManager: models.NPM,
			Ecosystem:      lockfile.NpmEcosystem,
			CompareAs:      lockfile.NpmEcosystem,
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 12, End: 20},
				Column:   models.Position{Start: 5, End: 6},
				Filename: filePath,
			},
			DepGroups: []string{"optional"},
			IsDirect:  true,
		},
	})
}
