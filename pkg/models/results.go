package models

// Combined vulnerabilities found for the scanned packages
type VulnerabilityResults struct {
	Results   []PackageSource   `json:"results"`
	Artifacts []ScannedArtifact `json:"artifacts,omitempty"`
}

type ArtifactDetail struct {
	Name      string
	Version   string
	Filename  string
	Ecosystem Ecosystem
}

type ScannedArtifact struct {
	ArtifactDetail
	DependsOn *ArtifactDetail
}

// ExperimentalAnalysisConfig is an experimental type intended to contain the
// types of analysis performed on packages found by the scanner.
type ExperimentalAnalysisConfig struct {
	Licenses ExperimentalLicenseConfig `json:"licenses"`
}

type ExperimentalLicenseConfig struct {
	Summary   bool      `json:"summary"`
	Allowlist []License `json:"allowlist"`
}

type SourceInfo struct {
	ScanPath string `json:"scan_path,omitempty"`
	Path     string `json:"path"`
	Type     string `json:"type"`
}

type Metadata struct {
	RepoURL   string   `json:"repo_url"`
	DepGroups []string `json:"-"`
}

func (s SourceInfo) String() string {
	return s.Type + ":" + s.Path
}

// Vulnerabilities grouped by sources
type PackageSource struct {
	Source   SourceInfo     `json:"source"`
	Packages []PackageVulns `json:"packages"`
}

// License is an SPDX license.
type License string

// Vulnerabilities grouped by package
// TODO: rename this to be Package as it now includes license information too.
type PackageVulns struct {
	Package   PackageInfo        `json:"package"`
	DepGroups []string           `json:"dependency_groups,omitempty"`
	Locations []PackageLocations `json:"locations,omitempty"`
	Metadata  PackageMetadata    `json:"metadata,omitempty"`
}

// Specific package information
type PackageInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Ecosystem string `json:"ecosystem"`
	Commit    string `json:"commit,omitempty"`
}
