package v11

import (
	"bytes"
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

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

func extractFleetsFromDoc(doc *goquery.Document, location *time.Location, lifeformEnabled bool) (res []ogame.Fleet) {
	res = make([]ogame.Fleet, 0)
	script := doc.Find("body script").Text()
	doc.Find("div.fleetDetails").Each(func(i int, s *goquery.Selection) {
		originText := s.Find("span.originCoords a").Text()
		origin := v6.ExtractCoord(originText)
		origin.Type = ogame.PlanetType
		if s.Find("span.originPlanet figure").HasClass("moon") {
			origin.Type = ogame.MoonType
		}

		destText := s.Find("span.destinationCoords a").Text()
		dest := v6.ExtractCoord(destText)
		dest.Type = ogame.PlanetType
		if s.Find("span.destinationPlanet figure").HasClass("moon") {
			dest.Type = ogame.MoonType
		} else if s.Find("span.destinationPlanet figure").HasClass("tf") {
			dest.Type = ogame.DebrisType
		}

		id := utils.DoParseI64(s.Find("a.openCloseDetails").AttrOr("data-mission-id", "0"))

		timerID := s.Find("span.timer").AttrOr("id", "")
		m := regexp.MustCompile(`new SimpleCountdownTimer\(\s*"#` + timerID + `",\s*(\d+)`).FindStringSubmatch(script)
		var arriveIn int64
		if len(m) == 2 {
			arriveIn = utils.DoParseI64(m[1])
		}

		timerNextID := s.Find("span.nextTimer").AttrOr("id", "")
		m = regexp.MustCompile(`getElementByIdWithCache\("` + timerNextID + `"\),\s*(\d+)\s*\);`).FindStringSubmatch(script)
		var backIn int64
		if len(m) == 2 {
			backIn = utils.DoParseI64(m[1])
		}

		missionType := utils.DoParseI64(s.AttrOr("data-mission-type", ""))
		returnFlight, _ := strconv.ParseBool(s.AttrOr("data-return-flight", ""))
		inDeepSpace := s.Find("span.fleetDetailButton a").HasClass("fleet_icon_forward_end")
		arrivalTime := utils.DoParseI64(s.AttrOr("data-arrival-time", ""))
		endTime := utils.DoParseI64(s.Find("a.openCloseDetails").AttrOr("data-end-time", ""))

		trs := s.Find("table.fleetinfo tr")
		shipment := ogame.Resources{}
		metalTrOffset := 3
		crystalTrOffset := 2
		DeuteriumTrOffset := 1
		if lifeformEnabled {
			metalTrOffset = 4
			crystalTrOffset = 3
			DeuteriumTrOffset = 2
		}
		shipment.Metal = utils.ParseInt(trs.Eq(trs.Size() - metalTrOffset).Find("td").Eq(1).Text())
		shipment.Crystal = utils.ParseInt(trs.Eq(trs.Size() - crystalTrOffset).Find("td").Eq(1).Text())
		shipment.Deuterium = utils.ParseInt(trs.Eq(trs.Size() - DeuteriumTrOffset).Find("td").Eq(1).Text())

		fedAttackHref := s.Find("span.fedAttack a").AttrOr("href", "")
		fedAttackURL, _ := url.Parse(fedAttackHref)
		fedAttackQuery := fedAttackURL.Query()
		targetPlanetID := utils.DoParseI64(fedAttackQuery.Get("target"))
		unionID := utils.DoParseI64(fedAttackQuery.Get("union"))

		fleet := ogame.Fleet{}
		fleet.ID = ogame.FleetID(id)
		fleet.Origin = origin
		fleet.Destination = dest
		fleet.Mission = ogame.MissionID(missionType)
		fleet.ReturnFlight = returnFlight
		fleet.InDeepSpace = inDeepSpace
		fleet.Resources = shipment
		fleet.TargetPlanetID = targetPlanetID
		fleet.UnionID = unionID
		fleet.ArrivalTime = time.Unix(endTime, 0)
		fleet.BackTime = time.Unix(arrivalTime, 0)

		var startTimeString string
		var startTimeStringExists bool
		if !returnFlight {
			fleet.ArriveIn = arriveIn
			fleet.BackIn = backIn
			startTimeString, startTimeStringExists = s.Find("div.origin img").Attr("title")
		} else {
			fleet.ArriveIn = -1
			fleet.BackIn = arriveIn
			startTimeString, startTimeStringExists = s.Find("div.destination img").Attr("title")
		}

		var startTime time.Time
		if startTimeStringExists {
			startTimeArray := strings.Split(startTimeString, ":| ")
			if len(startTimeArray) == 2 {
				startTime, _ = time.ParseInLocation("02.01.2006<br>15:04:05", startTimeArray[1], location)
			}
		}
		fleet.StartTime = startTime.Local()

		for i := 1; i < trs.Size()-5; i++ {
			tds := trs.Eq(i).Find("td")
			name := strings.ToLower(strings.Trim(strings.TrimSpace(tds.Eq(0).Text()), ":"))
			qty := utils.ParseInt(tds.Eq(1).Text())
			shipID := ogame.ShipName2ID(name)
			fleet.Ships.Set(shipID, qty)
		}

		res = append(res, fleet)
	})
	return
}
