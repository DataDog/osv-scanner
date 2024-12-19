package lockfile

import (
	"errors"
	"fmt"
	"github.com/google/osv-scanner/internal/cachedregexp"
	"github.com/google/osv-scanner/pkg/models"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
	"io"
	"strings"
)

type PnpmV9LockPackageResolution struct {
	Tarball string `yaml:"tarball"`
	Commit  string `yaml:"commit"`
	Repo    string `yaml:"repo"`
	Type    string `yaml:"type"`
}

type PnpmV9LockDependency struct {
	Specifier string `yaml:"specifier"`
	Version   string `yaml:"version"`
}

type PnpmV9Package struct {
	Resolution map[string]string `yaml:"resolution"`
	Version    string            `yaml:"version"`
}

type (
	PnpmV9LockPackages map[string]PnpmV9Package
	PnpmV9Dependencies map[string]PnpmV9LockDependency
)

type PnpmV9Importers struct {
	Dependencies         PnpmV9Dependencies `yaml:"dependencies,omitempty"`
	OptionalDependencies PnpmV9Dependencies `yaml:"optionalDependencies,omitempty"`
	DevDependencies      PnpmV9Dependencies `yaml:"devDependencies,omitempty"`
}

type PnpmSnapshot struct {
	Dependencies         map[string]string `yaml:"dependencies"`
	OptionalDependencies map[string]string `yaml:"optionalDependencies"`
}

type PnpmV9Lockfile struct {
	Version   string                     `yaml:"lockfileVersion"`
	Importers map[string]PnpmV9Importers `yaml:"importers,omitempty"`
	Packages  PnpmV9LockPackages         `yaml:"packages,omitempty"`
	Snapshots map[string]PnpmSnapshot    `yaml:"snapshots,omitempty"`
}

type PnpmDirectDependency struct {
	Pkg PackageDetails
	Dep PnpmV9LockDependency
}

func getCleanedVersion(lockfile PnpmV9Lockfile, name, version string) string {
	if strings.HasPrefix(version, "https://codeload.github.com") {
		// This is a link to a tarball, not a version we need to check the resolved version in the package section
		if pkg, ok := lockfile.Packages[name+"@"+version]; ok {
			return pkg.Version
		}
		return ""
	}
	return strings.Split(version, "(")[0]
}

func getCommitFromVersion(version string) string {
	if strings.HasPrefix(version, "https://codeload.github.com") {
		re := cachedregexp.MustCompile(`https://codeload\.github\.com(?:/[\w-.]+){2}/tar\.gz/(\w+)$`)
		matched := re.FindStringSubmatch(version)

		if matched != nil {
			return matched[1]
		}
	}
	return ""
}

func mergeSlices(slices ...[]string) []string {
	result := make(map[string]bool)
	for _, slice := range slices {
		for _, item := range slice {
			result[item] = true
		}
	}

	return maps.Keys(result)
}

func addDependencyToPackageDetails(dependency PackageDetails, deps map[string]PackageDetails) map[string]PackageDetails {
	key := dependency.Name + "@" + dependency.Version

	if dep, exists := deps[key]; exists {
		newDepGroups := mergeSlices(dep.DepGroups, dependency.DepGroups)
		newTargetedVersions := mergeSlices(dep.TargetVersions, dependency.TargetVersions)

		if len(newTargetedVersions) > 0 {
			dep.DepGroups = newDepGroups
		}
		if len(newTargetedVersions) > 0 {
			dep.TargetVersions = newTargetedVersions
		}
		dep.IsDirect = dep.IsDirect || dependency.IsDirect
		deps[key] = dep
	} else {
		deps[key] = dependency
	}
	return deps
}

