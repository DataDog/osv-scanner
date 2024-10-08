package lockfile

import (
	"errors"
	"strings"

	"github.com/google/osv-scanner/internal/cachedregexp"
	"github.com/google/osv-scanner/internal/utility/fileposition"
	"github.com/google/osv-scanner/pkg/models"
)

var spaceRegexp = cachedregexp.MustCompile(`\s+`)

// https://regex101.com/r/idZTt6/5
var commentsRegexp = cachedregexp.MustCompile(`\s*#[\s\S]*$`)

// https://regex101.com/r/djHuI8/3
var installRequiresRegexp = cachedregexp.MustCompile(`install_requires\s*=\s*(?P<requirements>.+)?\s*`)

// https://regex101.com/r/szEVdW/9
var requirementRegexp = cachedregexp.MustCompile(`^\s*(?P<pkgname>[\w.-]+)\s*(\[(?P<optnames>[\w.,\s-]+)])?\s*(\(?\s*(?P<requirement>(,?\s*(?P<constraint>~=|==|!=|<=|>=|<|>|===)\s*(?P<version>[\w.!*-]+))+|(@\s*(?P<wheel>[^;\s]+)))\s*\)?)?\s*(?P<options>((((--[\w-]+)="([^"]+)")|((--[\w-]+)=([^\s]+))|((--[\w-]+)\s+([^\s]+)))\s*)*)\s*(;\s*(?P<envmarkers>.*))?\s*$`)

// https://regex101.com/r/ppD7Uj/2
var wheelURLPattern = cachedregexp.MustCompile(
	`^\s*.*?\/(?P<distribution>[^-/]+)-(?P<version>[^-/]+)(-(?P<buildtag>[^-/]+))?-(?P<pythontag>[^-/]+)-(?P<abitag>[^-/]+)-(?P<platformtag>[^-/]+)\.whl\s*$`)

// ParseRequirementLine parses python requirement
// See: https://pip.pypa.io/en/stable/reference/requirements-file-format/#example
func ParseRequirementLine(path string, pkgManager models.PackageManager, line string, cleanLine string, lineNumber int, lineOffset int, columnStart int, columnEnd int) (*PackageDetails, error) {
	var name, versionRequirement, wheel string
	var nameLocation, versionLocation *models.FilePosition

	matchesList := requirementRegexp.FindAllStringSubmatch(cleanLine, -1)
	if len(matchesList) == 0 {
		return nil, errors.New("could not parse requirement line")
	}
	matches := matchesList[0]

	name = matches[requirementRegexp.SubexpIndex("pkgname")]
	nameLocation = fileposition.ExtractStringPositionInMultiline(line, name, lineNumber)
	if nameLocation != nil {
		nameLocation.Filename = path
	}

	versionRequirement = matches[requirementRegexp.SubexpIndex("requirement")]
	if versionRequirement != "" {
		versionLocation = fileposition.ExtractStringPositionInMultiline(line, versionRequirement, lineNumber)
		if versionLocation != nil {
			versionLocation.Filename = path
		}
	}

	wheel = matches[requirementRegexp.SubexpIndex("wheel")]
	if wheel != "" {
		if strings.HasSuffix(wheel, ".whl") {
			version := extractVersionFromWheelURL(wheel)
			versionRequirement = "==" + version
			versionLocation = fileposition.ExtractStringPositionInMultiline(line, wheel, lineNumber)
			if versionLocation != nil {
				versionLocation.Filename = path
			}
		} else {
			versionRequirement = ""
			versionLocation = nil
		}
	}

	blockLocation := models.FilePosition{
		Line:     models.Position{Start: lineNumber, End: lineNumber + lineOffset},
		Column:   models.Position{Start: columnStart, End: columnEnd},
		Filename: path,
	}

	versionRequirement = spaceRegexp.ReplaceAllString(versionRequirement, "")

	return &PackageDetails{
		Name:            normalizedRequirementName(name),
		Version:         versionRequirement,
		BlockLocation:   blockLocation,
		NameLocation:    nameLocation,
		VersionLocation: versionLocation,
		PackageManager:  pkgManager,
		Ecosystem:       PipEcosystem,
		CompareAs:       PipEcosystem,
	}, nil
}

// Please note the whl filename has been standardized here :
// https://packaging.python.org/en/latest/specifications/binary-distribution-format/#file-name-convention
func extractVersionFromWheelURL(wheelURL string) string {
	matchesList := wheelURLPattern.FindAllStringSubmatch(wheelURL, -1)
	if len(matchesList) == 0 {
		return ""
	}
	matches := matchesList[0]

	if version := matches[wheelURLPattern.SubexpIndex("version")]; version != "" {
		return version
	}

	return ""
}
