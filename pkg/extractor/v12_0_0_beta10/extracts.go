package v12_0_0_beta10

import (
	"errors"
	"math"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alaingilbert/ogame/pkg/ogame"
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
