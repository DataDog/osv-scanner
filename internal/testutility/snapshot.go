package testutility

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

type Snapshot struct{}

// NewSnapshot creates a snapshot that can be passed around within tests
func NewSnapshot() Snapshot {
	return Snapshot{}
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
