package v11

import (
	"bytes"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alaingilbert/clockwork"
	v104 "github.com/alaingilbert/ogame/pkg/extractor/v104"
	"github.com/alaingilbert/ogame/pkg/ogame"
)

// Extractor ...
type Extractor struct {
	v104.Extractor
}

// NewExtractor ...
func NewExtractor() *Extractor {
	return &Extractor{}
}

// ExtractResourceSettings ...
func (e Extractor) ExtractResourceSettings(pageHTML []byte) (ogame.ResourceSettings, string, error) {
	return extractResourceSettingsFromPage(pageHTML)
}

// ExtractConstructions ...
func (e *Extractor) ExtractConstructions(pageHTML []byte) (buildingID ogame.ID, buildingCountdown int64, researchID ogame.ID, researchCountdown int64, lfBuildingID ogame.ID, lfBuildingCountdown int64, lfResearchID ogame.ID, lfResearchCountdown int64) {
	return extractConstructions(pageHTML, clockwork.NewRealClock())
}

// ExtractProduction extracts ships/defenses production from the shipyard page
func (e *Extractor) ExtractProduction(pageHTML []byte) ([]ogame.Quantifiable, int64, error) {
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	shipSumCountdown := extractOverviewShipSumCountdownFromBytes(pageHTML)
	production, err := extractProductionFromDoc(doc)
	return production, shipSumCountdown, err
}

// ExtractProductionFromDoc extracts ships/defenses production from the shipyard page
func (e *Extractor) ExtractProductionFromDoc(doc *goquery.Document) ([]ogame.Quantifiable, error) {
	return extractProductionFromDoc(doc)
}

// ExtractOverviewShipSumCountdownFromBytes extracts production countdown
func (e *Extractor) ExtractOverviewShipSumCountdownFromBytes(pageHTML []byte) int64 {
	return extractOverviewShipSumCountdownFromBytes(pageHTML)
}

// ExtractFleets ...
func (e *Extractor) ExtractFleets(pageHTML []byte) (res []ogame.Fleet) {
	return e.extractFleets(pageHTML, e.GetLocation())
}

func (e *Extractor) extractFleets(pageHTML []byte, location *time.Location) (res []ogame.Fleet) {
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	return extractFleetsFromDoc(doc, location, e.GetLifeformEnabled())
}

// ExtractFleetsFromDoc ...
func (e *Extractor) ExtractFleetsFromDoc(doc *goquery.Document) (res []ogame.Fleet) {
	return e.extractFleetsFromDoc(doc, e.GetLocation())
}

func (e *Extractor) extractFleetsFromDoc(doc *goquery.Document, location *time.Location) (res []ogame.Fleet) {
	return extractFleetsFromDoc(doc, location, e.GetLifeformEnabled())
}
