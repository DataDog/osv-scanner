package json

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExtractPackageIndexes(t *testing.T) {
	type args struct {
		pkgName         string
		targetedVersion string
		content         string
	}
	tests := []struct {
		name     string
		args     args
		expected []int
	}{
		{
			name: "No targeted version set, matching all versions",
			args: args{
				pkgName:         "foo",
				targetedVersion: "",
				content:         `   "foo": "~1.2.3"`,
			},
			expected: []int{3, 18, 4, 7, 11, 17},
		},
		{
			name: "targeted version set, matching only targeted version",
			args: args{
				pkgName:         "foo",
				targetedVersion: "~1.2.3",
				content:         `   "foo": "~1.2.3"`,
			},
			expected: []int{3, 18, 4, 7, 11, 17},
		},
		{
			name: "package is not the one targeted",
			args: args{
				pkgName:         "bar",
				targetedVersion: "~1.2.3",
				content:         `   "foo": "~1.2.3"`,
			},
			expected: []int{},
		},
		{
			name: "package is not the one targeted, matching all versions",
			args: args{
				pkgName:         "bar",
				targetedVersion: "",
				content:         `   "foo": "~1.2.3"`,
			},
			expected: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractPackageIndexes(tt.args.pkgName, tt.args.targetedVersion, tt.args.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}
