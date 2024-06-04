package grouper

import (
	"github.com/google/osv-scanner/internal/utility/fileposition"
	"github.com/google/osv-scanner/pkg/models"
	"github.com/google/osv-scanner/pkg/reporter/purl"
)

func GroupByPURL(packageSources []models.PackageSource, considerScanPathAsRoot bool, pathRelativeToScanDir bool) map[string]models.PackageDetails {
	uniquePackages := make(map[string]models.PackageDetails)

	for _, packageSource := range packageSources {
		pkgLocations := make(map[string]bool)
		for _, pkg := range packageSource.Packages {
			packageURL := purl.From(pkg.Package)
			if packageURL == nil {
				continue
			}
			existingPackage, packageExists := uniquePackages[packageURL.ToString()]
			isLocationExtracted := isLocationExtractedSuccessfully(pkg.Package.LockfileLocations.Block)
			locations := extractPackageLocations(packageSource.Source.ScanPath, pkg.Package.LockfileLocations, considerScanPathAsRoot, pathRelativeToScanDir)
			if pkg.Package.SourcefileLocations != nil {
				sourcefileLocations := extractPackageLocations(packageSource.Source.ScanPath, *pkg.Package.SourcefileLocations, considerScanPathAsRoot, pathRelativeToScanDir)
				locations.Sourcefile = &models.SourcefilePackageLocations{
					Block:     sourcefileLocations.Block,
					Name:      sourcefileLocations.Name,
					Namespace: sourcefileLocations.Namespace,
					Version:   sourcefileLocations.Version,
				}
			}

			if packageExists && isLocationExtracted {
				locationHash := locations.Block.Hash()
				if _, isLocationDuplicated := pkgLocations[locationHash]; isLocationDuplicated {
					continue
				}

				// Package exists, locations exists and is not a duplicate => we need to add it
				existingPackage.Locations = append(existingPackage.Locations, locations)
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
					// We add locations only if it has been extracted successfully
					newPackage.Locations = append(newPackage.Locations, locations)
					pkgLocations[locations.Block.Hash()] = true
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

func extractPackageLocations(scanPath string, fileLocations models.FileLocations, considerScanPathAsRoot bool, pathRelativeToScanDir bool) models.PackageLocations {
	locations := models.PackageLocations{
		Block: models.PackageLocation{
			Filename:    fileposition.RemoveHostPath(scanPath, fileLocations.Block.Filename, considerScanPathAsRoot, pathRelativeToScanDir),
			LineStart:   fileLocations.Block.Line.Start,
			LineEnd:     fileLocations.Block.Line.End,
			ColumnStart: fileLocations.Block.Column.Start,
			ColumnEnd:   fileLocations.Block.Column.End,
		},
	}

	locations.Name = mapToPackageLocation(scanPath, fileLocations.Name, considerScanPathAsRoot, pathRelativeToScanDir)
	locations.Version = mapToPackageLocation(scanPath, fileLocations.Version, considerScanPathAsRoot, pathRelativeToScanDir)

	return locations
}

func mapToPackageLocation(scanPath string, location *models.FilePosition, considerScanPathAsRoot bool, pathRelativeToScanDir bool) *models.PackageLocation {
	if location == nil || !isLocationExtractedSuccessfully(*location) {
		return nil
	}

	return &models.PackageLocation{
		Filename:    fileposition.RemoveHostPath(scanPath, location.Filename, considerScanPathAsRoot, pathRelativeToScanDir),
		LineStart:   location.Line.Start,
		LineEnd:     location.Line.End,
		ColumnStart: location.Column.Start,
		ColumnEnd:   location.Column.End,
	}
}
