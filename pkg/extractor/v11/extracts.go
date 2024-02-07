package v11

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

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

func extractEmpire(pageHTML []byte) ([]ogame.EmpireCelestial, error) {
	var out []ogame.EmpireCelestial
	raw, err := v6.ExtractEmpireJSON(pageHTML)
	if err != nil {
		return nil, err
	}
	j, ok := raw.(map[string]any)
	if !ok {
		return nil, errors.New("failed to parse json")
	}
	planetsRaw, ok := j["planets"].([]any)
	if !ok {
		return nil, errors.New("failed to parse json")
	}
	for _, planetRaw := range planetsRaw {
		planet, ok := planetRaw.(map[string]any)
		if !ok {
			return nil, errors.New("failed to parse json")
		}

		var tempMin, tempMax int64
		temperatureStr := utils.DoCastStr(planet["temperature"])
		m := v6.TemperatureRgx.FindStringSubmatch(temperatureStr)
		if len(m) == 3 {
			tempMin = utils.DoParseI64(m[1])
			tempMax = utils.DoParseI64(m[2])
		}
		mm := v6.DiameterRgx.FindStringSubmatch(utils.DoCastStr(planet["diameter"]))
		energyStr := utils.DoCastStr(planet["energy"])
		energyDoc, _ := goquery.NewDocumentFromReader(strings.NewReader(energyStr))
		energy := utils.ParseInt(energyDoc.Find("div span").Text())

		energyProductionTitel := energyDoc.Find("div").AttrOr("title", "")
		energyProductionDoc, _ := goquery.NewDocumentFromReader(strings.NewReader(energyProductionTitel))
		energyProduction := utils.ParseInt(energyProductionDoc.Find("span").Text())

		resourcesDetails := ogame.ResourcesDetails{}
		resourcesDetails.Energy.CurrentProduction = energyProduction
		if energy < 0 {
			resourcesDetails.Energy.Consumption = energyProduction + energy
		} else {
			resourcesDetails.Energy.Consumption = energyProduction - energy
		}

		if reflect.TypeOf(planet["production"].(map[string]any)["hourly"]) == reflect.TypeOf([]any{}) {
			resourcesDetails.Metal.CurrentProduction = int64(utils.DoCastF64(planet["production"].(map[string]any)["hourly"].([]any)[0]))
			resourcesDetails.Crystal.CurrentProduction = int64(utils.DoCastF64(planet["production"].(map[string]any)["hourly"].([]any)[1]))
			resourcesDetails.Deuterium.CurrentProduction = int64(utils.DoCastF64(planet["production"].(map[string]any)["hourly"].([]any)[2]))
		} else {
			resourcesDetails.Metal.CurrentProduction = int64(utils.DoCastF64(planet["production"].(map[string]any)["hourly"].(map[string]any)["0"]))
			resourcesDetails.Crystal.CurrentProduction = int64(utils.DoCastF64(planet["production"].(map[string]any)["hourly"].(map[string]any)["1"]))
			resourcesDetails.Deuterium.CurrentProduction = int64(utils.DoCastF64(planet["production"].(map[string]any)["hourly"].(map[string]any)["2"]))
			resourcesDetails.Population.GrowthRate = utils.DoCastF64(planet["production"].(map[string]any)["hourly"].(map[string]any)["5"])
			resourcesDetails.Food.Overproduction = int64(utils.DoCastF64(planet["production"].(map[string]any)["hourly"].(map[string]any)["6"]))
		}

		resourcesDetails.Metal.Available = int64(utils.DoCastF64(planet["metal"]))
		resourcesDetails.Crystal.Available = int64(utils.DoCastF64(planet["crystal"]))
		resourcesDetails.Deuterium.Available = int64(utils.DoCastF64(planet["deuterium"]))
		resourcesDetails.Food.Available = int64(utils.DoCastF64(planet["food"]))
		resourcesDetails.Deuterium.Available = int64(utils.DoCastF64(planet["population"]))
		//resourcesDetails.Energy.Available = int64(utils.DoCastF64(planet["deuterium"]))

		resourcesDetails.Metal.StorageCapacity = int64(utils.DoCastF64(planet["metalStorage"]))
		resourcesDetails.Crystal.StorageCapacity = int64(utils.DoCastF64(planet["crystalStorage"]))
		resourcesDetails.Deuterium.StorageCapacity = int64(utils.DoCastF64(planet["deuteriumStorage"]))
		resourcesDetails.Population.LivingSpace = int64(utils.DoCastF64(planet["deuteriumStorage"]))
		resourcesDetails.Food.StorageCapacity = int64(utils.DoCastF64(planet["deuteriumStorage"]))

		celestialType := ogame.CelestialType(utils.DoCastF64(planet["type"]))

		out = append(out, ogame.EmpireCelestial{
			Name:     utils.DoCastStr(planet["name"]),
			ID:       ogame.CelestialID(utils.DoCastF64(planet["id"])),
			Diameter: utils.ParseInt(mm[1]),
			Img:      utils.DoCastStr(planet["image"]),
			Type:     celestialType,
			Fields: ogame.Fields{
				Built: utils.DoParseI64(utils.DoCastStr(planet["fieldUsed"])),
				Total: utils.DoParseI64(utils.DoCastStr(planet["fieldMax"])),
			},
			Temperature: ogame.Temperature{
				Min: tempMin,
				Max: tempMax,
			},
			Coordinate: ogame.Coordinate{
				Galaxy:   int64(utils.DoCastF64(planet["galaxy"])),
				System:   int64(utils.DoCastF64(planet["system"])),
				Position: int64(utils.DoCastF64(planet["position"])),
				Type:     celestialType,
			},
			Resources: ogame.Resources{
				Metal:     int64(utils.DoCastF64(planet["metal"])),
				Crystal:   int64(utils.DoCastF64(planet["crystal"])),
				Deuterium: int64(utils.DoCastF64(planet["deuterium"])),
				Energy:    energy,
			},
			ResourcesDetails: resourcesDetails,
			Supplies: ogame.ResourcesBuildings{
				MetalMine:            int64(utils.DoCastF64(planet["1"])),
				CrystalMine:          int64(utils.DoCastF64(planet["2"])),
				DeuteriumSynthesizer: int64(utils.DoCastF64(planet["3"])),
				SolarPlant:           int64(utils.DoCastF64(planet["4"])),
				FusionReactor:        int64(utils.DoCastF64(planet["12"])),
				SolarSatellite:       int64(utils.DoCastF64(planet["212"])),
				MetalStorage:         int64(utils.DoCastF64(planet["22"])),
				CrystalStorage:       int64(utils.DoCastF64(planet["23"])),
				DeuteriumTank:        int64(utils.DoCastF64(planet["24"])),
			},
			Facilities: ogame.Facilities{
				RoboticsFactory: int64(utils.DoCastF64(planet["14"])),
				Shipyard:        int64(utils.DoCastF64(planet["21"])),
				ResearchLab:     int64(utils.DoCastF64(planet["31"])),
				AllianceDepot:   int64(utils.DoCastF64(planet["34"])),
				MissileSilo:     int64(utils.DoCastF64(planet["44"])),
				NaniteFactory:   int64(utils.DoCastF64(planet["15"])),
				Terraformer:     int64(utils.DoCastF64(planet["33"])),
				SpaceDock:       int64(utils.DoCastF64(planet["36"])),
				LunarBase:       int64(utils.DoCastF64(planet["41"])),
				SensorPhalanx:   int64(utils.DoCastF64(planet["42"])),
				JumpGate:        int64(utils.DoCastF64(planet["43"])),
			},
			Defenses: ogame.DefensesInfos{
				RocketLauncher:         int64(utils.DoCastF64(planet["401"])),
				LightLaser:             int64(utils.DoCastF64(planet["402"])),
				HeavyLaser:             int64(utils.DoCastF64(planet["403"])),
				GaussCannon:            int64(utils.DoCastF64(planet["404"])),
				IonCannon:              int64(utils.DoCastF64(planet["405"])),
				PlasmaTurret:           int64(utils.DoCastF64(planet["406"])),
				SmallShieldDome:        int64(utils.DoCastF64(planet["407"])),
				LargeShieldDome:        int64(utils.DoCastF64(planet["408"])),
				AntiBallisticMissiles:  int64(utils.DoCastF64(planet["502"])),
				InterplanetaryMissiles: int64(utils.DoCastF64(planet["503"])),
			},
			Researches: ogame.Researches{
				EnergyTechnology:             int64(utils.DoCastF64(planet["113"])),
				LaserTechnology:              int64(utils.DoCastF64(planet["120"])),
				IonTechnology:                int64(utils.DoCastF64(planet["121"])),
				HyperspaceTechnology:         int64(utils.DoCastF64(planet["114"])),
				PlasmaTechnology:             int64(utils.DoCastF64(planet["122"])),
				CombustionDrive:              int64(utils.DoCastF64(planet["115"])),
				ImpulseDrive:                 int64(utils.DoCastF64(planet["117"])),
				HyperspaceDrive:              int64(utils.DoCastF64(planet["118"])),
				EspionageTechnology:          int64(utils.DoCastF64(planet["106"])),
				ComputerTechnology:           int64(utils.DoCastF64(planet["108"])),
				Astrophysics:                 int64(utils.DoCastF64(planet["124"])),
				IntergalacticResearchNetwork: int64(utils.DoCastF64(planet["123"])),
				GravitonTechnology:           int64(utils.DoCastF64(planet["199"])),
				WeaponsTechnology:            int64(utils.DoCastF64(planet["109"])),
				ShieldingTechnology:          int64(utils.DoCastF64(planet["110"])),
				ArmourTechnology:             int64(utils.DoCastF64(planet["111"])),
			},
			Ships: ogame.ShipsInfos{
				LightFighter:   int64(utils.DoCastF64(planet["204"])),
				HeavyFighter:   int64(utils.DoCastF64(planet["205"])),
				Cruiser:        int64(utils.DoCastF64(planet["206"])),
				Battleship:     int64(utils.DoCastF64(planet["207"])),
				Battlecruiser:  int64(utils.DoCastF64(planet["215"])),
				Bomber:         int64(utils.DoCastF64(planet["211"])),
				Destroyer:      int64(utils.DoCastF64(planet["213"])),
				Deathstar:      int64(utils.DoCastF64(planet["214"])),
				SmallCargo:     int64(utils.DoCastF64(planet["202"])),
				LargeCargo:     int64(utils.DoCastF64(planet["203"])),
				ColonyShip:     int64(utils.DoCastF64(planet["208"])),
				Recycler:       int64(utils.DoCastF64(planet["209"])),
				EspionageProbe: int64(utils.DoCastF64(planet["210"])),
				SolarSatellite: int64(utils.DoCastF64(planet["212"])),
				Crawler:        int64(utils.DoCastF64(planet["217"])),
				Reaper:         int64(utils.DoCastF64(planet["218"])),
				Pathfinder:     int64(utils.DoCastF64(planet["219"])),
			},
		})
	}
	return out, nil
}
func extractCancelBuildingInfos(pageHTML []byte) (token string, techID, listID int64, err error) {
	return ExtractCancelInfos(pageHTML, "cancelbuilding", 0)
}

