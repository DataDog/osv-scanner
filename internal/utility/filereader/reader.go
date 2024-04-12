package filereader

import (
	"bufio"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io"
)

func NewScanner(reader io.Reader) *bufio.Scanner {
	var transformer = unicode.BOMOverride(encoding.Nop.NewDecoder())
	decodedReader := transform.NewReader(reader, transformer)
	return bufio.NewScanner(decodedReader)
}
