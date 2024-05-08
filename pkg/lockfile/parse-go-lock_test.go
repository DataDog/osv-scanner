package lockfile_test

import (
	"fmt"
	"io/fs"
	"testing"

	"github.com/google/osv-scanner/pkg/models"

	"github.com/google/osv-scanner/pkg/lockfile"
)

func TestGoLockExtractor_ShouldExtract(t *testing.T) {
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
			path: "go.mod",
			want: true,
		},
		{
			name: "",
			path: "path/to/my/go.mod",
			want: true,
		},
		{
			name: "",
			path: "path/to/my/go.mod/file",
			want: false,
		},
		{
			name: "",
			path: "path/to/my/go.mod.file",
			want: false,
		},
		{
			name: "",
			path: "path.to.my.go.mod",
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := lockfile.GoLockExtractor{}
			got := e.ShouldExtract(tt.path)
			if got != tt.want {
				t.Errorf("Extract() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseGoLock_FileDoesNotExist(t *testing.T) {
	t.Parallel()

	packages, err := lockfile.ParseGoLock("fixtures/go/does-not-exist")

	expectErrIs(t, err, fs.ErrNotExist)
	expectPackages(t, packages, []lockfile.PackageDetails{})
}

func TestParseGoLock_Invalid(t *testing.T) {
	t.Parallel()

	packages, err := lockfile.ParseGoLock("fixtures/go/not-go-mod.txt")

	expectErrContaining(t, err, "could not extract from")
	expectPackages(t, packages, []lockfile.PackageDetails{})
}

func TestParseGoLock_NoPackages(t *testing.T) {
	t.Parallel()

	packages, err := lockfile.ParseGoLock("fixtures/go/empty.mod")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{})
}

func TestParseGoLock_WithPathMajor(t *testing.T) {
	t.Parallel()

	packages, err := lockfile.ParseGoLock("fixtures/go/with-path-major.mod")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	for _, p := range packages {
		fmt.Println(p.VersionLocation)
		fmt.Println(p.NameLocation)
	}
	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:      "github.com/elastic/go-elasticsearch/v8",
			Version:   "8",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
			BlockLocation: models.FilePosition{
				Line:   models.Position{Start: 2, End: 2},
				Column: models.Position{Start: 1, End: 54},
			},
			VersionLocation: &models.FilePosition{
				Line:   models.Position{Start: 2, End: 2},
				Column: models.Position{Start: 46, End: 47},
			},
			NameLocation: &models.FilePosition{
				Line:   models.Position{Start: 2, End: 2},
				Column: models.Position{Start: 9, End: 47},
			},
		},
		{
			Name:      "stdlib",
			Version:   "1.11.0",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
			BlockLocation: models.FilePosition{
				Line:   models.Position{Start: 0, End: 0},
				Column: models.Position{Start: 0, End: 0},
			},
		},
	})
}

func TestParseGoLock_WithoutSupportedVersioning(t *testing.T) {
	t.Parallel()

	packages, err := lockfile.ParseGoLock("fixtures/go/without-supported-versioning.mod")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:      "github.com/elastic/go-elasticsearch",
			Version:   "0.0.0",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
			BlockLocation: models.FilePosition{
				Line:   models.Position{Start: 2, End: 2},
				Column: models.Position{Start: 1, End: 51},
			},
			NameLocation: &models.FilePosition{
				Line:   models.Position{Start: 2, End: 2},
				Column: models.Position{Start: 9, End: 44},
			},
		},
		{
			Name:      "stdlib",
			Version:   "1.11.0",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
			BlockLocation: models.FilePosition{
				Line:   models.Position{Start: 0, End: 0},
				Column: models.Position{Start: 0, End: 0},
			},
		},
	})
}

func TestParseGoLock_OnePackage(t *testing.T) {
	t.Parallel()

	packages, err := lockfile.ParseGoLock("fixtures/go/one-package.mod")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:      "github.com/BurntSushi/toml",
			Version:   "1.0.0",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
			BlockLocation: models.FilePosition{
				Line:   models.Position{Start: 4, End: 4},
				Column: models.Position{Start: 2, End: 35},
			},
			VersionLocation: &models.FilePosition{
				Line:   models.Position{Start: 4, End: 4},
				Column: models.Position{Start: 30, End: 35},
			},
			NameLocation: &models.FilePosition{
				Line:   models.Position{Start: 4, End: 4},
				Column: models.Position{Start: 2, End: 28},
			},
		},
	})
}