func extractCancelResearchInfos(pageHTML []byte) (token string, techID, listID int64, err error) {
	return ExtractCancelInfos(pageHTML, "cancelresearch", 2)
}

func extractCancelLfBuildingInfos(pageHTML []byte) (token string, id, listID int64, err error) {
	return ExtractCancelInfos(pageHTML, "cancellfbuilding", 1)
}

func ExtractCancelInfos(pageHTML []byte, fnName string, tableIdx int) (token string, id, listID int64, err error) {
	r1 := regexp.MustCompile(`window\.token = '([^']+)'`)
	m1 := r1.FindSubmatch(pageHTML)
	if len(m1) < 2 {
		return "", 0, 0, errors.New("unable to find token")
	}
	token = string(m1[1])
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	t := doc.Find("table.construction").Eq(tableIdx)
	a, _ := t.Find("a").First().Attr("onclick")
	r := regexp.MustCompile(fnName + `\((\d+),\s?(\d+),`)
	m := r.FindStringSubmatch(a)
	if len(m) < 3 {
		return "", 0, 0, errors.New("unable to find id/listid")
	}
	id = utils.DoParseI64(m[1])
	listID = utils.DoParseI64(m[2])
	return
}

func extractCharacterClassFromDoc(doc *goquery.Document) ogame.CharacterClass {
	characterClassDiv := doc.Find("div#characterclass a div")
	characterClass := ogame.NoClass
	if characterClassDiv.HasClass("miner") {
		characterClass = ogame.Collector
	} else if characterClassDiv.HasClass("warrior") {
		characterClass = ogame.General
	} else if characterClassDiv.HasClass("explorer") {
		characterClass = ogame.Discoverer
	}
	return characterClass
}

