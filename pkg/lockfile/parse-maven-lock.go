package lockfile

import (
	"encoding/xml"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/osv-scanner/pkg/models"

	"github.com/google/osv-scanner/internal/cachedregexp"
)

const MaxParentDepth = 10

type MavenLockDependency struct {
	XMLName    xml.Name `xml:"dependency"`
	GroupID    string   `xml:"groupId"`
	ArtifactID string   `xml:"artifactId"`
	Version    string   `xml:"version"`
	Scope      string   `xml:"scope"`
	Start      models.FilePosition
	End        models.FilePosition
	SourceFile string
}

type MavenLockParent struct {
	XMLName      xml.Name `xml:"parent"`
	RelativePath string   `xml:"relativePath"`
}

type MavenLockDependencyHolder struct {
	Dependencies []MavenLockDependency `xml:"dependency"`
}

func (mld MavenLockDependency) parseResolvedVersion(version string) string {
	versionRequirementReg := cachedregexp.MustCompile(`[[(]?(.*?)(?:,|[)\]]|$)`)

	results := versionRequirementReg.FindStringSubmatch(version)

	if results == nil || results[1] == "" {
		return "0"
	}

	return results[1]
}

/*
You can see the regex working here : https://regex101.com/r/inAPiN/2
*/
func (mld MavenLockDependency) resolveVersionValue(lockfile MavenLockFile) string {
	interpolationReg := cachedregexp.MustCompile(`\${([^}]+)}`)
	result := interpolationReg.ReplaceAllFunc([]byte(mld.Version), func(bytes []byte) []byte {
		propStr := string(bytes)
		propName := propStr[2 : len(propStr)-1]
		property, ok := lockfile.Properties.m[propName]
		if !ok {
			fmt.Fprintf(
				os.Stderr,
				"Failed to resolve version of %s: property \"%s\" could not be found for \"%s\"\n",
				mld.GroupID+":"+mld.ArtifactID,
				string(bytes),
				lockfile.GroupID+":"+lockfile.ArtifactID,
			)

			return []byte("0")
		}

		return []byte(mld.parseResolvedVersion(property))
	})

	return mld.parseResolvedVersion(string(result))
}

func (mld MavenLockDependency) ResolveVersion(lockfile MavenLockFile) string {
	return mld.resolveVersionValue(lockfile)
}

type MavenLockFile struct {
	XMLName             xml.Name                  `xml:"project"`
	Parent              MavenLockParent           `xml:"parent"`
	ModelVersion        string                    `xml:"modelVersion"`
	GroupID             string                    `xml:"groupId"`
	ArtifactID          string                    `xml:"artifactId"`
	Properties          MavenLockProperties       `xml:"properties"`
	Dependencies        MavenLockDependencyHolder `xml:"dependencies"`
	ManagedDependencies MavenLockDependencyHolder `xml:"dependencyManagement>dependencies"`
}

const MavenEcosystem Ecosystem = "Maven"

type MavenLockProperties struct {
	m map[string]string
}

func (p *MavenLockProperties) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	p.m = map[string]string{}

	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		switch tt := t.(type) {
		case xml.StartElement:
			var s string

			if err := d.DecodeElement(&s, &tt); err != nil {
				return fmt.Errorf("%w", err)
			}

			p.m[tt.Name.Local] = s

		case xml.EndElement:
			if tt.Name == start.Name {
				return nil
			}
		}
	}
}

func (dependencyHolder *MavenLockDependencyHolder) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	dependencyHolder.Dependencies = make([]MavenLockDependency, 0)
DecodingLoop:
	for {
		startLine, startColumn := decoder.InputPos()
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch elem := token.(type) {
		case xml.StartElement:
			dependency := MavenLockDependency{}
			dependency.Start = models.FilePosition{Line: startLine, Column: startColumn}
			err := decoder.DecodeElement(&dependency, &elem)
			if err != nil {
				return err
			}
			endLine, endColumn := decoder.InputPos()
			dependency.End = models.FilePosition{Line: endLine, Column: endColumn}
			dependencyHolder.Dependencies = append(dependencyHolder.Dependencies, dependency)
		case xml.EndElement:
			if elem.Name == start.Name {
				break DecodingLoop
			}
		}
	}

	return nil
}

type MavenLockExtractor struct{}

func (e MavenLockExtractor) ShouldExtract(path string) bool {
	return filepath.Base(path) == "pom.xml"
}

/**
** This function merge a child lockfile into the parent one.
** It copies all information originating from the child in it, overriding any common properties/dependencies
**/
func (e MavenLockExtractor) mergeLockfiles(childLockfile *MavenLockFile, parentLockfile *MavenLockFile) *MavenLockFile {
	parentLockfile.Parent = childLockfile.Parent
	parentLockfile.ArtifactID = childLockfile.ArtifactID
	parentLockfile.GroupID = childLockfile.GroupID
	parentLockfile.ModelVersion = childLockfile.ModelVersion

	// Child properties take precedence over parent defined ones
	for key, value := range childLockfile.Properties.m {
		parentLockfile.Properties.m[key] = value
	}
	// We add child dependency at the end, this way they will override the parent ones during transformation to a map
	parentLockfile.Dependencies.Dependencies = append(parentLockfile.Dependencies.Dependencies, childLockfile.Dependencies.Dependencies...)
	parentLockfile.ManagedDependencies.Dependencies = append(parentLockfile.ManagedDependencies.Dependencies, childLockfile.ManagedDependencies.Dependencies...)

	return parentLockfile
}