func TestParseGoLock_TwoPackages(t *testing.T) {
	t.Parallel()

	packages, err := lockfile.ParseGoLock("fixtures/go/two-packages.mod")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:      "github.com/BurntSushi/toml",
			Version:   "1.0.0",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
			BlockLocation: models.FilePosition{
				Line:   models.Position{Start: 6, End: 6},
				Column: models.Position{Start: 2, End: 35},
			},
			VersionLocation: &models.FilePosition{
				Line:   models.Position{Start: 6, End: 6},
				Column: models.Position{Start: 30, End: 35},
			},
			NameLocation: &models.FilePosition{
				Line:   models.Position{Start: 6, End: 6},
				Column: models.Position{Start: 2, End: 28},
			},
		},
		{
			Name:      "gopkg.in/yaml.v2",
			Version:   "2.4.0",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
			BlockLocation: models.FilePosition{
				Line:   models.Position{Start: 7, End: 7},
				Column: models.Position{Start: 2, End: 25},
			},
			VersionLocation: &models.FilePosition{
				Line:   models.Position{Start: 7, End: 7},
				Column: models.Position{Start: 20, End: 25},
			},
			NameLocation: &models.FilePosition{
				Line:   models.Position{Start: 7, End: 7},
				Column: models.Position{Start: 2, End: 18},
			},
		},
		{
			Name:      "stdlib",
			Version:   "1.17.0",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
		},
	})
}

func TestParseGoLock_IndirectPackages(t *testing.T) {
	t.Parallel()

	packages, err := lockfile.ParseGoLock("fixtures/go/indirect-packages.mod")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:      "github.com/BurntSushi/toml",
			Version:   "1.0.0",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
			BlockLocation: models.FilePosition{
				Line:   models.Position{Start: 6, End: 6},
				Column: models.Position{Start: 2, End: 35},
			},
			VersionLocation: &models.FilePosition{
				Line:   models.Position{Start: 6, End: 6},
				Column: models.Position{Start: 30, End: 35},
			},
			NameLocation: &models.FilePosition{
				Line:   models.Position{Start: 6, End: 6},
				Column: models.Position{Start: 2, End: 28},
			},
		},
		{
			Name:      "gopkg.in/yaml.v2",
			Version:   "2.4.0",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
			BlockLocation: models.FilePosition{
				Line:   models.Position{Start: 7, End: 7},
				Column: models.Position{Start: 2, End: 25},
			},
			VersionLocation: &models.FilePosition{
				Line:   models.Position{Start: 7, End: 7},
				Column: models.Position{Start: 20, End: 25},
			},
			NameLocation: &models.FilePosition{
				Line:   models.Position{Start: 7, End: 7},
				Column: models.Position{Start: 2, End: 18},
			},
		},
		{
			Name:      "github.com/mattn/go-colorable",
			Version:   "0.1.9",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
			BlockLocation: models.FilePosition{
				Line:   models.Position{Start: 11, End: 11},
				Column: models.Position{Start: 2, End: 38},
			},
			VersionLocation: &models.FilePosition{
				Line:   models.Position{Start: 11, End: 11},
				Column: models.Position{Start: 33, End: 38},
			},
			NameLocation: &models.FilePosition{
				Line:   models.Position{Start: 11, End: 11},
				Column: models.Position{Start: 2, End: 31},
			},
		},
		{
			Name:      "github.com/mattn/go-isatty",
			Version:   "0.0.14",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
			BlockLocation: models.FilePosition{
				Line:   models.Position{Start: 12, End: 12},
				Column: models.Position{Start: 2, End: 36},
			},
			VersionLocation: &models.FilePosition{
				Line:   models.Position{Start: 12, End: 12},
				Column: models.Position{Start: 30, End: 36},
			},
			NameLocation: &models.FilePosition{
				Line:   models.Position{Start: 12, End: 12},
				Column: models.Position{Start: 2, End: 28},
			},
		},
		{
			Name:      "golang.org/x/sys",
			Version:   "0.0.0-20210630005230-0f9fa26af87c",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
			BlockLocation: models.FilePosition{
				Line:   models.Position{Start: 13, End: 13},
				Column: models.Position{Start: 2, End: 53},
			},
			VersionLocation: &models.FilePosition{
				Line:   models.Position{Start: 13, End: 13},
				Column: models.Position{Start: 20, End: 53},
			},
			NameLocation: &models.FilePosition{
				Line:   models.Position{Start: 13, End: 13},
				Column: models.Position{Start: 2, End: 18},
			},
		},
		{
			Name:      "stdlib",
			Version:   "1.17.0",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
		},
	})
}

func TestParseGoLock_Replacements_One(t *testing.T) {
	t.Parallel()

	packages, err := lockfile.ParseGoLock("fixtures/go/replace-one.mod")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:      "example.com/fork/net",
			Version:   "1.4.5",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
			BlockLocation: models.FilePosition{
				Line:   models.Position{Start: 5, End: 5},
				Column: models.Position{Start: 1, End: 63},
			},
			VersionLocation: &models.FilePosition{
				Line:   models.Position{Start: 5, End: 5},
				Column: models.Position{Start: 58, End: 63},
			},
			NameLocation: &models.FilePosition{
				Line:   models.Position{Start: 5, End: 5},
				Column: models.Position{Start: 36, End: 56},
			},
		},
	})
}

