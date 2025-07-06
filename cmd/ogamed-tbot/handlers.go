package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/alaingilbert/ogame/pkg/ogame"
	"github.com/alaingilbert/ogame/pkg/utils"
	"github.com/alaingilbert/ogame/pkg/wrapper"
	"github.com/labstack/echo/v4"
)

func PostToggleManualModeHandler(c echo.Context) error {
	bot := c.Get("bot").(*wrapper.OGame)
	botMap := c.Get("botMap").(map[string]wrapper.Prioritizable)
	txTimerMap := c.Get("manualModeTimeout").(map[string]*time.Timer)
	txTimer := txTimerMap["timeout"]

	locked, state := bot.GetState()
	if locked && state == "manual mode" {
		txTimer.Stop()
		tx := botMap["tx"]
		tx.Done()
		return c.JSON(http.StatusOK, false)
	} else {
		txTimer = time.NewTimer(30 * time.Second)
		txTimerMap["timeout"] = txTimer
		tx := bot.BeginNamed("manual mode")
		botMap["tx"] = tx
		go ManualTimerTx(botMap, txTimer)
		return c.JSON(http.StatusOK, true)
	}
}

// GetAllianceClassHandler ...
func GetAllianceClassHandler(c echo.Context) error {
	bot := c.Get("bot").(*wrapper.OGame)
	allianceClass, err := bot.GetCachedAllianceClass()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, wrapper.ErrorResp(http.StatusInternalServerError, err.Error()))
	}
	return c.JSON(http.StatusOK, wrapper.SuccessResp(allianceClass))

}

// GetExpeditionMessagesHandler ...
func GetExpeditionMessagesHandler(c echo.Context) error {
	bot := c.Get("bot").(*wrapper.OGame)
	messages, err := bot.GetExpeditionMessages(-1)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, wrapper.ErrorResp(http.StatusInternalServerError, err.Error()))
	}
	return c.JSON(http.StatusOK, wrapper.SuccessResp(messages))

}

// GetFromGameHandler ...
func GetFromGameHandler(c echo.Context) error {
	bot := c.Get("bot").(*wrapper.OGame)
	botMap := c.Get("botMap").(map[string]wrapper.Prioritizable)
	txTimerMap := c.Get("manualModeTimeout").(map[string]*time.Timer)
	txTimer := txTimerMap["timeout"]

	vals := url.Values{"page": {"ingame"}, "component": {"overview"}}
	if len(c.QueryParams()) > 0 {
		vals = c.QueryParams()
	}
	locked, state := bot.GetState()
	var pageHTML []byte
	if locked && state == "manual mode" {
		tx := botMap["tx"]
		txTimer.Reset(30 * time.Second)
		pageHTML, _ = tx.GetPageContent(vals)
	} else {
		pageHTML, _ = bot.GetPageContent(vals)
	}
	//pageHTML = replaceHostname(bot, pageHTML)
	pageHTML = replaceHostnameWithRegxp(bot, pageHTML, c.Request())
	pageHTML = removeCookiesBanner(pageHTML)
	pageHTML = ingameUi(pageHTML, state, locked)
	//pageHTML = injectJS(pageHTML)

	return c.HTMLBlob(http.StatusOK, pageHTML)
}

// PostToGameHandler ...
func PostToGameHandler(c echo.Context) error {
	bot := c.Get("bot").(*wrapper.OGame)
	botMap := c.Get("botMap").(map[string]wrapper.Prioritizable)
	txTimerMap := c.Get("manualModeTimeout").(map[string]*time.Timer)
	txTimer := txTimerMap["timeout"]

	vals := url.Values{"page": {"ingame"}, "component": {"overview"}}
	if len(c.QueryParams()) > 0 {
		vals = c.QueryParams()
	}
	payload, _ := c.FormParams()
	locked, state := bot.GetState()
	var pageHTML []byte
	if state == "manual mode" && locked {
		txTimer.Reset(30 * time.Second)
		tx := botMap["tx"]
		pageHTML, _ = tx.PostPageContent(vals, payload)
	} else {
		pageHTML, _ = bot.PostPageContent(vals, payload)
	}
	//pageHTML = replaceHostname(bot, pageHTML)
	pageHTML = replaceHostnameWithRegxp(bot, pageHTML, c.Request())
	pageHTML = removeCookiesBanner(pageHTML)
	pageHTML = ingameUi(pageHTML, state, locked)
	//pageHTML = injectJS(pageHTML)

	return c.HTMLBlob(http.StatusOK, pageHTML)
}

