package sbomjson

import (
	"encoding/json"
	"io"

	"github.com/nextlinux/sbom/sbom/sbom"
)

func encoder(output io.Writer, s sbom.SBOM) error {
	doc := ToFormatModel(s)

	enc := json.NewEncoder(output)
	// prevent > and < from being escaped in the payload
	enc.SetEscapeHTML(false)
	enc.SetIndent("", " ")

	return enc.Encode(&doc)
}
