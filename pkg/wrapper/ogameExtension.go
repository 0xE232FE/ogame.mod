package wrapper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alaingilbert/ogame/pkg/gameforge"
	"github.com/alaingilbert/ogame/pkg/ogame"
	"github.com/alaingilbert/ogame/pkg/taskRunner"
	"github.com/alaingilbert/ogame/pkg/utils"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	cookiejar "github.com/orirawlings/persistent-cookiejar"
	"golang.org/x/net/html"
)

func (b *OGame) GetUserAccounts() ([]gameforge.Account, error) {
	return gameforge.GetUserAccounts(b.device.GetClient(), b.ctx, b.lobby, b.GetBearerToken())
}

func (b *OGame) GetServers() ([]gameforge.Server, error) {
	return gameforge.GetServers(b.lobby, b.device.GetClient(), b.ctx)
}

func (b *OGame) GetPassword() string {
	return b.password
}

func (b *OGame) FindAccount(universe, lang string, playerID int64, accounts []gameforge.Account, servers []gameforge.Server) (gameforge.Account, gameforge.Server, error) {
	return findAccount(universe, lang, playerID, accounts, servers)
}

func (b *OGame) GetBearerToken() string {
	if b.bearerToken == "" {
		cookies := b.device.GetClient().Jar.(*cookiejar.Jar).AllCookies()
		for _, c := range cookies {
			if c.Name == gameforge.TokenCookieName {
				b.bearerToken = c.Value
				break
			}
		}
	}
	return b.bearerToken
}

func (b *OGame) SetQuiet(s bool) {
	b.quiet = s
}

// // Handlers.go
var lastActiveCelestialID ogame.CelestialID
var lastActiveCelestialIDMu sync.RWMutex

// HTMLCleaner ...
func HTMLCleaner(bot *OGame, method string, url1 string, params url.Values, payload url.Values, pageHTML []byte) []byte {
	extractor := bot.GetExtractor()
	tmpLastActiveCelestialID, err := extractor.ExtractPlanetID(pageHTML)
	if err != nil {

	} else {
		lastActiveCelestialIDMu.Lock()
		lastActiveCelestialID = tmpLastActiveCelestialID
		lastActiveCelestialIDMu.Unlock()
	}

	if (IsKnowFullPage(params) || IsEmpirePage(params) || len(params) == 0) && !IsAjaxPage(params) {
		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
		node, _ := html.Parse(strings.NewReader(`<style>.cookiebanner1 {display: none;}\n.cookiebanner2 {display: none;}\n.cookiebanner3 {display: none;}</style>`))
		doc.Find("head").AppendNodes(node)
		htmlString, _ := doc.Html()
		return []byte(htmlString)
	}
	/*
		if (params.Get("page") == "ingame" || params.Get("page") == "messages" || params.Get("page") == "messages" || params.Get("page") == "shop" || params.Get("page") == "premium" || params.Get("page") == "chat" || params.Get("page") == "resourceSettings" || params.Get("page") == "rewards" || params.Get("page") == "standalone" || params.Get("page") == "standalone") &&
			params.Get("ajax") == "" && params.Get("asJson") == "" {
			doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
			node, _ := html.Parse(strings.NewReader(`<style>.cookiebanner1 {display: none;}\n.cookiebanner2 {display: none;}\n.cookiebanner3 {display: none;}</style>`))
			doc.Find("head").AppendNodes(node)
			htmlString, _ := doc.Html()
			return []byte(htmlString)
		}
	*/

	if IsAjaxPage(params) {
		switch params.Get("component") {
		case "technologydetails":
			type techDetails struct {
				Target  string `json:"target"`
				Content struct {
					Technologydetails string `json:"technologydetails"`
				} `json:"content"`
				Files struct {
					Js  []string `json:"js"`
					Css []string `json:"css"`
				} `json:"files"`
				Page struct {
					StateObj interface{} `json:"stateObj"`
					Title    string      `json:"title"`
					Url      string      `json:"url"`
				} `json:"page"`
				ServerTime   int64  `json:"serverTime"`
				NewAjaxToken string `json:"newAjaxToken"`
			}

			var Data techDetails

			err := json.Unmarshal(pageHTML, &Data)
			if err != nil {
				log.Debug(err)
				break
			}

			id, _ := strconv.ParseInt(params.Get("technology"), 10, 64)
			obj := ogame.Objs.ByID(ogame.ID(id))

			lastActiveCelestialIDMu.RLock()
			res, _ := bot.getResourcesDetails(lastActiveCelestialID)
			lastActiveCelestialIDMu.RUnlock()

			if obj.GetID().IsShip() || obj.GetID().IsDefense() {
				s := strings.ReplaceAll(``+Data.Content.Technologydetails+``, "\\n", "")
				s = strings.ReplaceAll(``+s+``, "\\", "")

				node, _ := html.Parse(bytes.NewReader([]byte(s)))
				doc := goquery.NewDocumentFromNode(node)

				lfbonus, err := bot.GetCachedLfBonuses()

				max := res.Available().Div(obj.GetPrice(1, lfbonus))
				doc.Find("div.build_amount input").SetAttr("min", "0")
				doc.Find("div.build_amount input").SetAttr("max", strconv.FormatInt(max, 10))
				doc.Find("div.build_amount input").SetAttr("onfocus", `clearInput(this);"", "0"`)
				doc.Find("div.build_amount input").SetAttr("onkeyup", `checkIntInput(this, 1, `+strconv.FormatInt(max, 10)+`);event.stopPropagation();`)
				doc.Find("div.build_amount").AppendHtml("<button class=\"maximum\">[max. " + strconv.FormatInt(max, 10) + "]</button>")

				Data.Content.Technologydetails, err = doc.Html()
				if err != nil {
					log.Printf("Error occured %s", err.Error())
				}
				pageHTML, _ = json.Marshal(Data)
			}

			if obj.GetID().IsBuilding() {

			}
			break
		}
	}
	return pageHTML
}

var ninjaFleetToken string

