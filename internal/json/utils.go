package json

import (
	"github.com/google/osv-scanner/internal/cachedregexp"
	"strings"
)

/*
GetSectionOffset computes the start line of any section in the file.
To see the regex in action, check out https://regex101.com/r/3EHqB8/1 (it uses the dependencies section as an example)
*/
func GetSectionOffset(sectionName string, content string) int {
	sectionMatcher := cachedregexp.MustCompile(`(?m)^\s*"` + sectionName + `":\s*{\s*$`)
	sectionIndex := sectionMatcher.FindStringIndex(content)
	if len(sectionIndex) < 2 {
		return -1
	}

	return strings.Count(content[:sectionIndex[1]], "\n")
}
