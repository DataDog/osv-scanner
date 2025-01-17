package testutility

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

type Snapshot struct {
	windowsReplacements map[string]string
}

// NewSnapshot creates a snapshot that can be passed around within tests
func NewSnapshot() Snapshot {
	return Snapshot{windowsReplacements: map[string]string{}}
}

// WithWindowsReplacements adds a map of strings with values that they should be
// replaced within before comparing the snapshot when running on Windows
func (s Snapshot) WithWindowsReplacements(replacements map[string]string) Snapshot {
	s.windowsReplacements = replacements

	return s
}

// WithCRLFReplacement adds a Windows replacement for "\r\n" to "\n".
// This should be called after WithWindowsReplacements if used together.
func (s Snapshot) WithCRLFReplacement() Snapshot {
	s.windowsReplacements["\r\n"] = "\n"

	return s
}

// CleanSnapshots ensures that snapshots are relevant and sorted for consistency
func CleanSnapshots(m *testing.M) {
	snaps.Clean(m, snaps.CleanOpts{Sort: true})
}

// MatchText asserts the existing snapshot matches what was gotten in the test
func (s Snapshot) MatchText(t *testing.T, got string) {
	t.Helper()

	snaps.MatchSnapshot(t, got)
}
