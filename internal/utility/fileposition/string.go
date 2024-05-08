package fileposition

import (
	"github.com/google/osv-scanner/pkg/models"
	"strings"
)

func ExtractStringPositionInBlock(block []string, str string, blockStartLine int) *models.FilePosition {
	for i, line := range block {
		if strings.Contains(line, str) {
			linePosition := blockStartLine + i
			columnStart := strings.Index(line, str) + 1
			columnEnd := columnStart + len(str)

			return &models.FilePosition{
				Line:   models.Position{Start: linePosition, End: linePosition},
				Column: models.Position{Start: columnStart, End: columnEnd},
			}
		}
	}
	return nil
}
