package sbom

import (
	"slices"
	"strings"

	"github.com/datadog/osv-scanner/internal/utility/purl"

	"golang.org/x/exp/maps"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/datadog/osv-scanner/pkg/models"
)

type PackageProcessingHook = func(component *cyclonedx.Component, details models.PackageVulns)

func buildCycloneDXBom(uniquePackages map[string]models.PackageVulns, artifacts []models.ScannedArtifact, pkgProcessingHook PackageProcessingHook) *cyclonedx.BOM {
	bom := cyclonedx.NewBOM()
	components := make([]cyclonedx.Component, 0)
	bomVulnerabilities := make([]cyclonedx.Vulnerability, 0)
	vulnerabilities := make(map[string]cyclonedx.Vulnerability)

	fileComponents, dependsOn := addFileDependencies(artifacts)
	for packageURL, packageDetail := range uniquePackages {
		libraryComponent := createLibraryComponent(packageURL, packageDetail)
		artifact := findArtifact(packageDetail.Package.Name, packageDetail.Package.Version, artifacts)
		createFileComponents(packageDetail, artifact, dependsOn)

		pkgProcessingHook(&libraryComponent, packageDetail)

		components = append(components, libraryComponent)
	}
	components = append(components, maps.Values(fileComponents)...)
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

func buildProperties(metadatas models.PackageMetadata) []cyclonedx.Property {
	properties := make([]cyclonedx.Property, 0)

	for metadataType, value := range metadatas {
		if len(value) == 0 {
			continue
		}
		properties = append(properties, cyclonedx.Property{
			Name:  "osv-scanner:" + string(metadataType),
			Value: value,
		})
	}

	slices.SortFunc(properties, func(a, b cyclonedx.Property) int {
		return strings.Compare(a.Name, b.Name)
	})

	return properties
}

func findArtifact(name string, version string, artifacts []models.ScannedArtifact) *models.ScannedArtifact {
	for _, artifact := range artifacts {
		if artifact.Name == name && artifact.Version == version {
			return &artifact
		}
	}

	return nil
}

func createFileComponents(packageDetail models.PackageVulns, artifact *models.ScannedArtifact, dependsOn map[string]cyclonedx.Dependency) {
	for _, location := range packageDetail.Locations {
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
}

func createLibraryComponent(packageURL string, packageDetail models.PackageVulns) cyclonedx.Component {
	component := cyclonedx.Component{}

	component.Type = libraryComponentType
	component.BOMRef = packageURL
	component.PackageURL = packageURL
	component.Name = packageDetail.Package.Name
	component.Version = packageDetail.Package.Version

	properties := buildProperties(packageDetail.Metadata)
	component.Properties = &properties

	return component
}

func addFileDependencies(artifacts []models.ScannedArtifact) (map[string]cyclonedx.Component, map[string]cyclonedx.Dependency) {
	components := make(map[string]cyclonedx.Component)
	dependsOn := make(map[string]cyclonedx.Dependency)

	for _, artifact := range artifacts {
		artifactPURL, err := purl.From(models.PackageInfo{
			Name:      artifact.Name,
			Version:   artifact.Version,
			Ecosystem: string(artifact.Ecosystem),
		})
		if err != nil {
			continue
		}

		component := cyclonedx.Component{}
		properties := make([]cyclonedx.Property, 1)
		component.Name = artifact.Filename
		component.BOMRef = artifact.Filename
		component.Type = fileComponentType
		properties[0] = cyclonedx.Property{
			Name:  "osv-scanner:package",
			Value: artifactPURL.String(),
		}
		component.Properties = &properties
		components[component.BOMRef] = component

		// Computing parent dependency
		if artifact.DependsOn != nil {
			if dependency, ok := dependsOn[artifact.Filename]; ok {
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
