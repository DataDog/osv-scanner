package sbom

import (
	"slices"
	"strings"
	"time"

	"golang.org/x/exp/maps"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/google/osv-scanner/pkg/models"
)

type PackageProcessingHook = func(component *cyclonedx.Component, details models.PackageVulns)

func buildCycloneDXBom(uniquePackages map[string]models.PackageVulns, artifacts []models.ScannedArtifact, pkgProcessingHook PackageProcessingHook) *cyclonedx.BOM {
	bom := cyclonedx.NewBOM()
	components := make([]cyclonedx.Component, 0)
	bomVulnerabilities := make([]cyclonedx.Vulnerability, 0)
	vulnerabilities := make(map[string]cyclonedx.Vulnerability)

	fileComponents, dependsOn := addFileDependencies(artifacts)
	components = append(components, fileComponents...)
	for packageURL, packageDetail := range uniquePackages {
		libraryComponent := createLibraryComponent(packageURL, packageDetail)
		artifact := findArtifact(packageDetail.Package.Name, packageDetail.Package.Version, artifacts)
		pkgFileComponents := createFileComponents(packageDetail, artifact, dependsOn)

		pkgProcessingHook(&libraryComponent, packageDetail)
		addVulnerabilities(vulnerabilities, packageDetail)

		components = append(components, libraryComponent)
		components = append(components, pkgFileComponents...)
	}

	slices.SortFunc(components, func(a, b cyclonedx.Component) int {
		return strings.Compare(a.BOMRef, b.BOMRef)
	})

	for _, vulnerability := range vulnerabilities {
		bomVulnerabilities = append(bomVulnerabilities, vulnerability)
	}

	slices.SortFunc(bomVulnerabilities, func(a, b cyclonedx.Vulnerability) int {
		return strings.Compare(a.ID, b.ID)
	})

	dependencies := maps.Values(dependsOn)
	slices.SortFunc(dependencies, func(a, b cyclonedx.Dependency) int {
		return strings.Compare(a.Ref, b.Ref)
	})

	bom.Components = &components
	bom.Dependencies = &dependencies
	bom.Vulnerabilities = &bomVulnerabilities

	return bom
}

func findArtifact(name string, version string, artifacts []models.ScannedArtifact) *models.ScannedArtifact {
	for _, artifact := range artifacts {
		if artifact.Name == name && artifact.Version == version {
			return &artifact
		}
	}

	return nil
}

func createFileComponents(packageDetail models.PackageVulns, artifact *models.ScannedArtifact, dependsOn map[string]cyclonedx.Dependency) []cyclonedx.Component {
	components := make([]cyclonedx.Component, 0)

	for _, location := range packageDetail.Locations {
		component := cyclonedx.Component{}

		component.Type = fileComponentType
		component.BOMRef = location.Block.Filename
		component.Name = location.Block.Filename
		components = append(components, component)

		if artifact != nil {
			// The current component is a repository artifact, meaning it is an internal dependency, we should report a dependsOn on the location
			if dependency, ok := dependsOn[location.Block.Filename]; !ok {
				dependencies := make([]string, 1)
				dependencies[0] = artifact.Filename
				dependsOn[location.Block.Filename] = cyclonedx.Dependency{
					Ref:          location.Block.Filename,
					Dependencies: &dependencies,
				}
			} else {
				dependencies := append(*dependency.Dependencies, artifact.Filename)
				slices.Sort(dependencies)
				dependency.Dependencies = &dependencies
				dependsOn[location.Block.Filename] = dependency
			}
		}
	}

	return components
}

func createLibraryComponent(packageURL string, packageDetail models.PackageVulns) cyclonedx.Component {
	component := cyclonedx.Component{}

	component.Type = libraryComponentType
	component.BOMRef = packageURL
	component.PackageURL = packageURL
	component.Name = packageDetail.Package.Name
	component.Version = packageDetail.Package.Version

	fillLicenses(&component, packageDetail)

	return component
}

func fillLicenses(component *cyclonedx.Component, packageDetail models.PackageVulns) {
	licenses := make(cyclonedx.Licenses, len(packageDetail.Licenses))

	for index, license := range packageDetail.Licenses {
		licenses[index] = cyclonedx.LicenseChoice{
			License: &cyclonedx.License{
				ID: string(license),
			},
		}
	}
	component.Licenses = &licenses
}

