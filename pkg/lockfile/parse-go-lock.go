package lockfile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/module"

	"github.com/google/osv-scanner/pkg/models"

	"github.com/google/osv-scanner/internal/semantic"
	"golang.org/x/mod/modfile"
)

const GoEcosystem Ecosystem = "Go"
const unknownVersion string = "v0.0.0-unknown-version"

func deduplicatePackages(packages map[string]PackageDetails) map[string]PackageDetails {
	details := map[string]PackageDetails{}

	for _, detail := range packages {
		details[detail.Name+"@"+detail.Version] = detail
	}

	return details
}

type GoLockExtractor struct{}

func defaultNonCanonicalVersions(path, version string) (string, error) {
	resolvedVersion := module.CanonicalVersion(version)

	// If the resolvedVersion is not canonical, we try to find the major resolvedVersion in the path and report that
	if resolvedVersion == "" {
		_, pathMajor, ok := module.SplitPathVersion(path)
		if ok {
			resolvedVersion = module.PathMajorPrefix(pathMajor)
		}
	}

	if resolvedVersion == "" {
		// When a version is not resolved, we still have to default to at least v0 and then filter it as the mod package check for the major path which have to be valid
		_, _ = fmt.Fprintf(os.Stderr, "%s@%s is not a canonical path, defaulting to an empty version\n", path, resolvedVersion)
		resolvedVersion = unknownVersion
	}

	return resolvedVersion, nil
}

func (e GoLockExtractor) ShouldExtract(path string) bool {
	return filepath.Base(path) == "go.mod"
}

func (e GoLockExtractor) Extract(f DepFile) ([]PackageDetails, error) {
	var parsedLockfile *modfile.File

	b, err := io.ReadAll(f)

	if err == nil {
		parsedLockfile, err = modfile.Parse(f.Path(), b, defaultNonCanonicalVersions)
	}

	if err != nil {
		return []PackageDetails{}, fmt.Errorf("could not extract from %s: %w", f.Path(), err)
	}

	packages := map[string]PackageDetails{}

	for _, require := range parsedLockfile.Require {
		var start = require.Syntax.Start
		var end = require.Syntax.End
		var version = strings.TrimPrefix(require.Mod.Version, "v")

		if require.Mod.Version == unknownVersion {
			version = ""
		}
		packages[require.Mod.Path+"@"+require.Mod.Version] = PackageDetails{
			Name:      require.Mod.Path,
			Version:   version,
			Ecosystem: GoEcosystem,
			CompareAs: GoEcosystem,
			Line:      models.Position{Start: start.Line, End: end.Line},
			Column:    models.Position{Start: start.LineRune, End: end.LineRune},
		}
	}

	for _, replace := range parsedLockfile.Replace {
		var start = replace.Syntax.Start
		var end = replace.Syntax.End
		var replacements []string

		if replace.Old.Version == "" {
			// If the left version is omitted, all versions of the module are replaced.
			for k, pkg := range packages {
				if pkg.Name == replace.Old.Path {
					replacements = append(replacements, k)
				}
			}
		} else {
			// If a version is present on the left side of the arrow (=>),
			// only that specific version of the module is replaced
			s := replace.Old.Path + "@" + replace.Old.Version

			// A `replace` directive has no effect if the module version on the left side is not required.
			if _, ok := packages[s]; ok {
				replacements = []string{s}
			}
		}

		for _, replacement := range replacements {
			version := strings.TrimPrefix(replace.New.Version, "v")

			if len(version) == 0 || version == unknownVersion {
				// There is no version specified on the replacement, it means the artifact is directly accessible
				// the package itself will then be scanned so there is no need to keep it
				delete(packages, replacement)
				continue
			}
			packages[replacement] = PackageDetails{
				Name:      replace.New.Path,
				Version:   version,
				Ecosystem: GoEcosystem,
				CompareAs: GoEcosystem,
				Line:      models.Position{Start: start.Line, End: end.Line},
				Column:    models.Position{Start: start.LineRune, End: end.LineRune},
			}
		}
	}

	if parsedLockfile.Go != nil && parsedLockfile.Go.Version != "" {
		v := semantic.ParseSemverLikeVersion(parsedLockfile.Go.Version, 3)

		goVersion := fmt.Sprintf(
			"%d.%d.%d",
			v.Components.Fetch(0),
			v.Components.Fetch(1),
			v.Components.Fetch(2),
		)

		packages["stdlib"] = PackageDetails{
			Name:      "stdlib",
			Version:   goVersion,
			Ecosystem: GoEcosystem,
			CompareAs: GoEcosystem,
		}
	}

	return pkgDetailsMapToSlice(deduplicatePackages(packages)), nil
}

var _ Extractor = GoLockExtractor{}

//nolint:gochecknoinits
func init() {
	registerExtractor("go.mod", GoLockExtractor{})
}

func ParseGoLock(pathToLockfile string) ([]PackageDetails, error) {
	return extractFromFile(pathToLockfile, GoLockExtractor{})
}
