package grouper

import (
	"github.com/google/osv-scanner/internal/utility/purl"
	"github.com/google/osv-scanner/pkg/models"
	"slices"
)

func GroupByPURL(packageSources []models.PackageSource) map[string]models.PackageVulns {
	uniquePackages := make(map[string]models.PackageVulns)

	for _, packageSource := range packageSources {
		for _, pkg := range packageSource.Packages {
			packageURL := purl.From(pkg.Package)
			if packageURL == nil {
				continue
			}
			packageVulns, packageExists := uniquePackages[packageURL.ToString()]
			if packageExists {
				// Entry already exists, we need to merge slices which are not expected to be the exact same
				packageVulns.DepGroups = append(packageVulns.DepGroups, pkg.DepGroups...)

				uniquePackages[packageURL.ToString()] = packageVulns
			} else {
				// Entry does not exists yet, lets create it
				newPackageVuln := models.PackageVulns{
					Package: models.PackageInfo{
						Name:      pkg.Package.Name,
						Version:   pkg.Package.Version,
						Ecosystem: pkg.Package.Ecosystem,
					},
					DepGroups:         slices.Clone(pkg.DepGroups),
					Vulnerabilities:   slices.Clone(pkg.Vulnerabilities),
					Groups:            slices.Clone(pkg.Groups),
					Licenses:          slices.Clone(pkg.Licenses),
					LicenseViolations: slices.Clone(pkg.LicenseViolations),
				}

				uniquePackages[packageURL.ToString()] = newPackageVuln
			}
		}
	}

	return uniquePackages
}
