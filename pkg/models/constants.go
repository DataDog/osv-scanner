package models

type Ecosystem string

const (
	EcosystemGo        Ecosystem = "Go"
	EcosystemNPM       Ecosystem = "npm"
	EcosystemPyPI      Ecosystem = "PyPI"
	EcosystemRubyGems  Ecosystem = "RubyGems"
	EcosystemPackagist Ecosystem = "Packagist"
	EcosystemMaven     Ecosystem = "Maven"
	EcosystemNuGet     Ecosystem = "NuGet"
)
