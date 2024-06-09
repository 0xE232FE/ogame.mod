package v11_9_0

import (
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alaingilbert/clockwork"
	v6 "github.com/alaingilbert/ogame/pkg/extractor/v6"
	"github.com/alaingilbert/ogame/pkg/ogame"
	"github.com/alaingilbert/ogame/pkg/utils"
)

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

func extractCombatReportMessagesFromDoc(doc *goquery.Document) ([]ogame.CombatReportSummary, int64, error) {
	msgs := make([]ogame.CombatReportSummary, 0)
	nbPage := utils.DoParseI64(doc.Find("ul.pagination li").Last().AttrOr("data-page", "1"))
	doc.Find("li.msg").Each(func(i int, s *goquery.Selection) {
		if idStr, exists := s.Attr("data-msg-id"); exists {
			if id, err := utils.ParseI64(idStr); err == nil {
				report := ogame.CombatReportSummary{ID: id}
				report.Destination = v6.ExtractCoord(s.Find("div.msg_head a").Text())
				if s.Find("div.msg_head figure").HasClass("planet") {
					report.Destination.Type = ogame.PlanetType
				} else if s.Find("div.msg_head figure").HasClass("moon") {
					report.Destination.Type = ogame.MoonType
				} else {
					report.Destination.Type = ogame.PlanetType
				}
				apiKeyTitle := s.Find("span.icon_apikey").AttrOr("title", "")
				m := regexp.MustCompile(`'(cr-[^']+)'`).FindStringSubmatch(apiKeyTitle)
				if len(m) == 2 {
					report.APIKey = m[1]
				}
				resTitle := s.Find("span.msg_content div.combatLeftSide span").Eq(1).AttrOr("title", "")
				m = regexp.MustCompile(`([\d.,]+)<br/>[^\d]*([\d.,]+)<br/>[^\d]*([\d.,]+)`).FindStringSubmatch(resTitle)
				if len(m) == 4 {
					report.Metal = utils.ParseInt(m[1])
					report.Crystal = utils.ParseInt(m[2])
					report.Deuterium = utils.ParseInt(m[3])
				}
				debrisFieldTitle := s.Find("span.msg_content div.combatLeftSide span").Eq(2).AttrOr("title", "0")
				report.DebrisField = utils.ParseInt(debrisFieldTitle)
				resText := s.Find("span.msg_content div.combatLeftSide span").Eq(1).Text()
				m = regexp.MustCompile(`[\d.,]+[^\d]*([\d.,]+)`).FindStringSubmatch(resText)
				if len(m) == 2 {
					report.Loot = utils.ParseInt(m[1])
				}
				msgDate, _ := time.Parse("02.01.2006 15:04:05", s.Find("span.msg_date").Text())
				report.CreatedAt = msgDate

				link := s.Find("message-footer.msg_actions button.msgAttackBtn").AttrOr("onclick", "")
				m = regexp.MustCompile(`page=ingame&component=fleetdispatch&galaxy=(\d+)&system=(\d+)&position=(\d+)&type=(\d+)&`).FindStringSubmatch(link)
				if len(m) != 5 {
					return
				}
				galaxy := utils.DoParseI64(m[1])
				system := utils.DoParseI64(m[2])
				position := utils.DoParseI64(m[3])
				planetType := utils.DoParseI64(m[4])
				report.Origin = &ogame.Coordinate{Galaxy: galaxy, System: system, Position: position, Type: ogame.CelestialType(planetType)}
				if report.Origin.Equal(report.Destination) {
					report.Origin = nil
				}

				msgs = append(msgs, report)
			}
		}
	})
	return msgs, nbPage, nil
}

