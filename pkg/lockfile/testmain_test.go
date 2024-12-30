package lockfile_test

import (
	"os"
	"testing"

	"github.com/datadog/osv-scanner/internal/testutility"
)

func TestMain(m *testing.M) {
	MockAllMatchers()
	code := m.Run()

	testutility.CleanSnapshots(m)
	os.Exit(code)
}
