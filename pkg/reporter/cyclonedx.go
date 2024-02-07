package reporter

import (
	"encoding/json"
	"fmt"
	"github.com/CycloneDX/cyclonedx-go"
	"github.com/google/osv-scanner/pkg/models"
	"github.com/google/osv-scanner/pkg/reporter/purl"
	"io"
	"strings"
)

const componentType = "library"
type CycloneDXVersion int

type bomCreator func(stderr io.Writer, packageSources []models.PackageSource) *cyclonedx.BOM

type CycloneDXReporter struct {
	hasPrintedError bool
	stdout          io.Writer
	stderr          io.Writer
	version         CycloneDXVersion
}

type packageDetails struct {
	Name      string
	Version   string
	Ecosystem string
	Locations []packageLocations
}

type packageLocation struct {
	Filename    string `json:"file_name"`
	LineStart   int    `json:"line_start"`
	LineEnd     int    `json:"line_end"`
	ColumnStart int    `json:"column_start"`
	ColumnEnd   int    `json:"column_end"`
}

type packageLocations struct {
	Block     packageLocation `json:"block"`
	Namespace packageLocation `json:"namespace,omitempty"`
	Name      packageLocation `json:"name,omitempty"`
	Version   packageLocation `json:"version,omitempty"`
}

const (
	CycloneDXVersion14 CycloneDXVersion = iota
	CycloneDXVersion15
)

const (
	cycloneDx14Schema = "http://cyclonedx.org/schema/bom-1.4.schema.json"
	cycloneDx15Schema = "http://cyclonedx.org/schema/bom-1.5.schema.json"
)

var specVersionToBomCreator = map[CycloneDXVersion]bomCreator{
	CycloneDXVersion14: toCycloneDX14Bom,
	CycloneDXVersion15: toCycloneDX15Bom,
}

func NewCycloneDXReporter(stdout io.Writer, stderr io.Writer, version CycloneDXVersion) *CycloneDXReporter {
	return &CycloneDXReporter{
		stdout:          stdout,
		stderr:          stderr,
		hasPrintedError: false,
		version:         version,
	}
}

func (r *CycloneDXReporter) PrintError(msg string) {
	_, _ = fmt.Fprint(r.stderr, msg)
	r.hasPrintedError = true
}

func (r *CycloneDXReporter) PrintErrorf(msg string, a ...any) {
	fmt.Fprintf(r.stderr, msg, a...)
	r.hasPrintedError = true
}

func (r *CycloneDXReporter) HasPrintedError() bool {
	return r.hasPrintedError
}

func (r *CycloneDXReporter) PrintText(msg string) {
	_, _ = fmt.Fprint(r.stderr, msg)
}

func (r *CycloneDXReporter) PrintTextf(msg string, a ...any) {
	fmt.Fprintf(r.stderr, msg, a...)
}

func (r *CycloneDXReporter) PrintResult(vulnResults *models.VulnerabilityResults) error {
	bomCreator := specVersionToBomCreator[r.version]
	bom := bomCreator(r.stderr, vulnResults.Results)
	encoder := cyclonedx.NewBOMEncoder(r.stdout, cyclonedx.BOMFileFormatJSON)

	return encoder.Encode(bom)
}

func toCycloneDX14Bom(_ io.Writer, packageSources []models.PackageSource) *cyclonedx.BOM {
	bom := cyclonedx.NewBOM()
	components := make([]cyclonedx.Component, 0)
	bom.JSONSchema = cycloneDx14Schema
	bom.SpecVersion = cyclonedx.SpecVersion1_4
	bom.Components = &components

	uniquePackages := groupByPackage(packageSources)

	for packageURL, packageDetail := range uniquePackages {
		component := cyclonedx.Component{}
		component.Name = packageDetail.Name
		component.Version = packageDetail.Version
		component.BOMRef = packageURL
		component.PackageURL = packageURL
		component.Type = componentType
		components = append(components, component)
	}

	return bom
}

func toCycloneDX15Bom(stderr io.Writer, packageSources []models.PackageSource) *cyclonedx.BOM {
	bom := cyclonedx.NewBOM()
	components := make([]cyclonedx.Component, 0)
	bom.JSONSchema = cycloneDx15Schema
	bom.SpecVersion = cyclonedx.SpecVersion1_5
	bom.Components = &components

	uniquePackages := groupByPackage(packageSources)

	for packageURL, packageDetail := range uniquePackages {
		component := cyclonedx.Component{}
		occurrences := make([]cyclonedx.EvidenceOccurrence, len(packageDetail.Locations))
		component.Name = packageDetail.Name
		component.Version = packageDetail.Version
		component.BOMRef = packageURL
		component.PackageURL = packageURL
		component.Evidence = &cyclonedx.Evidence{Occurrences: &occurrences}

		for index, location := range packageDetail.Locations {
			err, location := createLocationString(location)
			if err != nil {
				_, _ = fmt.Fprintf(stderr, fmt.Sprintf("An error occured when creating the location structure : %v", err.Error()))
				continue
			}

			occurrence := cyclonedx.EvidenceOccurrence{
				Location: location,
			}
			(*component.Evidence.Occurrences)[index] = occurrence
		}
	}

	return bom
}

func createLocationString(location packageLocations) (error, string) {
	buffer := strings.Builder{}
	encoder := json.NewEncoder(&buffer)

	err := encoder.Encode(location)
	if err != nil {
		return err, ""
	}

	return nil, buffer.String()
}

func groupByPackage(packageSources []models.PackageSource) map[string]packageDetails {
	uniquePackages := make(map[string]packageDetails)

	for _, packageSource := range packageSources {
		for _, pkg := range packageSource.Packages {
			packageURL := purl.From(pkg.Package)
			if packageURL == nil {
				continue
			}
			existingPackage, packageExists := uniquePackages[packageURL.ToString()]
			location := extractPackageLocations(packageSource.Source, pkg.Package)
			if packageExists {
				// Package exists we need to add a location
				existingPackage.Locations = append(existingPackage.Locations, location)
			} else {
				// Create a new package and update the map
				newPackage := packageDetails{
					Name:      pkg.Package.Name,
					Version:   pkg.Package.Version,
					Ecosystem: pkg.Package.Ecosystem,
					Locations: make([]packageLocations, 1),
				}
				newPackage.Locations[0] = location
				uniquePackages[packageURL.ToString()] = newPackage
			}
		}
	}

	return uniquePackages
}

func extractPackageLocations(pkgSource models.SourceInfo, pkgInfos models.PackageInfo) packageLocations {
	return packageLocations{
		Block: packageLocation{
			Filename:  pkgSource.Path,
			LineStart: pkgInfos.LinePosition.Start,
			LineEnd:   pkgInfos.LinePosition.End,
		},
	}
}