// NinjaSendFleet (With Checks)...
func (b *OGame) ninjaSendFleet(celestialID ogame.CelestialID, ships []ogame.Quantifiable, speed ogame.Speed, where ogame.Coordinate,
	mission ogame.MissionID, resources ogame.Resources, holdingTime, unionID int64, ensure bool) (ogame.Fleet, error) {

	BeginTime := time.Now()
	origin, _ := b.getCachedCelestial(celestialID)
	originCoords := origin.GetCoordinate()
	// /game/index.php?page=ajax&component=fleetdispatch&ajax=1&asJson=1
	if ninjaFleetToken == "" {
		// GetToken
		nToken := url.Values{}
		nToken.Add("page", "ajax")
		nToken.Add("component", "fleetdispatch")
		nToken.Add("ajax", "1")
		nToken.Add("asJson", "1")
		tokenRsp := struct {
			NewAjaxToken string `json:"newAjaxToken"`
		}{}
		pageHTMLToken, err := b.getPageContent(nToken)
		if err != nil {
			return ogame.Fleet{}, err
		}
		err = json.Unmarshal(pageHTMLToken, &tokenRsp)
		if err != nil {
			return ogame.Fleet{}, err
		}
		ninjaFleetToken = tokenRsp.NewAjaxToken
	}

	payload := url.Values{}
	for _, s := range ships {
		if s.ID.IsFlyableShip() && s.Nbr > 0 {
			payload.Set("am"+strconv.FormatInt(int64(s.ID), 10), strconv.FormatInt(s.Nbr, 10))
		}
	}

	payload.Set("token", ninjaFleetToken)
	payload.Set("galaxy", strconv.FormatInt(where.Galaxy, 10))
	payload.Set("system", strconv.FormatInt(where.System, 10))
	payload.Set("position", strconv.FormatInt(where.Position, 10))
	if mission == ogame.RecycleDebrisField {
		where.Type = ogame.DebrisType // Send to debris field
	} else if mission == ogame.Colonize || mission == ogame.Expedition {
		where.Type = ogame.PlanetType
	}
	payload.Set("type", strconv.FormatInt(int64(where.Type), 10))
	payload.Set("union", "0")

	if unionID != 0 {
		found := false
		if !found {
			return ogame.Fleet{}, ogame.ErrUnionNotFound
		}
	}
	lfbonus, err := b.getCachedLfBonuses()
	if err != nil {
		return ogame.Fleet{}, err
	}
	multiplier := float64(b.GetServerData().CargoHyperspaceTechMultiplier) / 100.0
	cargo := ogame.ShipsInfos{}.FromQuantifiables(ships).Cargo(b.getCachedResearch(), lfbonus, b.characterClass, multiplier, b.server.Settings.EspionageProbeRaids == 1)
	newResources := ogame.Resources{}
	if resources.Total() > cargo {
		newResources.Deuterium = int64(math.Min(float64(resources.Deuterium), float64(cargo)))
		cargo -= newResources.Deuterium
		newResources.Crystal = int64(math.Min(float64(resources.Crystal), float64(cargo)))
		cargo -= newResources.Crystal
		newResources.Metal = int64(math.Min(float64(resources.Metal), float64(cargo)))
	} else {
		newResources = resources
	}

	newResources.Metal = utils.MaxInt(newResources.Metal, 0)
	newResources.Crystal = utils.MaxInt(newResources.Crystal, 0)
	newResources.Deuterium = utils.MaxInt(newResources.Deuterium, 0)

	// Page 3 : select coord, mission, speed
	if b.IsV8() {
		payload.Set("token", ninjaFleetToken)
	}
	payload.Set("speed", strconv.FormatInt(int64(speed), 10))
	payload.Set("crystal", strconv.FormatInt(newResources.Crystal, 10))
	payload.Set("deuterium", strconv.FormatInt(newResources.Deuterium, 10))
	payload.Set("metal", strconv.FormatInt(newResources.Metal, 10))
	payload.Set("mission", strconv.FormatInt(int64(mission), 10))
	payload.Set("prioMetal", "1")
	payload.Set("prioCrystal", "2")
	payload.Set("prioDeuterium", "3")
	payload.Set("retreatAfterDefenderRetreat", "0")
	if mission == ogame.ParkInThatAlly || mission == ogame.Expedition {
		if mission == ogame.Expedition { // ogame.Expedition 1 to 18
			holdingTime = utils.Clamp(holdingTime, 1, 18)
		} else if mission == ogame.ParkInThatAlly { // ogame.ParkInThatAlly 0, 1, 2, 4, 8, 16, 32
			holdingTime = utils.Clamp(holdingTime, 0, 32)
		}
		payload.Set("holdingtime", strconv.FormatInt(holdingTime, 10))
	}

	// Page 4 : send the fleet
	res, _ := b.postPageContent(url.Values{"page": {"ingame"}, "component": {"fleetdispatch"}, "action": {"sendFleet"}, "ajax": {"1"}, "asJson": {"1"}, "cp": {strconv.FormatInt(int64(celestialID), 10)}}, payload)
	// {"success":true,"message":"Your fleet has been successfully sent.","redirectUrl":"https:\/\/s801-en.ogame.gameforge.com\/game\/index.php?page=ingame&component=fleetdispatch","components":[]}
	// Insufficient resources. (4060)
	// {"success":false,"errors":[{"message":"Not enough cargo space!","error":4029}],"fleetSendingToken":"b4786751c6d5e64e56d8eb94807fbf88","components":[]}
	// {"success":false,"errors":[{"message":"Fleet launch failure: The fleet could not be launched. Please try again later.","error":4047}],"fleetSendingToken":"1507c7228b206b4a298dec1d34a5a207","components":[]} // bad token I think
	// {"success":false,"errors":[{"message":"Recyclers must be sent to recycle this debris field!","error":4013}],"fleetSendingToken":"b826ff8c3d4e04066c28d10399b32ab8","components":[]}
	// {"success":false,"errors":[{"message":"Error, no ships available","error":4059}],"fleetSendingToken":"b369e37ce34bb64e3a59fa26bd8d5602","components":[]}
	// {"success":false,"errors":[{"message":"You have to select a valid target.","error":4049}],"fleetSendingToken":"19218f446d0985dfd79e03c3ec008514","components":[]} // colonize debris field
	// {"success":false,"errors":[{"message":"Planet is already inhabited!","error":4053}],"fleetSendingToken":"3281f9ad5b4cba6c0c26a24d3577bd4c","components":[]}
	// {"success":false,"errors":[{"message":"Colony ships must be sent to colonise this planet!","error":4038}],"fleetSendingToken":"8700c275a055c59ca276a7f66c81b205","components":[]}
	// fetch("https://s801-en.ogame.gameforge.com/game/index.php?page=ingame&component=fleetdispatch&action=sendFleet&ajax=1&asJson=1", {"credentials":"include","headers":{"content-type":"application/x-www-form-urlencoded; charset=UTF-8","sec-fetch-mode":"cors","sec-fetch-site":"same-origin","x-requested-with":"XMLHttpRequest"},"body":"token=414847e59344881d5c71303023735ab8&am209=1&am202=10&galaxy=9&system=297&position=7&type=2&metal=0&crystal=0&deuterium=0&prioMetal=1&prioCrystal=2&prioDeuterium=3&mission=8&speed=1&retreatAfterDefenderRetreat=0&union=0&holdingtime=0","method":"POST","mode":"cors"}).then(res => res.json()).then(r => console.log(r));
	StartTime := time.Now()
	b.debug("Send ogame.Fleet: " + strconv.FormatInt(time.Now().Sub(BeginTime).Milliseconds(), 10) + " ms")
	var resStruct struct {
		Success           bool          `json:"success"`
		Message           string        `json:"message"`
		FleetSendingToken string        `json:"fleetSendingToken"`
		Components        []interface{} `json:"components"`
		RedirectURL       string        `json:"redirectUrl"`
		Errors            []struct {
			Message string `json:"message"`
			Error   int64  `json:"error"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(res, &resStruct); err != nil {
		return ogame.Fleet{}, errors.New("failed to unmarshal response: " + err.Error())
	}
	ninjaFleetToken = resStruct.FleetSendingToken

	if len(resStruct.Errors) > 0 {
		return ogame.Fleet{}, errors.New(resStruct.Errors[0].Message + " (" + strconv.FormatInt(resStruct.Errors[0].Error, 10) + ")")
	}
	allianceClass, _ := b.getCachedAllianceClass()
	origin, _ = b.getCachedCelestial(celestialID)
	secs, _ := CalcFlightTime2(
		origin.GetCoordinate(), where,
		b.serverData.Galaxies, b.serverData.Systems, b.serverData.DonutGalaxy, b.serverData.DonutSystem, b.serverData.GlobalDeuteriumSaveFactor,
		float64(speed)/10, GetFleetSpeedForMission(b.serverData, mission), ogame.ShipsInfos{}.FromQuantifiables(ships), b.getCachedResearch(), lfbonus, b.characterClass, allianceClass, holdingTime)

	if resStruct.Success == true {
		return ogame.Fleet{
			Mission:      mission,
			ReturnFlight: false,
			InDeepSpace:  false,
			ID:           0,
			Resources:    resources,
			Origin:       originCoords,
			Destination:  where,
			Ships:        ogame.ShipsInfos{}.FromQuantifiables(ships),
			StartTime:    StartTime,
			ArrivalTime:  StartTime.Add(time.Duration(secs) * time.Second),
			ArriveIn:     int64(StartTime.Add(time.Duration(secs) * time.Second).Sub(StartTime).Seconds()),
			BackIn:       int64(StartTime.Add(time.Duration(secs)*time.Second).Sub(StartTime).Seconds()) * 2,
		}, nil
	}
	now := time.Now().Unix()
	b.error(errors.New("could not find new fleet ID").Error()+", planetID:", celestialID, ", ts: ", now)
	return ogame.Fleet{}, errors.New("could not find new fleet ID")
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// NinjaSendFleet (With Checks)...
func (b *OGame) ninjaSendFleetWithChecks(celestialID ogame.CelestialID, ships []ogame.Quantifiable, speed ogame.Speed, where ogame.Coordinate,
	mission ogame.MissionID, resources ogame.Resources, holdingTime, unionID int64, ensure bool) (ogame.Fleet, error) {

	b.debug("Begin NinjaSendFleet")
	b.debug(ships)

	BeginTime := time.Now()
	origin, _ := b.getCachedCelestialByID(celestialID)
	originCoords := origin.GetCoordinate()
	// /game/index.php?page=ajax&component=fleetdispatch&ajax=1&asJson=1
	// GetToken
	nToken := url.Values{}
	nToken.Add("page", "ajax")
	nToken.Add("component", "fleetdispatch")
	nToken.Add("ajax", "1")
	nToken.Add("asJson", "1")
	tokenRsp := struct {
		NewAjaxToken string `json:"newAjaxToken"`
	}{}
	pageHTMLToken, err := b.getPageContent(nToken)
	if err != nil {
		return ogame.Fleet{}, err
	}
	err = json.Unmarshal(pageHTMLToken, &tokenRsp)
	if err != nil {
		return ogame.Fleet{}, err
	}

	b.debug("Get Token: " + strconv.FormatInt(time.Now().Sub(BeginTime).Milliseconds(), 10) + " ms")

	_, _, availableShips, _, techs, _, _, err := b.getTechs(celestialID)
	if err != nil {
		return ogame.Fleet{}, err
	}
	b.debug("Get Techs: " + strconv.FormatInt(time.Now().Sub(BeginTime).Milliseconds(), 10) + " ms")

	// /game/index.php?page=json&component=eventList&ajax=1
	type Events struct {
		Time   int64 `json:"time"`
		Events []struct {
			EventID   int64 `json:"eventId"`
			Timestamp int64 `json:"timestamp"`
			Type      int64 `json:"type"`
			FleetId   int64 `json:"fleetId"`
			OwnFleet  bool  `json:"ownFleet"`
			MissionId int64 `json:"missionId"`

			UnionId             int64  `json:"UnionId"`
			IsUnionOwner        bool   `json:"isUnionOwner"`
			IsUnion             bool   `json:"isUnion"`
			IsUnionMember       bool   `json:"isUnionMember"`
			OriginId            int64  `json:"originId"`
			OriginPlayerId      int64  `json:"originPlayerId"`
			OriginPlayerDeleted bool   `json:"originPlayerDeleted"`
			OriginPlayerName    string `json:"originPlayerName"`
			OriginName          string `json:"originName"` // Colony Name
			OriginGalaxy        int64  `json:"originGalaxy"`
			OriginSystem        int64  `json:"originSystem"`
			OriginPosition      int64  `json:"originPosition"`
			OriginType          int64  `json:"originType"`
			OriginCoordinates   string `json:"originCoordinates"`
			// Target
			TargetId            int64  `json:"targetId"`
			TargetPlayerDeleted bool   `json:"targetPlayerDeleted"`
			TargetPlayerId      int64  `json:"targetPlayerId"`
			TargetName          string `json:"targetName"`
			TargetGalaxy        int64  `json:"targetGalaxy"`
			TargetSystem        int64  `json:"targetSystem"`
			TargetPosition      int64  `json:"targetPosition"`
			TargetType          int64  `json:"targetType"`
			TargetCoordinates   string `json:"targetCoordinates"`
			IsReturnFlight      bool   `json:"isReturnFlight"`
			MissionType         string `json:"missionType"` // friendly
			Ships               struct {
				Num202 struct {
					ID     int64 `json:"id"`
					Number int64 `json:"number"`
				} `json:"202"`
				Num203 struct {
					ID     int64 `json:"id"`
					Number int64 `json:"number"`
				} `json:"203"`
				Num204 struct {
					ID     int64 `json:"id"`
					Number int64 `json:"number"`
				} `json:"204"`
				Num205 struct {
					ID     int64 `json:"id"`
					Number int64 `json:"number"`
				} `json:"205"`
				Num206 struct {
					ID     int64 `json:"id"`
					Number int64 `json:"number"`
				} `json:"206"`
				Num207 struct {
					ID     int64 `json:"id"`
					Number int64 `json:"number"`
				} `json:"207"`
				Num208 struct {
					ID     int64 `json:"id"`
					Number int64 `json:"number"`
				} `json:"208"`
				Num209 struct {
					ID     int64 `json:"id"`
					Number int64 `json:"number"`
				} `json:"209"`
				Num210 struct {
					ID     int64 `json:"id"`
					Number int64 `json:"number"`
				} `json:"210"`
				Num211 struct {
					ID     int64 `json:"id"`
					Number int64 `json:"number"`
				} `json:"211"`
				Num212 struct {
					ID     int64 `json:"id"`
					Number int64 `json:"number"`
				} `json:"212"`
				Num213 struct {
					ID     int64 `json:"id"`
					Number int64 `json:"number"`
				} `json:"213"`
				Num214 struct {
					ID     int64 `json:"id"`
					Number int64 `json:"number"`
				} `json:"214"`
				Num215 struct {
					ID     int64 `json:"id"`
					Number int64 `json:"number"`
				} `json:"215"`
				Num217 struct {
					ID     int64 `json:"id"`
					Number int64 `json:"number"`
				} `json:"217"`
				Num218 struct {
					ID     int64 `json:"id"`
					Number int64 `json:"number"`
				} `json:"218"`
				Num219 struct {
					ID     int64 `json:"id"`
					Number int64 `json:"number"`
				} `json:"219"`
			} `json:"ships"`
			ShipCountUncensored int64 `json:"shipCountUncensored"`
			ShipCount           int64 `json:"shipCount"`
			Cargo               []struct {
				Name   string `json:"name"`
				Amount int64  `json:"amount"`
			} `json:"cargo"`
		} `json:"events"`
	}

	eventResp := Events{}
	eventVals := url.Values{}
	eventVals.Add("page", "json")
	eventVals.Add("component", "eventList")
	eventVals.Add("ajax", "1")
	pageHTMLEventList, err := b.getPageContent(eventVals)
	if err != nil {
		return ogame.Fleet{}, err
	}
	err = json.Unmarshal(pageHTMLEventList, &eventResp)
	if err != nil {
		return ogame.Fleet{}, err
	}
	b.debug("Get EventList 1: " + strconv.FormatInt(time.Now().Sub(BeginTime).Milliseconds(), 10) + " ms")

	maxInitialFleetID := ogame.FleetID(0)
	for _, f := range eventResp.Events {
		if ogame.FleetID(f.FleetId) > maxInitialFleetID {
			maxInitialFleetID = ogame.FleetID(f.FleetId)
		}
	}

	lfbonus, err := b.getCachedLfBonuses()
	multiplier := float64(b.GetServerData().CargoHyperspaceTechMultiplier) / 100.0
	fuelCapacity := ogame.ShipsInfos{}.FromQuantifiables(ships).Cargo(b.getCachedResearch(), lfbonus, b.characterClass, multiplier, b.server.Settings.EspionageProbeRaids == 1)

	allianceClass, _ := b.getCachedAllianceClass()

	orogin, _ := b.getCachedCelestial(celestialID)
	_, fuel := CalcFlightTime2(
		orogin.GetCoordinate(), where,
		b.serverData.Galaxies, b.serverData.Systems, b.serverData.DonutGalaxy, b.serverData.DonutSystem, b.serverData.GlobalDeuteriumSaveFactor,
		float64(speed)/10, GetFleetSpeedForMission(b.serverData, mission), ogame.ShipsInfos{}.FromQuantifiables(ships), techs, lfbonus, b.characterClass, allianceClass, holdingTime)
	if fuelCapacity < fuel {
		return ogame.Fleet{}, fmt.Errorf("not enough fuel capacity, available " + strconv.FormatInt(fuelCapacity, 10) + " but needed " + strconv.FormatInt(fuel, 10))
	}

	// Ensure we're not trying to attack/spy ourselves
	destinationIsMyOwnPlanet := false
	myCelestials := b.getCachedCelestials()
	for _, c := range myCelestials {
		if c.GetCoordinate().Equal(where) && c.GetID() == celestialID {
			return ogame.Fleet{}, errors.New("origin and destination are the same")
		}
		if c.GetCoordinate().Equal(where) {
			destinationIsMyOwnPlanet = true
			break
		}
	}
	if destinationIsMyOwnPlanet {
		switch mission {
		case ogame.Spy:
			return ogame.Fleet{}, errors.New("you cannot spy yourself")
		case ogame.Attack:
			return ogame.Fleet{}, errors.New("you cannot attack yourself")
		}
	}

	atLeastOneShipSelected := false
	if !ensure {
		for i := range ships {
			avail := availableShips.ByID(ships[i].ID)
			ships[i].Nbr = int64(math.Min(float64(ships[i].Nbr), float64(avail)))
			if ships[i].Nbr > 0 {
				atLeastOneShipSelected = true
			}
		}
	} else {
		if ships != nil {
			for _, ship := range ships {
				if ship.Nbr > availableShips.ByID(ship.ID) {
					return ogame.Fleet{}, fmt.Errorf("not enough ships to send, %s", ogame.Objs.ByID(ship.ID).GetName())
				}
				atLeastOneShipSelected = true
			}
		}
	}
	if !atLeastOneShipSelected {
		return ogame.Fleet{}, ogame.ErrNoShipSelected
	}

	payload := url.Values{}
	for _, s := range ships {
		if s.ID.IsFlyableShip() && s.Nbr > 0 {
			payload.Set("am"+strconv.FormatInt(int64(s.ID), 10), strconv.FormatInt(s.Nbr, 10))
		}
	}

	payload.Set("token", tokenRsp.NewAjaxToken)
	payload.Set("galaxy", strconv.FormatInt(where.Galaxy, 10))
	payload.Set("system", strconv.FormatInt(where.System, 10))
	payload.Set("position", strconv.FormatInt(where.Position, 10))
	if mission == ogame.RecycleDebrisField {
		where.Type = ogame.DebrisType // Send to debris field
	} else if mission == ogame.Colonize || mission == ogame.Expedition {
		where.Type = ogame.PlanetType
	}
	payload.Set("type", strconv.FormatInt(int64(where.Type), 10))
	payload.Set("union", "0")

	if unionID != 0 {
		found := false
		if !found {
			return ogame.Fleet{}, ogame.ErrUnionNotFound
		}
	}

	// Check
	by1, err := b.postPageContent(url.Values{"page": {"ingame"}, "component": {"fleetdispatch"}, "action": {"checkTarget"}, "ajax": {"1"}, "asJson": {"1"}}, payload)
	if err != nil {
		b.error(err.Error())
		return ogame.Fleet{}, err
	}

	b.debug("Get Check: " + strconv.FormatInt(time.Now().Sub(BeginTime).Milliseconds(), 10) + " ms")

	var checkRes CheckTargetResponse
	if err := json.Unmarshal(by1, &checkRes); err != nil {
		b.error(err.Error())
		return ogame.Fleet{}, err
	}

	if !checkRes.TargetOk {
		if len(checkRes.Errors) > 0 {
			return ogame.Fleet{}, errors.New(checkRes.Errors[0].Message + " (" + strconv.Itoa(checkRes.Errors[0].Error) + ")")
		}
		return ogame.Fleet{}, errors.New("target is not ok")
	}

	lfbonus, err = b.getCachedLfBonuses()
	multiplier = float64(b.GetServerData().CargoHyperspaceTechMultiplier) / 100.0
	cargo := ogame.ShipsInfos{}.FromQuantifiables(ships).Cargo(b.getCachedResearch(), lfbonus, b.characterClass, multiplier, b.server.Settings.EspionageProbeRaids == 1)
	newResources := ogame.Resources{}
	if resources.Total() > cargo {
		newResources.Deuterium = int64(math.Min(float64(resources.Deuterium), float64(cargo)))
		cargo -= newResources.Deuterium
		newResources.Crystal = int64(math.Min(float64(resources.Crystal), float64(cargo)))
		cargo -= newResources.Crystal
		newResources.Metal = int64(math.Min(float64(resources.Metal), float64(cargo)))
	} else {
		newResources = resources
	}

	newResources.Metal = utils.MaxInt(newResources.Metal, 0)
	newResources.Crystal = utils.MaxInt(newResources.Crystal, 0)
	newResources.Deuterium = utils.MaxInt(newResources.Deuterium, 0)

	// Page 3 : select coord, mission, speed
	if b.IsV8() {
		payload.Set("token", checkRes.NewAjaxToken)
	}
	payload.Set("speed", strconv.FormatInt(int64(speed), 10))
	payload.Set("crystal", strconv.FormatInt(newResources.Crystal, 10))
	payload.Set("deuterium", strconv.FormatInt(newResources.Deuterium, 10))
	payload.Set("metal", strconv.FormatInt(newResources.Metal, 10))
	payload.Set("mission", strconv.FormatInt(int64(mission), 10))
	payload.Set("prioMetal", "1")
	payload.Set("prioCrystal", "2")
	payload.Set("prioDeuterium", "3")
	payload.Set("retreatAfterDefenderRetreat", "0")
	if mission == ogame.ParkInThatAlly || mission == ogame.Expedition {
		if mission == ogame.Expedition { // ogame.Expedition 1 to 18
			holdingTime = utils.Clamp(holdingTime, 1, 18)
		} else if mission == ogame.ParkInThatAlly { // ogame.ParkInThatAlly 0, 1, 2, 4, 8, 16, 32
			holdingTime = utils.Clamp(holdingTime, 0, 32)
		}
		payload.Set("holdingtime", strconv.FormatInt(holdingTime, 10))
	}

	// Page 4 : send the fleet
	res, _ := b.postPageContent(url.Values{"page": {"ingame"}, "component": {"fleetdispatch"}, "action": {"sendFleet"}, "ajax": {"1"}, "asJson": {"1"}}, payload)
	// {"success":true,"message":"Your fleet has been successfully sent.","redirectUrl":"https:\/\/s801-en.ogame.gameforge.com\/game\/index.php?page=ingame&component=fleetdispatch","components":[]}
	// Insufficient resources. (4060)
	// {"success":false,"errors":[{"message":"Not enough cargo space!","error":4029}],"fleetSendingToken":"b4786751c6d5e64e56d8eb94807fbf88","components":[]}
	// {"success":false,"errors":[{"message":"Fleet launch failure: The fleet could not be launched. Please try again later.","error":4047}],"fleetSendingToken":"1507c7228b206b4a298dec1d34a5a207","components":[]} // bad token I think
	// {"success":false,"errors":[{"message":"Recyclers must be sent to recycle this debris field!","error":4013}],"fleetSendingToken":"b826ff8c3d4e04066c28d10399b32ab8","components":[]}
	// {"success":false,"errors":[{"message":"Error, no ships available","error":4059}],"fleetSendingToken":"b369e37ce34bb64e3a59fa26bd8d5602","components":[]}
	// {"success":false,"errors":[{"message":"You have to select a valid target.","error":4049}],"fleetSendingToken":"19218f446d0985dfd79e03c3ec008514","components":[]} // colonize debris field
	// {"success":false,"errors":[{"message":"Planet is already inhabited!","error":4053}],"fleetSendingToken":"3281f9ad5b4cba6c0c26a24d3577bd4c","components":[]}
	// {"success":false,"errors":[{"message":"Colony ships must be sent to colonise this planet!","error":4038}],"fleetSendingToken":"8700c275a055c59ca276a7f66c81b205","components":[]}
	// fetch("https://s801-en.ogame.gameforge.com/game/index.php?page=ingame&component=fleetdispatch&action=sendFleet&ajax=1&asJson=1", {"credentials":"include","headers":{"content-type":"application/x-www-form-urlencoded; charset=UTF-8","sec-fetch-mode":"cors","sec-fetch-site":"same-origin","x-requested-with":"XMLHttpRequest"},"body":"token=414847e59344881d5c71303023735ab8&am209=1&am202=10&galaxy=9&system=297&position=7&type=2&metal=0&crystal=0&deuterium=0&prioMetal=1&prioCrystal=2&prioDeuterium=3&mission=8&speed=1&retreatAfterDefenderRetreat=0&union=0&holdingtime=0","method":"POST","mode":"cors"}).then(res => res.json()).then(r => console.log(r));
	StartTime := time.Now()
	b.debug("Send ogame.Fleet: " + strconv.FormatInt(time.Now().Sub(BeginTime).Milliseconds(), 10) + " ms")
	var resStruct struct {
		Success           bool          `json:"success"`
		Message           string        `json:"message"`
		FleetSendingToken string        `json:"fleetSendingToken"`
		Components        []interface{} `json:"components"`
		RedirectURL       string        `json:"redirectUrl"`
		Errors            []struct {
			Message string `json:"message"`
			Error   int64  `json:"error"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(res, &resStruct); err != nil {
		return ogame.Fleet{}, errors.New("failed to unmarshal response: " + err.Error())
	}

	if len(resStruct.Errors) > 0 {
		return ogame.Fleet{}, errors.New(resStruct.Errors[0].Message + " (" + strconv.FormatInt(resStruct.Errors[0].Error, 10) + ")")
	}

	origin, _ = b.getCachedCelestial(celestialID)
	secs, _ := CalcFlightTime2(
		origin.GetCoordinate(), where,
		b.serverData.Galaxies, b.serverData.Systems, b.serverData.DonutGalaxy, b.serverData.DonutSystem, b.serverData.GlobalDeuteriumSaveFactor,
		float64(speed)/10, GetFleetSpeedForMission(b.serverData, mission), ogame.ShipsInfos{}.FromQuantifiables(ships), techs, lfbonus, b.characterClass, allianceClass, holdingTime)

	// Check latest fleetID
	pageHTMLEventList2, err := b.getPageContent(eventVals)
	if err != nil {
		return ogame.Fleet{}, err
	}
	eventResp2 := Events{}
	err = json.Unmarshal(pageHTMLEventList2, &eventResp2)
	if err != nil {
		return ogame.Fleet{}, err
	}
	max := ogame.Fleet{}
	if len(eventResp2.Events) > 0 {
		max := ogame.Fleet{}

		for i, fleet := range eventResp2.Events {
			origin := ogame.Coordinate{fleet.OriginGalaxy, fleet.OriginSystem, fleet.OriginPosition, ogame.CelestialType(fleet.OriginType)}
			destination := ogame.Coordinate{fleet.TargetGalaxy, fleet.TargetSystem, fleet.TargetPosition, ogame.CelestialType(fleet.TargetType)}

			if ogame.FleetID(fleet.FleetId) > max.ID &&
				origin.Equal(originCoords) &&
				destination.Equal(where) &&
				ogame.MissionID(fleet.MissionId) == mission &&
				!fleet.IsReturnFlight {
				max.ID = ogame.FleetID(eventResp2.Events[i].FleetId)
			}
		}
		if max.ID > maxInitialFleetID {
			return max, nil
		}
	}

	if resStruct.Success == true {
		return ogame.Fleet{
			Mission:      mission,
			ReturnFlight: false,
			InDeepSpace:  false,
			ID:           max.ID,
			Resources:    resources,
			Origin:       originCoords,
			Destination:  where,
			Ships:        ogame.ShipsInfos{}.FromQuantifiables(ships),
			StartTime:    StartTime,
			ArrivalTime:  StartTime.Add(time.Duration(secs) * time.Second),
			ArriveIn:     int64(StartTime.Add(time.Duration(secs) * time.Second).Sub(StartTime).Seconds()),
			BackIn:       int64(StartTime.Add(time.Duration(secs)*time.Second).Sub(StartTime).Seconds()) * 2,
		}, nil
	}
	now := time.Now().Unix()
	b.error(errors.New("could not find new fleet ID").Error()+", planetID:", celestialID, ", ts: ", now)
	return ogame.Fleet{}, errors.New("could not find new fleet ID")

}

// SendFleetHandler ...
// curl 127.0.0.1:1234/bot/planets/123/send-fleet -d 'ships=203,1&ships=204,10&speed=10&galaxy=1&system=1&type=1&position=1&mission=3&metal=1&crystal=2&deuterium=3'
// curl 10.156.176.2:8080/bot/planets/35699346/ninja-send-fleet -d 'ships=210,1&speed=10&galaxy=12&system=178&type=1&position=9&mission=3&metal=0&crystal=0&deuterium=0'
func NinjaSendFleetHandler(c echo.Context) error {
	bot := c.Get("bot").(*OGame)
	planetID, err := strconv.ParseInt(c.Param("planetID"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResp(400, "invalid planet id"))
	}

	if err := c.Request().ParseForm(); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResp(400, "invalid form"))
	}

	var ships []ogame.Quantifiable
	where := ogame.Coordinate{Type: ogame.PlanetType}
	mission := ogame.Transport
	var duration int64
	var unionID int64
	payload := ogame.Resources{}
	speed := ogame.HundredPercent
	for key, values := range c.Request().PostForm {
		switch key {
		case "ships":
			for _, s := range values {
				a := strings.Split(s, ",")
				shipID, err := strconv.ParseInt(a[0], 10, 64)
				if err != nil || !ogame.ID(shipID).IsShip() {
					return c.JSON(http.StatusBadRequest, ErrorResp(400, "invalid ship id "+a[0]))
				}
				nbr, err := strconv.ParseInt(a[1], 10, 64)
				if err != nil || nbr < 0 {
					return c.JSON(http.StatusBadRequest, ErrorResp(400, "invalid nbr "+a[1]))
				}
				ships = append(ships, ogame.Quantifiable{ID: ogame.ID(shipID), Nbr: nbr})
			}
		case "speed":
			speedInt, err := strconv.ParseInt(values[0], 10, 64)
			if err != nil || speedInt < 0 || speedInt > 10 {
				return c.JSON(http.StatusBadRequest, ErrorResp(400, "invalid speed"))
			}
			speed = ogame.Speed(speedInt)
		case "galaxy":
			galaxy, err := strconv.ParseInt(values[0], 10, 64)
			if err != nil {
				return c.JSON(http.StatusBadRequest, ErrorResp(400, "invalid galaxy"))
			}
			where.Galaxy = galaxy
		case "system":
			system, err := strconv.ParseInt(values[0], 10, 64)
			if err != nil {
				return c.JSON(http.StatusBadRequest, ErrorResp(400, "invalid system"))
			}
			where.System = system
		case "position":
			position, err := strconv.ParseInt(values[0], 10, 64)
			if err != nil {
				return c.JSON(http.StatusBadRequest, ErrorResp(400, "invalid position"))
			}
			where.Position = position
		case "type":
			t, err := strconv.ParseInt(values[0], 10, 64)
			if err != nil {
				return c.JSON(http.StatusBadRequest, ErrorResp(400, "invalid type"))
			}
			where.Type = ogame.CelestialType(t)
		case "mission":
			missionInt, err := strconv.ParseInt(values[0], 10, 64)
			if err != nil {
				return c.JSON(http.StatusBadRequest, ErrorResp(400, "invalid mission"))
			}
			mission = ogame.MissionID(missionInt)
		case "duration":
			duration, err = strconv.ParseInt(values[0], 10, 64)
			if err != nil {
				return c.JSON(http.StatusBadRequest, ErrorResp(400, "invalid duration"))
			}
		case "union":
			unionID, err = strconv.ParseInt(values[0], 10, 64)
			if err != nil {
				return c.JSON(http.StatusBadRequest, ErrorResp(400, "invalid union id"))
			}
		case "metal":
			metal, err := strconv.ParseInt(values[0], 10, 64)
			if err != nil || metal < 0 {
				return c.JSON(http.StatusBadRequest, ErrorResp(400, "invalid metal"))
			}
			payload.Metal = metal
		case "crystal":
			crystal, err := strconv.ParseInt(values[0], 10, 64)
			if err != nil || crystal < 0 {
				return c.JSON(http.StatusBadRequest, ErrorResp(400, "invalid crystal"))
			}
			payload.Crystal = crystal
		case "deuterium":
			deuterium, err := strconv.ParseInt(values[0], 10, 64)
			if err != nil || deuterium < 0 {
				return c.JSON(http.StatusBadRequest, ErrorResp(400, "invalid deuterium"))
			}
			payload.Deuterium = deuterium
		}
	}

	fleet, err := bot.WithPriority(taskRunner.Critical).NinjaSendFleet(ogame.CelestialID(planetID), ships, speed, where, mission, payload, duration, unionID, false)
	if err != nil &&
		(err == ogame.ErrInvalidPlanetID ||
			err == ogame.ErrNoShipSelected ||
			err == ogame.ErrUninhabitedPlanet ||
			err == ogame.ErrNoDebrisField ||
			err == ogame.ErrPlayerInVacationMode ||
			err == ogame.ErrAdminOrGM ||
			err == ogame.ErrNoAstrophysics ||
			err == ogame.ErrNoobProtection ||
			err == ogame.ErrPlayerTooStrong ||
			err == ogame.ErrNoMoonAvailable ||
			err == ogame.ErrNoRecyclerAvailable ||
			err == ogame.ErrNoEventsRunning ||
			err == ogame.ErrPlanetAlreadyReservedForRelocation) {
		return c.JSON(http.StatusBadRequest, ErrorResp(400, err.Error()))
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResp(500, err.Error()))
	}
	return c.JSON(http.StatusOK, SuccessResp(fleet))
}