func (e MavenLockExtractor) enrichDependencies(f DepFile, dependencies []MavenLockDependency) MavenLockDependencyHolder {
	result := make([]MavenLockDependency, len(dependencies))
	for index, dependency := range dependencies {
		if len(dependency.SourceFile) == 0 {
			dependency.SourceFile = f.Path()
		}
		result[index] = dependency
	}

	return MavenLockDependencyHolder{Dependencies: result}
}

func (e MavenLockExtractor) decodeMavenFile(f DepFile, depth int) (*MavenLockFile, error) {
	var parsedLockfile *MavenLockFile
	if depth >= MaxParentDepth {
		return nil, fmt.Errorf("maven file decoding reached the max depth (%d/%d), check for a circular dependency", depth, MaxParentDepth)
	}
	// Decoding the original lockfile and enrich its dependencies
	err := xml.NewDecoder(f).Decode(&parsedLockfile)
	if err != nil {
		return nil, err
	}
	parsedLockfile.Dependencies = e.enrichDependencies(f, parsedLockfile.Dependencies.Dependencies)
	parsedLockfile.ManagedDependencies = e.enrichDependencies(f, parsedLockfile.ManagedDependencies.Dependencies)
	if parsedLockfile.Parent == (MavenLockParent{}) {
		return parsedLockfile, nil
	}

	// If a parent is defined, use its relative path to find the file, then recurse to decode it properly and enrich its dependencies
	// If the relativePath is not defined, default to ../pom.xml
	parentRelativePath := parsedLockfile.Parent.RelativePath
	if len(parentRelativePath) == 0 {
		parentRelativePath = "../pom.xml"
	} else if !strings.HasSuffix(parentRelativePath, ".xml") {
		// It means we only have a path, we should append the default pom.xml
		parentRelativePath = path.Join(parentRelativePath, "pom.xml")
	}
	parentPath, _ := filepath.Abs(filepath.FromSlash(path.Join(path.Dir(f.Path()), parentRelativePath)))
	fmt.Printf("Opening parent file %v\n", parentPath)
	parentFile, err := OpenLocalDepFile(parentPath)
	if err != nil {
		return nil, err
	}
	parentLockfile, parentErr := e.decodeMavenFile(parentFile, depth+1)
	if parentErr != nil {
		return nil, parentErr
	}
	parentLockfile.Dependencies = e.enrichDependencies(parentFile, parentLockfile.Dependencies.Dependencies)
	parentLockfile.ManagedDependencies = e.enrichDependencies(parentFile, parentLockfile.ManagedDependencies.Dependencies)

	// Once everything is decoded and enriched, merge them together
	return e.mergeLockfiles(parsedLockfile, parentLockfile), nil
}

func (e MavenLockExtractor) Extract(f DepFile) ([]PackageDetails, error) {
	parsedLockfile, err := e.decodeMavenFile(f, 0)
	if err != nil {
		return []PackageDetails{}, fmt.Errorf("could not extract from %s: %w", f.Path(), err)
	}

	details := map[string]PackageDetails{}

	for _, lockPackage := range parsedLockfile.Dependencies.Dependencies {
		finalName := lockPackage.GroupID + ":" + lockPackage.ArtifactID

		pkgDetails := PackageDetails{
			Name:       finalName,
			Version:    lockPackage.ResolveVersion(*parsedLockfile),
			Ecosystem:  MavenEcosystem,
			CompareAs:  MavenEcosystem,
			Start:      lockPackage.Start,
			End:        lockPackage.End,
			SourceFile: lockPackage.SourceFile,
		}
		if strings.TrimSpace(lockPackage.Scope) != "" {
			pkgDetails.DepGroups = append(pkgDetails.DepGroups, lockPackage.Scope)
		}
		details[finalName] = pkgDetails
	}

	// managed dependencies take precedent over standard dependencies
	for _, lockPackage := range parsedLockfile.ManagedDependencies.Dependencies {
		finalName := lockPackage.GroupID + ":" + lockPackage.ArtifactID
		pkgDetails := PackageDetails{
			Name:       finalName,
			Version:    lockPackage.ResolveVersion(*parsedLockfile),
			Ecosystem:  MavenEcosystem,
			CompareAs:  MavenEcosystem,
			Start:      lockPackage.Start,
			End:        lockPackage.End,
			SourceFile: lockPackage.SourceFile,
		}
		if strings.TrimSpace(lockPackage.Scope) != "" {
			pkgDetails.DepGroups = append(pkgDetails.DepGroups, lockPackage.Scope)
		}
		details[finalName] = pkgDetails
	}

	return pkgDetailsMapToSlice(details), nil
}

var _ Extractor = MavenLockExtractor{}

//nolint:gochecknoinits
func init() {
	registerExtractor("pom.xml", MavenLockExtractor{})
}

func ParseMavenLock(pathToLockfile string) ([]PackageDetails, error) {
	return extractFromFile(pathToLockfile, MavenLockExtractor{})
}
