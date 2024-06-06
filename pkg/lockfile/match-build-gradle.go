package lockfile

import (
	"github.com/google/osv-scanner/internal/utility/fileposition"
	"github.com/google/osv-scanner/pkg/models"
	"io"
	"strings"
)

type BuildGradleMatcher struct{}

func (m BuildGradleMatcher) GetSourceFile(lockfile DepFile) (DepFile, error) {
	fileName := "build.gradle"

	// lockfile (default, groovy)
	sourcefile, err := lockfile.Open(fileName)
	if err != nil {
		// kotlin
		sourcefile, err = lockfile.Open(fileName + ".kts")
	}

	// gradle verification metadata (<rootdir>/gradle/verification-metadata.xml)
	relativePath := "../" + fileName
	if err != nil {
		// groovy
		sourcefile, err = lockfile.Open(relativePath)
	}
	if err != nil {
		// kotlin
		sourcefile, err = lockfile.Open(relativePath + ".kts")
	}

	return sourcefile, err
}

func (m BuildGradleMatcher) Match(sourcefile DepFile, packages []PackageDetails) error {
	content, err := io.ReadAll(sourcefile)
	if err != nil {
		return err
	}

	lines := fileposition.BytesToLines(content)

	for index, line := range lines {
		lineNumber := index + 1
		for key, pkg := range packages {
			group, artifact, _ := strings.Cut(pkg.Name, ":")
			// TODO: what to do if, while using extended format, components are split in multiple lines?
			if strings.Contains(line, group) && strings.Contains(line, artifact) {
				if strings.Contains(line, pkg.Version) {
					startColumn := fileposition.GetFirstNonEmptyCharacterIndexInLine(line)
					endColumn := fileposition.GetLastNonEmptyCharacterIndexInLine(line)

					packages[key].BlockLocation = models.FilePosition{
						Line:     models.Position{Start: lineNumber, End: lineNumber},
						Column:   models.Position{Start: startColumn, End: endColumn},
						Filename: sourcefile.Path(),
					}

					nameLocation := fileposition.ExtractDelimitedRegexpPositionInBlock([]string{line}, artifact, lineNumber, "['\":]", "['\":]")
					if nameLocation != nil {
						nameLocation.Filename = sourcefile.Path()
						packages[key].NameLocation = nameLocation
					}

					versionLocation := fileposition.ExtractDelimitedRegexpPositionInBlock([]string{line}, pkg.Version, lineNumber, "['\":]", "['\"]")
					if versionLocation != nil {
						versionLocation.Filename = sourcefile.Path()
						packages[key].VersionLocation = versionLocation
					}

					// If the dep is using version range, it won't include the exact resolved version
					// It could also happen if there are multiple versions conflicting and the transitive one has a higher version
				} else {
					// TODO: how to resolve version ranges
					// A) Only one pkg with this group+artifact -> We know it was resolved from this line
					// B) Multiple pkgs with this group+artifcat -> TODO: How to be sure of the one that was resolved from this line
				}
			}
		}
	}

	return nil
}

var _ Matcher = BuildGradleMatcher{}
