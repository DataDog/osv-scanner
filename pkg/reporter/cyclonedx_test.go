package reporter_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/google/osv-scanner/pkg/models"
	"github.com/google/osv-scanner/pkg/reporter"
	"github.com/stretchr/testify/assert"
)

type JSONMap = map[string]interface{}

var vulnResults = models.VulnerabilityResults{
	Results: []models.PackageSource{
		{
			Source: models.SourceInfo{
				Path: "/path/to/lockfile.xml",
				Type: "",
			},
			Packages: []models.PackageVulns{
				{
					Package: models.PackageInfo{
						Name:      "com.foo:the-greatest-package",
						Version:   "1.0.0",
						Ecosystem: string(models.EcosystemMaven),
						Line: models.Position{
							Start: 1,
							End:   3,
						},
						Column: models.Position{
							Start: 30,
							End:   35,
						},
					},
				},
			},
		},
		{
			Source: models.SourceInfo{
				Path: "/path/to/another-lockfile.xml",
				Type: "",
			},
			Packages: []models.PackageVulns{
				{
					Package: models.PackageInfo{
						Name:      "com.foo:the-greatest-package",
						Version:   "1.0.0",
						Ecosystem: string(models.EcosystemMaven),
						Line: models.Position{
							Start: 11,
							End:   13,
						},
					},
				},
			},
		},
		{
			Source: models.SourceInfo{
				Path: "/path/to/npm/lockfile.lock",
				Type: "",
			},
			Packages: []models.PackageVulns{
				{
					Package: models.PackageInfo{
						Name:      "the-npm-package",
						Version:   "1.1.0",
						Ecosystem: string(models.EcosystemNPM),
						Line: models.Position{
							Start: 12,
							End:   15,
						},
					},
				},
			},
		},
	},
}

func TestEncoding_EncodeComponentsInValidCycloneDX1_4(t *testing.T) {
	t.Parallel()
	var stdout, stderr strings.Builder
	cycloneDXReporter := reporter.NewCycloneDXReporter(&stdout, &stderr, reporter.CycloneDXVersion14)

	// First we format packages in CycloneDX format
	err := cycloneDXReporter.PrintResult(&vulnResults)
	require.NoError(t, err, "an error occurred when formatting")

	// Then we try to decode it using the CycloneDX library directly to check the content
	var bom cyclonedx.BOM
	decoder := cyclonedx.NewBOMDecoder(strings.NewReader(stdout.String()), cyclonedx.BOMFileFormatJSON)
	err = decoder.Decode(&bom)
	require.NoError(t, err, "an error occurred when decoding")

	expectedBOM := cyclonedx.BOM{
		JSONSchema:  "https://cyclonedx.org/schema/bom-1.4.schema.json",
		Version:     1,
		BOMFormat:   cyclonedx.BOMFormat,
		SpecVersion: cyclonedx.SpecVersion1_4,
		Components: &[]cyclonedx.Component{
			{
				BOMRef:     "pkg:maven/com.foo/the-greatest-package@1.0.0",
				PackageURL: "pkg:maven/com.foo/the-greatest-package@1.0.0",
				Name:       "com.foo:the-greatest-package",
				Version:    "1.0.0",
				Type:       "library",
			},
			{
				BOMRef:     "pkg:npm/the-npm-package@1.1.0",
				PackageURL: "pkg:npm/the-npm-package@1.1.0",
				Name:       "the-npm-package",
				Version:    "1.1.0",
				Type:       "library",
			},
		},
	}
	assertBaseBomEquals(t, expectedBOM, bom)
	for _, expectedComponent := range *expectedBOM.Components {
		assertComponentsContains(t, expectedComponent, *bom.Components)
	}
}