func extractTransitiveDeps(lockfile PnpmV9Lockfile, root PnpmDirectDependency, targetedKey string, deps map[string]PackageDetails) map[string]PackageDetails {
	// Need to look at dependencies
	visitedSnapshots := make(map[string]bool)
	snapshotQueue := make([]string, 1)
	snapshotQueue[0] = targetedKey

	for len(snapshotQueue) > 0 {
		targetedKey = snapshotQueue[0]
		snapshotQueue = snapshotQueue[1:]

		if _, visited := visitedSnapshots[targetedKey]; visited {
			continue
		}

		visitedSnapshots[targetedKey] = true
		snapshot, ok := lockfile.Snapshots[targetedKey]

		if !ok {
			continue
		}

		for depName, depVersion := range snapshot.Dependencies {
			transitiveDep := PackageDetails{
				Name:           depName,
				Version:        getCleanedVersion(lockfile, depName, depVersion),
				Commit:         getCommitFromVersion(depVersion),
				Ecosystem:      PnpmEcosystem,
				CompareAs:      PnpmEcosystem,
				DepGroups:      root.Pkg.DepGroups,
				PackageManager: models.Pnpm,
				IsDirect:       false,
			}
			addDependencyToPackageDetails(transitiveDep, deps)
			childKey := depName + "@" + depVersion
			snapshotQueue = append(snapshotQueue, childKey)
		}

		for depName, depVersion := range snapshot.OptionalDependencies {
			transitiveDep := PackageDetails{
				Name:           depName,
				Version:        getCleanedVersion(lockfile, depName, depVersion),
				Commit:         getCommitFromVersion(depVersion),
				Ecosystem:      PnpmEcosystem,
				CompareAs:      PnpmEcosystem,
				DepGroups:      root.Pkg.DepGroups,
				PackageManager: models.Pnpm,
				IsDirect:       false,
			}
			addDependencyToPackageDetails(transitiveDep, deps)
			childKey := depName + "@" + depVersion
			snapshotQueue = append(snapshotQueue, childKey)
		}
	}

	return deps
}

func extractDirectDependencies(lockfile PnpmV9Lockfile, roots []PnpmDirectDependency, dependencies PnpmV9Dependencies, depGroup string) []PnpmDirectDependency {
	for dependencyName, dependency := range dependencies {
		roots = append(roots, PnpmDirectDependency{
			Pkg: PackageDetails{
				Name:           dependencyName,
				Version:        getCleanedVersion(lockfile, dependencyName, dependency.Version),
				Commit:         getCommitFromVersion(dependency.Version),
				TargetVersions: []string{dependency.Specifier},
				Ecosystem:      PnpmEcosystem,
				CompareAs:      PnpmEcosystem,
				DepGroups:      []string{depGroup},
				PackageManager: models.Pnpm,
				IsDirect:       true,
			},
			Dep: dependency,
		})
	}
	return roots
}

func parsePnpmV9Lock(lockfile PnpmV9Lockfile) []PackageDetails {
	// First create the deps tree
	// To do so, first look at the packages list, for each package, look into the importers
	// If present in the importers => its direct and we know its scope
	// Then looking at snapshot, we can build its branch

	// Going through the importers to get a direct (prod or dev), then finding the transitives in the snapshot
	directDependencies := make([]PnpmDirectDependency, 0)
	for _, importer := range lockfile.Importers {
		directDependencies = extractDirectDependencies(lockfile, directDependencies, importer.Dependencies, "prod")
		directDependencies = extractDirectDependencies(lockfile, directDependencies, importer.OptionalDependencies, "optional")
		directDependencies = extractDirectDependencies(lockfile, directDependencies, importer.DevDependencies, "dev")
	}

	packages := make(map[string]PackageDetails)
	for _, direct := range directDependencies {
		packages = addDependencyToPackageDetails(direct.Pkg, packages)
		packages = extractTransitiveDeps(lockfile, direct, direct.Pkg.Name+"@"+direct.Dep.Version, packages)
	}
	return maps.Values(packages)
}

func (e PnpmLockExtractor) extractV9(f DepFile) ([]PackageDetails, error) {
	var parsedLockfile *PnpmV9Lockfile

	err := yaml.NewDecoder(f).Decode(&parsedLockfile)

	if err != nil && !errors.Is(err, io.EOF) {
		return []PackageDetails{}, fmt.Errorf("could not extract from %s: %w", f.Path(), err)
	}

	// this will happen if the file is empty
	if parsedLockfile == nil {
		parsedLockfile = &PnpmV9Lockfile{}
	}

	return parsePnpmV9Lock(*parsedLockfile), nil
}