// GetStaticHEADHandler ...
func GetStaticHEADHandler(c echo.Context) error {
	bot := c.Get("bot").(*wrapper.OGame)
	newURL := "/api/" + strings.Join(c.ParamValues(), "") // + "?" + c.QueryString()
	if len(c.QueryString()) > 0 {
		newURL = newURL + "?" + c.QueryString()
	}
	headers, err := bot.HeadersForPage(newURL)
	if err != nil {
		return c.JSON(http.StatusBadRequest, wrapper.ErrorResp(400, err.Error()))
	}
	if len(headers) < 1 {
		return c.NoContent(http.StatusFailedDependency)
	}
	// Copy the original HTTP HEAD headers to our client
	for k, vv := range headers { // duplicate headers are acceptable in HTTP spec, so add all of them individually: https://stackoverflow.com/questions/4371328/are-duplicate-http-response-headers-acceptable
		k = http.CanonicalHeaderKey(k)
		for _, v := range vv {
			c.Response().Header().Add(k, v)
		}
	}
	return c.NoContent(http.StatusOK)
}

// GetStaticHandler ...
func GetStaticHandler(c echo.Context) error {
	bot := c.Get("bot").(*wrapper.OGame)

	newURL := bot.ServerURL() + c.Request().URL.String()
	req, err := http.NewRequest(http.MethodGet, newURL, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, wrapper.ErrorResp(500, err.Error()))
	}
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	resp, err := bot.GetDevice().GetClient().Do(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, wrapper.ErrorResp(500, err.Error()))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, wrapper.ErrorResp(500, err.Error()))
	}

	// Copy the original HTTP headers to our client
	for k, vv := range resp.Header { // duplicate headers are acceptable in HTTP spec, so add all of them individually: https://stackoverflow.com/questions/4371328/are-duplicate-http-response-headers-acceptable
		k = http.CanonicalHeaderKey(k)
		if k != "Content-Length" && k != "Content-Encoding" { // https://github.com/alaingilbert/ogame/pull/80#issuecomment-674559853
			for _, v := range vv {
				c.Response().Header().Add(k, v)
			}
		}
	}

	if strings.Contains(c.Request().URL.String(), ".xml") {
		//body = replaceHostname(bot, body)
		body = replaceHostnameWithRegxp(bot, body, c.Request())
		return c.Blob(http.StatusOK, "application/xml", body)
	}

	contentType := http.DetectContentType(body)
	if strings.Contains(newURL, ".css") {
		contentType = "text/css"
	} else if strings.Contains(newURL, ".js") {
		contentType = "application/javascript"
	} else if strings.Contains(newURL, ".gif") {
		contentType = "image/gif"
	}

	return c.Blob(http.StatusOK, contentType, body)
}

func replaceHostnameWithRegxp(bot *wrapper.OGame, pageHTML []byte, r *http.Request) []byte {
	requestHostname := "http://" + r.Host
	if r.TLS != nil {
		requestHostname = "https://" + r.Host
	}
	regexes := []string{
		`https://s\d+-[a-z]{2}\.ogame\.gameforge\.com`,
		`https:\\/\\/s\d+-[a-z]{2}\.ogame\.gameforge\.com`,
		`https:\\\\\\\/\\\\\\\/s\d+-[a-z]{2}\.ogame\.gameforge\.com`,
	}
	for idx, regex := range regexes {
		replaceStr := requestHostname
		// if bot.apiNewHostname != "" {
		// 	replaceStr = bot.apiNewHostname
		// }

		if idx > 0 {
			replaceStr = strings.ReplaceAll(replaceStr, `/`, `\/`)
			if idx > 1 {
				replaceStr = strings.ReplaceAll(replaceStr, `/`, `\\\/`)
			}
		}

		pageHTML = regexp.MustCompile(regex).ReplaceAll(pageHTML, []byte(replaceStr))
	}
	return pageHTML
}

