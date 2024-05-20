package grouper

import (
	"path/filepath"
	"strings"

	"github.com/google/osv-scanner/pkg/models"
	"github.com/google/osv-scanner/pkg/reporter/purl"
)

func GroupByPURL(packageSources []models.PackageSource) map[string]models.PackageDetails {
	uniquePackages := make(map[string]models.PackageDetails)

	for _, packageSource := range packageSources {
		pkgLocations := make(map[string]bool)
		for _, pkg := range packageSource.Packages {
			packageURL := purl.From(pkg.Package)
			if packageURL == nil {
				continue
			}
			existingPackage, packageExists := uniquePackages[packageURL.ToString()]
			isLocationExtracted := isLocationExtractedSuccessfully(pkg.Package.BlockLocation)
			location := extractPackageLocations(packageSource.Source.HostPath, pkg.Package)

			if packageExists && isLocationExtracted {
				locationHash := location.Block.Hash()
				if _, isLocationDuplicated := pkgLocations[locationHash]; isLocationDuplicated {
					continue
				}

				// Package exists, location exists and is not a duplicate => we need to add it
				existingPackage.Locations = append(existingPackage.Locations, location)
				pkgLocations[locationHash] = true
				uniquePackages[packageURL.ToString()] = existingPackage
			} else if !packageExists {
				// The package does not exists we need to add it
				// Create a new package and update the map
				newPackage := models.PackageDetails{
					Name:      pkg.Package.Name,
					Version:   pkg.Package.Version,
					Ecosystem: pkg.Package.Ecosystem,
					Locations: make([]models.PackageLocations, 0),
				}

				if isLocationExtracted {
					// We add location only if it has been extracted successfully
					newPackage.Locations = append(newPackage.Locations, location)
					pkgLocations[location.Block.Hash()] = true
				}
				uniquePackages[packageURL.ToString()] = newPackage
			}
		}
	}

	return uniquePackages
}

func isLocationExtractedSuccessfully(filePosition models.FilePosition) bool {
	return filePosition.Line.Start > 0 && filePosition.Line.End > 0 && filePosition.Column.Start > 0 && filePosition.Column.End > 0 && filePosition.Filename != ""
}

func extractPackageLocations(hostPath string, pkgInfos models.PackageInfo) models.PackageLocations {
	blockFilename := filepath.ToSlash(strings.TrimPrefix(pkgInfos.BlockLocation.Filename, hostPath))
	locations := models.PackageLocations{
		Block: models.PackageLocation{
			Filename:    blockFilename,
			LineStart:   pkgInfos.BlockLocation.Line.Start,
			LineEnd:     pkgInfos.BlockLocation.Line.End,
			ColumnStart: pkgInfos.BlockLocation.Column.Start,
			ColumnEnd:   pkgInfos.BlockLocation.Column.End,
		},
	}

	locations.Name = mapToPackageLocation(hostPath, pkgInfos.NameLocation)
	locations.Version = mapToPackageLocation(hostPath, pkgInfos.VersionLocation)

	return locations
}

func mapToPackageLocation(hostPath string, location *models.FilePosition) *models.PackageLocation {
	if location == nil || !isLocationExtractedSuccessfully(*location) {
		return nil
	}
	filename := filepath.ToSlash(strings.TrimPrefix(location.Filename, hostPath))

	return &models.PackageLocation{
		Filename:    filename,
		LineStart:   location.Line.Start,
		LineEnd:     location.Line.End,
		ColumnStart: location.Column.Start,
		ColumnEnd:   location.Column.End,
	}
}
