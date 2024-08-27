package output

import (
	"errors"
	"io"
	"testing"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/google/osv-scanner/internal/output/sbom"
	"github.com/google/osv-scanner/internal/utility/purl"
	"github.com/google/osv-scanner/pkg/models"
)

// PrintCycloneDXResults writes results to the provided writer in CycloneDX format
func PrintCycloneDXResults(vulnResult *models.VulnerabilityResults, cycloneDXVersion models.CycloneDXVersion, outputWriter io.Writer) error {
	bomCreator := sbom.SpecVersionToBomCreator[cycloneDXVersion]
	resultsByPurl, errs := purl.Group(vulnResult.Results)

	bom := bomCreator(resultsByPurl, vulnResult.Artifacts)
	encoder := cyclonedx.NewBOMEncoder(outputWriter, cyclonedx.BOMFileFormatJSON)
	encoder.SetPretty(testing.Testing())

	err := encoder.EncodeVersion(bom, bom.SpecVersion)

	return errors.Join(err, errors.Join(errs...))
}
