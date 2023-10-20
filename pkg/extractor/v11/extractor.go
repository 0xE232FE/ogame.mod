package v11

import (
	v104 "github.com/alaingilbert/ogame/pkg/extractor/v104"
)

// Extractor ...
type Extractor struct {
	v104.Extractor
}

// NewExtractor ...
func NewExtractor() *Extractor {
	return &Extractor{}
}
