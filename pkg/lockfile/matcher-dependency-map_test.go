package lockfile

import (
	"github.com/google/osv-scanner/pkg/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMatcherDependencyMap_UpdatePackageDetails(t *testing.T) {
	type fields struct {
		LineOffset int
	}
	type args struct {
		pkg      *PackageDetails
		content  string
		indexes  []int
		depGroup string
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		expected *PackageDetails
	}{
		{
			name: "Updating position with line offset to 0",
			fields: fields{
				LineOffset: 0,
			},
			args: args{
				pkg: &PackageDetails{
					Name:    "Foobar",
					Version: "1.2.3",
				},
				content: `
          "Foobar": "^1.2.3"
				`,
				indexes: []int{11, 29, 12, 18, 22, 28},
			},
			expected: &PackageDetails{
				Name:     "Foobar",
				Version:  "1.2.3",
				IsDirect: true,
				BlockLocation: models.FilePosition{
					Filename: "file-path",
					Line: models.Position{
						Start: 2,
						End:   2,
					},
					Column: models.Position{
						Start: 11,
						End:   29,
					},
				},
				NameLocation: &models.FilePosition{
					Filename: "file-path",
					Line: models.Position{
						Start: 2,
						End:   2,
					},
					Column: models.Position{
						Start: 12,
						End:   18,
					},
				},
				VersionLocation: &models.FilePosition{
					Filename: "file-path",
					Line: models.Position{
						Start: 2,
						End:   2,
					},
					Column: models.Position{
						Start: 22,
						End:   28,
					},
				},
			},
		},
		{
			name: "Updating position  with positive line offset",
			fields: fields{
				LineOffset: 1,
			},
			args: args{
				pkg: &PackageDetails{
					Name:    "Foobar",
					Version: "1.2.3",
				},
				content: `
            "Foobar": "^1.2.3"
				`,
				indexes: []int{13, 31, 14, 20, 24, 30},
			},
			expected: &PackageDetails{
				Name:     "Foobar",
				Version:  "1.2.3",
				IsDirect: true,
				BlockLocation: models.FilePosition{
					Filename: "file-path",
					Line: models.Position{
						Start: 3,
						End:   3,
					},
					Column: models.Position{
						Start: 13,
						End:   31,
					},
				},
				NameLocation: &models.FilePosition{
					Filename: "file-path",
					Line: models.Position{
						Start: 3,
						End:   3,
					},
					Column: models.Position{
						Start: 14,
						End:   20,
					},
				},
				VersionLocation: &models.FilePosition{
					Filename: "file-path",
					Line: models.Position{
						Start: 3,
						End:   3,
					},
					Column: models.Position{
						Start: 24,
						End:   30,
					},
				},
			},
		},
		{
			name: "No package, no updates",
			fields: fields{
				LineOffset: 1,
			},
			args: args{
				pkg:     nil,
				indexes: []int{14, 32, 15, 19, 25, 29},
			},
		},
		{
			name: "No index is set, position is not updated",
			fields: fields{
				LineOffset: 1,
			},
			args: args{
				pkg: &PackageDetails{
					Name:    "Foobar",
					Version: "1.2.3",
				},
				content: `{
					"dependencies": {
						"Foobar": "^1.2.3"
					}
				}`,
				indexes: []int{},
			},
			expected: &PackageDetails{
				Name:     "Foobar",
				Version:  "1.2.3",
				IsDirect: true,
			},
		},
		{
			name: "depgroup is empty, no dep group is added",
			fields: fields{
				LineOffset: 1,
			},
			args: args{
				pkg: &PackageDetails{
					Name:    "Foobar",
					Version: "1.2.3",
				},
				content:  "",
				indexes:  []int{},
				depGroup: "",
			},
			expected: &PackageDetails{
				Name:     "Foobar",
				Version:  "1.2.3",
				IsDirect: true,
			},
		},
		{
			name: "depgroup is not empty, dep group is added",
			fields: fields{
				LineOffset: 1,
			},
			args: args{
				pkg: &PackageDetails{
					Name:    "Foobar",
					Version: "1.2.3",
				},
				content:  "",
				indexes:  []int{},
				depGroup: "prod",
			},
			expected: &PackageDetails{
				Name:      "Foobar",
				Version:   "1.2.3",
				IsDirect:  true,
				DepGroups: []string{"prod"},
			},
		},
		{
			name: "depgroup should be propagated to child dependencies",
			fields: fields{
				LineOffset: 1,
			},
			args: args{
				pkg: &PackageDetails{
					Name:    "Foobar",
					Version: "1.2.3",
					Dependencies: []*PackageDetails{
						{
							Name:    "Bar",
							Version: "2.1.3",
						},
					},
				},
				content:  "",
				indexes:  []int{},
				depGroup: "prod",
			},
			expected: &PackageDetails{
				Name:      "Foobar",
				Version:   "1.2.3",
				IsDirect:  true,
				DepGroups: []string{"prod"},
				Dependencies: []*PackageDetails{
					{
						Name:      "Bar",
						Version:   "2.1.3",
						DepGroups: []string{"prod"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			depMap := &MatcherDependencyMap{
				RootType:   0,
				FilePath:   "file-path",
				LineOffset: tt.fields.LineOffset,
				Packages:   []*PackageDetails{},
			}
			if tt.args.pkg != nil {
				depMap.Packages = append(depMap.Packages, tt.args.pkg)
			} else {
				depMap.Packages = append(depMap.Packages, &PackageDetails{
					Name:    "Foobar",
					Version: "1.2.3",
				})
			}
			depMap.UpdatePackageDetails(tt.args.pkg, tt.args.content, tt.args.indexes, tt.args.depGroup)

			if tt.expected != nil {
				assert.EqualExportedValues(t, *tt.expected, *depMap.Packages[0])
			} else {
				assert.EqualExportedValues(t, *depMap.Packages[0], PackageDetails{
					Name:    "Foobar",
					Version: "1.2.3",
				})
			}
		})
	}
}
