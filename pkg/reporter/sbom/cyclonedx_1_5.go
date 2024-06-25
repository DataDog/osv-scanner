package sbom

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/google/osv-scanner/pkg/models"
)

func ToCycloneDX15Bom(stderr io.Writer, uniquePackages map[string]models.PackageDetails, artifacts []models.ScannedArtifact) *cyclonedx.BOM {
	bom := cyclonedx.NewBOM()
	components := make([]cyclonedx.Component, 0)
	bom.JSONSchema = cycloneDx15Schema
	bom.SpecVersion = cyclonedx.SpecVersion1_5

	for packageURL, packageDetail := range uniquePackages {
		component := cyclonedx.Component{}
		occurrences := make([]cyclonedx.EvidenceOccurrence, len(packageDetail.Locations))
		component.Name = packageDetail.Name
		component.BOMRef = packageURL
		component.PackageURL = packageURL
		component.Type = componentTypeLibrary
		component.Evidence = &cyclonedx.Evidence{Occurrences: &occurrences}

		if packageDetail.Version != "" {
			component.Version = packageDetail.Version
		}

		for index, packageLocations := range packageDetail.Locations {
			jsonLocation, err := packageLocations.MarshalToJSONString()
			if err != nil {
				_, _ = fmt.Fprintf(stderr, "An error occurred when creating the jsonLocation structure : %v", err.Error())
				continue
			}

			occurrence := cyclonedx.EvidenceOccurrence{
				Location: jsonLocation,
			}
			(*component.Evidence.Occurrences)[index] = occurrence
		}
		components = append(components, component)
	}

	for _, artifact := range artifacts {
		component := cyclonedx.Component{}
		occurrences := make([]cyclonedx.EvidenceOccurrence, 1)
		component.Name = artifact.Name
		component.BOMRef = artifact.Filename
		component.Type = componentTypeFile
		component.Evidence = &cyclonedx.Evidence{Occurrences: &occurrences}

		if artifact.Version != "" {
			component.Version = artifact.Version
		}

		jsonLocation, err := json.Marshal(artifact.FilePosition)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "An error occurred when creating the jsonLocation structure : %v", err.Error())
			continue
		}
		occurrence := cyclonedx.EvidenceOccurrence{
			Location: string(jsonLocation),
		}
		occurrences[0] = occurrence
		components = append(components, component)
	}
	bom.Components = &components

	return bom
}
