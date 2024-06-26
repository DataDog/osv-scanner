package lockfile

import (
	"encoding/json"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"golang.org/x/exp/maps"

	"github.com/google/osv-scanner/internal/utility/fileposition"

	"github.com/google/osv-scanner/pkg/models"
)

type NpmLockDependency struct {
	// For an aliased package, Version is like "npm:[name]@[version]"
	Version      string                        `json:"version"`
	Dependencies map[string]*NpmLockDependency `json:"dependencies,omitempty"`

	Dev      bool `json:"dev,omitempty"`
	Optional bool `json:"optional,omitempty"`

	Requires map[string]string `json:"requires,omitempty"`

	models.FilePosition
}

func (dep *NpmLockDependency) GetNestedDependencies() map[string]*models.FilePosition {
	result := make(map[string]*models.FilePosition)
	for key, value := range dep.Dependencies {
		result[key] = &value.FilePosition
	}

	return result
}

type NpmLockPackage struct {
	// For an aliased package, Name is the real package name
	Name     string `json:"name"`
	Version  string `json:"version"`
	Resolved string `json:"resolved"`

	Dependencies         map[string]string `json:"dependencies,omitempty"`
	DevDependencies      map[string]string `json:"devDependencies,omitempty"`
	OptionalDependencies map[string]string `json:"optionalDependencies,omitempty"`
	PeerDependencies     map[string]string `json:"peerDependencies,omitempty"`

	Dev         bool `json:"dev,omitempty"`
	DevOptional bool `json:"devOptional,omitempty"`
	Optional    bool `json:"optional,omitempty"`

	Link bool `json:"link,omitempty"`

	models.FilePosition
}

type NpmLockfile struct {
	Version    int `json:"lockfileVersion"`
	SourceFile string
	// npm v1- lockfiles use "dependencies"
	Dependencies map[string]*NpmLockDependency `json:"dependencies"`
	// npm v2+ lockfiles use "packages"
	Packages map[string]*NpmLockPackage `json:"packages,omitempty"`
}

const NpmEcosystem Ecosystem = "npm"

func (dep *NpmLockDependency) depGroups() []string {
	if dep.Dev && dep.Optional {
		return []string{"dev", "optional"}
	}
	if dep.Dev {
		return []string{"dev"}
	}
	if dep.Optional {
		return []string{"optional"}
	}

	return nil
}

func parseNpmLockDependencies(dependencies map[string]*NpmLockDependency, path string) map[string]PackageDetails {
	details := map[string]PackageDetails{}

	keys := reflect.ValueOf(dependencies).MapKeys()
	keysOrder := func(i, j int) bool { return keys[i].Interface().(string) < keys[j].Interface().(string) }
	sort.Slice(keys, keysOrder)

	for _, key := range keys {
		name := key.Interface().(string)
		detail := dependencies[name]
		if detail.Dependencies != nil {
			maps.Copy(details, parseNpmLockDependencies(detail.Dependencies, path))
		}

		version := detail.Version
		finalVersion := version
		commit := ""

		// If the package is aliased, get the name and version
		if strings.HasPrefix(detail.Version, "npm:") {
			i := strings.LastIndex(detail.Version, "@")
			name = detail.Version[4:i]
			finalVersion = detail.Version[i+1:]
		}

		// we can't resolve a version from a "file:" dependency
		if strings.HasPrefix(detail.Version, "file:") {
			finalVersion = ""
			version = ""
		} else {
			commit = tryExtractCommit(detail.Version)

			// if there is a commit, we want to deduplicate based on that rather than
			// the version (the versions must match anyway for the commits to match)
			//
			// we also don't actually know what the "version" is, so blank it
			if commit != "" {
				finalVersion = ""
				version = commit
			}
		}

		details[name+"@"+version] = PackageDetails{
			Name:      name,
			Version:   finalVersion,
			Ecosystem: NpmEcosystem,
			CompareAs: NpmEcosystem,
			BlockLocation: models.FilePosition{
				Line:     detail.Line,
				Column:   detail.Column,
				Filename: path,
			},
			Commit:    commit,
			DepGroups: detail.depGroups(),
		}
	}

	return details
}

func extractNpmPackageName(name string) string {
	maybeScope := path.Base(path.Dir(name))
	pkgName := path.Base(name)

	if strings.HasPrefix(maybeScope, "@") {
		pkgName = maybeScope + "/" + pkgName
	}

	return pkgName
}

func extractRootKeyPackageName(name string) string {
	_, right, _ := strings.Cut(name, "/")
	return right
}

func (pkg NpmLockPackage) depGroups() []string {
	if pkg.Dev {
		return []string{"dev"}
	}
	if pkg.Optional {
		return []string{"optional"}
	}
	if pkg.DevOptional {
		return []string{"dev", "optional"}
	}

	return nil
}

