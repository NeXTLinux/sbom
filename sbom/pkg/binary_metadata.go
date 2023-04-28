package pkg

import "github.com/anchore/sbom/sbom/source"

type BinaryMetadata struct {
	Matches []ClassifierMatch `mapstructure:"Matches" json:"matches"`
}

type ClassifierMatch struct {
	Classifier string          `mapstructure:"Classifier" json:"classifier"`
	Location   source.Location `mapstructure:"Location" json:"location"`
}
