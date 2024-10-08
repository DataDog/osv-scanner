package lockfile_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/google/osv-scanner/internal/output"
	"github.com/google/osv-scanner/pkg/lockfile"
)

func expectErrContaining(t *testing.T, err error, str string) {
	t.Helper()

	if err == nil {
		t.Errorf("Expected to get error, but did not")
	}

	if !strings.Contains(err.Error(), str) {
		t.Errorf("Expected to get \"%s\" error, but got \"%v\"", str, err)
	}
}

func expectErrIs(t *testing.T, err error, expected error) {
	t.Helper()

	if err == nil {
		t.Errorf("Expected to get error, but did not")
	}

	if !errors.Is(err, expected) {
		t.Errorf("Expected to get \"%v\" error but got \"%v\" instead", expected, err)
	}
}

func packageToString(pkg lockfile.PackageDetails) string {
	commit := pkg.Commit

	if commit == "" {
		commit = "<no commit>"
	}

	groups := strings.Join(pkg.DepGroups, ", ")

	if groups == "" {
		groups = "<no groups>"
	}

	blockLocation := fmt.Sprintf("L%d-%d;C%d-%d", pkg.BlockLocation.Line.Start, pkg.BlockLocation.Line.End, pkg.BlockLocation.Column.Start, pkg.BlockLocation.Column.End)

	nameLocation := "<no location>"
	if pkg.NameLocation != nil {
		nameLocation = fmt.Sprintf("L%d-%d;C%d-%d", pkg.NameLocation.Line.Start, pkg.NameLocation.Line.End, pkg.NameLocation.Column.Start, pkg.NameLocation.Column.End)
	}

	versionLocation := "<no location>"
	if pkg.VersionLocation != nil {
		versionLocation = fmt.Sprintf("L%d-%d;C%d-%d", pkg.VersionLocation.Line.Start, pkg.VersionLocation.Line.End, pkg.VersionLocation.Column.Start, pkg.VersionLocation.Column.End)
	}

	return fmt.Sprintf("%s@%s (%s, %s, %s) [block:%s] [name:%s] [version:%s]", pkg.Name, pkg.Version, pkg.Ecosystem, commit, groups, blockLocation, nameLocation, versionLocation)
}

func hasPackage(t *testing.T, packages []lockfile.PackageDetails, pkg lockfile.PackageDetails, ignoreLocations bool) bool {
	t.Helper()

	for _, details := range packages {
		var ignore []string
		if ignoreLocations {
			ignore = []string{"BlockLocation", "NameLocation", "VersionLocation"}
		}

		if cmp.Equal(details, pkg, cmpopts.IgnoreFields(lockfile.PackageDetails{}, ignore...)) {
			return true
		}
	}

	return false
}

func innerExpectPackage(t *testing.T, packages []lockfile.PackageDetails, pkg lockfile.PackageDetails, ignoreLocations bool) {
	t.Helper()

	if !hasPackage(t, packages, pkg, ignoreLocations) {
		t.Errorf(
			"Expected packages to include %s@%s (%s, %s), but it did not",
			pkg.Name,
			pkg.Version,
			pkg.Ecosystem,
			pkg.CompareAs,
		)
	}
}

func expectPackage(t *testing.T, packages []lockfile.PackageDetails, pkg lockfile.PackageDetails) {
	t.Helper()

	innerExpectPackage(t, packages, pkg, false)
}

func findMissingPackages(t *testing.T, actualPackages []lockfile.PackageDetails, expectedPackages []lockfile.PackageDetails, ignoreLocations bool) []lockfile.PackageDetails {
	t.Helper()
	var missingPackages []lockfile.PackageDetails

	for _, pkg := range actualPackages {
		if !hasPackage(t, expectedPackages, pkg, ignoreLocations) {
			missingPackages = append(missingPackages, pkg)
		}
	}

	return missingPackages
}

func innerExpectPackages(t *testing.T, actualPackages []lockfile.PackageDetails, expectedPackages []lockfile.PackageDetails, ignoreLocations bool) {
	t.Helper()

	if len(expectedPackages) != len(actualPackages) {
		t.Errorf(
			"Expected to get %d %s, but got %d",
			len(expectedPackages),
			output.Form(len(expectedPackages), "package", "packages"),
			len(actualPackages),
		)
	}

	missingActualPackages := findMissingPackages(t, actualPackages, expectedPackages, ignoreLocations)
	missingExpectedPackages := findMissingPackages(t, expectedPackages, actualPackages, ignoreLocations)

	if len(missingActualPackages) != 0 {
		for _, unexpectedPackage := range missingActualPackages {
			t.Errorf("Did not expect %s", packageToString(unexpectedPackage))
		}
	}

	if len(missingExpectedPackages) != 0 {
		for _, unexpectedPackage := range missingExpectedPackages {
			t.Errorf("Did not find %s", packageToString(unexpectedPackage))
		}
	}
}

func expectPackages(t *testing.T, actualPackages []lockfile.PackageDetails, expectedPackages []lockfile.PackageDetails) {
	t.Helper()

	innerExpectPackages(t, actualPackages, expectedPackages, false)
}

func expectPackagesWithoutLocations(t *testing.T, actualPackages []lockfile.PackageDetails, expectedPackages []lockfile.PackageDetails) {
	t.Helper()

	innerExpectPackages(t, actualPackages, expectedPackages, true)
}
