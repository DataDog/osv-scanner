package models

import (
	"encoding/xml"
	"strings"
)

/*
StringHolder is a structure meant to deserialize string data along with the position of the data in the file when they can be mixed with other such as :
  - spaces, tabs and newline for formatting
  - comments (for example with XML based files)

It supports XML based files through the UnmarshalXML method
*/
type StringHolder struct {
	Value string
	FilePosition
}

func (stringHolder *StringHolder) computePositions(content, trimmedString string, lineStart, columnStart int) {
	// Lets compute where it starts
	stringStart := strings.Index(content, trimmedString)
	endOfLineCount := strings.Count(content[:stringStart], "\n")

	if !stringHolder.IsStartSet() {
		stringHolder.SetLineStart(lineStart + endOfLineCount)
	}
	stringHolder.SetLineEnd(lineStart + endOfLineCount)

	if endOfLineCount == 0 {
		// content is on the same line than tag start, we need to take the existing offset into account
		contentPrefixSize := len(content[:stringStart])
		if !stringHolder.IsStartSet() {
			stringHolder.SetColumnStart(columnStart + contentPrefixSize)
		}
		stringHolder.SetColumnEnd(columnStart + contentPrefixSize + len(trimmedString))
	} else {
		// content is not on the same line, column count is reset to 0
		contentLineStart := strings.LastIndex(content[:stringStart], "\n") + 1
		contentPrefixSize := len(content[contentLineStart:stringStart])

		if !stringHolder.IsStartSet() {
			stringHolder.SetColumnStart(contentPrefixSize + 1)
		}
		stringHolder.SetColumnEnd(contentPrefixSize + len(trimmedString) + 1)
	}
}

func (stringHolder *StringHolder) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	for {
		lineStart, columnStart := decoder.InputPos()
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch se := token.(type) {
		case xml.EndElement:
			if se.Name == start.Name {
				return nil
			}
		case xml.CharData:
			content := string(se)
			trimmedString := strings.TrimSpace(content)
			if len(trimmedString) > 0 {
				// We have string content in there (not space, not a comment)
				stringHolder.Value += trimmedString
				stringHolder.computePositions(content, trimmedString, lineStart, columnStart)
			}
		}
	}
}
