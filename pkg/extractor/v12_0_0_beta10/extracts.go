package v12_0_0_beta10

import (
	"errors"
	"fmt"
	"log"
	"math"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alaingilbert/ogame/pkg/ogame"
	"github.com/alaingilbert/ogame/pkg/utils"
)

func extractServerTimeFromDoc(doc *goquery.Document) (time.Time, error) {
	txt := doc.Find("div.OGameClock").First().Text()
	serverTime, err := time.Parse("02.01.2006 15:04:05", txt)
	if err != nil {
		return time.Time{}, err
	}

	u1 := time.Now().UTC().Unix()
	u2 := serverTime.Unix()
	n := int(math.Round(float64(u2-u1)/900)) * 900 // u2-u1 should be close to 0, round to nearest 15min difference

	serverTime = serverTime.Add(time.Duration(-n) * time.Second).In(time.FixedZone("OGT", n))

	return serverTime, nil
}

func extractAllianceClassFromDoc(doc *goquery.Document) (ogame.AllianceClass, error) {
	allianceClassTd := doc.Find("td.alliance_class").First()
	if allianceClassTd.HasClass("warrior") { // TODO: untested
		return ogame.Warrior, nil
	} else if allianceClassTd.HasClass("trader") {
		return ogame.Trader, nil
	} else if allianceClassTd.HasClass("explorer") {
		return ogame.Researcher, nil
	}
	return ogame.NoAllianceClass, errors.New("alliance class not found")
}

func extractTechinfoFromDoc(doc *goquery.Document) error {

	var out []ogame.TechinfoMineProduction = []ogame.TechinfoMineProduction{}

	hrefTxt, exist := doc.Find("a.overlay").First().Attr("href")
	if !exist {
		return errors.New("TechnologyID could not be found")
	}

	var technologyID int64
	re := regexp.MustCompile(`technologyId=(\d+)`)
	match := re.FindSubmatch([]byte(hrefTxt))
	if match != nil {
		tmpid, err := utils.ParseI64(string(match[1]))
		if err != nil {
			return err
		}
		technologyID = tmpid
		fmt.Println("technologyId:", technologyID)
	} else {
		fmt.Println("No match found regexp")
		return errors.New("TechnologyID could not be found with regexp")
	}

	ogameid := ogame.ID(technologyID)

	if ogameid == ogame.MetalMineID || ogameid == ogame.CrystalMineID || ogameid == ogame.DeuteriumSynthesizerID {
		doc.Find("table.general_details tr").EachWithBreak(func(i int, s *goquery.Selection) bool {
			var item ogame.TechinfoMineProduction
			item.ID = ogameid

			level, exist := s.Find("td.level").Attr("data-value")
			if exist {
				item.Level = utils.DoParseI64(level)
			}

			prodDiff, exist := s.Find("td.production_difference").Attr("data-value")
			if exist {
				item.ProductionDifference = utils.DoParseI64(prodDiff)
			}

			prodDiffLevel, exist := s.Find("td.production_level_difference").Attr("data-value")
			if exist {
				item.ProductionLevelDifference = utils.DoParseI64(prodDiffLevel)
			}

			energyConsumption, exist := s.Find("td.energy_consumption").Attr("data-value")
			if exist {
				item.EnergyConsumption = utils.DoParseI64(energyConsumption)
			}

			energyConsumptionDifference, exist := s.Find("td.energy_consumption_difference").Attr("data-value")
			if exist {
				item.EnergyConsumptionDifference = utils.DoParseI64(energyConsumptionDifference)
			}

			protected, exist := s.Find("td.protected").Attr("data-value")
			if exist {
				item.Protected = utils.DoParseI64(protected)
			}

			out = append(out, item)
			log.Print(item)
			return true
		})
	}

	return nil
}
