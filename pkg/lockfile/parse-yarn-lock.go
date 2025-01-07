package lockfile

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/osv-scanner/pkg/models"

	"github.com/google/osv-scanner/internal/cachedregexp"
)

const YarnEcosystem = NpmEcosystem

type YarnDependency struct {
	Name     string
	Version  string
	Registry string
}

type YarnPackage struct {
	Name           string
	Version        string
	TargetVersions []string
	Resolution     string
	Dependencies   []YarnDependency
}

func shouldSkipYarnLine(line string) bool {
	return line == "" || strings.HasPrefix(line, "#")
}

func parseYarnPackageGroup(group []string) YarnPackage {
	name, targetVersions := extractYarnPackageNameAndTargetVersions(group[0])

	return YarnPackage{
		Name:           name,
		Version:        determineYarnPackageVersion(group),
		TargetVersions: targetVersions,
		Resolution:     determineYarnPackageResolution(group),
		Dependencies:   determineYarnPackageDependencies(group),
	}
}

func groupYarnPackageLines(scanner *bufio.Scanner) []YarnPackage {
	var groups []YarnPackage
	var group []string

	var line string
	for scanner.Scan() {
		line = scanner.Text()

		if shouldSkipYarnLine(line) {
			continue
		}

		// represents the lineStart of a new dependency
		if !strings.HasPrefix(line, " ") {
			if len(group) > 0 {
				groups = append(groups, parseYarnPackageGroup(group))
			}
			group = make([]string, 0)
		}

		group = append(group, line)
	}

	if len(group) > 0 {
		groups = append(groups, parseYarnPackageGroup(group))
	}

	return groups
}

func extractYarnPackageNameAndTargetVersions(str string) (string, []string) {
	str = strings.ReplaceAll(str, "\"", "")
	str = strings.TrimSuffix(str, ":")
	parts := strings.Split(str, ",")

	var name, right string
	var isScoped bool
	targetVersions := make([]string, 0)

	for _, part := range parts {
		part = strings.TrimPrefix(part, " ")
		partIsScoped := strings.HasPrefix(part, "@")
		isScoped = isScoped || partIsScoped

		if partIsScoped {
			part = strings.TrimPrefix(part, "@")
		}

		_name, _right, _ := strings.Cut(part, "@")
		if len(name) == 0 {
			name = _name
		}
		right = _right

		if strings.HasPrefix(right, "npm:") {
			right = strings.TrimPrefix(right, "npm:")
			if strings.Contains(right, "@") {
				resolvedName, resolvedTargetVersions := extractYarnPackageNameAndTargetVersions(right)
				name = resolvedName
				targetVersions = append(targetVersions, resolvedTargetVersions...)

				continue
			}
		}

		// for yarn v2 - it could include these prefixes even when they are not included in package.json
		prefixes := []string{"file", "link", "portal"}
		for _, prefix := range prefixes {
			if strings.HasPrefix(right, prefix+":") {
				right = strings.TrimPrefix(right, prefix+":")
			}
		}

		// for yarn v2 - "file:path/to/dir::locator=...%40workspace%3A.": -> file:path/to/dir
		right, _, _ = strings.Cut(right, "::locator")

		targetVersions = append(targetVersions, right)
	}

	if isScoped {
		name = "@" + name
	}

	return name, targetVersions
}

func determineYarnPackageVersion(group []string) string {
	re := cachedregexp.MustCompile(`^ {2}"?version"?:? "?([\w-.+]+)"?$`)

	for _, s := range group {
		matched := re.FindStringSubmatch(s)

		if matched != nil {
			return matched[1]
		}
	}

	// todo: decide what to do here - maybe panic...?
	return ""
}

// You can find the line parsing regex in action here: https://regex101.com/r/QoJ3b7/3
func determineYarnPackageDependencies(group []string) []YarnDependency {
	indentCount := -1
	results := make([]YarnDependency, 0)
	lineParsing := cachedregexp.MustCompile(`^"?(?P<package_name>[^\s":]+)"?\s*:?\s*"?(?P<targeted_version>[^"\n]+)"?$`)

	for _, line := range group {
		cleaned := strings.TrimSpace(line)
		if strings.HasPrefix(cleaned, "dependencies") {
			// start of the dependencies section
			indentCount = len(line) - len(cleaned)
		} else if indentCount != -1 && len(line)-len(cleaned) == indentCount {
			// end of the dependencies section, we can stop there
			break
		} else if indentCount != -1 {
			// A line inside the dependencies section, lets parse it
			match := lineParsing.FindStringSubmatch(cleaned)
			name, version, registry := match[1], match[2], ""

			if strings.Contains(version, ":") {
				registry, version, _ = strings.Cut(version, ":")
			} else {
				registry = "npm"
			}

			results = append(results, YarnDependency{
				Name:     name,
				Version:  version,
				Registry: registry,
			})
		}
	}

	return results
}

func determineYarnPackageResolution(group []string) string {
	re := cachedregexp.MustCompile(`^ {2}"?(?:resolution:|resolved)"? "([^ '"]+)"$`)

	for _, s := range group {
		matched := re.FindStringSubmatch(s)

		if matched != nil {
			return matched[1]
		}
	}

	// todo: decide what to do here - maybe panic...?
	return ""
}