func (b *OGame) HasEngineer() bool {
	return b.hasEngineer
}

func (b *OGame) HasCommander() bool {
	return b.hasCommander
}

func (b *OGame) HasAdmiral() bool {
	return b.hasAdmiral
}

func (b *OGame) HasGeologist() bool {
	return b.hasGeologist
}

func (b *OGame) HasTechnocrat() bool {
	return b.hasTechnocrat
}

// never obfuscate this type
var _ = reflect.TypeOf(GiftCodePayload{})

type GiftCodePayload struct {
	GameAccountID int64      `json:"gameAccountId"`
	Server        GiftServer `json:"server"`
}

type GiftServer struct {
	Language string `json:"language"`
	Number   int64  `json:"number"`
} //`json:"server"`

func GetUserAccountsWithBearerToken(client *http.Client, lobby, token string) ([]gameforge.Account, error) {
	var userAccounts []gameforge.Account
	req, err := http.NewRequest("GET", "https://"+lobby+".ogame.gameforge.com/api/users/me/accounts", nil)
	if err != nil {
		return userAccounts, err
	}
	req.Header.Add("authorization", "Bearer "+token)
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	resp, err := client.Do(req)
	if err != nil {
		return userAccounts, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Print(err)
		}
	}()
	by, err := utils.ReadBody(resp)
	if err != nil {
		return userAccounts, err
	}
	if err := json.Unmarshal(by, &userAccounts); err != nil {
		if string(by) == `{"error":"not logged in"}` {
			return userAccounts, ogame.ErrNotLogged
		}
		return userAccounts, errors.New("failed to get user accounts : " + err.Error() + " : " + string(by))
	}
	return userAccounts, nil
}

