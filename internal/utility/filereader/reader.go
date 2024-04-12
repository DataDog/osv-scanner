package filereader

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io"
)

func CharsetDecoder(_ string, input io.Reader) (io.Reader, error) {
	var transformer = unicode.BOMOverride(encoding.Nop.NewDecoder())
	return transform.NewReader(input, transformer), nil
}