func removeCookiesBanner(pageHTML []byte) []byte {
	pageHTMLstr := string(pageHTML)
	pageHTMLstr = strings.ReplaceAll(pageHTMLstr, "https://consent.gameforge.com/cookiebanner.js", "")
	pageHTMLstr = strings.ReplaceAll(pageHTMLstr, "https://consent-sandbox.gameforge.com/cookiebanner.js", "")
	// regex := `<script[^>]*id="cookiebanner"[^>]*>[\s\S]*?</script>`
	// re := regexp.MustCompile(regex)
	// return re.ReplaceAll(pageHTML, []byte(""))
	return []byte(pageHTMLstr)
}

func ingameUi(pageHTML []byte, state string, locked bool) []byte {
	tmpHTML := string(pageHTML)

	toReplace := `<script>
	var btn1 = document.createElement("BUTTON");`
	if state == "manual mode" && locked {
		toReplace += `
        var t = document.createTextNode("Manual mode (on)");
		`
	} else {
		toReplace += `
        var t = document.createTextNode("Manual mode (off)");
		`
	}
	toReplace += `
        btn1.style.height = '20px';
        btn1.style.width = '130px';
        btn1.style.display = 'block';
        btn1.style.marginBottom = '3px';
        btn1.appendChild(t);
        btn1.onclick = function() {
            var formData = new FormData();
            $.ajax({
                url: "/bot/toggle-manual-mode", data: formData, type: 'POST', processData: false, contentType: false,
                success: function(res) {
                    $(btn1).text("Manual mode (" + (res ? "on" : "off") + ")");
                },
                error: function(req) { console.log(req.responseText); },
            });
        };

		var div = document.createElement("div");
        div.style.position = 'fixed';
        div.style.right = '10px';
        div.style.top = '10px';
        div.style.zIndex = '3001';
        div.appendChild(btn1);

		document.body.prepend(div);

		</script>`

	toReplace += `</body>`

	tmpHTML = strings.ReplaceAll(tmpHTML, `</body>`, toReplace)
	return []byte(tmpHTML)
}

// GetLfBonusesHandler ...
func GetLfBonusesTbotHandler(c echo.Context) error {
	bot := c.Get("bot").(*wrapper.OGame)
	planetID, err := utils.ParseI64(c.Param("planetID"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, wrapper.ErrorResp(400, "invalid planet id"))
	}
	opts := getOptions()
	opts.ChangePlanet = ogame.CelestialID(planetID)
	res, err := getLfBonuses(bot, opts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, wrapper.ErrorResp(500, err.Error()))
	}
	return c.JSON(http.StatusOK, wrapper.SuccessResp(res))
}

func getLfBonuses(b *wrapper.OGame, opts wrapper.Options) (out ogame.LfBonuses, err error) {
	page := wrapper.LfBonusesPageName
	vals := url.Values{"page": {"ingame"}, "component": {page}, "cp": {utils.DoCastStr(opts.ChangePlanet)}}
	if page == wrapper.FetchResourcesPageName || page == wrapper.FetchTechsName || page == wrapper.LogoutPageName {
		vals = url.Values{"page": {page}}
	}
	pageHtml, err := b.GetPageContent(vals)
	if err != nil {
		return
	}
	bonuses, err := ExtractLfBonuses(pageHtml)
	if err != nil {
		return
	}
	return bonuses, nil
}

func getOptions(opts ...wrapper.Option) (out wrapper.Options) {
	for _, opt := range opts {
		if opt != nil {
			opt(&out)
		}
	}
	return
}

func ManualTimerTx(botMap map[string]wrapper.Prioritizable, txTimer *time.Timer) {
	tx := botMap["tx"]
	select {
	case <-txTimer.C:
		tx.Done()
		log.Println("Manual Mode Expired!")
		// do something for timeout, like change state
	}
}