// extractAuctionFromDoc extract auction information from page "traderAuctioneer"
func extractAuctionFromDoc(doc *goquery.Document) (ogame.Auction, error) {
	auction := ogame.Auction{}
	auction.HasFinished = false

	// Detect if Auction has already finished
	nextAuction := doc.Find("#nextAuction")
	if nextAuction.Size() > 0 {
		// Find time until next auction starts
		auction.Endtime = utils.DoParseI64(nextAuction.Text())
		auction.HasFinished = true
	} else {
		endAtApprox := doc.Find("p.auction_info span").Text()
		m := regexp.MustCompile(`[^\d]+(\d+).*`).FindStringSubmatch(endAtApprox)
		if len(m) != 2 {
			return ogame.Auction{}, errors.New("failed to find end time approx")
		}
		endTimeMinutes, err := utils.ParseI64(m[1])
		if err != nil {
			return ogame.Auction{}, errors.New("invalid end time approx: " + err.Error())
		}
		auction.Endtime = endTimeMinutes * 60
	}

	auction.HighestBidder = strings.TrimSpace(doc.Find("a.currentPlayer").Text())
	auction.HighestBidderUserID = utils.DoParseI64(doc.Find("a.currentPlayer").AttrOr("data-player-id", ""))
	auction.NumBids = utils.DoParseI64(doc.Find("div.numberOfBids").Text())
	auction.CurrentBid = utils.ParseInt(doc.Find("div.currentSum").Text())
	auction.Inventory = utils.DoParseI64(doc.Find("span.level.amount").Text())
	auction.Ref = doc.Find("a.detail_button").First().AttrOr("ref", "")
	auction.CurrentItem = strings.ToLower(doc.Find("img").First().AttrOr("alt", ""))
	auction.CurrentItemLong = strings.ToLower(doc.Find("div.image_140px").First().Find("a").First().AttrOr("title", ""))
	multiplierRegex := regexp.MustCompile(`multiplier\s?=\s?([^;]+);`).FindStringSubmatch(doc.Text())
	if len(multiplierRegex) != 2 {
		return ogame.Auction{}, errors.New("failed to find auction multiplier")
	}
	if err := json.Unmarshal([]byte(multiplierRegex[1]), &auction.ResourceMultiplier); err != nil {
		return ogame.Auction{}, errors.New("failed to json parse auction multiplier: " + err.Error())
	}

	// Find auctioneer token
	tokenRegex := regexp.MustCompile(`token\s?=\s?"([^"]+)";`).FindStringSubmatch(doc.Text())
	if len(tokenRegex) != 2 {
		return ogame.Auction{}, errors.New("failed to find auctioneer token")
	}
	auction.Token = tokenRegex[1]

	// Find Planet / Moon resources JSON
	planetMoonResources := regexp.MustCompile(`planetResources\s?=\s?([^;]+);`).FindStringSubmatch(doc.Text())
	if len(planetMoonResources) != 2 {
		return ogame.Auction{}, errors.New("failed to find planetResources")
	}
	if err := json.Unmarshal([]byte(planetMoonResources[1]), &auction.Resources); err != nil {
		return ogame.Auction{}, errors.New("failed to json unmarshal planetResources: " + err.Error())
	}

	// Find already-bid
	m := regexp.MustCompile(`var playerBid\s?=\s?([^;]+);`).FindStringSubmatch(doc.Text())
	if len(m) != 2 {
		return ogame.Auction{}, errors.New("failed to get playerBid")
	}
	var alreadyBid int64
	if m[1] != "false" {
		alreadyBid = utils.DoParseI64(m[1])
	}
	auction.AlreadyBid = alreadyBid

	// Find min-bid
	auction.MinimumBid = utils.ParseInt(doc.Find("table.table_ressources_sum tr td.auctionInfo.js_price").Text())

	// Find deficit-bid
	auction.DeficitBid = utils.ParseInt(doc.Find("table.table_ressources_sum tr td.auctionInfo.js_deficit").Text())

	// Note: Don't just bid the min-bid amount. It will keep doubling the total bid and grow exponentially...
	// DeficitBid is 1000 when another player has outbid you or if nobody has bid yet.
	// DeficitBid seems to be filled by Javascript in the browser. We're parsing it anyway. Correct Bid calculation would be:
	// bid = max(auction.DeficitBid, auction.MinimumBid - auction.AlreadyBid)

	return auction, nil
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

// func extractConstructions(pageHTML []byte, clock clockwork.Clock) (buildingID ogame.ID, buildingCountdown int64,
// 	// OGame Version as of 11.13.0
// 	researchID ogame.ID, researchCountdown int64,
// 	lfBuildingID ogame.ID, lfBuildingCountdown int64,
// 	lfResearchID ogame.ID, lfResearchCountdown int64) {

// 	//new CountdownTimer('buildingCountdown', 4567, 'http://10.156.176.26:7001/game/index.php?page=ingame&component=overview', null, true, 3)
// 	matches := regexp.MustCompile(`new CountdownTimer\('buildingCountdown', (\d+),`).FindSubmatch(pageHTML)
// 	log.Println("==============")
// 	log.Println(len(matches))

// 	log.Println("==============")

// 	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))

// 	buildingCountdownMatch, ok := doc.Find(".buildingCountdown").Attr("data-end")
// 	if ok {
// 		buildingCountdown = utils.ParseInt(buildingCountdownMatch) - clock.Now().Unix()
// 		buildingIDInt := utils.ToInt(regexp.MustCompile(`onclick="cancelbuilding\((\d+),`).FindSubmatch(pageHTML)[1])
// 		buildingID = ogame.ID(buildingIDInt)
// 	}
// 	researchCountdownMatch, ok := doc.Find(".researchCountdown").Attr("data-end")
// 	if ok {
// 		researchCountdown = utils.ParseInt(researchCountdownMatch) - clock.Now().Unix()
// 		researchIDInt := utils.ToInt(regexp.MustCompile(`onclick="cancelresearch\((\d+),`).FindSubmatch(pageHTML)[1])
// 		researchID = ogame.ID(researchIDInt)
// 	}

// 	lfBuildingCountdownMatch, ok := doc.Find(".lfbuildingCountdown").Attr("data-end")
// 	if ok {
// 		lfBuildingCountdown = utils.ParseInt(lfBuildingCountdownMatch) - clock.Now().Unix()
// 		lfBuildingIDInt := utils.ToInt(regexp.MustCompile(`onclick="cancellfbuilding\((\d+),`).FindSubmatch(pageHTML)[1])
// 		lfBuildingID = ogame.ID(lfBuildingIDInt)
// 	}

// 	lfResearchCountdownMatch, ok := doc.Find(".lfresearchCountdown").Attr("data-end")
// 	if ok {
// 		lfResearchCountdown = utils.ParseInt(lfResearchCountdownMatch) - clock.Now().Unix()
// 		lfResearchIDInt := utils.ToInt(regexp.MustCompile(`onclick="cancellfbuilding\((\d+),`).FindSubmatch(pageHTML)[1])
// 		lfResearchID = ogame.ID(lfResearchIDInt)
// 	}
// 	return
// }

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
