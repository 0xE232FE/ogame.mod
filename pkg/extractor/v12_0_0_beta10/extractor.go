package v12_0_0_beta10

import (
	"bytes"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alaingilbert/ogame/pkg/extractor/v11_15_0"
	"github.com/alaingilbert/ogame/pkg/ogame"
)

// Extractor ...
type Extractor struct {
	v11_15_0.Extractor
}

// NewExtractor ...
func NewExtractor() *Extractor {
	return &Extractor{}
}

// ExtractAllianceClass ...
func (e *Extractor) ExtractAllianceClass(pageHTML []byte) (ogame.AllianceClass, error) {
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	return e.ExtractAllianceClassFromDoc(doc)
}

// ExtractAllianceClassFromDoc ...
func (e *Extractor) ExtractAllianceClassFromDoc(doc *goquery.Document) (ogame.AllianceClass, error) {
	return extractAllianceClassFromDoc(doc)
}

// ExtractServerTime ...
func (e *Extractor) ExtractServerTime(pageHTML []byte) (time.Time, error) {
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	return e.ExtractServerTimeFromDoc(doc)
}

// ExtractServerTimeFromDoc ...
func (e *Extractor) ExtractServerTimeFromDoc(doc *goquery.Document) (time.Time, error) {
	return extractServerTimeFromDoc(doc)
}

func (e *Extractor) ExtractTechinfo(pageHTML []byte) error {
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	return e.ExtractTechinfoFromDoc(doc)
}

func (e *Extractor) ExtractTechinfoFromDoc(doc *goquery.Document) error {
	return extractTechinfoFromDoc(doc)
}
