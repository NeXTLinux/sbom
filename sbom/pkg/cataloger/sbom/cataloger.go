package sbom

import (
	"github.com/anchore/sbom/internal/log"
	"github.com/anchore/sbom/sbom/artifact"
	"github.com/anchore/sbom/sbom/formats"
	"github.com/anchore/sbom/sbom/pkg"
	"github.com/anchore/sbom/sbom/pkg/cataloger/generic"
	"github.com/anchore/sbom/sbom/source"
)

const catalogerName = "sbom-cataloger"

// NewSBOMCataloger returns a new SBOM cataloger object loaded from saved SBOM JSON.
func NewSBOMCataloger() *generic.Cataloger {
	return generic.NewCataloger(catalogerName).
		WithParserByGlobs(parseSBOM,
			"**/*.sbom.json",
			"**/*.bom.*",
			"**/*.bom",
			"**/bom",
			"**/*.sbom.*",
			"**/*.sbom",
			"**/sbom",
			"**/*.cdx.*",
			"**/*.cdx",
			"**/*.spdx.*",
			"**/*.spdx",
		)
}

func parseSBOM(_ source.FileResolver, _ *generic.Environment, reader source.LocationReadCloser) ([]pkg.Package, []artifact.Relationship, error) {
	s, _, err := formats.Decode(reader)
	if err != nil {
		return nil, nil, err
	}

	if s == nil {
		log.WithFields("path", reader.Location.RealPath).Trace("file is not an SBOM")
		return nil, nil, nil
	}

	var pkgs []pkg.Package
	var relationships []artifact.Relationship
	for _, p := range s.Artifacts.PackageCatalog.Sorted() {
		// replace all locations on the package with the location of the SBOM file.
		// Why not keep the original list of locations? Since the "locations" field is meant to capture
		// where there is evidence of this file, and the catalogers have not run against any file other than,
		// the SBOM, this is the only location that is relevant for this cataloger.
		p.Locations = source.NewLocationSet(
			reader.Location.WithAnnotation(pkg.EvidenceAnnotationKey, pkg.PrimaryEvidenceAnnotation),
		)
		p.FoundBy = catalogerName

		pkgs = append(pkgs, p)
		relationships = append(relationships, artifact.Relationship{
			From: p,
			To:   reader.Location.Coordinates,
			Type: artifact.DescribedByRelationship,
		})
	}

	return pkgs, relationships, nil
}
