package lockfile

import (
	"bufio"
	"fmt"
	"github.com/google/osv-scanner/internal/utility/fileposition"
	"github.com/google/osv-scanner/pkg/models"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/osv-scanner/internal/cachedregexp"
)

const YarnEcosystem = NpmEcosystem

type YarnPackage struct {
	Name           string
	Version        string
	TargetVersions []string
	Resolution     string
	models.FilePosition
}

func shouldSkipYarnLine(line string) bool {
	return line == "" || strings.HasPrefix(line, "#")
}

func parseYarnPackageGroup(group []string, lineStart int, lineEnd int, columnStart int, columnEnd int) YarnPackage {
	name, targetVersions := extractYarnPackageNameAndTargetVersions(group[0])
	return YarnPackage{
		Name:           name,
		Version:        determineYarnPackageVersion(group),
		TargetVersions: targetVersions,
		Resolution:     determineYarnPackageResolution(group),
		FilePosition: models.FilePosition{
			Line:   models.Position{Start: lineStart, End: lineEnd},
			Column: models.Position{Start: columnStart, End: columnEnd},
		},
	}
}

func groupYarnPackageLines(scanner *bufio.Scanner) []YarnPackage {
	var groups []YarnPackage
	var group []string
	var lineNumber, lineStart, lineEnd, columnStart, columnEnd int

	var line string
	for scanner.Scan() {
		lineNumber++
		line = scanner.Text()

		if shouldSkipYarnLine(line) {
			continue
		}

		// represents the lineStart of a new dependency
		if !strings.HasPrefix(line, " ") {
			if len(group) > 0 {
				groups = append(groups, parseYarnPackageGroup(group, lineStart, lineEnd, columnStart, columnEnd))
			}
			lineStart = lineNumber
			columnStart = fileposition.GetFirstNonEmptyCharacterIndexInLine(line)
			group = make([]string, 0)
		}

		lineEnd = lineNumber
		columnEnd = fileposition.GetLastNonEmptyCharacterIndexInLine(line)
		group = append(group, line)
	}

	if len(group) > 0 {
		lineEnd = lineNumber
		columnEnd = fileposition.GetLastNonEmptyCharacterIndexInLine(line)
		groups = append(groups, parseYarnPackageGroup(group, lineStart, lineEnd, columnStart, columnEnd))
	}

	return groups
}

func extractYarnPackageNameAndTargetVersions(str string) (string, []string) {
	str = strings.ReplaceAll(str, "\"", "")
	str = strings.TrimSuffix(str, ":")
	parts := strings.Split(str, ",")

	var name, right string
	var isScoped bool
	var targetVersions = make([]string, 0)

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
		Ecosystem:      YarnEcosystem,
		CompareAs:      YarnEcosystem,
		Commit:         tryExtractCommit(dependency.Resolution),
		LockfileLocations: Locations{
			Block: models.FilePosition{
				Line:     dependency.Line,
				Column:   dependency.Column,
				Filename: dependency.Filename,
			},
		},
	}
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

	if err := scanner.Err(); err != nil {
		return []PackageDetails{}, fmt.Errorf("error while scanning %s: %w", f.Path(), err)
	}

	packages := make([]PackageDetails, 0, len(yarnPackages))

	for _, yarnPackage := range yarnPackages {
		if yarnPackage.Name == "__metadata" {
			continue
		}

		yarnPackage.FilePosition.Filename = f.Path()
		packages = append(packages, parseYarnPackage(yarnPackage))
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
