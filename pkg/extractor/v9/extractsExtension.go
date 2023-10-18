package v9

import (
	"bytes"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alaingilbert/ogame/pkg/ogame"
)

// ExtractMessages ...
func (e *Extractor) ExtractMessages(pageHTML []byte) ([]ogame.Message, int64, error) {
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(pageHTML))
	return e.ExtractMessagesFromDoc(doc)
}

// ExtractMessagesFromDoc ...
func (e *Extractor) ExtractMessagesFromDoc(doc *goquery.Document) ([]ogame.Message, int64, error) {
	return extractMessagesFromDoc(doc, e.GetLocation())
}

func extractMessagesFromDoc(doc *goquery.Document, location *time.Location) ([]ogame.Message, int64, error) {
	msgs := make([]ogame.Message, 0)
	nbPage, _ := strconv.ParseInt(doc.Find("ul.pagination li").Last().AttrOr("data-page", "1"), 10, 64)
	doc.Find("li.msg").Each(func(i int, s *goquery.Selection) {
		if idStr, exists := s.Attr("data-msg-id"); exists {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				msg := ogame.Message{ID: id}
				msg.CreatedAt, _ = time.ParseInLocation("02.01.2006 15:04:05", s.Find(".msg_date").Text(), location)
				msg.Title = s.Find(".msg_title a").Text()
				msg.Content, _ = s.Find("span.msg_content").Html()
				msg.Content = strings.TrimSpace(msg.Content)
				msgs = append(msgs, msg)
			}
		}
	})
	return msgs, nbPage, nil
}

// ExtractUpgradeToken ...
func (e *Extractor) ExtractUpgradeToken(pageHTML []byte) (string, error) {
	return extractUpgradeToken(pageHTML)
}

func extractUpgradeToken(pageHTML []byte) (string, error) {
	rgx := regexp.MustCompile(`var token = "([^"]+)"`)
	m := rgx.FindSubmatch(pageHTML)
	if len(m) != 2 {
		return "", errors.New("unable to find form token")
	}
	return string(m[1]), nil
}
