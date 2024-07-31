package models

import (
	"encoding/json"
	"strconv"
	"strings"
)

type PackageDetails struct {
	Name      string
	Version   string
	Ecosystem string
	Locations []PackageLocations
}

type PackageLocation struct {
	Filename    string `json:"file_name"`
	LineStart   int    `json:"line_start"`
	LineEnd     int    `json:"line_end"`
	ColumnStart int    `json:"column_start"`
	ColumnEnd   int    `json:"column_end"`
}

type PackageLocations struct {
	Block     PackageLocation  `json:"block"`
	Namespace *PackageLocation `json:"namespace,omitempty"`
	Name      *PackageLocation `json:"name,omitempty"`
	Version   *PackageLocation `json:"version,omitempty"`
}

func NewPackageLocations(block FilePosition, name *FilePosition, version *FilePosition) PackageLocations {
	result := PackageLocations{
		Block: PackageLocation{
			Filename:    block.Filename,
			LineStart:   block.Line.Start,
			LineEnd:     block.Line.End,
			ColumnStart: block.Column.Start,
			ColumnEnd:   block.Column.End,
		},
	}

	if name != nil {
		result.Name = &PackageLocation{
			Filename:    name.Filename,
			LineStart:   name.Line.Start,
			LineEnd:     name.Line.End,
			ColumnStart: name.Column.Start,
			ColumnEnd:   name.Column.End,
		}
	}
	if version != nil {
		result.Version = &PackageLocation{
			Filename:    version.Filename,
			LineStart:   version.Line.Start,
			LineEnd:     version.Line.End,
			ColumnStart: version.Column.Start,
			ColumnEnd:   version.Column.End,
		}
	}

	return result
}

func (location PackageLocations) MarshalToJSONString() (string, error) {
	str, err := json.Marshal(location)
	if err != nil {
		return "", err
	}

	return string(str), nil
}

func (location PackageLocation) Hash() string {
	return strings.Join([]string{
		location.Filename,
		strconv.Itoa(location.LineStart),
		strconv.Itoa(location.LineEnd),
		strconv.Itoa(location.ColumnStart),
		strconv.Itoa(location.ColumnEnd),
	}, "#")
}
