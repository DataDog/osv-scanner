package output_test

import (
	"testing"

	"github.com/datadog/osv-scanner/pkg/models"
)

type outputTestCaseArgs struct {
	vulnResult *models.VulnerabilityResults
}

type outputTestCase struct {
	name string
	args outputTestCaseArgs
}

type outputTestRunner = func(t *testing.T, args outputTestCaseArgs)

func testOutputWithArtifacts(t *testing.T, run outputTestRunner) {
	t.Helper()

	tests := []outputTestCase{
		{
			name: "no sources",
			args: outputTestCaseArgs{
				vulnResult: &models.VulnerabilityResults{
					Results: []models.PackageSource{},
				},
			},
		},
		{
			name: "one source with no packages",
			args: outputTestCaseArgs{
				vulnResult: &models.VulnerabilityResults{
					Results: []models.PackageSource{
						{
							Source:   models.SourceInfo{Path: "path/to/my/first/lockfile"},
							Packages: []models.PackageVulns{},
						},
					},
				},
			},
		},
		{
			name: "multiple sources with no packages",
			args: outputTestCaseArgs{
				vulnResult: &models.VulnerabilityResults{
					Results: []models.PackageSource{
						{
							Source:   models.SourceInfo{Path: "path/to/my/first/lockfile"},
							Packages: []models.PackageVulns{},
						},
						{
							Source:   models.SourceInfo{Path: "path/to/my/second/lockfile"},
							Packages: []models.PackageVulns{},
						},
						{
							Source:   models.SourceInfo{Path: "path/to/my/third/lockfile"},
							Packages: []models.PackageVulns{},
						},
					},
				},
			},
		},
		{
			name: "one source with one package, no artifacts",
			args: outputTestCaseArgs{
				vulnResult: &models.VulnerabilityResults{
					Artifacts: make([]models.ScannedArtifact, 0),
					Results: []models.PackageSource{
						{
							Source: models.SourceInfo{Path: "path/to/my/first/lockfile"},
							Packages: []models.PackageVulns{
								{
									Package: models.PackageInfo{
										Name:      "com.nobody:mine1",
										Version:   "1.2.3",
										Ecosystem: "npm",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "one source with one artifact, no dependsOn",
			args: outputTestCaseArgs{
				vulnResult: &models.VulnerabilityResults{
					Artifacts: []models.ScannedArtifact{
						{
							ArtifactDetail: models.ArtifactDetail{
								Name:     "dev.foo:artifact",
								Version:  "1.0-SNAPSHOT",
								Filename: "pom.xml",
							},
						},
					},
					Results: []models.PackageSource{
						{
							Source: models.SourceInfo{Path: "pom.xml"},
							Packages: []models.PackageVulns{
								{
									Package: models.PackageInfo{
										Name:      "com.nobody:mine1",
										Version:   "1.2.3",
										Ecosystem: "Maven",
									},
									Locations: []models.PackageLocations{
										{
											Block: models.PackageLocation{
												Filename:    "pom.xml",
												LineStart:   0,
												LineEnd:     0,
												ColumnStart: 0,
												ColumnEnd:   0,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "one source with one artifact, one dependsOn",
			args: outputTestCaseArgs{
				vulnResult: &models.VulnerabilityResults{
					Artifacts: []models.ScannedArtifact{
						{
							ArtifactDetail: models.ArtifactDetail{
								Name:     "dev.foo:artifact",
								Version:  "1.0-SNAPSHOT",
								Filename: "pom.xml",
							},
							DependsOn: &models.ArtifactDetail{
								Name:     "deb.foo:parent",
								Version:  "1.0-SNAPSHOT",
								Filename: "parent/pom.xml",
							},
						},
						{
							ArtifactDetail: models.ArtifactDetail{
								Name:     "deb.foo:parent",
								Version:  "1.0-SNAPSHOT",
								Filename: "parent/pom.xml",
							},
						},
					},
					Results: []models.PackageSource{
						{
							Source: models.SourceInfo{Path: "pom.xml"},
							Packages: []models.PackageVulns{
								{
									Package: models.PackageInfo{
										Name:      "com.nobody:mine1",
										Version:   "1.2.3",
										Ecosystem: "Maven",
									},
									Locations: []models.PackageLocations{
										{
											Block: models.PackageLocation{
												Filename:    "pom.xml",
												LineStart:   0,
												LineEnd:     0,
												ColumnStart: 0,
												ColumnEnd:   0,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "one source with multiple artifacts, used by one of them",
			args: outputTestCaseArgs{
				vulnResult: &models.VulnerabilityResults{
					Artifacts: []models.ScannedArtifact{
						{
							ArtifactDetail: models.ArtifactDetail{
								Name:     "dev.foo:artifact",
								Version:  "1.0-SNAPSHOT",
								Filename: "pom.xml",
							},
						},
						{
							ArtifactDetail: models.ArtifactDetail{
								Name:     "dev.foo:lib1",
								Version:  "1.0-SNAPSHOT",
								Filename: "lib1/pom.xml",
							},
						},
						{
							ArtifactDetail: models.ArtifactDetail{
								Name:     "dev.foo:lib2",
								Version:  "1.0-SNAPSHOT",
								Filename: "lib2/pom.xml",
							},
						},
					},
					Results: []models.PackageSource{
						{
							Source: models.SourceInfo{Path: "pom.xml"},
							Packages: []models.PackageVulns{
								{
									Package: models.PackageInfo{
										Name:      "dev.foo:lib1",
										Version:   "1.0-SNAPSHOT",
										Ecosystem: "Maven",
									},
									Locations: []models.PackageLocations{
										{
											Block: models.PackageLocation{
												Filename:    "pom.xml",
												LineStart:   0,
												LineEnd:     0,
												ColumnStart: 0,
												ColumnEnd:   0,
											},
										},
									},
								},
								{
									Package: models.PackageInfo{
										Name:      "dev.foo:lib2",
										Version:   "1.0-SNAPSHOT",
										Ecosystem: "Maven",
									},
									Locations: []models.PackageLocations{
										{
											Block: models.PackageLocation{
												Filename:    "pom.xml",
												LineStart:   0,
												LineEnd:     0,
												ColumnStart: 0,
												ColumnEnd:   0,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			run(t, tt.args)
		})
	}
}
