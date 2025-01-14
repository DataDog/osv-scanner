package lockfile

import (
	"github.com/datadog/osv-scanner/pkg/models"
)

type Packages []PackageDetails

type Lockfile struct {
	FilePath string                  `json:"filePath"`
	ParsedAs string                  `json:"parsedAs"`
	Packages Packages                `json:"packages"`
	Artifact *models.ScannedArtifact `json:"artifact,omitempty"`
}
