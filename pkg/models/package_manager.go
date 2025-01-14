package models

type PackageManager string

const (
	Maven        PackageManager = "Maven"
	Gradle       PackageManager = "Gradle"
	NPM          PackageManager = "NPM"
	Yarn         PackageManager = "Yarn"
	Pnpm         PackageManager = "Pnpm"
	Requirements PackageManager = "Requirements"
	Pipfile      PackageManager = "Pipfile"
	Poetry       PackageManager = "Poetry"
	NuGet        PackageManager = "NuGet"
	Bundler      PackageManager = "Bundler"
	Golang       PackageManager = "Golang"
	Composer     PackageManager = "Composer"
	Unknown      PackageManager = "Unknown"
)
