package v11_9_0

import (
	"bytes"

	"github.com/PuerkitoBio/goquery"
	"github.com/alaingilbert/clockwork"
	v11 "github.com/alaingilbert/ogame/pkg/extractor/v11"
	"github.com/alaingilbert/ogame/pkg/ogame"
)

// Extractor ...
type Extractor struct {
	v11.Extractor
}

// NewExtractor ...
func NewExtractor() *Extractor {
	return &Extractor{}
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

// ExtractCombatReportMessagesSummary ...
func (e *Extractor) ExtractCombatReportMessagesSummary(pageHTML []byte) ([]ogame.CombatReportSummary, int64, error) {
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	return e.ExtractCombatReportMessagesFromDoc(doc)
}

// ExtractCombatReportMessagesFromDoc ...
func (e *Extractor) ExtractCombatReportMessagesFromDoc(doc *goquery.Document) ([]ogame.CombatReportSummary, int64, error) {
	return extractCombatReportMessagesFromDoc(doc)
}

// ExtractAuction ...
func (e *Extractor) ExtractAuction(pageHTML []byte) (ogame.Auction, error) {
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	return extractAuctionFromDoc(doc)
}

// ExtractConstructions ...
func (e *Extractor) ExtractConstructions(pageHTML []byte) (buildingID ogame.ID, buildingCountdown int64, researchID ogame.ID, researchCountdown int64, lfBuildingID ogame.ID, lfBuildingCountdown int64, lfResearchID ogame.ID, lfResearchCountdown int64) {
	return extractConstructions(pageHTML, clockwork.NewRealClock())
}