func TestEncoding_EncodeComponentsInValidCycloneDX1_5(t *testing.T) {
	t.Parallel()
	var stdout, stderr strings.Builder
	cycloneDXReporter := reporter.NewCycloneDXReporter(&stdout, &stderr, reporter.CycloneDXVersion15)

	// First we format packages in CycloneDX format
	err := cycloneDXReporter.PrintResult(&vulnResults)
	require.NoError(t, err, "an error occurred when formatting")

	// Then we try to decode it using the CycloneDX library directly to check the content
	var bom cyclonedx.BOM
	decoder := cyclonedx.NewBOMDecoder(strings.NewReader(stdout.String()), cyclonedx.BOMFileFormatJSON)
	err = decoder.Decode(&bom)
	require.NoError(t, err, "an error occurred when decoding")

	expectedJSONLocations := map[string][]JSONMap{
		"pkg:maven/com.foo/the-greatest-package@1.0.0": {
			{
				"block": JSONMap{
					"file_name":    "/path/to/lockfile.xml",
					"line_start":   1,
					"line_end":     3,
					"column_start": 30,
					"column_end":   35,
				},
			},
			{
				"block": JSONMap{
					"file_name":    "/path/to/another-lockfile.xml",
					"line_start":   11,
					"line_end":     13,
					"column_start": 0,
					"column_end":   0,
				},
			},
		},
		"pkg:npm/the-npm-package@1.1.0": {
			{
				"block": JSONMap{
					"file_name":    "/path/to/npm/lockfile.lock",
					"line_start":   12,
					"line_end":     15,
					"column_start": 0,
					"column_end":   0,
				},
			},
		},
	}

	expectedBOM := cyclonedx.BOM{
		JSONSchema:  "https://cyclonedx.org/schema/bom-1.5.schema.json",
		Version:     1,
		BOMFormat:   cyclonedx.BOMFormat,
		SpecVersion: cyclonedx.SpecVersion1_5,
		Components: &[]cyclonedx.Component{
			{
				BOMRef:     "pkg:maven/com.foo/the-greatest-package@1.0.0",
				PackageURL: "pkg:maven/com.foo/the-greatest-package@1.0.0",
				Name:       "com.foo:the-greatest-package",
				Version:    "1.0.0",
				Type:       "library",
				Evidence: &cyclonedx.Evidence{
					Occurrences: buildOccurrences(t, "pkg:maven/com.foo/the-greatest-package@1.0.0", expectedJSONLocations),
				},
			},
			{
				BOMRef:     "pkg:npm/the-npm-package@1.1.0",
				PackageURL: "pkg:npm/the-npm-package@1.1.0",
				Name:       "the-npm-package",
				Version:    "1.1.0",
				Type:       "library",
				Evidence: &cyclonedx.Evidence{
					Occurrences: buildOccurrences(t, "pkg:npm/the-npm-package@1.1.0", expectedJSONLocations),
				},
			},
		},
	}

	assertBaseBomEquals(t, expectedBOM, bom)
	for _, expectedComponent := range *expectedBOM.Components {
		actualComponent := assertComponentsContains(t, expectedComponent, *bom.Components)
		expectedLocations, ok := expectedJSONLocations[actualComponent.PackageURL]
		if !ok {
			continue
		}
		assertLocationsExactlyContains(t, expectedLocations, *actualComponent.Evidence.Occurrences)
	}
}

func buildOccurrences(t *testing.T, purl string, expectedLocations map[string][]JSONMap) *[]cyclonedx.EvidenceOccurrence {
	t.Helper()
	locations, ok := expectedLocations[purl]

	if !ok {
		return nil
	}

	result := make([]cyclonedx.EvidenceOccurrence, len(locations))
	for index, location := range locations {
		builder := strings.Builder{}
		err := json.NewEncoder(&builder).Encode(location)
		require.NoError(t, err)
		result[index] = cyclonedx.EvidenceOccurrence{
			Location: builder.String(),
		}
	}

	return &result
}

func assertLocationsExactlyContains(t *testing.T, expectedLocations []JSONMap, actualLocations []cyclonedx.EvidenceOccurrence) {
	t.Helper()
	assert.EqualValues(t, len(expectedLocations), len(actualLocations), "Size between expected and actual is different")
	for _, occurrence := range actualLocations {
		var actualLocation JSONMap
		err := json.NewDecoder(strings.NewReader(occurrence.Location)).Decode(&actualLocation)
		require.NoError(t, err)

		assert.Condition(t, assertContainsEquals(expectedLocations, actualLocation))
	}
}

func assertContainsEquals(expectedLocations []JSONMap, actualLocation JSONMap) assert.Comparison {
	return func() bool {
		for _, expected := range expectedLocations {
			if fmt.Sprint(expected) == fmt.Sprint(actualLocation) {
				return true
			}
		}

		return false
	}
}

func assertBaseBomEquals(t *testing.T, expected, actual cyclonedx.BOM) {
	t.Helper()
	assert.EqualValues(t, expected.JSONSchema, actual.JSONSchema)
	assert.EqualValues(t, expected.Version, actual.Version)
	assert.EqualValues(t, expected.BOMFormat, actual.BOMFormat)
	assert.EqualValues(t, expected.SpecVersion, actual.SpecVersion)
}

func assertComponentsContains(t *testing.T, expected cyclonedx.Component, actual []cyclonedx.Component) *cyclonedx.Component {
	t.Helper()

	for _, component := range actual {
		if component.PackageURL != expected.PackageURL {
			continue
		}
		assert.EqualValues(t, expected.Name, component.Name)
		assert.EqualValues(t, expected.Version, component.Version)
		assert.EqualValues(t, expected.BOMRef, component.BOMRef)
		assert.EqualValues(t, expected.PackageURL, component.PackageURL)
		assert.EqualValues(t, expected.Type, component.Type)

		return &component
	}
	assert.FailNowf(t, "Received component array does not contains expected component", "%v", expected)

	return nil
}