func addVulnerabilities(vulnerabilities map[string]cyclonedx.Vulnerability, packageDetail models.PackageVulns) {
	for _, vulnerability := range packageDetail.Vulnerabilities {
		if _, exists := vulnerabilities[vulnerability.ID]; exists {
			continue
		}

		// It doesn't exists yet, lets add it
		vulnerabilities[vulnerability.ID] = cyclonedx.Vulnerability{
			ID:          vulnerability.ID,
			Updated:     formatDateIfExists(vulnerability.Modified),
			Published:   formatDateIfExists(vulnerability.Published),
			Rejected:    formatDateIfExists(vulnerability.Withdrawn),
			References:  buildReferences(vulnerability),
			Description: vulnerability.Summary,
			Detail:      vulnerability.Details,
			Affects:     buildAffectedPackages(vulnerability),
			Ratings:     buildRatings(vulnerability),
			Advisories:  buildAdvisories(vulnerability),
			Credits:     buildCredits(vulnerability),
		}
	}
}

func addFileDependencies(artifacts []models.ScannedArtifact) ([]cyclonedx.Component, map[string]cyclonedx.Dependency) {
	components := make([]cyclonedx.Component, 0)
	dependsOn := make(map[string]cyclonedx.Dependency)

	for _, artifact := range artifacts {
		component := cyclonedx.Component{}
		occurrences := make([]cyclonedx.EvidenceOccurrence, 1)
		component.Name = artifact.Filename
		component.BOMRef = artifact.Filename
		component.Type = fileComponentType
		component.Evidence = &cyclonedx.Evidence{Occurrences: &occurrences}

		components = append(components, component)

		// Computing parent dependency
		if artifact.DependsOn != nil {
			if dependency, ok := dependsOn[artifact.DependsOn.Filename]; ok {
				dependencies := append(*dependency.Dependencies, artifact.DependsOn.Filename)
				slices.Sort(dependencies)

				dependency.Dependencies = &dependencies
				dependsOn[artifact.Filename] = dependency
			} else {
				dependsOn[artifact.Filename] = cyclonedx.Dependency{
					Ref: component.BOMRef,
					Dependencies: &[]string{
						artifact.DependsOn.Filename,
					},
				}
			}
		}
	}

	return components, dependsOn
}

func formatDateIfExists(date time.Time) string {
	if date.IsZero() {
		return ""
	}

	return date.Format(time.RFC3339)
}

func buildCredits(vulnerability models.Vulnerability) *cyclonedx.Credits {
	organizations := make([]cyclonedx.OrganizationalEntity, len(vulnerability.Credits))

	for index, credit := range vulnerability.Credits {
		organizations[index] = cyclonedx.OrganizationalEntity{
			Name: credit.Name,
			URL:  &vulnerability.Credits[index].Contact,
		}
	}

	return &cyclonedx.Credits{
		Organizations: &organizations,
	}
}

func buildAffectedPackages(vulnerability models.Vulnerability) *[]cyclonedx.Affects {
	affectedPackages := make([]cyclonedx.Affects, len(vulnerability.Affected))

	for index, affected := range vulnerability.Affected {
		affectedPackages[index] = cyclonedx.Affects{
			Ref: affected.Package.Purl,
		}
	}

	return &affectedPackages
}

func buildRatings(vulnerability models.Vulnerability) *[]cyclonedx.VulnerabilityRating {
	ratings := make([]cyclonedx.VulnerabilityRating, len(vulnerability.Severity))
	for index, severity := range vulnerability.Severity {
		ratings[index] = cyclonedx.VulnerabilityRating{
			Method: SeverityMapper[severity.Type],
			Vector: severity.Score,
		}
	}

	return &ratings
}

func buildReferences(vulnerability models.Vulnerability) *[]cyclonedx.VulnerabilityReference {
	references := make([]cyclonedx.VulnerabilityReference, len(vulnerability.Aliases))

	for index, alias := range vulnerability.Aliases {
		references[index] = cyclonedx.VulnerabilityReference{
			ID:     alias,
			Source: &cyclonedx.Source{},
		}
	}

	return &references
}

func buildAdvisories(vulnerability models.Vulnerability) *[]cyclonedx.Advisory {
	advisories := make([]cyclonedx.Advisory, 0)
	for _, reference := range vulnerability.References {
		if reference.Type != models.ReferenceAdvisory {
			continue
		}
		advisories = append(advisories, cyclonedx.Advisory{
			URL: reference.URL,
		})
	}

	return &advisories
}
