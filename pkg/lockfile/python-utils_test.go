package lockfile_test

import (
	"github.com/google/osv-scanner/pkg/models"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/google/osv-scanner/pkg/lockfile"
)

const testPath = "path"

func TestParseRequirementLine(t *testing.T) {
	t.Parallel()

	detail, err := lockfile.ParseRequirementLine(testPath, models.SetupTools, "    package-name[request,foo]~=1.2.3,<10.2.3 # one comment", "package-name[request,foo]~=1.2.3,<10.2.3", 3, 0, 5, 45)

	assert.Nil(t, err)
	assert.EqualValues(t, &lockfile.PackageDetails{
		Name:           "package-name",
		Version:        "~=1.2.3,<10.2.3",
		TargetVersions: []string(nil),
		Commit:         "",
		PackageManager: models.SetupTools,
		Ecosystem:      lockfile.PipEcosystem,
		CompareAs:      lockfile.PipEcosystem,
		DepGroups:      []string(nil),
		IsDirect:       false,
		BlockLocation: models.FilePosition{
			Line:     models.Position{Start: 3, End: 3},
			Column:   models.Position{Start: 5, End: 45},
			Filename: testPath,
		},
		NameLocation: &models.FilePosition{
			Line:     models.Position{Start: 3, End: 3},
			Column:   models.Position{Start: 5, End: 17},
			Filename: testPath,
		},
		VersionLocation: &models.FilePosition{
			Line:     models.Position{Start: 3, End: 3},
			Column:   models.Position{Start: 30, End: 45},
			Filename: testPath,
		},
	}, detail)

}

func TestParseRequirementLineWheel(t *testing.T) {
	t.Parallel()

	detail, err := lockfile.ParseRequirementLine(testPath, models.Requirements, "    wxPython_Phoenix @ http://wxpython.org/Phoenix/snapshot-builds/wxPython_Phoenix-3.0.3.dev1820+49a8884-cp34-none-win_amd64.whl # one comment", "wxPython_Phoenix @ http://wxpython.org/Phoenix/snapshot-builds/wxPython_Phoenix-3.0.3.dev1820+49a8884-cp34-none-win_amd64.whl", 3, 0, 5, 129)

	assert.Nil(t, err)
	assert.EqualValues(t, &lockfile.PackageDetails{
		Name:           "wxpython-phoenix",
		Version:        "==3.0.3.dev1820+49a8884",
		TargetVersions: []string(nil),
		Commit:         "",
		PackageManager: models.Requirements,
		Ecosystem:      lockfile.PipEcosystem,
		CompareAs:      lockfile.PipEcosystem,
		DepGroups:      []string(nil),
		IsDirect:       false,
		BlockLocation: models.FilePosition{
			Line:     models.Position{Start: 3, End: 3},
			Column:   models.Position{Start: 5, End: 129},
			Filename: testPath,
		},
		NameLocation: &models.FilePosition{
			Line:     models.Position{Start: 3, End: 3},
			Column:   models.Position{Start: 5, End: 21},
			Filename: testPath,
		},
		VersionLocation: &models.FilePosition{
			Line:     models.Position{Start: 3, End: 3},
			Column:   models.Position{Start: 24, End: 130},
			Filename: testPath,
		},
	}, detail)

}
