package cyclonedxxml

import (
	"github.com/CycloneDX/cyclonedx-go"

	"github.com/nextlinux/sbom/sbom/formats/common/cyclonedxhelpers"
	"github.com/nextlinux/sbom/sbom/sbom"
)

const ID sbom.FormatID = "cyclonedx-xml"

func Format() sbom.Format {
	return sbom.NewFormat(
		sbom.AnyVersion,
		encoder,
		cyclonedxhelpers.GetDecoder(cyclonedx.BOMFileFormatXML),
		cyclonedxhelpers.GetValidator(cyclonedx.BOMFileFormatXML),
		ID, "cyclonedx", "cyclone",
	)
}
