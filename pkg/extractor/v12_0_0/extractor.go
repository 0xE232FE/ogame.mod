package v12_0_0

import (
	"bytes"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alaingilbert/clockwork"
	"github.com/alaingilbert/ogame/pkg/extractor/v12_0_0_beta10"
	"github.com/alaingilbert/ogame/pkg/ogame"
)

// Extractor ...
type Extractor struct {
	v12_0_0_beta10.Extractor
}

// NewExtractor ...
func NewExtractor() *Extractor {
	return &Extractor{}
}

// ExtractServerTime ...
func (e *Extractor) ExtractServerTime(pageHTML []byte) (time.Time, error) {
	clock := clockwork.NewRealClock()
	return e.extractServerTime(pageHTML, clock)
}

func (e *Extractor) extractServerTime(pageHTML []byte, clock clockwork.Clock) (time.Time, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	if err != nil {
		return time.Time{}, err
	}
	return extractServerTimeFromDoc(doc, clock)
}

// ExtractServerTimeFromDoc ...
func (e *Extractor) ExtractServerTimeFromDoc(doc *goquery.Document) (time.Time, error) {
	clock := clockwork.NewRealClock()
	return extractServerTimeFromDoc(doc, clock)
}

// ExtractHighscore ...
func (e *Extractor) ExtractHighscore(pageHTML []byte) (ogame.Highscore, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	if err != nil {
		return ogame.Highscore{}, err
	}
	return e.ExtractHighscoreFromDoc(doc)
}

// ExtractHighscoreFromDoc ...
func (e *Extractor) ExtractHighscoreFromDoc(doc *goquery.Document) (ogame.Highscore, error) {
	return extractHighscoreFromDoc(doc)
}
