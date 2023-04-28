/*
Package kernel provides a concrete Cataloger implementation for linux kernel and module files.
*/
package kernel

import (
	"github.com/hashicorp/go-multierror"

	"github.com/anchore/sbom/internal/log"
	"github.com/anchore/sbom/sbom/artifact"
	"github.com/anchore/sbom/sbom/pkg"
	"github.com/anchore/sbom/sbom/pkg/cataloger/generic"
	"github.com/anchore/sbom/sbom/source"
)

var _ pkg.Cataloger = (*LinuxKernelCataloger)(nil)

type LinuxCatalogerConfig struct {
	CatalogModules bool
}

type LinuxKernelCataloger struct {
	cfg LinuxCatalogerConfig
}

func DefaultLinuxCatalogerConfig() LinuxCatalogerConfig {
	return LinuxCatalogerConfig{
		CatalogModules: true,
	}
}

var kernelArchiveGlobs = []string{
	"**/kernel",
	"**/kernel-*",
	"**/vmlinux",
	"**/vmlinux-*",
	"**/vmlinuz",
	"**/vmlinuz-*",
}

var kernelModuleGlobs = []string{
	"**/lib/modules/**/*.ko",
}

// NewLinuxKernelCataloger returns a new kernel files cataloger object.
func NewLinuxKernelCataloger(cfg LinuxCatalogerConfig) *LinuxKernelCataloger {
	return &LinuxKernelCataloger{
		cfg: cfg,
	}
}

func (l LinuxKernelCataloger) Name() string {
	return "linux-kernel-cataloger"
}

func (l LinuxKernelCataloger) Catalog(resolver source.FileResolver) ([]pkg.Package, []artifact.Relationship, error) {
	var allPackages []pkg.Package
	var allRelationships []artifact.Relationship
	var errs error

	kernelPackages, kernelRelationships, err := generic.NewCataloger(l.Name()).WithParserByGlobs(parseLinuxKernelFile, kernelArchiveGlobs...).Catalog(resolver)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	allRelationships = append(allRelationships, kernelRelationships...)
	allPackages = append(allPackages, kernelPackages...)

	if l.cfg.CatalogModules {
		modulePackages, moduleRelationships, err := generic.NewCataloger(l.Name()).WithParserByGlobs(parseLinuxKernelModuleFile, kernelModuleGlobs...).Catalog(resolver)
		if err != nil {
			errs = multierror.Append(errs, err)
		}

		allPackages = append(allPackages, modulePackages...)

		moduleToKernelRelationships := createKernelToModuleRelationships(kernelPackages, modulePackages)
		allRelationships = append(allRelationships, moduleRelationships...)
		allRelationships = append(allRelationships, moduleToKernelRelationships...)
	}

	return allPackages, allRelationships, errs
}

func createKernelToModuleRelationships(kernelPackages, modulePackages []pkg.Package) []artifact.Relationship {
	// organize kernel and module packages by kernel version
	kernelPackagesByVersion := make(map[string][]*pkg.Package)
	for idx, p := range kernelPackages {
		kernelPackagesByVersion[p.Version] = append(kernelPackagesByVersion[p.Version], &kernelPackages[idx])
	}

	modulesByKernelVersion := make(map[string][]*pkg.Package)
	for idx, p := range modulePackages {
		m, ok := p.Metadata.(pkg.LinuxKernelModuleMetadata)
		if !ok {
			log.Debug("linux-kernel-module package found without metadata: %s@%s", p.Name, p.Version)
			continue
		}
		modulesByKernelVersion[m.KernelVersion] = append(modulesByKernelVersion[m.KernelVersion], &modulePackages[idx])
	}

	// create relationships between kernel and modules: [module] --(depends on)--> [kernel]
	// since we try to use singular directions for relationships, we'll use "dependency of" here instead:
	// [kernel] --(dependency of)--> [module]
	var moduleToKernelRelationships []artifact.Relationship
	for kernelVersion, modules := range modulesByKernelVersion {
		kps, ok := kernelPackagesByVersion[kernelVersion]
		if !ok {
			// it's ok if there is a module that has no installed kernel...
			continue
		}

		// we don't know which kernel is the "right" one, so we'll create a relationship for each one
		for _, kp := range kps {
			for _, mp := range modules {
				moduleToKernelRelationships = append(moduleToKernelRelationships, artifact.Relationship{
					// note: relationships should have Package objects, not pointers
					From: *kp,
					To:   *mp,
					Type: artifact.DependencyOfRelationship,
				})
			}
		}
	}

	return moduleToKernelRelationships
}
