package v11

import (
	"bytes"
	"errors"
	"regexp"

	"github.com/alaingilbert/clockwork"
	v6 "github.com/alaingilbert/ogame/pkg/extractor/v6"
	"github.com/alaingilbert/ogame/pkg/ogame"
	"github.com/alaingilbert/ogame/pkg/utils"

	"github.com/PuerkitoBio/goquery"
)

func extractResourceSettingsFromPage(pageHTML []byte) (ogame.ResourceSettings, string, error) {
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	bodyID := v6.ExtractBodyIDFromDoc(doc)
	if bodyID == "overview" {
		return ogame.ResourceSettings{}, "", ogame.ErrInvalidPlanetID
	}
	vals := make([]int64, 0)
	doc.Find("option").Each(func(i int, s *goquery.Selection) {
		_, selectedExists := s.Attr("selected")
		if selectedExists {
			a, _ := s.Attr("value")
			val := utils.DoParseI64(a)
			vals = append(vals, val)
		}
	})
	if len(vals) != 7 {
		return ogame.ResourceSettings{}, "", errors.New("failed to find all resource settings")
	}

	res := ogame.ResourceSettings{}
	res.MetalMine = vals[0]
	res.CrystalMine = vals[1]
	res.DeuteriumSynthesizer = vals[2]
	res.SolarPlant = vals[3]
	res.FusionReactor = vals[4]
	res.SolarSatellite = vals[5]
	res.Crawler = vals[6]

	getToken := func(pageHTML []byte) (string, error) {
		m := regexp.MustCompile(`var token = "([^"]+)"`).FindSubmatch(pageHTML)
		if len(m) != 2 {
			return "", errors.New("unable to find token")
		}
		return string(m[1]), nil
	}
	token, _ := getToken(pageHTML)

	return res, token, nil
}

func extractConstructions(pageHTML []byte, clock clockwork.Clock) (buildingID ogame.ID, buildingCountdown int64,
	// OGame Version as of 11.13.0
	researchID ogame.ID, researchCountdown int64,
	lfBuildingID ogame.ID, lfBuildingCountdown int64,
	lfResearchID ogame.ID, lfResearchCountdown int64) {
	buildingCountdownMatch := regexp.MustCompile(`new CountdownTimer\('buildingCountdown', (\d+),`).FindSubmatch(pageHTML)
	if len(buildingCountdownMatch) > 0 {
		buildingCountdown = int64(utils.ToInt(buildingCountdownMatch[1]))
		buildingIDInt := utils.ToInt(regexp.MustCompile(`onclick="cancelbuilding\((\d+),`).FindSubmatch(pageHTML)[1])
		buildingID = ogame.ID(buildingIDInt)
	}
	researchCountdownMatch := regexp.MustCompile(`new CountdownTimer\('researchCountdown', (\d+),`).FindSubmatch(pageHTML)
	if len(researchCountdownMatch) > 0 {
		researchCountdown = int64(utils.ToInt(researchCountdownMatch[1]))
		researchIDInt := utils.ToInt(regexp.MustCompile(`onclick="cancelresearch\((\d+),`).FindSubmatch(pageHTML)[1])
		researchID = ogame.ID(researchIDInt)
	}
	lfBuildingCountdownMatch := regexp.MustCompile(`new CountdownTimer\('lfbuildingCountdown', (\d+),`).FindSubmatch(pageHTML)
	if len(lfBuildingCountdownMatch) > 0 {
		lfBuildingCountdown = int64(utils.ToInt(lfBuildingCountdownMatch[1]))
		lfBuildingIDInt := utils.ToInt(regexp.MustCompile(`onclick="cancellfbuilding\((\d+),`).FindSubmatch(pageHTML)[1])
		lfBuildingID = ogame.ID(lfBuildingIDInt)
	}
	lfResearchCountdownMatch := regexp.MustCompile(`new CountdownTimer\('lfresearchCountdown', (\d+),`).FindSubmatch(pageHTML)
	if len(lfResearchCountdownMatch) > 0 {
		lfResearchCountdown = int64(utils.ToInt(lfResearchCountdownMatch[1]))
		lfResearchIDInt := utils.ToInt(regexp.MustCompile(`onclick="cancellfresearch\((\d+),`).FindSubmatch(pageHTML)[1])
		lfResearchID = ogame.ID(lfResearchIDInt)
	}
	return
}

func extractProductionFromDoc(doc *goquery.Document) ([]ogame.Quantifiable, error) {
	res := make([]ogame.Quantifiable, 0)
	active := doc.Find("div.productionBoxShips  table.construction")
	href, _ := active.Find("td a").Attr("href")
	m := regexp.MustCompile(`openTech=(\d+)`).FindStringSubmatch(href)
	if len(m) == 0 {
		return []ogame.Quantifiable{}, nil
	}
	idInt := utils.DoParseI64(m[1])
	activeID := ogame.ID(idInt)
	activeNbr := utils.DoParseI64(active.Find("div.shipSumCount").Text())
	res = append(res, ogame.Quantifiable{ID: activeID, Nbr: activeNbr})
	doc.Find("div.productionBoxShips  table.queue td").Each(func(i int, s *goquery.Selection) {
		link := s.Find("img")
		alt := link.AttrOr("alt", "")
		var itemID ogame.ID
		if id := ogame.DefenceName2ID(alt); id.IsValid() {
			itemID = id
		} else if id := ogame.ShipName2ID(alt); id.IsValid() {
			itemID = id
		}
		if itemID.IsValid() {
			itemNbr := utils.ParseInt(s.Text())
			res = append(res, ogame.Quantifiable{ID: ogame.ID(itemID), Nbr: itemNbr})
		}
	})
	return res, nil
}

func extractOverviewShipSumCountdownFromBytes(pageHTML []byte) int64 {
	var shipSumCountdown int64
	shipSumCountdownMatch := regexp.MustCompile(`new CountdownTimer\('shipyardCountdown', (\d+),`).FindSubmatch(pageHTML)
	if len(shipSumCountdownMatch) > 0 {
		shipSumCountdown = int64(utils.ToInt(shipSumCountdownMatch[1]))
	}
	return shipSumCountdown
}
