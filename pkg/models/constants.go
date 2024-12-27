package models

type Ecosystem string

const (
	EcosystemGo          Ecosystem = "Go"
	EcosystemNPM         Ecosystem = "npm"
	EcosystemOSSFuzz     Ecosystem = "OSS-Fuzz"
	EcosystemPyPI        Ecosystem = "PyPI"
	EcosystemRubyGems    Ecosystem = "RubyGems"
	EcosystemCratesIO    Ecosystem = "crates.io"
	EcosystemPackagist   Ecosystem = "Packagist"
	EcosystemMaven       Ecosystem = "Maven"
	EcosystemNuGet       Ecosystem = "NuGet"
	EcosystemDebian      Ecosystem = "Debian"
	EcosystemAlpine      Ecosystem = "Alpine"
	EcosystemHex         Ecosystem = "Hex"
	EcosystemPub         Ecosystem = "Pub"
	EcosystemConanCenter Ecosystem = "ConanCenter"
	EcosystemCRAN        Ecosystem = "CRAN"
)

type SeverityType string

const (
	SeverityCVSSV2 SeverityType = "CVSS_V2"
	SeverityCVSSV3 SeverityType = "CVSS_V3"
	SeverityCVSSV4 SeverityType = "CVSS_V4"
)

type RangeType string

const (
	RangeSemVer    RangeType = "SEMVER"
	RangeEcosystem RangeType = "ECOSYSTEM"
)

type ReferenceType string

const (
	ReferenceAdvisory ReferenceType = "ADVISORY"
)

type CreditType string