func CreateGiftCodeWithBearerToken(lobby, bearerToken string, client *http.Client) string {
	var payload struct {
		Accounts []GiftCodePayload `json:"accounts"`
	}

	accounts, _ := GetUserAccountsWithBearerToken(client, lobby, bearerToken)
	for _, account := range accounts {
		payload.Accounts = append(payload.Accounts, GiftCodePayload{
			GameAccountID: account.ID,
			Server: GiftServer{
				Language: account.Server.Language,
				Number:   account.Server.Number,
			},
		})
	}
	jsonPayloadBytes, err := json.Marshal(&payload)
	if err != nil {
		return ""
	}
	//log.Print(string(jsonPayloadBytes))
	req, err := http.NewRequest("PUT", "https://"+lobby+".ogame.gameforge.com/api/users/me/accountTrading", strings.NewReader(string(jsonPayloadBytes)))
	if err != nil {
		return ""
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("authorization", "Bearer "+bearerToken)
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode == 409 {
		gfChallengeID := resp.Header.Get(gameforge.TokenCookieName) // c434aa65-a064-498f-9ca4-98054bab0db8;https://challenge.gameforge.com
		if gfChallengeID != "" {
			parts := strings.Split(gfChallengeID, ";")
			challengeID := parts[0]
			return "error" + challengeID
		}
	}

	by, err := utils.ReadBody(resp)
	if err != nil {
		return ""
	}
	var res struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(by, &res); err != nil {

	}
	return res.Token
}

func (b *OGame) CreateGiftCode() string {
	client := b.GetClient()

	// var payload struct {
	// 	Accounts []struct {
	// 		GameAccountID int64 `json:"gameAccountId"`
	// 		Server        struct {
	// 			Language string `json:"language"`
	// 			Number   int64  `json:"number"`
	// 		} `json:"server"`
	// 	} `json:"accounts"`
	// }

	var payload struct {
		Accounts []GiftCodePayload `json:"accounts"`
	}

	accounts, _ := b.GetUserAccounts()
	for _, account := range accounts {
		payload.Accounts = append(payload.Accounts, GiftCodePayload{
			GameAccountID: account.ID,
			Server: GiftServer{
				Language: account.Server.Language,
				Number:   account.Server.Number,
			},
		})

	}
	jsonPayloadBytes, err := json.Marshal(&payload)
	if err != nil {
		return ""
	}
	req, err := http.NewRequest("PUT", "https://"+b.lobby+".ogame.gameforge.com/api/users/me/accountTrading", strings.NewReader(string(jsonPayloadBytes)))
	if err != nil {
		return ""
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("authorization", "Bearer "+b.GetBearerToken())
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode == 409 {
		gfChallengeID := resp.Header.Get(gameforge.TokenCookieName) // c434aa65-a064-498f-9ca4-98054bab0db8;https://challenge.gameforge.com
		if gfChallengeID != "" {
			parts := strings.Split(gfChallengeID, ";")
			challengeID := parts[0]
			return "error" + challengeID
		}
	}

	by, err := utils.ReadBody(resp)
	if err != nil {
		return ""
	}
	var res struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(by, &res); err != nil {

	}
	return res.Token
}

func (b *OGame) CreateGiftCodeSingleAccount(accountID int64, number int64, lang string) string {
	// client := b.GetClient()

	// var payload struct {
	// 	Accounts []struct {
	// 		GameAccountID int64 `json:"gameAccountId"`
	// 		Server        struct {
	// 			Language string `json:"language"`
	// 			Number   int64  `json:"number"`
	// 		} `json:"server"`
	// 	} `json:"accounts"`
	// }
	// payload.Accounts = append(payload.Accounts, struct {
	// 	GameAccountID int64 `json:"gameAccountId"`
	// 	Server        struct {
	// 		Language string `json:"language"`
	// 		Number   int64  `json:"number"`
	// 	} `json:"server"`
	// }{
	// 	GameAccountID: accountID,
	// 	Server: struct {
	// 		Language string `json:"language"`
	// 		Number   int64  `json:"number"`
	// 	}{
	// 		Language: lang,
	// 		Number:   number,
	// 	},
	// })

	// jsonPayloadBytes, err := json.Marshal(&payload)
	// if err != nil {
	// 	return ""
	// }
	// req, err := http.NewRequest("PUT", "https://"+b.lobby+".ogame.gameforge.com/api/users/me/accountTrading", strings.NewReader(string(jsonPayloadBytes)))
	// if err != nil {
	// 	return ""
	// }

	// req.Header.Add("Content-Type", "application/json")
	// req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	// req.Header.Add("authorization", "Bearer "+b.GetBearerToken())
	// resp, err := client.Do(req)
	// if err != nil {
	// 	return ""
	// }
	// defer resp.Body.Close()
	// if resp.StatusCode == 409 {
	// 	gfChallengeID := resp.Header.Get(TokenCookieName) // c434aa65-a064-498f-9ca4-98054bab0db8;https://challenge.gameforge.com
	// 	if gfChallengeID != "" {
	// 		parts := strings.Split(gfChallengeID, ";")
	// 		challengeID := parts[0]
	// 		return "error" + challengeID
	// 	}
	// }

	// by, err := utils.ReadBody(resp)
	// if err != nil {
	// 	return ""
	// }
	// var res struct {
	// 	Token string `json:"token"`
	// }
	// if err := json.Unmarshal(by, &res); err != nil {

	// }
	// return res.Token

	return ""
}

func (b *OGame) SelectCharacterClass(c ogame.CharacterClass) error {
	//{"POST":{"scheme":"https","host":"s133-cz.ogame.gameforge.com","filename":"/game/index.php","query":{"page":"ingame","component":"characterclassselection","characterClassId":"3","action":"selectClass","ajax":"1","asJson":"1"},"remote":{"Address":"0.0.0.0:443"}}}
	class := strconv.FormatInt(int64(c), 10)
	vals := url.Values{
		"page":             {"ingame"},
		"component":        {"characterclassselection"},
		"characterClassId": {class},
		"action":           {"selectClass"},
		"ajax":             {"1"},
		"asJson":           {"1"},
	}

	payload := url.Values{}
	by, err := b.PostPageContent(vals, payload)
	if err != nil {
		return err
	}
	var result struct {
		Status string `json:"status"`
	}
	err = json.Unmarshal(by, &result)
	if err != nil {
		return err
	}
	if result.Status == "success" {
		return nil
	}
	return nil
}

// CalcCargo ...
func (b *OGame) CalcCargo(total int64) (sc, lc, rc, pf, ds int64) {
	lfbonus, _ := b.getCachedLfBonuses()
	multiplier := float64(b.GetServerData().CargoHyperspaceTechMultiplier) / 100.0

	lc = int64(math.Ceil(float64(total) / float64(ogame.LargeCargo.GetCargoCapacity(b.getCachedResearch(), lfbonus, b.characterClass, multiplier, b.server.Settings.EspionageProbeRaids == 1))))
	sc = int64(math.Ceil(float64(total) / float64(ogame.SmallCargo.GetCargoCapacity(b.getCachedResearch(), lfbonus, b.characterClass, multiplier, b.server.Settings.EspionageProbeRaids == 1))))
	rc = int64(math.Ceil(float64(total) / float64(ogame.Recycler.GetCargoCapacity(b.getCachedResearch(), lfbonus, b.characterClass, multiplier, b.server.Settings.EspionageProbeRaids == 1))))
	pf = int64(math.Ceil(float64(total) / float64(ogame.Pathfinder.GetCargoCapacity(b.getCachedResearch(), lfbonus, b.characterClass, multiplier, b.server.Settings.EspionageProbeRaids == 1))))
	ds = int64(math.Ceil(float64(total) / float64(ogame.Deathstar.GetCargoCapacity(b.getCachedResearch(), lfbonus, b.characterClass, multiplier, b.server.Settings.EspionageProbeRaids == 1))))
	return
}

// RedeemCode ...
func RedeemCodeWithBearerToken(lobby, bearerToken, token string, client *http.Client) error {
	var payload struct {
		Token string `json:"token"`
	}
	payload.Token = token
	jsonPayloadBytes, err := json.Marshal(&payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", "https://"+lobby+".ogame.gameforge.com/api/token", strings.NewReader(string(jsonPayloadBytes)))
	if err != nil {
		return err
	}
	req.Header.Add("authorization", "Bearer "+bearerToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// {"tokenType":"accountTrading"}
	type respStruct struct {
		TokenType string `json:"tokenType"`
	}
	var respParsed respStruct
	by, err := utils.ReadBody(resp)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusBadRequest {
		return errors.New("invalid request, token invalid ?")
	}
	if err := json.Unmarshal(by, &respParsed); err != nil {
		return errors.New(err.Error() + " : " + string(by))
	}
	if respParsed.TokenType != "accountTrading" {
		return errors.New("tokenType is not accountTrading")
	}
	return nil
}

var cancelFleetToken = ""

func (b *OGame) njaCancelFleet(fleetID ogame.FleetID) error {
	params := url.Values{
		"page":      {"ajax"},
		"component": {MovementPageName},
		"ajax":      {"1"},
	}
	pageHTML, err := b.getPageContent(params)

	token, err := b.extractor.ExtractCancelFleetToken(pageHTML, fleetID)
	if err != nil {
		return err
	}
	if pageHTML, err = b.getPageContent(url.Values{"page": {"ajax"}, "component": {"movement"}, "return": {fleetID.String()}, "token": {token}, "ajax": {"1"}}); err != nil {
		return err
	}
	fleets := b.extractor.ExtractFleets(pageHTML)
	token, err = b.extractor.ExtractCancelFleetToken(pageHTML, fleetID)
	if err == nil {
		cancelFleetToken = token
	}

	var ok bool
	for _, f := range fleets {
		if f.ID == fleetID && f.ReturnFlight {
			ok = true
			break
		}
	}
	if !ok {
		return errors.New("fleet cancel ogame.Error")
	}
	return nil
}

// GetMaxExpeditionPoints returns the max ogame.Expedition Points for ogame.Fleet and Resources finds.
func (b *OGame) GetMaxExpeditionPoints() (int64, int64) {
	var top1 int64
	h, err := b.Highscore(1, 1, 1)
	if err != nil {
		return 0, 0
	}
	for _, p := range h.Players {
		if p.Position == 1 {
			top1 = p.Score
			break
		}
	}
	//  less than 100.000
	if top1 < 100000 {
		return 1250, 2400
	}
	//  100.000–1.000.000
	if top1 >= 100000 && top1 < 1000000 {
		return 3000, 6000
	}
	//  1.000.000–5.000.000
	if top1 >= 1000000 && top1 < 5000000 {
		return 4500, 9000
	}
	// 5.000.000-25.000.000
	if top1 >= 5000000 && top1 < 25000000 {
		return 6000, 12000
	}
	// 25.000.000-50.000.000
	if top1 >= 25000000 && top1 < 50000000 {
		return 7500, 15000
	}
	// 50.000.000-75.000.000
	if top1 >= 50000000 && top1 < 75000000 {
		return 9000, 18000
	}
	// 75.000.000-100.000.000
	if top1 >= 75000000 && top1 < 100000000 {
		return 10500, 21000
	}
	// more than 100.000.000
	if top1 >= 100000000 {
		return 12500, 25000
	}
	return 0, 0
}

func (b *OGame) BuyItem(ref string, celestialID ogame.CelestialID) error {
	return b.WithPriority(taskRunner.Normal).BuyItem(ref, celestialID)
}

func (b *OGame) buyItem(ref string, celestialID ogame.CelestialID) error {
	params := url.Values{"page": {"shop"}, "ajax": {"1"}, "type": {ref}}
	if celestialID != 0 {
		params.Set("cp", strconv.FormatInt(int64(celestialID), 10))
	}
	darkmatter, err := b.fetchResources(celestialID)
	if err != nil {
		return err
	}
	pageHTML, err := b.getPageContent(params)
	if err != nil {
		return err
	}
	items, err := b.getItems(celestialID)
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.Ref == ref {
			if item.Costs > darkmatter.Darkmatter.Available {
				costs := strconv.FormatInt(item.Costs, 10)
				dm := strconv.FormatInt(darkmatter.Darkmatter.Available, 10)
				return errors.New("not enough Darkmatter " + costs + " needed " + dm + " available")
			}
		}
	}

	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	scriptTxt := doc.Find("script").Text()
	r := regexp.MustCompile(`var token="([^"]+)"`)
	m := r.FindStringSubmatch(scriptTxt)
	if len(m) != 2 {
		err := errors.New("failed to find buy token")
		return err
	}
	token := m[1]
	params = url.Values{"page": {"componentOnly"}, "component": {"itemactions"}, "action": {"buy"}, "itemUuid": {ref}, "asJson": {"1"}}
	payload := url.Values{
		"token": {token},
	}
	var res struct {
		Message  interface{} `json:"message"`
		Status   string      `json:"status"`
		NewToken string      `json:"newAjaxToken"`
	}
	by, err := b.postPageContent(params, payload)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(by, &res); err != nil {
		return err
	}
	if res.Status != "success" {
		if msg, ok := res.Message.(string); ok {
			return errors.New(msg)
		}
		return errors.New("unknown error")
	}
	return err
}

// func (b *OGame) SetPreferences() error {
// 	payload := url.Values{}
// 	payload.Add("page", "ingame")
// 	payload.Add("component", PreferencesPageName)

// 	p := b.BeginNamed("SetPreferences")
// 	defer p.Done()
// 	pageHTML, err := p.GetPageContent(payload) // Will update preferences cached values
// 	if err != nil {
// 		return err
// 	}
// 	var changeSettingsToken string
// 	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
// 	if doc.Find("form#prefs input").Eq(2).AttrOr("name", "") == "token" {
// 		changeSettingsToken = doc.Find("form#prefs input").Eq(2).AttrOr("value", "")
// 	}
// 	if changeSettingsToken == "" {
// 		return errors.New("Token not found")
// 	}
// 	//#prefs > input:nth-child(3)
// 	//    var changeSettingsToken = "aaa71ec0484386d40100ad6a93950aa1";
// 	// r := regexp.MustCompile(`var changeSettingsToken = "([^"]+)"`)
// 	// m := r.FindStringSubmatch(string(pageHTML))
// 	// if len(m) != 2 {
// 	// 	err := errors.New("failed to find buy token")
// 	// 	return err
// 	// }
// 	// changeSettingsToken := m[1]

// 	fmt.Println("changeSettingsToken::", changeSettingsToken)

// 	// POST https://s180-de.ogame.gameforge.com/game/index.php?page=ingame&component=preferences

// 	payloadData := url.Values{}
// 	payloadData.Add("mode", "save")
// 	payloadData.Add("selectedTab", "0")
// 	payloadData.Add("token", changeSettingsToken)
// 	//payloadData.Add("db_character", "")
// 	payloadData.Add("spio_anz", "1")
// 	payloadData.Add("spySystemAutomaticQuantity", "1")
// 	payloadData.Add("spySystemTargetPlanetTypes", "0")
// 	payloadData.Add("spySystemTargetPlayerTypes", "0")
// 	payloadData.Add("spySystemIgnoreSpiedInLastXMinutes", "0")
// 	payloadData.Add("activateAutofocus", "on")
// 	payloadData.Add("eventsShow", "2")
// 	payloadData.Add("settings_sort", "0")
// 	payloadData.Add("settings_order", "0")
// 	payloadData.Add("showDetailOverlay", "on")
// 	//payloadData.Add("animatedSliders", "off")
// 	//payloadData.Add("animatedOverview", "off")
// 	payloadData.Add("msgResultsPerPage", "50")
// 	payloadData.Add("auctioneerNotifications", "on")
// 	payloadData.Add("showActivityMinutes", "1")

// 	_, err = p.PostPageContent(payload, payloadData)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// CalcFlightTime ...
func CalcFlightTime2(origin, destination ogame.Coordinate, universeSize, nbSystems int64, donutGalaxy, donutSystem bool,
	fleetDeutSaveFactor, speed float64, universeSpeedFleet int64, ships ogame.ShipsInfos, techs ogame.Researches, lfBonuses ogame.LfBonuses, characterClass ogame.CharacterClass, allianceClass ogame.AllianceClass, holdingTime int64) (secs, fuel int64) {
	if !ships.HasFlyableShips() {
		return
	}
	v := findSlowestSpeed(ships, techs, lfBonuses, characterClass, allianceClass)
	d := float64(Distance(origin, destination, universeSize, nbSystems, donutGalaxy, donutSystem))

	secs = CalcFlightTimeWithBaseSpeed(origin, destination, universeSize, nbSystems, donutGalaxy, donutSystem, speed, v, universeSpeedFleet)
	//secs = int64(math.Round(((3500/s)*math.Sqrt(d*10/v) + 10) / a))
	fuel = calcFuel2(ships, int64(d), secs, float64(universeSpeedFleet), fleetDeutSaveFactor, techs, lfBonuses, characterClass, allianceClass, holdingTime)

	return
}

// CalcFlightTime calculates the flight time and the fuel consumption
func (b *OGame) CalcFlightTime2(origin, destination ogame.Coordinate, speed float64, ships ogame.ShipsInfos, missionID ogame.MissionID, holdingTime int64) (secs, fuel int64) {
	lfbonus, _ := b.getCachedLfBonuses()
	allianceClass, _ := b.getCachedAllianceClass()
	return CalcFlightTime2(origin, destination, b.serverData.Galaxies, b.serverData.Systems, b.serverData.DonutGalaxy,
		b.serverData.DonutSystem, b.serverData.GlobalDeuteriumSaveFactor, speed, GetFleetSpeedForMission(b.serverData, missionID), ships,
		b.GetCachedResearch(), lfbonus, b.characterClass, allianceClass, holdingTime)
}

func calcFuel2(ships ogame.ShipsInfos, dist, duration int64, universeSpeedFleet, fleetDeutSaveFactor float64, techs ogame.Researches, lfBonuses ogame.LfBonuses, characterClass ogame.CharacterClass, allianceClass ogame.AllianceClass, holdingTime int64) (fuel int64) {
	tmpFn := func(baseFuel, nbr, shipSpeed, holdingTime int64) float64 {
		tmpSpeed := (35000 / (float64(duration)*universeSpeedFleet - 10)) * math.Sqrt(float64(dist)*10/float64(shipSpeed))
		if holdingTime > 0 {
			return float64(baseFuel*nbr*dist)/35000*math.Pow(tmpSpeed/10+1, 2) + math.Floor(float64(baseFuel*nbr*holdingTime)/10)
		}
		return float64(baseFuel*nbr*dist) / 35000 * math.Pow(tmpSpeed/10+1, 2)
	}
	tmpFuel := 0.0
	for _, ship := range ogame.Ships {
		if ship.GetID() == ogame.SolarSatelliteID || ship.GetID() == ogame.CrawlerID {
			continue
		}
		nbr := ships.ByID(ship.GetID())
		if nbr > 0 {
			tmpFuel += tmpFn(ship.GetFuelConsumption(techs, lfBonuses, characterClass, fleetDeutSaveFactor), nbr, ship.GetSpeed(techs, lfBonuses, characterClass, allianceClass), holdingTime)
		}
	}
	fuel = int64(1 + math.Round(tmpFuel))
	return
}

// TradeScraper get enemy fleets attacking you
func (b *OGame) TradeScraper(ships ogame.ShipsInfos, opts ...Option) error {
	return b.WithPriority(taskRunner.Normal).TradeScraper(ships, opts...)
}

// NinjaSendFleet get enemy fleets attacking you
func (b *OGame) NinjaSendFleet(celestialID ogame.CelestialID, ships []ogame.Quantifiable, speed ogame.Speed, where ogame.Coordinate,
	mission ogame.MissionID, resources ogame.Resources, holdingTime, unionID int64, ensure bool) (ogame.Fleet, error) {
	return b.WithPriority(taskRunner.Critical).NinjaSendFleet(celestialID, ships, speed, where, mission, resources, holdingTime, unionID, ensure)
}

// NjaCancelFleet ...
func (b *OGame) NjaCancelFleet(fleetID ogame.FleetID) error {
	return b.WithPriority(taskRunner.Normal).NjaCancelFleet(fleetID)
}

func (b *OGame) tradeScraper(ships ogame.ShipsInfos, opts ...Option) error {
	//http://10.156.176.2:8080/game/index.php?page=ajax&component=traderscrap
	tradersOverview := url.Values{"page": {"ajax"}, "component": {"traderscrap"}}
	payloadTrader := url.Values{"show": {"scrap"}, "ajax": {"1"}}
	pageHTML, err := b.postPageContent(tradersOverview, payloadTrader)
	if err != nil {
		return err
	}
	getToken := func(pageHTML []byte) (string, error) {
		m := regexp.MustCompile(`var token = "([^"]+)"`).FindSubmatch(pageHTML)
		if len(m) != 2 {
			return "", errors.New("unable to find token")
		}
		return string(m[1]), nil
	}
	token, _ := getToken(pageHTML)

	var cfg Options
	for _, opt := range opts {
		opt(&cfg)
	}
	var cp string
	if cfg.ChangePlanet != 0 {
		cp = strconv.FormatInt(int64(cfg.ChangePlanet), 10)
	}
	payload := url.Values{}
	for _, s := range ogame.Ships {
		if ships.ByID(s.GetID()) > 0 {
			if payload.Get("lastTechId") == "" {
				payload.Add("lastTechId", strconv.FormatInt(s.GetID().Int64(), 10))
			}
			id := strconv.FormatInt(s.GetID().Int64(), 10)
			nbr := strconv.FormatInt(ships.ByID(s.GetID()), 10)
			payload.Add("trade["+id+"]", nbr)

		}
	}
	queryString := url.Values{}
	queryString.Add("page", "ajax")
	queryString.Add("component", "traderscrap")
	queryString.Add("ajax", "1")
	queryString.Add("asJson", "1")
	queryString.Add("action", "trade")
	if token != "" {
		queryString.Add("token", token)
	}
	if cp != "" {
		queryString.Add("cp", cp)
	}
	resp, err := b.postPageContent(queryString, payload)
	if err != nil {
		return err
	}
	data := struct {
		Data struct {
			Error     bool   `json:"error"`
			Message   string `json:"message"`
			Resources struct {
				Metal     int64 `json:"metal"`
				Crystal   int64 `json:"crystal"`
				Deuterium int64 `json:"deuterium"`
			} `json:"resources"`
			TechAmount struct {
				Num1   int64 `json:"1"`
				Num2   int64 `json:"2"`
				Num3   int64 `json:"3"`
				Num4   int64 `json:"4"`
				Num12  int64 `json:"12"`
				Num14  int64 `json:"14"`
				Num15  int64 `json:"15"`
				Num21  int64 `json:"21"`
				Num22  int64 `json:"22"`
				Num23  int64 `json:"23"`
				Num24  int64 `json:"24"`
				Num31  int64 `json:"31"`
				Num33  int64 `json:"33"`
				Num34  int64 `json:"34"`
				Num36  int64 `json:"36"`
				Num41  int64 `json:"41"`
				Num42  int64 `json:"42"`
				Num43  int64 `json:"43"`
				Num44  int64 `json:"44"`
				Num106 int64 `json:"106"`
				Num108 int64 `json:"108"`
				Num109 int64 `json:"109"`
				Num110 int64 `json:"110"`
				Num111 int64 `json:"111"`
				Num113 int64 `json:"113"`
				Num114 int64 `json:"114"`
				Num115 int64 `json:"115"`
				Num117 int64 `json:"117"`
				Num118 int64 `json:"118"`
				Num120 int64 `json:"120"`
				Num121 int64 `json:"121"`
				Num122 int64 `json:"122"`
				Num123 int64 `json:"123"`
				Num124 int64 `json:"124"`
				Num199 int64 `json:"199"`
				Num202 int64 `json:"202"`
				Num203 int64 `json:"203"`
				Num204 int64 `json:"204"`
				Num205 int64 `json:"205"`
				Num206 int64 `json:"206"`
				Num207 int64 `json:"207"`
				Num208 int64 `json:"208"`
				Num209 int64 `json:"209"`
				Num210 int64 `json:"210"`
				Num211 int64 `json:"211"`
				Num212 int64 `json:"212"`
				Num213 int64 `json:"213"`
				Num214 int64 `json:"214"`
				Num215 int64 `json:"215"`
				Num217 int64 `json:"217"`
				Num218 int64 `json:"218"`
				Num219 int64 `json:"219"`
				Num401 int64 `json:"401"`
				Num402 int64 `json:"402"`
				Num403 int64 `json:"403"`
				Num404 int64 `json:"404"`
				Num405 int64 `json:"405"`
				Num406 int64 `json:"406"`
				Num407 int64 `json:"407"`
				Num408 int64 `json:"408"`
				Num502 int64 `json:"502"`
				Num503 int64 `json:"503"`
			} `json:"techAmount"`
			Percentage   int64  `json:"percentage"`
			BargainPrice int64  `json:"bargainPrice"`
			Quote        string `json:"quote"`
			NewToken     string `json:"newToken"`
		} `json:"data"`
		NewAjaxToken string `json:"newAjaxToken"`
	}{}
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return err
	}
	if data.Data.Error {
		return errors.New(data.Data.Message)
	}

	//{"data":{"error":false,"message":null,"resources":{"metal":700,"crystal":700,"deuterium":0},"techAmount":{"202":1,"203":0,"204":0,"205":0,"206":0,"207":0,"208":0,"210":0,"211":0,"213":0,"215":0,"218":0},"percentage":35,"bargainPrice":2000,"quote":"Dann zeigen Sie mal, was Sie loswerden wollen.","newToken":"a187c24bb2f75765f38cbea3fb0be139"},"components":[],"newAjaxToken":"f3bbbfc178daada1bf1d29fe41cf23ed"}

	// POST URL
	// http://10.156.176.2:8080/game/index.php?page=ajax&component=traderscrap&ajax=1&asJson=1&action=trade&cp=35695585

	// Request Data finishTrade=1&lastTechId=210&trade%5B210%5D=10
	/*
		finishTrade	"1"
		trade[210]	"10"
		trade[202]	"1"
		trade[203]	"0"
		trade[204]	"0"
		trade[205]	"0"
		trade[206]	"0"
		trade[207]	"0"
		trade[208]	"0"
		trade[210]	"0"
		trade[211]	"0"
		trade[213]	"0"
		trade[215]	"0"
		trade[218]	"0"
	*/

	// Query String
	/*
		page=ajax
		component=traderscrap
		ajax=1
		asJson=1
		action=trade
		cp=35695585
	*/

	// Request Header
	// Host: 10.156.176.2:8080
	// User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:92.0) Gecko/20100101 Firefox/92.0
	// Accept: application/json, text/javascript, */*; q=0.01
	// Accept-Language: en-US,en;q=0.5
	// Accept-Encoding: gzip, deflate
	// Referer: http://10.156.176.2:8080/game/index.php?page=ingame&component=traderOverview
	// Content-Type: application/x-www-form-urlencoded; charset=UTF-8
	// X-Requested-With: XMLHttpRequest
	// Content-Length: 46
	// Origin: http://10.156.176.2:8080
	// DNT: 1
	// Connection: keep-alive
	// Cookie: maximizeId=null; tabBoxFleets=%7B%2287242368%22%3A%5B1%2C%221631031607%22%5D%2C%2287242377%22%3A%5B1%2C%221631031616%22%5D%2C%2287242384%22%3A%5B1%2C%221631031623%22%5D%2C%2287242392%22%3A%5B1%2C%221631031629%22%5D%2C%2287242398%22%3A%5B1%2C%221631031636%22%5D%7D
	// Pragma: no-cache
	// Cache-Control: no-cache
	return nil
}

// GetMessages get enemy fleets attacking you
func (b *OGame) GetMessages() ([]ogame.Message, error) {
	return b.WithPriority(taskRunner.Normal).GetMessages()
}

func (b *OGame) getMessages() ([]ogame.Message, error) {
	msgs := make([]ogame.Message, 0)
	/*
		?konomie
		tabid: 3 => ?konomie (economy)
		tabid: 4 => OGame
		tabid: 5 => Spielewelt (Game)
		tabid: 6 => Favoriten (Favorites)

		Communication
		tabid: 10 => Beitr?ge (Contributions)
		tabid: 11 => Shared EspionageReports
		tabid: 12 => Shared CombatReports
		tabid: 13 => Shared ExpeditionReports
		tabid: 14 => Information

		Fleet
		tabid: 20 => Espionage
		tabid: 21 => Combat Reports
		tabid: 22 => Expeditions
		tabid: 23 => Unions/Transport
		tabid: 24 => Other
	*/
	var tabids []int64
	// ?konomie
	tabids = append(tabids, 3)
	tabids = append(tabids, 4)
	tabids = append(tabids, 5)
	tabids = append(tabids, 6)
	// Communication
	tabids = append(tabids, 10)
	tabids = append(tabids, 11)
	tabids = append(tabids, 12)
	tabids = append(tabids, 13)
	tabids = append(tabids, 14)
	// Fleet
	tabids = append(tabids, 20)
	tabids = append(tabids, 21)
	tabids = append(tabids, 22)
	tabids = append(tabids, 23)
	tabids = append(tabids, 24)
	var page int64 = 1
	var nbPage int64 = 1
	for _, tabid := range tabids {
		page = 1
		nbPage = 1
		for page <= nbPage {
			pageHTML, _ := b.getPageMessages(page, ogame.MessagesTabID(tabid))
			newMessages, newNbPage, _ := b.extractor.ExtractMessages(pageHTML)
			msgs = append(msgs, newMessages...)
			nbPage = newNbPage
			page++
		}
	}
	return msgs, nil
}

func (b *OGame) ChangeUsername(newUsername string) {
	// POST
	// https://s182-de.ogame.gameforge.com/game/index.php?page=ingame&component=preferences

	//mode=save&selectedTab=0&token=0ca38d65c41dd963494865a3a2f6c619&db_character=Legend&db_character_password=6M4duzUcja2CU0Q4AZne&urlaubs_modus=on

	/*
		mode "save"
		selectedTab	"0"
		token	"0ca38d65c41dd963494865a3a2f6c619"
		db_character	"Legend"
		db_character_password	"6M4duzUcja2CU0Q4AZne"
		urlaubs_modus	"on"
	*/
	prio := b.BeginNamed("Change Username")
	defer prio.Done()
	pageHTML, _ := prio.GetPageContent(url.Values{"page": {"ingame"}, "component": {PreferencesPageName}})
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	token := doc.Find("form#prefs input").Eq(2).AttrOr("value", "")

	vals := url.Values{}
	vals.Add("mode", "save")
	vals.Add("selectedTab", "0")
	vals.Add("token", token)
	vals.Add("db_character", newUsername)
	vals.Add("db_character_password", b.password)
	//vals.Add("urlaubs_modus", "off")
	//resultJSON, _ := prio.PostPageContent(url.Values{"page": {"ingame"}, "component": {PreferencesPage}}, vals)
	//log.Println(string(resultJSON))
	prio.PostPageContent(url.Values{"page": {"ingame"}, "component": {PreferencesPageName}}, vals)
}

func (b *OGame) CreateAlliance(tag, name string) error {
	bot := b.BeginNamed("Create Alliance")
	defer bot.Done()

	params := url.Values{"page": {"ingame"}, "component": {"alliance"}}

	pageHTML, err := bot.GetPageContent(params)
	if err != nil {
		return err
	}
	//doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	rgx := regexp.MustCompile(`var token = "([^"]+)";`)
	m := rgx.FindSubmatch(pageHTML)
	if len(m) != 2 {
		return errors.New("unable to find form token")
	}
	token := string(m[1])

	//  var urlCreateAlliance = "http:\/\/127.0.0.1:8080\/195\/game\/index.php?page=ingame&component=alliance&tab=createNewAlliance&action=createAlliance&asJson=1";
	//  var urlSendApplication = "http:\/\/127.0.0.1:8080\/195\/game\/index.php?page=ingame&component=alliance&tab=handleApplication&action=createApplication&asJson=1";
	//	var urlCancelApplication = "http:\/\/127.0.0.1:8080\/195\/game\/index.php?page=ingame&component=alliance&tab=handleApplication&action=cancelApplication&asJson=1";
	params = url.Values{"page": {"ingame"}, "component": {"alliance"}, "tab": {"createNewAlliance"}, "action": {"createAlliance"}, "asJson": {"1"}}
	payload := url.Values{"createTag": {tag}, "createName": {name}, "token": {token}}
	//createTag=Nuclear&createName=Nuclear&token=043a9f44af42e0da6da9b7f4c72fdb75
	pageHTML, err = bot.PostPageContent(params, payload)
	log.Printf("Payload: %s", payload.Encode())

	type resultJson struct {
		Status string `json:"status"`
	}
	var result resultJson
	err = json.Unmarshal(pageHTML, &result)
	if err != nil {
		return err
	}
	if result.Status == "success" {
		return nil
	}
	return errors.New("Status " + result.Status)
}

func (b *OGame) SetAllianceClass(class ogame.AllianceClass) error {
	bot := b.BeginNamed("Set Alliance Class")
	defer bot.Done()
	//https://s801-en.ogame.gameforge.com/game/index.php?page=ingame&component=alliance&tab=classselection&action=fetchClasses&ajax=1&token=917c290ab03a9e989ce851a207d27dd3
	// Warriors
	// Traders
	// Researchers
	//ogame.Warrior
	//ogame.Trader
	//ogame.Researcher

	//https://s801-en.ogame.gameforge.com/game/index.php?page=ingame&component=allianceclassselection&action=fetchDataAboutCurrentAllianceClass&ajax=1&asJson=1
	/*
		Result:
		{"currentAllianceClass":"-","dateOfLastAllianceClassChange":"16.08.2022 13:52:15","status":"success","message":"OK","components":[],"allianceClasses":[{"id":1,"name":"Warriors","price":500000,"icon":"warrior","isActive":true,"boni":[{"name":"Faster Ships","icon":"allymembershipspeed","longDescription":"+10% speed for ships flying between alliance members","shortDescription":"+10%"},{"name":"More Combat Research Levels","icon":"allycombatresearch","longDescription":"+1 combat research levels","shortDescription":"1 additional combat research levels"},{"name":"More Espionage Research Levels","icon":"allyespionageresearch","longDescription":"+1 espionage research levels","shortDescription":"1 additional espionage research levels"},{"name":"Espionage System","icon":"usespysystem","longDescription":"The espionage system can be used to scan whole systems.","shortDescription":"Use espionage system"}],"titleText":"Alliance Class: Warriors|+10% speed for ships flying between alliance members<br>+1 combat research levels<br>+1 espionage research levels<br>The espionage system can be used to scan whole systems.","isSelected":true,"infoLink":"https:\/\/s801-en.ogame.gameforge.com\/game\/index.php?page=allianceclassinfo&ajax=1&allianceClassId=1","button":{"type":"freeselect","darkmatter":500000,"label":"Free Activation","link":"https:\/\/s801-en.ogame.gameforge.com\/game\/index.php?page=ingame&component=alliance&tab=classselection&allianceClassId=1&action=selectClass&ajax=1&asJson=1&type=freeselect","disabled":false,"changeTitle":""}},{"id":2,"name":"Traders","price":500000,"icon":"trader","isActive":false,"boni":[{"name":"Rapid Transporters","icon":"allyshipspeed","longDescription":"+10% speed for transporters","shortDescription":"+10%"},{"name":"Increased Production","icon":"allyresource","longDescription":"+5% mine production","shortDescription":"+5%"},{"name":"Increased Energy Production","icon":"allyenergy","longDescription":"+5% energy production","shortDescription":"+5%"},{"name":"Increased Planet Storage Capacity","icon":"allyresource","longDescription":"+10% planet storage capacity","shortDescription":"+10%"},{"name":"Increased Moon Storage Capacity","icon":"allyresource","longDescription":"+10% moon storage capacity","shortDescription":"+10%"}],"titleText":"Alliance Class: Traders|+10% speed for transporters<br>+5% mine production<br>+5% energy production<br>+10% planet storage capacity<br>+10% moon storage capacity","isSelected":false,"infoLink":"https:\/\/s801-en.ogame.gameforge.com\/game\/index.php?page=allianceclassinfo&ajax=1&allianceClassId=2","button":{"type":"freeselect","darkmatter":500000,"label":"Free Activation","link":"https:\/\/s801-en.ogame.gameforge.com\/game\/index.php?page=ingame&component=alliance&tab=classselection&allianceClassId=2&action=selectClass&ajax=1&asJson=1&type=freeselect","disabled":false,"changeTitle":""}},{"id":3,"name":"Researchers","price":500000,"icon":"explorer","isActive":false,"boni":[{"name":"Planet Size","icon":"allycolonization","longDescription":"+5% larger planets on colonisation","shortDescription":"+5%"},{"name":"Faster Expeditions","icon":"allyexpeditionspeed","longDescription":"+10% speed to expedition destination","shortDescription":"+10%"},{"name":"System Phalanx","icon":"usephalanxsystem","longDescription":"The system phalanx can be used to scan fleet movements in whole systems.","shortDescription":"Use system phalanx"}],"titleText":"Alliance Class: Researchers|+5% larger planets on colonisation<br>+10% speed to expedition destination<br>The system phalanx can be used to scan fleet movements in whole systems.","isSelected":false,"infoLink":"https:\/\/s801-en.ogame.gameforge.com\/game\/index.php?page=allianceclassinfo&ajax=1&allianceClassId=3","button":{"type":"freeselect","darkmatter":500000,"label":"Free Activation","link":"https:\/\/s801-en.ogame.gameforge.com\/game\/index.php?page=ingame&component=alliance&tab=classselection&allianceClassId=3&action=selectClass&ajax=1&asJson=1&type=freeselect","disabled":false,"changeTitle":""}}],"premiumLink":"https:\/\/s801-en.ogame.gameforge.com\/game\/index.php?page=premium&showDarkMatter=1","allianceClassId":1,"token":"d5e705f2d350435c1e3d17cca9f9dc0b","newAjaxToken":"d5e705f2d350435c1e3d17cca9f9dc0b"}
	*/
	params := url.Values{}
	pageHTML, err := bot.GetPageContent(params)
	if err != nil {
		return err
	}

	type jsonResult struct {
		NewAjaxToken string `json:"newAjaxToken"`
	}
	var result jsonResult
	err = json.Unmarshal(pageHTML, &result)
	if err != nil {
		return err
	}

	params = url.Values{
		"page":            {"ingame"},
		"component":       {"alliance"},
		"tab":             {"classselection"},
		"allianceClassId": {strconv.FormatInt(int64(class), 10)},
		"action":          {"selectClass"},
		"ajax":            {"1"},
		"asJson":          {"1"},
		"type":            {"freeselect"},
	}
	payload := url.Values{
		"token": {result.NewAjaxToken},
	}
	pageHTML, err = bot.PostPageContent(params, payload)
	if err != nil {
		return err
	}

	type result2Json struct {
		Success string `json:"success"`
	}
	var result2 result2Json
	err = json.Unmarshal(pageHTML, &result2)
	if err != nil {
		return err
	}

	if result2.Success == "success" {
		return nil
	}

	// POST
	//https://s801-en.ogame.gameforge.com/game/index.php?page=ingame&component=alliance&tab=classselection&allianceClassId=3&action=selectClass&ajax=1&asJson=1&type=freeselect
	// Payload: {token: ""abcd}

	/*
		Result:{"tab":"classselection","allianceClassEnabled":true,"tabs":{"overview":{"name":"Overview","cssClass":"navi overview","enabled":true,"url":"https:\/\/s801-en.ogame.gameforge.com\/game\/index.php?page=ingame&component=alliance&tab=overview&action=fetchOverview&ajax=1","active":false,"tab":"overview"},"management":{"name":"Management","cssClass":"navi management","enabled":true,"url":"https:\/\/s801-en.ogame.gameforge.com\/game\/index.php?page=ingame&component=alliance&tab=management&action=fetchManagement&ajax=1","active":false,"tab":"management"},"broadcast":{"name":"Communication","cssClass":"navi broadcast","enabled":true,"url":"https:\/\/s801-en.ogame.gameforge.com\/game\/index.php?page=ingame&component=alliance&tab=broadcast&action=fetchBroadcast&ajax=1","active":false,"tab":"broadcast"},"applications":{"name":"Applications","cssClass":"navi applications","enabled":true,"applicationCount":0,"url":"https:\/\/s801-en.ogame.gameforge.com\/game\/index.php?page=ingame&component=alliance&tab=applications&action=fetchApplications&ajax=1","active":false,"tab":"applications"},"classselection":{"name":"Alliance Classes","cssClass":"navi classselection","enabled":true,"url":"https:\/\/s801-en.ogame.gameforge.com\/game\/index.php?page=ingame&component=alliance&tab=classselection&action=fetchClasses&ajax=1","active":true,"tab":"classselection"}},"status":"failure","errors":[{"message":"A previously unknown error has occurred. Unfortunately your last action couldn`t be executed!","error":100001}],"components":[],"newAjaxToken":"f7c9ea655d2cffa7aae7abf8eea45e02"}
	*/
	return errors.New("Error Selecting Alliance Class Success: " + result2.Success)
}

// FlightTime calculate flight time and fuel needed
func (b *OGame) FlightTime2(origin, destination ogame.Coordinate, speed ogame.Speed, ships ogame.ShipsInfos, missionID ogame.MissionID, holdingTime int64) (secs, fuel int64) {
	return b.WithPriority(taskRunner.Normal).FlightTime2(origin, destination, speed, ships, missionID, holdingTime)
}

func (b *OGame) GetLobby() string {
	return b.lobby
}
