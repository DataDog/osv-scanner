package grouper_test

import (
	"testing"

	"github.com/google/osv-scanner/pkg/grouper"
	"github.com/google/osv-scanner/pkg/lockfile"
	"github.com/google/osv-scanner/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestGroupPackageByPURL_ShouldUnifyPackages(t *testing.T) {
	t.Parallel()
	input := []models.PackageSource{
		{
			Source: models.SourceInfo{
				ScanPath: "/dir",
				Path:     "/dir/lockfile.xml",
				Type:     "",
			},
			Packages: []models.PackageVulns{
				{
					Package: models.PackageInfo{
						Name:      "foo.bar:the-first-package",
						Version:   "1.0.0",
						Ecosystem: string(lockfile.MavenEcosystem),
						LockfileLocations: models.FileLocations{
							Block: models.FilePosition{
								Line:     models.Position{Start: 1, End: 2},
								Column:   models.Position{Start: 10, End: 21},
								Filename: "/dir/lockfile.xml",
							},
							Name: &models.FilePosition{
								Line:     models.Position{Start: 3, End: 3},
								Column:   models.Position{Start: 5, End: 14},
								Filename: "/dir/nested/other-lockfile.xml",
							},
						},
						SourcefileLocations: &models.FileLocations{
							Block: models.FilePosition{
								Line:     models.Position{Start: 2, End: 2},
								Column:   models.Position{Start: 15, End: 35},
								Filename: "/dir/sourcefile",
							},
							Name: &models.FilePosition{
								Line:     models.Position{Start: 2, End: 2},
								Column:   models.Position{Start: 16, End: 24},
								Filename: "/dir/sourcefile",
							},
						},
					},
				},
				{
					Package: models.PackageInfo{
						Name:      "foo.bar:the-first-package",
						Version:   "1.0.0",
						Ecosystem: string(lockfile.MavenEcosystem),
						LockfileLocations: models.FileLocations{
							Block: models.FilePosition{
								Line:     models.Position{Start: 1, End: 2},
								Column:   models.Position{Start: 10, End: 21},
								Filename: "/dir/lockfile.xml",
							},
						},
					},
				},
				{
					Package: models.PackageInfo{
						Name:      "foo.bar:the-first-package",
						Version:   "1.0.0",
						Ecosystem: string(lockfile.MavenEcosystem),
						LockfileLocations: models.FileLocations{
							Block: models.FilePosition{
								Line:     models.Position{Start: 1, End: 2},
								Column:   models.Position{Start: 10, End: 21},
								Filename: "/dir/lockfile.xml",
							},
						},
					},
				},
				{
					Package: models.PackageInfo{
						Name:      "foo.bar:package-2",
						Ecosystem: string(lockfile.MavenEcosystem),
						Version:   "1.0.0",
						LockfileLocations: models.FileLocations{
							Block: models.FilePosition{
								Line:     models.Position{Start: 11, End: 22},
								Column:   models.Position{Start: 10, End: 21},
								Filename: "/dir/lockfile.xml",
							},
						},
					},
				},
			},
		},
		{
			Source: models.SourceInfo{
				ScanPath: "/dir2",
				Path:     "/dir2/lockfile.json",
				Type:     "",
			},
			Packages: []models.PackageVulns{
				{
					Package: models.PackageInfo{
						Name:      "foo.bar:the-first-package",
						Version:   "1.0.0",
						Ecosystem: string(lockfile.MavenEcosystem),
						LockfileLocations: models.FileLocations{
							Block: models.FilePosition{
								Line:     models.Position{Start: 1, End: 2},
								Column:   models.Position{Start: 10, End: 21},
								Filename: "/dir2/lockfile.json",
							},
						},
					},
				},
				{
					Package: models.PackageInfo{
						Name:      "foo.bar:package-2",
						Ecosystem: string(lockfile.MavenEcosystem),
						Version:   "1.0.0",
					},
				},
			},
		},
	}

	result := grouper.GroupByPURL(input, false, false)

	expected := map[string]models.PackageDetails{
		"pkg:maven/foo.bar/the-first-package@1.0.0": {
			Name:      "foo.bar:the-first-package",
			Version:   "1.0.0",
			Ecosystem: string(lockfile.MavenEcosystem),
			Locations: []models.PackageLocations{
				{
					Block: models.PackageLocation{
						Filename:    "/dir/lockfile.xml",
						LineStart:   1,
						LineEnd:     2,
						ColumnStart: 10,
						ColumnEnd:   21,
					},
					Name: &models.PackageLocation{
						Filename:    "/dir/nested/other-lockfile.xml",
						LineStart:   3,
						LineEnd:     3,
						ColumnStart: 5,
						ColumnEnd:   14,
					},
					Sourcefile: &models.SourcefilePackageLocations{
						Block: models.PackageLocation{
							Filename:    "/dir/sourcefile",
							LineStart:   2,
							LineEnd:     2,
							ColumnStart: 15,
							ColumnEnd:   35,
						},
						Name: &models.PackageLocation{
							Filename:    "/dir/sourcefile",
							LineStart:   2,
							LineEnd:     2,
							ColumnStart: 16,
							ColumnEnd:   24,
						},
					},
				},
				{
					Block: models.PackageLocation{
						Filename:    "/dir2/lockfile.json",
						LineStart:   1,
						LineEnd:     2,
						ColumnStart: 10,
						ColumnEnd:   21,
					},
				},
			},
		},
		"pkg:maven/foo.bar/package-2@1.0.0": {
			Name:      "foo.bar:package-2",
			Version:   "1.0.0",
			Ecosystem: string(lockfile.MavenEcosystem),
			Locations: []models.PackageLocations{
				{
					Block: models.PackageLocation{
						Filename:    "/dir/lockfile.xml",
						LineStart:   11,
						LineEnd:     22,
						ColumnStart: 10,
						ColumnEnd:   21,
					},
				},
			},
		},
	}
	assert.Equal(t, expected, result)
}
