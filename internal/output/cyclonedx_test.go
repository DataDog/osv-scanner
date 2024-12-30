package output_test

import (
	"bytes"
	"testing"

	"github.com/datadog/osv-scanner/internal/output"
	"github.com/datadog/osv-scanner/internal/testutility"
	"github.com/datadog/osv-scanner/pkg/models"
)

func TestPrintCycloneDX14Results_WithDependencies(t *testing.T) {
	t.Parallel()

	testOutputWithArtifacts(t, func(t *testing.T, args outputTestCaseArgs) {
		t.Helper()

		outputWriter := &bytes.Buffer{}
		err := output.PrintCycloneDXResults(args.vulnResult, models.CycloneDXVersion14, outputWriter)

		if err != nil {
			t.Errorf("%v", err)
		}

		testutility.NewSnapshot().MatchText(t, outputWriter.String())
	})
}

func TestPrintCycloneDX15Results_WithDependencies(t *testing.T) {
	t.Parallel()

	testOutputWithArtifacts(t, func(t *testing.T, args outputTestCaseArgs) {
		t.Helper()

		outputWriter := &bytes.Buffer{}
		err := output.PrintCycloneDXResults(args.vulnResult, models.CycloneDXVersion15, outputWriter)

		if err != nil {
			t.Errorf("%v", err)
		}

		testutility.NewSnapshot().MatchText(t, outputWriter.String())
	})
}