func extractLifeformTypeFromDoc(doc *goquery.Document) ogame.LifeformType {
	lfDiv := doc.Find("div#lifeform div.lifeform-item-icon")
	if lfDiv.HasClass("lifeform1") {
		return ogame.Humans
	} else if lfDiv.HasClass("lifeform2") {
		return ogame.Rocktal
	} else if lfDiv.HasClass("lifeform3") {
		return ogame.Mechas
	} else if lfDiv.HasClass("lifeform4") {
		return ogame.Kaelesh
	}
	return ogame.NoneLfType
}

func extractJumpGate(pageHTML []byte) (ogame.ShipsInfos, string, []ogame.MoonID, int64) {
	m := regexp.MustCompile(`\$\("#cooldown"\), (\d+),`).FindSubmatch(pageHTML)
	ships := ogame.ShipsInfos{}
	var destinations []ogame.MoonID
	if len(m) > 0 {
		waitTime := int64(utils.ToInt(m[1]))
		return ships, "", destinations, waitTime
	}
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	for _, s := range ogame.Ships {
		ships.Set(s.GetID(), utils.ParseInt(doc.Find("input#ship_"+utils.FI64(s.GetID())).AttrOr("rel", "0")))
	}
	token := doc.Find("input[name=token]").AttrOr("value", "")

	doc.Find("select[name=targetSpaceObjectId] option").Each(func(i int, s *goquery.Selection) {
		moonID := utils.ParseInt(s.AttrOr("value", "0"))
		if moonID > 0 {
			destinations = append(destinations, ogame.MoonID(moonID))
		}
	})

	return ships, token, destinations, 0
}