func tryExtractCommit(resolution string) string {
	// language=GoRegExp
	matchers := []string{
		// ssh://...
		// git://...
		// git+ssh://...
		// git+https://...
		`(?:^|.+@)(?:git(?:\+(?:ssh|https))?|ssh)://.+#(\w+)$`,
		// https://....git/...
		`(?:^|.+@)https://.+\.git#(\w+)$`,
		`https://codeload\.github\.com(?:/[\w-.]+){2}/tar\.gz/(\w+)$`,
		`.+#commit[:=](\w+)$`,
		// github:...
		// gitlab:...
		// bitbucket:...
		`^(?:github|gitlab|bitbucket):.+#(\w+)$`,
	}

	for _, matcher := range matchers {
		re := cachedregexp.MustCompile(matcher)
		matched := re.FindStringSubmatch(resolution)

		if matched != nil {
			return matched[1]
		}
	}

	u, err := url.Parse(resolution)

	if err == nil {
		gitRepoHosts := []string{
			"bitbucket.org",
			"github.com",
			"gitlab.com",
		}

		for _, host := range gitRepoHosts {
			if u.Host != host {
				continue
			}

			if u.RawQuery != "" {
				queries := u.Query()

				if queries.Has("ref") {
					return queries.Get("ref")
				}
			}

			return u.Fragment
		}
	}

	return ""
}

func buildDependencyTree(rootPkgName, rootPkgTargetVersion, rootPkgRegistry string, dependencies map[string]YarnPackage, packagesIndex map[string]*PackageDetails) []*PackageDetails {
	results := make([]*PackageDetails, 0)
	pkg, ok := dependencies[rootPkgName+"@"+rootPkgTargetVersion]
	if !ok {
		pkg, ok = dependencies[rootPkgName+"@"+rootPkgRegistry+":"+rootPkgTargetVersion]
		if !ok {
			return []*PackageDetails{}
		}
	}

	for _, dependency := range pkg.Dependencies {
		dependentPackage, ok := dependencies[dependency.Name+"@"+dependency.Version]
		if !ok {
			dependentPackage, ok = dependencies[dependency.Name+"@"+dependency.Registry+":"+dependency.Version]
			if !ok {
				continue
			}
		}
		dep, exists := packagesIndex[dependentPackage.Name+"@"+dependentPackage.Version]
		if exists {
			results = append(results, dep)
		}
	}

	return results
}

func parseYarnPackage(dependency YarnPackage) PackageDetails {
	if dependency.Version == "" {
		_, _ = fmt.Fprintf(
			os.Stderr,
			"Failed to determine version of %s while parsing a yarn.lock - please report this!\n",
			dependency.Name,
		)
	}

	return PackageDetails{
		Name:           dependency.Name,
		Version:        dependency.Version,
		TargetVersions: dependency.TargetVersions,
		PackageManager: models.Yarn,
		Ecosystem:      YarnEcosystem,
		CompareAs:      YarnEcosystem,
		Commit:         tryExtractCommit(dependency.Resolution),
	}
}

func indexByTargetVersion(packages []YarnPackage) map[string]YarnPackage {
	index := make(map[string]YarnPackage)

	for _, pkg := range packages {
		for _, targetVersion := range pkg.TargetVersions {
			index[pkg.Name+"@"+targetVersion] = pkg
		}
	}
	return index
}

func indexByNameAndVersions(packages []PackageDetails) map[string]*PackageDetails {
	result := make(map[string]*PackageDetails)
	for index, pkg := range packages {
		result[pkg.Name+"@"+pkg.Version] = &packages[index]
	}
	return result
}

type YarnLockExtractor struct {
	WithMatcher
}

func (e YarnLockExtractor) ShouldExtract(path string) bool {
	return filepath.Base(path) == "yarn.lock"
}

func (e YarnLockExtractor) Extract(f DepFile) ([]PackageDetails, error) {
	scanner := bufio.NewScanner(f)

	yarnPackages := groupYarnPackageLines(scanner)
	yarnPackageIndex := indexByTargetVersion(yarnPackages)

	// Use this index to build all subtrees (trees from each package)
	// Then use all this in the matcher to know is-dev / is-direct and propagate it everywhere

	if err := scanner.Err(); err != nil {
		return []PackageDetails{}, fmt.Errorf("error while scanning %s: %w", f.Path(), err)
	}

	packages := make([]PackageDetails, 0, len(yarnPackages))

	for _, yarnPackage := range yarnPackages {
		if yarnPackage.Name == "__metadata" {
			continue
		}

		packages = append(packages, parseYarnPackage(yarnPackage))
	}
	pkgIndex := indexByNameAndVersions(packages)
	for index, pkg := range packages {
		packages[index].Dependencies = buildDependencyTree(pkg.Name, pkg.TargetVersions[0], "npm", yarnPackageIndex, pkgIndex)
	}

	return packages, nil
}

var YarnExtractor = YarnLockExtractor{
	WithMatcher{Matcher: PackageJSONMatcher{}},
}

//nolint:gochecknoinits
func init() {
	registerExtractor("yarn.lock", YarnExtractor)
}

func ParseYarnLock(pathToLockfile string) ([]PackageDetails, error) {
	return extractFromFile(pathToLockfile, YarnExtractor)
}