// GetLfBonusesHandler ...
func GetLfBonusesHandler(c echo.Context) error {
	bot := c.Get("bot").(*wrapper.OGame)
	lfBonuses, err := bot.GetLfBonuses()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, wrapper.ErrorResp(http.StatusInternalServerError, err.Error()))
	}
	return c.JSON(http.StatusOK, wrapper.SuccessResp(lfBonuses))
}

// GetPositionsAvailableForDiscoveryFleet...
// curl 127.0.0.1:1234/bot/planets/:planetID/get-system-available-discovery -d 'galaxy=6&system=473'
func GetPositionsAvailableForDiscoveryFleet(c echo.Context) error {
	bot := c.Get("bot").(*wrapper.OGame)
	planetID, err := utils.ParseI64(c.Param("planetID"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, wrapper.ErrorResp(400, "invalid planet id"))
	}
	var opts wrapper.Option
	if planetID != 0 {
		opts = wrapper.ChangePlanet(ogame.CelestialID(planetID))
	}

	if err := c.Request().ParseForm(); err != nil {
		return c.JSON(http.StatusBadRequest, wrapper.ErrorResp(400, "invalid form"))
	}

	where := ogame.Coordinate{Type: ogame.PlanetType}
	for key, values := range c.Request().PostForm {
		switch key {
		case "galaxy":
			galaxy, err := utils.ParseI64(values[0])
			if err != nil {
				return c.JSON(http.StatusBadRequest, wrapper.ErrorResp(400, "invalid galaxy"))
			}
			where.Galaxy = galaxy
		case "system":
			system, err := utils.ParseI64(values[0])
			if err != nil {
				return c.JSON(http.StatusBadRequest, wrapper.ErrorResp(400, "invalid system"))
			}
			where.System = system
		case "position":
			position, err := utils.ParseI64(values[0])
			if err != nil {
				return c.JSON(http.StatusBadRequest, wrapper.ErrorResp(400, "invalid position"))
			}
			where.Position = position
		}
	}
	var coords []ogame.Coordinate
	if coords, err = bot.GetPositionsAvailableForDiscoveryFleet(where.Galaxy, where.System, opts); err != nil {
		return c.JSON(http.StatusBadRequest, wrapper.ErrorResp(400, err.Error()))
	}
	return c.JSON(http.StatusOK, wrapper.SuccessResp(coords))
}

// GetAvailableDiscoveries...
// curl 127.0.0.1:1234/bot/planets/:planetID/get-available-discoveries
func GetAvailableDiscoveries(c echo.Context) error {
	bot := c.Get("bot").(*wrapper.OGame)
	bot.GetAvailableDiscoveries()

	planetID, err := utils.ParseI64(c.Param("planetID"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, wrapper.ErrorResp(400, "invalid planet id"))
	}
	var opts wrapper.Option
	if planetID != 0 {
		opts = wrapper.ChangePlanet(ogame.CelestialID(planetID))
	}

	if err := c.Request().ParseForm(); err != nil {
		return c.JSON(http.StatusBadRequest, wrapper.ErrorResp(400, "invalid form"))
	}

	where := ogame.Coordinate{Type: ogame.PlanetType}
	for key, values := range c.Request().PostForm {
		switch key {
		case "galaxy":
			galaxy, err := utils.ParseI64(values[0])
			if err != nil {
				return c.JSON(http.StatusBadRequest, wrapper.ErrorResp(400, "invalid galaxy"))
			}
			where.Galaxy = galaxy
		case "system":
			system, err := utils.ParseI64(values[0])
			if err != nil {
				return c.JSON(http.StatusBadRequest, wrapper.ErrorResp(400, "invalid system"))
			}
			where.System = system
		case "position":
			position, err := utils.ParseI64(values[0])
			if err != nil {
				return c.JSON(http.StatusBadRequest, wrapper.ErrorResp(400, "invalid position"))
			}
			where.Position = position
		}
	}
	var count int64
	if count, err = bot.GetAvailableDiscoveries(opts); err != nil {
		return c.JSON(http.StatusBadRequest, wrapper.ErrorResp(400, err.Error()))
	}
	return c.JSON(http.StatusOK, wrapper.SuccessResp(count))
}
