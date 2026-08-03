package v13_0_0

import (
	"errors"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/alaingilbert/ogame/pkg/ogame"
)

// OGame v13 movement page: the recall link exposes data-recall-action and the
// csrf token is stored in the page script (var token = "...").
func extractCancelFleetTokenFromDoc(doc *goquery.Document, _ ogame.FleetID) (string, error) {
	rgx := regexp.MustCompile(`var token = "([^"]+)"`)
	for _, s := range doc.Find("script").EachIter() {
		if m := rgx.FindStringSubmatch(s.Text()); len(m) == 2 {
			return m[1], nil
		}
	}
	return "", errors.New("cancel fleet token not found")
}