func extractPreferencesFromDoc(doc *goquery.Document) ogame.Preferences {
	prefs := v6.ExtractPreferencesFromDoc(doc)
	prefs.Language = extractLanguageFromDoc(doc)
	return prefs
}

func extractLanguageFromDoc(doc *goquery.Document) string {
	return doc.Find("select[name=language] option[selected]").AttrOr("value", "en")
}

func extractBuffActivationFromDoc(doc *goquery.Document) (token string, items []ogame.Item, err error) {
	scriptTxt := doc.Find("script").Text()
	r := regexp.MustCompile(`token = "([^"]+)"`)
	m := r.FindStringSubmatch(scriptTxt)
	if len(m) != 2 {
		err = errors.New("failed to find activate token")
		return
	}
	token = m[1]
	r = regexp.MustCompile(`inventoryObj\.items_inventory = ([^*]+\});`)
	m = r.FindStringSubmatch(scriptTxt)
	if len(m) != 2 {
		err = errors.New("failed to find items inventory")
		return
	}
	var inventoryMap map[string]ogame.Item = map[string]ogame.Item{}
	if err = json.Unmarshal([]byte(m[1]), &inventoryMap); err != nil {
		fmt.Println(err)
		return
	}
	for _, item := range inventoryMap {
		items = append(items, item)
	}
	return
}
