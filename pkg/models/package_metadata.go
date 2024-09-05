package models

type PackageMetadataType string

const (
	PackageManagerMetadata     PackageMetadataType = "package-manager"
	IsDirectDependencyMetadata PackageMetadataType = "is-direct-dependency"
)

type PackageMetadata map[PackageMetadataType]string