func TestParseGoLock_Replacements_Mixed(t *testing.T) {
	t.Parallel()

	packages, err := lockfile.ParseGoLock("fixtures/go/replace-mixed.mod")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:      "example.com/fork/net",
			Version:   "1.4.5",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
			BlockLocation: models.FilePosition{
				Line:   models.Position{Start: 7, End: 7},
				Column: models.Position{Start: 5, End: 59},
			},
			VersionLocation: &models.FilePosition{
				Line:   models.Position{Start: 7, End: 7},
				Column: models.Position{Start: 54, End: 59},
			},
			NameLocation: &models.FilePosition{
				Line:   models.Position{Start: 7, End: 7},
				Column: models.Position{Start: 32, End: 52},
			},
		},
		{
			Name:      "golang.org/x/net",
			Version:   "0.5.6",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
			BlockLocation: models.FilePosition{
				Line:   models.Position{Start: 3, End: 3},
				Column: models.Position{Start: 5, End: 28},
			},
			VersionLocation: &models.FilePosition{
				Line:   models.Position{Start: 3, End: 3},
				Column: models.Position{Start: 23, End: 28},
			},
			NameLocation: &models.FilePosition{
				Line:   models.Position{Start: 3, End: 3},
				Column: models.Position{Start: 5, End: 21},
			},
		},
	})
}

func TestParseGoLock_Replacements_Local(t *testing.T) {
	t.Parallel()

	packages, err := lockfile.ParseGoLock("fixtures/go/replace-local.mod")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:      "github.com/BurntSushi/toml",
			Version:   "1.0.0",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
			BlockLocation: models.FilePosition{
				Line:   models.Position{Start: 3, End: 3},
				Column: models.Position{Start: 5, End: 38},
			},
			VersionLocation: &models.FilePosition{
				Line:   models.Position{Start: 3, End: 3},
				Column: models.Position{Start: 33, End: 38},
			},
			NameLocation: &models.FilePosition{
				Line:   models.Position{Start: 3, End: 3},
				Column: models.Position{Start: 5, End: 31},
			},
		},
	})
}

func TestParseGoLock_Replacements_Different(t *testing.T) {
	t.Parallel()

	packages, err := lockfile.ParseGoLock("fixtures/go/replace-different.mod")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:      "example.com/fork/foe",
			Version:   "1.4.5",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
			BlockLocation: models.FilePosition{
				Line:   models.Position{Start: 7, End: 7},
				Column: models.Position{Start: 5, End: 59},
			},
			VersionLocation: &models.FilePosition{
				Line:   models.Position{Start: 7, End: 7},
				Column: models.Position{Start: 54, End: 59},
			},
			NameLocation: &models.FilePosition{
				Line:   models.Position{Start: 7, End: 7},
				Column: models.Position{Start: 32, End: 52},
			},
		},
		{
			Name:      "example.com/fork/foe",
			Version:   "1.4.2",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
			BlockLocation: models.FilePosition{
				Line:   models.Position{Start: 8, End: 8},
				Column: models.Position{Start: 5, End: 59},
			},
			VersionLocation: &models.FilePosition{
				Line:   models.Position{Start: 8, End: 8},
				Column: models.Position{Start: 54, End: 59},
			},
			NameLocation: &models.FilePosition{
				Line:   models.Position{Start: 8, End: 8},
				Column: models.Position{Start: 32, End: 52},
			},
		},
	})
}

func TestParseGoLock_Replacements_NotRequired(t *testing.T) {
	t.Parallel()

	packages, err := lockfile.ParseGoLock("fixtures/go/replace-not-required.mod")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:      "golang.org/x/net",
			Version:   "0.5.6",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
			BlockLocation: models.FilePosition{
				Line:   models.Position{Start: 2, End: 2},
				Column: models.Position{Start: 5, End: 28},
			},
			VersionLocation: &models.FilePosition{
				Line:   models.Position{Start: 2, End: 2},
				Column: models.Position{Start: 23, End: 28},
			},
			NameLocation: &models.FilePosition{
				Line:   models.Position{Start: 2, End: 2},
				Column: models.Position{Start: 5, End: 21},
			},
		},
		{
			Name:      "github.com/BurntSushi/toml",
			Version:   "1.0.0",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
			BlockLocation: models.FilePosition{
				Line:   models.Position{Start: 3, End: 3},
				Column: models.Position{Start: 5, End: 38},
			},
			VersionLocation: &models.FilePosition{
				Line:   models.Position{Start: 3, End: 3},
				Column: models.Position{Start: 33, End: 38},
			},
			NameLocation: &models.FilePosition{
				Line:   models.Position{Start: 3, End: 3},
				Column: models.Position{Start: 5, End: 31},
			},
		},
	})
}

func TestParseGoLock_Replacements_NoVersion(t *testing.T) {
	t.Parallel()

	packages, err := lockfile.ParseGoLock("fixtures/go/replace-no-version.mod")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:      "example.com/fork/net",
			Version:   "1.4.5",
			Ecosystem: lockfile.GoEcosystem,
			CompareAs: lockfile.GoEcosystem,
			BlockLocation: models.FilePosition{
				Line:   models.Position{Start: 7, End: 7},
				Column: models.Position{Start: 5, End: 52},
			},
			VersionLocation: &models.FilePosition{
				Line:   models.Position{Start: 7, End: 7},
				Column: models.Position{Start: 47, End: 52},
			},
			NameLocation: &models.FilePosition{
				Line:   models.Position{Start: 7, End: 7},
				Column: models.Position{Start: 25, End: 45},
			},
		},
	})
}
