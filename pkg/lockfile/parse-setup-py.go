package lockfile

import (
	"bufio"
	"errors"
	"io"
	"path/filepath"

	"github.com/google/osv-scanner/pkg/models"
	"golang.org/x/exp/maps"
)

// Adds support for parsing the `install_requires` key
// Only supports arrays of plain string values being passed as a named argument
// Any dependencies described in other requires keys are not scanned
// Fails fast on unsupported inputs

type SetupPyExtractor struct{}

func (e SetupPyExtractor) ShouldExtract(path string) bool {
	return filepath.Base(path) == "setup.py"
}

const InstallRequiresKeyword = "install_requires"

var SkipRunes = map[rune]struct{}{
	' ':  {},
	'\t': {},
	'\r': {},
	'\f': {},
	',':  {},
}

func (e SetupPyExtractor) Extract(f DepFile) ([]PackageDetails, error) {
	var lineNumber, columnStart = 1, 1
	var lineContents string

	packages := map[string]PackageDetails{}

	inInstallRequires := false
	inEqual := false
	inArray := false
	r := bufio.NewReader(f)

out:
	for {
		var rn rune
		var runeSize int
		var err error
		if rn, runeSize, err = r.ReadRune(); err != nil {
			return nil, err
		}
		columnStart += runeSize
		lineContents += string(rn)

		// Skip comments, even before install_requires, as they are not relevant
		// and might incorrectly trigger install_requires start
		skippedComment, err := skipComment(rn, r)
		if err != nil {
			return nil, err
		} else if skippedComment {
			lineContents = ""
			lineNumber++
			columnStart = 1

			continue out
		}

		if rn == '\n' {
			lineContents = ""
			lineNumber++
			columnStart = 1

			continue
		}

		if !inInstallRequires {
			isInstallRequires, err := checkString(rn, r, InstallRequiresKeyword)
			if err != nil {
				return nil, err
			} else if isInstallRequires {
				inInstallRequires = true
			}

			continue
		}

		if _, ok := SkipRunes[rn]; ok {
			// skip
		} else if rn == '=' {
			if inEqual {
				return nil, errors.New("unexpected equal inside already started equal")
			}
			inEqual = true
		} else if rn == '[' {
			if !inEqual {
				return nil, errors.New("unexpected array start without =")
			}
			if inArray {
				return nil, errors.New("unexpected array start inside already started array")
			}
			inArray = true
		} else if rn == ']' {
			if !inEqual || !inArray {
				return nil, errors.New("unexpected array end without start and/or equal")
			}

			return maps.Values(packages), nil
		} else if rn == '\'' || rn == '"' {
			if !inArray {
				return nil, errors.New("unexpected string outside of install_requires with equal array")
			}

			requirement, err := readRemainingStringUntil(r, &rn)
			if err != nil {
				return nil, err
			}
			lineContents += requirement

			details, err := ParseRequirementLine(f.Path(), models.SetupTools, lineContents, requirement, lineNumber, 0, columnStart, columnStart+len(requirement))
			if err != nil {
				return nil, err
			}

			packages[details.Name] = *details
		} else {
			text, err := readRemainingStringUntil(r, nil)
			if err != nil {
				return nil, err
			}

			return nil, errors.New("unexpected text=" + string(rn) + text)
		}
	}
}

func readRemainingStringUntil(r *bufio.Reader, end *rune) (string, error) {
	var text string
	for {
		rn, _, err := r.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return "", err
		}
		if end != nil && rn == *end {
			break
		}
		text += string(rn)
	}

	return text, nil
}

func skipComment(current rune, r *bufio.Reader) (bool, error) {
	// Skip comments, even before install_requires, as they are not relevant
	// and might incorrectly trigger install_requires start
	if current == '#' {
		for {
			rn, _, err := r.ReadRune()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return false, errors.New("unexpected end of file")
				}

				return false, err
			}
			if rn == '\n' {
				return true, nil
			}
		}
	}

	return false, nil
}

func checkString(current rune, r *bufio.Reader, str string) (bool, error) {
	for pos, keywordRune := range str {
		if pos == 0 {
			if current != keywordRune {
				return false, nil
			}

			continue
		}

		bufferRune, _, err := r.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return false, nil
			}

			return false, err
		}
		if bufferRune != keywordRune {
			return false, nil
		}
	}

	return true, nil
}

var _ Extractor = SetupPyExtractor{}

//nolint:gochecknoinits
func init() {
	registerExtractor("setup.py", SetupPyExtractor{})
}

func ParseSetupPy(pathToLockfile string) ([]PackageDetails, error) {
	return extractFromFile(pathToLockfile, SetupPyExtractor{})
}