func parseNpmLockPackages(packages map[string]*NpmLockPackage, path string) map[string]PackageDetails {
	details := map[string]PackageDetails{}

	keys := reflect.ValueOf(packages).MapKeys()
	keysOrder := func(i, j int) bool { return keys[i].Interface().(string) < keys[j].Interface().(string) }
	sort.Slice(keys, keysOrder)

	for _, key := range keys {
		namePath := key.Interface().(string)
		detail := packages[namePath]
		if namePath == "" {
			continue
		}

		finalName := detail.Name
		if finalName == "" {
			finalName = extractNpmPackageName(namePath)
		}

		finalVersion := detail.Version

		commit := tryExtractCommit(detail.Resolved)

		// if there is a commit, we want to deduplicate based on that rather than
		// the version (the versions must match anyway for the commits to match)
		if commit != "" {
			finalVersion = commit
		}

		if finalVersion == "" {
			// If version and commit are not set in the lockfile, it means the package is defined locally
			// with its own package.json, without any version defined for it, lets default on 0.0.0
			detail.Version = "0.0.0"
		}

		// Element "" in packages, contains in its dependencies/devDependencies
		// the dependencies with the version written as it appears in the package.json
		var targetVersions []string
		var targetVersion string
		rootKey := extractRootKeyPackageName(namePath)
		if p, ok := packages[""]; ok {
			if dep, ok := p.Dependencies[rootKey]; ok {
				targetVersion = dep
			} else if devDep, ok := p.DevDependencies[rootKey]; ok {
				targetVersion = devDep
			}
		}

		if len(targetVersion) > 0 {
			// Clean aliased target version
			if strings.HasPrefix(targetVersion, "npm:") {
				_, targetVersion, _ = strings.Cut(targetVersion, "@")
			}

			// Clean some prefixes that may not be included in package.json
			prefixes := []string{"file", "link", "portal"}
			for _, prefix := range prefixes {
				if strings.HasPrefix(targetVersion, prefix+":") {
					targetVersion = strings.TrimPrefix(targetVersion, prefix+":")
					targetVersion = strings.TrimPrefix(targetVersion, "./")
				}
			}

			targetVersions = []string{targetVersion}
		}

		_, exists := details[finalName+"@"+finalVersion]
		if !exists && !detail.Link {
			details[finalName+"@"+finalVersion] = PackageDetails{
				Name:           finalName,
				Version:        detail.Version,
				TargetVersions: targetVersions,
				Ecosystem:      NpmEcosystem,
				CompareAs:      NpmEcosystem,
				BlockLocation: models.FilePosition{
					Line:     detail.Line,
					Column:   detail.Column,
					Filename: path,
				},
				Commit:    commit,
				DepGroups: detail.depGroups(),
			}
		}
	}

	return details
}

func parseNpmLock(lockfile NpmLockfile, lines []string) map[string]PackageDetails {
	if lockfile.Packages != nil {
		fileposition.InJSON("packages", lockfile.Packages, lines, 0)

		return parseNpmLockPackages(lockfile.Packages, lockfile.SourceFile)
	}

	fileposition.InJSON("dependencies", lockfile.Dependencies, lines, 0)

	return parseNpmLockDependencies(lockfile.Dependencies, lockfile.SourceFile)
}

type NpmLockExtractor struct {
	WithMatcher
}

func (e NpmLockExtractor) ShouldExtract(path string) bool {
	return filepath.Base(path) == "package-lock.json"
}

func (e NpmLockExtractor) Extract(f DepFile) ([]PackageDetails, error) {
	var parsedLockfile *NpmLockfile

	contentBytes, err := io.ReadAll(f)
	if err != nil {
		return []PackageDetails{}, fmt.Errorf("could not read from %s: %w", f.Path(), err)
	}
	contentString := string(contentBytes)
	lines := strings.Split(contentString, "\n")
	decoder := json.NewDecoder(strings.NewReader(contentString))

	if err := decoder.Decode(&parsedLockfile); err != nil {
		return []PackageDetails{}, fmt.Errorf("could not extract from %s: %w", f.Path(), err)
	}
	parsedLockfile.SourceFile = f.Path()

	return maps.Values(parseNpmLock(*parsedLockfile, lines)), nil
}

var NpmExtractor = NpmLockExtractor{
	WithMatcher{Matcher: PackageJSONMatcher{}},
}

//nolint:gochecknoinits
func init() {
	registerExtractor("package-lock.json", NpmExtractor)
}

func ParseNpmLock(pathToLockfile string) ([]PackageDetails, error) {
	return extractFromFile(pathToLockfile, NpmExtractor)
}
