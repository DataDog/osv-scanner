package lockfile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/osv-scanner/pkg/lockfile"
	"github.com/google/osv-scanner/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestParseSetupCfg(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/pip/simple/setup.cfg"))
	packages, err := lockfile.ParseSetupCfg(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackages(t, packages, []lockfile.PackageDetails{
		{
			Name:           "jinja2",
			Version:        "~=2.7.2",
			PackageManager: models.SetupTools,
			Ecosystem:      lockfile.PipEcosystem,
			CompareAs:      lockfile.PipEcosystem,
			DepGroups:      []string{"setup"},
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 18, End: 18},
				Column:   models.Position{Start: 5, End: 20},
				Filename: path,
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 18, End: 18},
				Column:   models.Position{Start: 5, End: 11},
				Filename: path,
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 18, End: 18},
				Column:   models.Position{Start: 12, End: 20},
				Filename: path,
			},
		},
		{
			Name:           "django",
			Version:        ">=1.6.1",
			PackageManager: models.SetupTools,
			Ecosystem:      lockfile.PipEcosystem,
			CompareAs:      lockfile.PipEcosystem,
			DepGroups:      []string{"setup"},
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 19, End: 19},
				Column:   models.Position{Start: 5, End: 20},
				Filename: path,
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 19, End: 19},
				Column:   models.Position{Start: 5, End: 11},
				Filename: path,
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 19, End: 19},
				Column:   models.Position{Start: 12, End: 20},
				Filename: path,
			},
		},
		{
			Name:           "python-etcd",
			Version:        "<=0.4.5",
			PackageManager: models.SetupTools,
			Ecosystem:      lockfile.PipEcosystem,
			CompareAs:      lockfile.PipEcosystem,
			DepGroups:      []string{"setup"},
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 20, End: 20},
				Column:   models.Position{Start: 5, End: 24},
				Filename: path,
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 20, End: 20},
				Column:   models.Position{Start: 5, End: 16},
				Filename: path,
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 20, End: 20},
				Column:   models.Position{Start: 17, End: 24},
				Filename: path,
			},
		},
		{
			Name:           "django-select2",
			Version:        ">6.0.1",
			PackageManager: models.SetupTools,
			Ecosystem:      lockfile.PipEcosystem,
			CompareAs:      lockfile.PipEcosystem,
			DepGroups:      []string{"setup"},
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 22, End: 22},
				Column:   models.Position{Start: 5, End: 113},
				Filename: path,
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 22, End: 22},
				Column:   models.Position{Start: 5, End: 19},
				Filename: path,
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 22, End: 22},
				Column:   models.Position{Start: 20, End: 26},
				Filename: path,
			},
		},
		{
			Name:           "irc",
			Version:        "<16.2",
			PackageManager: models.SetupTools,
			Ecosystem:      lockfile.PipEcosystem,
			CompareAs:      lockfile.PipEcosystem,
			DepGroups:      []string{"setup"},
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 23, End: 23},
				Column:   models.Position{Start: 5, End: 92},
				Filename: path,
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 23, End: 23},
				Column:   models.Position{Start: 5, End: 8},
				Filename: path,
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 23, End: 23},
				Column:   models.Position{Start: 9, End: 14},
				Filename: path,
			},
		},
		{
			Name:           "testtools",
			Version:        "===2.3.0",
			PackageManager: models.SetupTools,
			Ecosystem:      lockfile.PipEcosystem,
			CompareAs:      lockfile.PipEcosystem,
			DepGroups:      []string{"setup"},
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 24, End: 24},
				Column:   models.Position{Start: 5, End: 67},
				Filename: path,
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 24, End: 24},
				Column:   models.Position{Start: 5, End: 14},
				Filename: path,
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 24, End: 24},
				Column:   models.Position{Start: 15, End: 23},
				Filename: path,
			},
		},
		{
			Name:           "requests",
			Version:        "!=2.3.0",
			PackageManager: models.SetupTools,
			Ecosystem:      lockfile.PipEcosystem,
			CompareAs:      lockfile.PipEcosystem,
			DepGroups:      []string{"setup"},
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 25, End: 25},
				Column:   models.Position{Start: 5, End: 20},
				Filename: path,
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 25, End: 25},
				Column:   models.Position{Start: 5, End: 13},
				Filename: path,
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 25, End: 25},
				Column:   models.Position{Start: 13, End: 20},
				Filename: path,
			},
		},
		{
			Name:           "tensorflow",
			Version:        "==2.17.0",
			PackageManager: models.SetupTools,
			Ecosystem:      lockfile.PipEcosystem,
			CompareAs:      lockfile.PipEcosystem,
			DepGroups:      []string{"setup"},
			BlockLocation: models.FilePosition{
				Line:     models.Position{Start: 27, End: 27},
				Column:   models.Position{Start: 5, End: 24},
				Filename: path,
			},
			NameLocation: &models.FilePosition{
				Line:     models.Position{Start: 27, End: 27},
				Column:   models.Position{Start: 5, End: 15},
				Filename: path,
			},
			VersionLocation: &models.FilePosition{
				Line:     models.Position{Start: 27, End: 27},
				Column:   models.Position{Start: 15, End: 24},
				Filename: path,
			},
		},
	})
}

func TestParseSetupCfg_MissingInstallRequiresInsideComment(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/pip/install_requires_comment/setup.cfg"))
	_, err = lockfile.ParseSetupCfg(path)

	assert.ErrorContains(t, err, "could not find options.install_requires")
}

func TestParseSetupCfg_DoubleEqual(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/pip/double_equals/setup.cfg"))
	_, err = lockfile.ParseSetupCfg(path)

	assert.ErrorContains(t, err, "could not parse requirement line")
}

func TestParseSetupCfg_BadStringNoArray(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/pip/bad_string_no_array/setup.cfg"))
	_, err = lockfile.ParseSetupCfg(path)

	assert.ErrorContains(t, err, "could not find options.install_requires")
}
