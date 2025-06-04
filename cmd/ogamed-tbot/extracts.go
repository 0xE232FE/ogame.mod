package main

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/alaingilbert/ogame/pkg/ogame"
	"github.com/alaingilbert/ogame/pkg/utils"
	"golang.org/x/net/html"
)

func extractLfBonusesFromDoc(doc *goquery.Document) (ogame.LfBonuses, error) {
	b := ogame.NewLfBonuses()
	doc.Find("bonus-item-content[data-toggable-target^=category]").Each(func(_ int, s *goquery.Selection) {
		category := s.AttrOr("data-toggable-target", "")
		if category == "categoryShips" || category == "categoryCostAndTime" {
			s.Find("inner-bonus-item-heading[data-toggable^=subcategory]").Each(func(_ int, g *goquery.Selection) {
				category, subcategory := extractCategories(g, category)
				if category != "" && subcategory != "" {
					assignBonusValue(g, b, category, subcategory)
				}
			})
		}
	})
	return *b, nil
}

func extractCategories(g *goquery.Selection, category string) (string, string) {
	c := strings.Replace(category, "category", "", 1)
	s, _ := g.Attr("data-toggable")
	v := "sub" + category
	return c, strings.Replace(s, v, "", 1)
}

// ExtractLfBonuses ...
func ExtractLfBonuses(pageHTML []byte) (ogame.LfBonuses, error) {
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	return ExtractLfBonusesFromDoc(doc)
}

// ExtractLfBonusesFromDoc ...
func ExtractLfBonusesFromDoc(doc *goquery.Document) (ogame.LfBonuses, error) {
	return extractLfBonusesFromDoc(doc)
}

func assignBonusValue(s *goquery.Selection, b *ogame.LfBonuses, category string, subcategory string) {
	switch category {
	case "Ships":
		extractShipStatBonus(s, b, subcategory)
	case "CostAndTime":
		extractCostReductionBonus(s, b, subcategory)
		extractTimeReductionBonus(s, b, subcategory)
	}
}

// Extracts ships stats fixed
func extractShipStatBonus(s *goquery.Selection, b *ogame.LfBonuses, subcategory string) {
	i := utils.DoParseI64(subcategory)
	id := ogame.ID(i)
	if !id.IsShip() {
		return
	}
	s.Find("bonus-items").Each(func(_ int, s *goquery.Selection) {
		extractFn := func(idx int) float64 {
			txt := s.Children().Eq(idx).Contents().FilterFunction(func(i int, s *goquery.Selection) bool {
				return s.Nodes[0].Type == html.TextNode
			}).Text()
			return extractBonusFromStringPercentage(txt)
		}
		shipBonus := ogame.LfShipBonus{
			ID:                  id,
			StructuralIntegrity: extractFn(0),
			ShieldPower:         extractFn(1),
			WeaponPower:         extractFn(2),
			Speed:               extractFn(3),
			CargoCapacity:       extractFn(4),
			FuelConsumption:     extractFn(5),
		}
		b.LfShipBonuses[id] = shipBonus
	})
	return
}

// Extracts cost reduction
func extractCostReductionBonus(s *goquery.Selection, l *ogame.LfBonuses, subcategory string) {
	i := utils.DoParseI64(subcategory)
	id := ogame.ID(i)
	s.Find("bonus-items").Each(func(_ int, s *goquery.Selection) {
		txt := s.Eq(0).Children().Eq(0).Contents().FilterFunction(func(i int, s *goquery.Selection) bool {
			return s.Nodes[0].Type == html.TextNode
		}).Text()
		costTimeBonus := l.CostTimeBonuses[id]
		costTimeBonus.Cost = extractBonusFromStringPercentage(txt)
		l.CostTimeBonuses[id] = costTimeBonus
	})
}

// Extracts time reduction
func extractTimeReductionBonus(s *goquery.Selection, l *ogame.LfBonuses, subcategory string) {
	i := utils.DoParseI64(subcategory)
	id := ogame.ID(i)
	s.Find("bonus-items").Each(func(_ int, s *goquery.Selection) {
		txt := s.Eq(0).Children().Eq(1).Contents().FilterFunction(func(i int, s *goquery.Selection) bool {
			return s.Nodes[0].Type == html.TextNode
		}).Text()
		costTimeBonus := l.CostTimeBonuses[id]
		costTimeBonus.Duration = extractBonusFromStringPercentage(txt)
		l.CostTimeBonuses[id] = costTimeBonus
	})
}

// Extract bonus value from a string with percentage sign [ex: 1.056% -> 0.01056]
func extractBonusFromStringPercentage(s string) float64 {
	v := strings.Replace(s, "%", "", 1)
	return extractBonusFromString(v) / 100.0
}

// Extract bonus value from a string [ex: 1.056]
func extractBonusFromString(s string) float64 {
	v := strings.TrimSpace(s)
	v = strings.Replace(v, ",", ".", 1)
	b, _ := strconv.ParseFloat(v, 64)
	return utils.RoundThousandth(b)
}
