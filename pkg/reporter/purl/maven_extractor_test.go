package purl_test

import (
	"testing"

	"github.com/google/osv-scanner/pkg/models"
	"github.com/google/osv-scanner/pkg/reporter/purl"
)

func TestMavenExtraction_shouldExtractPackages(t *testing.T) {
	t.Parallel()
	testCase := struct {
		packageInfo       models.PackageInfo
		expectedNamespace string
		expectedName      string
	}{
		packageInfo: models.PackageInfo{
			Name:      "log4j:log4j-core",
			Version:   "1.2.17",
			Ecosystem: string(models.EcosystemMaven),
			Commit:    "",
			BlockLocation: models.FilePosition{
				Line: models.Position{Start: 0, End: 0},
			},
		},
		expectedNamespace: "log4j",
		expectedName:      "log4j-core",
	}

	namespace, name, ok := purl.ExtractPURLFromMaven(testCase.packageInfo)

	if !ok {
		t.Errorf("Extraction didn't succeed, package has been wrongfully filtered")
	}
	if namespace != testCase.expectedNamespace {
		t.Errorf("got %s; want %s", namespace, testCase.expectedNamespace)
	}
	if name != testCase.expectedName {
		t.Errorf("got %s; want %s", name, testCase.expectedName)
	}
}

func TestMavenExtraction_shouldFilterPackages(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name        string
		packageInfo models.PackageInfo
	}{
		{
			name: "when_package_contains_less_than_2_parts",
			packageInfo: models.PackageInfo{
				Name:      "log4j",
				Version:   "1.2.17",
				Ecosystem: string(models.EcosystemMaven),
				Commit:    "",
				BlockLocation: models.FilePosition{
					Line: models.Position{Start: 0, End: 0},
				},
			},
		},
		{
			name: "when_package_have_no_name",
			packageInfo: models.PackageInfo{
				Name:      "",
				Version:   "1.2.17",
				Ecosystem: string(models.EcosystemMaven),
				Commit:    "",
				BlockLocation: models.FilePosition{
					Line: models.Position{Start: 0, End: 0},
				},
			},
		},
	}

	for _, test := range testCases {
		testCase := test
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			_, _, ok := purl.ExtractPURLFromMaven(testCase.packageInfo)

			if ok {
				t.Errorf("Package %v should have been filtered\n", testCase.packageInfo)
			}
		})
	}
}
