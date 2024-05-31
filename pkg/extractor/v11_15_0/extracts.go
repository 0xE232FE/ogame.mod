package v11_15_0

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	v6 "github.com/alaingilbert/ogame/pkg/extractor/v6"
	v71 "github.com/alaingilbert/ogame/pkg/extractor/v71"
	"github.com/alaingilbert/ogame/pkg/ogame"
	"github.com/alaingilbert/ogame/pkg/utils"
)

func ExtractCoord(v string) (coord ogame.Coordinate) {
	coordRgx := regexp.MustCompile(`\[(\d+):(\d+):(\d+)]`)
	m := coordRgx.FindStringSubmatch(v)
	if len(m) == 4 {
		coord.Galaxy = utils.DoParseI64(m[1])
		coord.System = utils.DoParseI64(m[2])
		coord.Position = utils.DoParseI64(m[3])
	}
	return
}

// OK, PLANET TYPE RETURNS ALWAYS PLANET
func extractEspionageReportMessageIDsFromDoc(doc *goquery.Document) ([]ogame.EspionageReportSummary, int64, error) {
	msgs := make([]ogame.EspionageReportSummary, 0)
	doc.Find(".msg").Each(func(i int, s *goquery.Selection) {
		if idStr, exists := s.Attr("data-msg-id"); exists {
			if id, err := utils.ParseI64(idStr); err == nil {
				messageType := ogame.Report
				if s.Find("span.espionageDefText").Size() > 0 {
					messageType = ogame.Action
				}
				report := ogame.EspionageReportSummary{ID: id, Type: messageType}
				report.From = s.Find(".msgSender").Text()
				spanLink := s.Find(".msgTitle a")
				targetStr := spanLink.Text()
				report.Target = ExtractCoord(targetStr)

				// TODO: implement moon type, the following extractor is not working 'cause the element always has the "planet" class
				report.Target.Type = ogame.PlanetType

				if messageType == ogame.Report {
					s.Find(".lootPercentage").Each(func(i int, s *goquery.Selection) {
						if regexp.MustCompile(`%`).MatchString(s.Text()) {
							report.LootPercentage, _ = strconv.ParseFloat(regexp.MustCompile(`: (\d+)%`).FindStringSubmatch(s.Text())[1], 64)
							report.LootPercentage /= 100
						}
					})
				}
				msgs = append(msgs, report)
			}
		}
	})
	return msgs, 0, nil
}

// OK, MISSING A COUPLE OF NOT VERY IMPORTANT THINGS
func extractEspionageReportFromDoc(doc *goquery.Document, location *time.Location) (ogame.EspionageReport, error) {
	report := ogame.EspionageReport{}
	report.ID = utils.DoParseI64(doc.Find("div.detail_msg").AttrOr("data-msg-id", "0"))
	spanLink := doc.Find("span.msg_title a").First()
	txt := spanLink.Text()
	report.Coordinate = ExtractCoord(txt)
	figure := spanLink.Find("figure").First()
	if figure.HasClass("planet") {
		report.Coordinate.Type = ogame.PlanetType
	} else if figure.HasClass("moon") {
		report.Coordinate.Type = ogame.MoonType
	}
	messageType := ogame.Report
	if doc.Find("span.espionageDefText").Size() > 0 {
		messageType = ogame.Action
	}
	report.Type = messageType
	msgDateRaw := doc.Find("span.msg_date").Text()
	msgDate, _ := time.ParseInLocation("02.01.2006 15:04:05", msgDateRaw, location)
	report.Date = msgDate.In(time.Local)

	username := doc.Find(".playerInfo .playerName").Find("span").First().Text()
	report.Username = strings.TrimSpace(username)
	characterClassStr := doc.Find(".playerInfo .characterClassInfo").First().Text()
	characterClassStr = strings.TrimSpace(characterClassStr)
	characterClassSplit := strings.Split(characterClassStr, ":")
	if len(characterClassSplit) > 1 {
		report.CharacterClass = v71.GetCharacterClass(strings.TrimSpace(characterClassSplit[1]))
	}

	report.AllianceClass = ogame.NoAllianceClass
	allianceClassStr := doc.Find(".playerInfo .allianceClassInfo").Text()
	allianceClassStr = strings.TrimSpace(allianceClassStr)
	allianceClassSplit := strings.Split(allianceClassStr, ":")
	if len(allianceClassSplit) > 0 {
		// TODO: implement an extractor for alliance class
		report.AllianceClass = ogame.NoAllianceClass
	}

	// Bandit, Starlord
	banditstarlord := doc.Find(".playerInfo .playerName").First().Find("span").First().Find("span").First()
	if banditstarlord.HasClass("honorRank") {
		report.IsBandit = banditstarlord.HasClass("rank_bandit1") || banditstarlord.HasClass("rank_bandit2") || banditstarlord.HasClass("rank_bandit3")
		report.IsStarlord = banditstarlord.HasClass("rank_starlord1") || banditstarlord.HasClass("rank_starlord2") || banditstarlord.HasClass("rank_starlord3")
	}

	honorableFound := doc.Find(".playerInfo .playerName").First().Find("span.status_abbr_honorableTarget")
	report.HonorableTarget = honorableFound.Length() > 0

	// IsInactive, IsLongInactive
	inactive := doc.Find(".playerInfo .playerName").First().Find("span")
	if inactive.HasClass("status_abbr_longinactive") {
		report.IsInactive = true
		report.IsLongInactive = true
	} else if inactive.HasClass("status_abbr_inactive") {
		report.IsInactive = true
	}

	// APIKey
	apikey, _ := doc.Find("button.icon_apikey").Attr("title")
	apiDoc, _ := goquery.NewDocumentFromReader(strings.NewReader(apikey))
	report.APIKey = apiDoc.Find("input").First().AttrOr("value", "")

	// Inactivity timer
	report.LastActivity = utils.DoParseI64(doc.Find(".detailsActivity span").Text())

	// CounterEspionage
	ceTxt := doc.Find(".miscInfo .counterInfo").Text()
	m1 := regexp.MustCompile(`(\d+)%`).FindStringSubmatch(ceTxt)
	if len(m1) == 2 {
		report.CounterEspionage = utils.DoParseI64(m1[1])
	}

	hasError := false
	report.HasFleetInformation = false
	report.HasDefensesInformation = false
	report.HasBuildingsInformation = false
	report.HasResearchesInformation = false

	res := doc.Find(".resourceLootInfo .loot-row")
	if res.Size() > 0 {
		report.Metal = utils.ParseInt(res.Find("resource-icon.metal .amount").First().AttrOr("title", "0"))
		report.Crystal = utils.ParseInt(res.Find("resource-icon.crystal .amount").First().AttrOr("title", "0"))
		report.Deuterium = utils.ParseInt(res.Find("resource-icon.deuterium .amount").First().AttrOr("title", "0"))
		report.Food = utils.ParseInt(res.Find("resource-icon.food .amount").First().AttrOr("title", "0"))
		report.Population = utils.ParseInt(res.Find("resource-icon.population .amount").First().AttrOr("title", "0"))
		report.Energy = utils.ParseInt(res.Find(".energy .amount").First().Text())
	}

	fleet := doc.Find(".fleetSection")
	if fleet.Size() > 0 {
		report.HasFleetInformation = true
		lightFighter := utils.ParseInt(fleet.Find(".loot-item technology-icon[fighterlight]").Siblings().First().Find(".loot-amount").Text())
		report.LightFighter = &lightFighter
		heavyFighter := utils.ParseInt(fleet.Find(".loot-item technology-icon[fighterheavy]").Siblings().First().Find(".loot-amount").Text())
		report.HeavyFighter = &heavyFighter
		cruiser := utils.ParseInt(fleet.Find(".loot-item technology-icon[cruiser]").Siblings().First().Find(".loot-amount").Text())
		report.Cruiser = &cruiser
		battleship := utils.ParseInt(fleet.Find(".loot-item technology-icon[battleship]").Siblings().First().Find(".loot-amount").Text())
		report.Battleship = &battleship
		battlecruiser := utils.ParseInt(fleet.Find(".loot-item technology-icon[interceptor]").Siblings().First().Find(".loot-amount").Text())
		report.Battlecruiser = &battlecruiser
		bomber := utils.ParseInt(fleet.Find(".loot-item technology-icon[bomber]").Siblings().First().Find(".loot-amount").Text())
		report.Bomber = &bomber
		destroyer := utils.ParseInt(fleet.Find(".loot-item technology-icon[destroyer]").Siblings().First().Find(".loot-amount").Text())
		report.Destroyer = &destroyer
		deathstar := utils.ParseInt(fleet.Find(".loot-item technology-icon[deathstar]").Siblings().First().Find(".loot-amount").Text())
		report.Deathstar = &deathstar
		reaper := utils.ParseInt(fleet.Find(".loot-item technology-icon[reaper]").Siblings().First().Find(".loot-amount").Text())
		report.Reaper = &reaper
		pathfinder := utils.ParseInt(fleet.Find(".loot-item technology-icon[explorer]").Siblings().First().Find(".loot-amount").Text())
		report.Pathfinder = &pathfinder
		smallCargo := utils.ParseInt(fleet.Find(".loot-item technology-icon[transportersmall]").Siblings().First().Find(".loot-amount").Text())
		report.SmallCargo = &smallCargo
		largeCargo := utils.ParseInt(fleet.Find(".loot-item technology-icon[transporterlarge]").Siblings().First().Find(".loot-amount").Text())
		report.LargeCargo = &largeCargo
		colonyShip := utils.ParseInt(fleet.Find(".loot-item technology-icon[colonyship]").Siblings().First().Find(".loot-amount").Text())
		report.ColonyShip = &colonyShip
		recycler := utils.ParseInt(fleet.Find(".loot-item technology-icon[recycler]").Siblings().First().Find(".loot-amount").Text())
		report.Recycler = &recycler
		espionageProbe := utils.ParseInt(fleet.Find(".loot-item technology-icon[espionageprobe]").Siblings().First().Find(".loot-amount").Text())
		report.EspionageProbe = &espionageProbe
		solarSatellite := utils.ParseInt(fleet.Find(".loot-item technology-icon[solarsatellite]").Siblings().First().Find(".loot-amount").Text())
		report.SolarSatellite = &solarSatellite
		crawler := utils.ParseInt(fleet.Find(".loot-item technology-icon[resbuggy]").Siblings().First().Find(".loot-amount").Text())
		report.Crawler = &crawler
	}

	defenses := doc.Find(".defenseSection")
	if defenses.Size() > 0 {
		report.HasDefensesInformation = true
		rocketLauncher := utils.ParseInt(defenses.Find(".loot-item technology-icon[rocketlauncher]").Siblings().First().Find(".loot-amount").Text())
		report.RocketLauncher = &rocketLauncher
		lightLaser := utils.ParseInt(defenses.Find(".loot-item technology-icon[lasercannonlight]").Siblings().First().Find(".loot-amount").Text())
		report.LightLaser = &lightLaser
		heavyLaser := utils.ParseInt(defenses.Find(".loot-item technology-icon[lasercannonheavy]").Siblings().First().Find(".loot-amount").Text())
		report.HeavyLaser = &heavyLaser
		gaussCannon := utils.ParseInt(defenses.Find(".loot-item technology-icon[gausscannon]").Siblings().First().Find(".loot-amount").Text())
		report.GaussCannon = &gaussCannon
		ionCannon := utils.ParseInt(defenses.Find(".loot-item technology-icon[ioncannon]").Siblings().First().Find(".loot-amount").Text())
		report.IonCannon = &ionCannon
		plasmaTurret := utils.ParseInt(defenses.Find(".loot-item technology-icon[plasmacannon]").Siblings().First().Find(".loot-amount").Text())
		report.PlasmaTurret = &plasmaTurret
		smallShieldDome := utils.ParseInt(defenses.Find(".loot-item technology-icon[shielddomesmall]").Siblings().First().Find(".loot-amount").Text())
		report.SmallShieldDome = &smallShieldDome
		largeShieldDome := utils.ParseInt(defenses.Find(".loot-item technology-icon[shielddomelarge]").Siblings().First().Find(".loot-amount").Text())
		report.LargeShieldDome = &largeShieldDome
		antiBallisticMissiles := utils.ParseInt(defenses.Find(".loot-item technology-icon[missileinterceptor]").Siblings().First().Find(".loot-amount").Text())
		report.AntiBallisticMissiles = &antiBallisticMissiles
		interplanetaryMissiles := utils.ParseInt(defenses.Find(".loot-item technology-icon[missileinterplanetary]").Siblings().First().Find(".loot-amount").Text())
		report.InterplanetaryMissiles = &interplanetaryMissiles
	}

	buildings := doc.Find(".buildingsSection")
	if buildings.Size() > 0 {
		report.HasBuildingsInformation = true
		// TODO: implement building levels (there seems to be no IDs on the elements)
	}

	researches := doc.Find(".researchSection")
	if researches.Size() > 0 {
		report.HasResearchesInformation = true
		// TODO: implement research levels (there seems to be no IDs on the elements)
	}

	if hasError {
		return report, ogame.ErrDeactivateHidePictures
	}
	return report, nil
}

// OK, MISSING DETAILED RESOURCES INFO
func extractCombatReportMessagesFromDoc(doc *goquery.Document) ([]ogame.CombatReportSummary, int64, error) {
	msgs := make([]ogame.CombatReportSummary, 0)
	doc.Find(".msg").Each(func(i int, s *goquery.Selection) {
		if idStr, exists := s.Attr("data-msg-id"); exists {
			if id, err := utils.ParseI64(idStr); err == nil {
				report := ogame.CombatReportSummary{ID: id}
				report.Destination = v6.ExtractCoord(s.Find("div.msgHead a").Text())
				report.Destination.Type = ogame.PlanetType
				if s.Find("div.msgHead figure").HasClass("moon") {
					report.Destination.Type = ogame.MoonType
				}

				attackerName := s.Find(".basicInfo .msg_ctn.msg_ctn2").Text()
				m := regexp.MustCompile(`\([^']+\)`).FindStringSubmatch(attackerName)
				if len(m) == 1 {
					report.AttackerName = m[0][1 : len(m[0])-1]
				}

				defenderName := s.Find(".miscInfo .msg_ctn.msg_ctn2").Text()
				m = regexp.MustCompile(`\([^']+\)`).FindStringSubmatch(defenderName)
				if len(m) == 1 {
					report.DefenderName = m[0][1 : len(m[0])-1]
				}

				apiKeyTitle := s.Find(".icon_apikey").AttrOr("title", "")
				m = regexp.MustCompile(`'(cr-[^']+)'`).FindStringSubmatch(apiKeyTitle)
				if len(m) == 2 {
					report.APIKey = m[1]
				}

				// There are no more detailed infos about looted resources in the combat report summary
				rel := s.Find(".basicInfo .msg_ctn.msg_ctn3:nth-child(2)").Text()
				m = regexp.MustCompile(`[\d.,]+[^\d]*([\d.,]+)`).FindStringSubmatch(rel)
				if len(m) == 2 {
					report.Loot = utils.ParseInt(m[1])
				}
				m = regexp.MustCompile(`: ([\d.,]+)`).FindStringSubmatch(rel)
				if len(m) == 2 {
					report.Resources = utils.ParseInt(m[1][0 : len(m[1])-1])
				}

				report.DebrisField = utils.ParseInt(s.Find(".basicInfo .msg_ctn.msg_ctn3:nth-child(3)").AttrOr("title", "0"))

				msgDate, _ := time.Parse("02.01.2006 15:04:05", s.Find("div.msgDate").Text())
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
	return msgs, 0, nil
}

// OK
func extractExpeditionMessagesFromDoc(doc *goquery.Document, location *time.Location) ([]ogame.ExpeditionMessage, int64, error) {
	msgs := make([]ogame.ExpeditionMessage, 0)
	doc.Find(".msg").Each(func(i int, s *goquery.Selection) {
		if idStr, exists := s.Attr("data-msg-id"); exists {
			if id, err := utils.ParseI64(idStr); err == nil {
				msg := ogame.ExpeditionMessage{ID: id}

				msg.CreatedAt, _ = time.ParseInLocation("02.01.2006 15:04:05", s.Find("div.msgDate").Text(), location)
				msg.ExpeditionResult = ogame.ExpeditionResultUnknown

				msg.Coordinate = v6.ExtractCoord(s.Find("div.msgTitle a").Text())
				msg.Coordinate.Type = ogame.PlanetType

				msg.Content, _ = s.Find("div.msgContent").Html()
				msg.Content = strings.TrimSpace(msg.Content)

				msg.Resources.Metal = utils.ParseInt(s.Find("resource-icon.metal .amount").First().AttrOr("title", "0"))
				msg.Resources.Crystal = utils.ParseInt(s.Find("resource-icon.crystal .amount").First().AttrOr("title", "0"))
				msg.Resources.Deuterium = utils.ParseInt(s.Find("resource-icon.deuterium .amount").First().AttrOr("title", "0"))
				if msg.Resources.Metal > 0 || msg.Resources.Crystal > 0 || msg.Resources.Deuterium > 0 {
					msg.ExpeditionResult = ogame.ExpeditionResultResources
				}

				msg.Resources.Darkmatter = utils.ParseInt(s.Find("resource-icon.darkmatter .amount").First().Text())
				if msg.Resources.Darkmatter > 0 {
					msg.ExpeditionResult = ogame.ExpeditionResuldDarkmatter
				}

				type ShipInfo struct {
					Amount int64  `json:"amount"`
					Name   string `json:"name"`
				}
				var msgDataStruct struct {
					SmallCargo     ShipInfo `json:"202"`
					LargeCargo     ShipInfo `json:"203"`
					LightFighter   ShipInfo `json:"204"`
					HeavyFighter   ShipInfo `json:"205"`
					Cruiser        ShipInfo `json:"206"`
					Battleship     ShipInfo `json:"207"`
					EspionageProbe ShipInfo `json:"210"`
					Bomber         ShipInfo `json:"211"`
					Destroyer      ShipInfo `json:"213"`
					Battlecruiser  ShipInfo `json:"215"`
					Reaper         ShipInfo `json:"218"`
					Pathfinder     ShipInfo `json:"219"`
				}
				techGained := s.Find("div.rawMessageData").AttrOr("data-raw-technologiesgained", "{}")
				json.Unmarshal([]byte(techGained), &msgDataStruct)
				msg.Ships.SmallCargo = msgDataStruct.SmallCargo.Amount
				msg.Ships.LargeCargo = msgDataStruct.LargeCargo.Amount
				msg.Ships.LightFighter = msgDataStruct.LightFighter.Amount
				msg.Ships.HeavyFighter = msgDataStruct.HeavyFighter.Amount
				msg.Ships.Cruiser = msgDataStruct.Cruiser.Amount
				msg.Ships.Battleship = msgDataStruct.Battleship.Amount
				msg.Ships.EspionageProbe = msgDataStruct.EspionageProbe.Amount
				msg.Ships.Bomber = msgDataStruct.Bomber.Amount
				msg.Ships.Destroyer = msgDataStruct.Destroyer.Amount
				msg.Ships.Battlecruiser = msgDataStruct.Battlecruiser.Amount
				msg.Ships.Reaper = msgDataStruct.Reaper.Amount
				msg.Ships.Pathfinder = msgDataStruct.Pathfinder.Amount
				if msg.Ships.SmallCargo > 0 || msg.Ships.LargeCargo > 0 || msg.Ships.LightFighter > 0 || msg.Ships.HeavyFighter > 0 || msg.Ships.Cruiser > 0 || msg.Ships.Battleship > 0 || msg.Ships.EspionageProbe > 0 || msg.Ships.Bomber > 0 || msg.Ships.Destroyer > 0 || msg.Ships.Battlecruiser > 0 || msg.Ships.Reaper > 0 || msg.Ships.Pathfinder > 0 {
					msg.ExpeditionResult = ogame.ExpeditionResultShips
				}

				itemGained := s.Find("div.rawMessageData[data-raw-itemsgained]")
				if itemGained.Length() > 0 {
					msg.ExpeditionResult = ogame.ExpeditionResultItem
				}

				content := strings.ToLower(msg.Content)
				content = regexp.MustCompile(`\.|,|\(|\)|:`).ReplaceAllString(content, "")
				content = regexp.MustCompile(`( de | de$)`).ReplaceAllString(content, "")

				if msg.ExpeditionResult == ogame.ExpeditionResultUnknown {
					nothing := [...]string{"bueno ahora sabemos que esas anomalías r", "tu flota expedición siguió unas peculiar", "un ser pura energía causó que toda la tr", "a pesar los prometedores escaneos prelim", "quizás no deberíamos haber celebrado el ", "podivný počítačový virus nakazil navigač", "dziwny wirus komputerowy zaatakował syst", "krátko po opustení našej slnečnej sústav", "udover nogle underlige smådyr på en uken", "un virus calculator a atacat sistemele n", "een onbekend computervirus is in het nav", "egy ismeretlen számítógép vírus támadta ", "uno strano virus per computer ha intacca", "a strange computer virus attacked the na", "un virus informatique a fait planter vot", "un extraño virus informático atacó el si", "ein seltsames computervirus legte kurz n", "um virus computador atacou o sistema nav", "το σύστημα ναυσιπλοϊας προσβλήθηκε από κ", "evrenin tüm yönbulma sistemleri garip bi", "экспедиция не принесла ничего особого кр", "celá expedice strávila mnoho času zírání", "żywa istota zbudowana z czystej energii ", "neznáma životná forma z čistej energie s", "en skabelse af ren energi sikrede at eks", "o fiinta formata din energie pura s-a as", "een levensvorm van pure energie hypnotis", "egy élőlény akit tiszta energiából csiná", "una forma di vita di pura energia ha fat", "a living being made out of pure energy c", "une forme vie composée d`énergie pure a ", "un ser pura energía se aseguró que todos", "eine lebensform aus reiner energie hat d", "um ser feito energia hipnotizou toda a t", "μια οντότητα αποτελούμενη από καθαρή ενέ", "bu kesif seni kelimenin tam anlamiyla ev", "несмотря на первые многообещающие сканы ", "porucha v reaktoru velitelské lodi málem", "awaria reaktora głównego statku ekspedyc", "kaptajnens fødselsdagsfest burde nok ikk", "un esec la reactorul navei lider aproape", "een storing in reactor van het moedersch", "a vezető hajó reaktorának meghibásodása ", "un problema al reattore della nave ammir", "a failure in the flagships reactor core ", "un problème réacteur a failli détruire t", "un fallo en los motores la nave insignia", "ein reaktorfehler des führungsschiffes h", "um problema no reactor da nave principal", "μια αποτυχία στον αντιδραστήρα του ηγούμ", "bu kesif sonucunda kücük garip bir yarat", "неполадка в реакторе ведущего корабля чу", "někdo nainstaloval na palubní počítače e", "ekspedycja niemalże została złapana prze", "oslava kapitánových narodenín na povrchu", "det nye navigationsmodul har stadigvæk n", "iemand heeft een oud strategiespel geïns", "valaki feltelepített egy régi stratégiai", "qualcuno ha installato un antico gioco d", "due to a failure in the central computer", "quelqu`un a installé un jeu stratégie su", "alguien instaló un viejo juego estrategi", "irgendjemand hat auf allen schiffscomput", "alguém instalou um velho jogo estratégia", "κάποιος εγκατέστησε ένα παλιό παιχνίδι σ", "filo merkez gemisinin reaktöründeki hata", "жизненная форма состоящая из чистой энер", "tvá expedice se poučila o prázdnosti tét", "ekspedycja napotkała na rozległą pustkę ", "din ekspedition har lært om store tomrum", "expeditia ta a invatat despre pustietate", "de expeditie heeft uitgebreid onderzoek ", "az expedíciód megtanulta mi a nagy üres ", "la tua spedizione si è imbattuta nel vuo", "your expedition has learnt about the ext", "votre expédition a découvert le vide pas", "tu expedición aprendió acerca del extens", "deine expedition hat wortwörtlich mit de", "a tripulação descobriu o significado da ", "η αποστολή έμαθε για το αχανές κενό του ", "bu bölgeden gelen ilk raporlar cok ilgi ", "ваша экспедиция в прямом смысле слова по", "あなたの艦隊はこの空間にはなにもないことを学んだ、小惑星や何か分子でさえ存在しな", "zdá se že jsme na té osamělé planetě nem", "naš ekspedicijski tim je naišao na čudnu", "ktoś zainstalował w komputerach statku s", "en fejl i moderskibets reaktor ødelagde ", "echipa noastră expediție a ajuns la o co", "waarschijnlijk was het toch geen goed id", "valószínűleg a kapitányok szülinapi ünne", "forse le celebrazioni per il compleanno ", "our expedition team came across a strang", "nous n`aurions peut-être pas dû fêter l`", "probablemente la celebración del cumplea", "vielleicht hätte man den geburtstag des ", "provavelmente a festa aniversário do cap", "τα γενέθλια του καπετάνιου μάλλον δεν έπ", "ilginc bir sekilde gemi mürettabindan bi", "ваша экспедиция сделала замечательные сн", "až na pár prvních velmi slibných scannů ", "iako su prva skeniranja sektora bila dob", "mimo pierwszych obiecujących skanów tego", "den nye og vovede kommandør har med succ", "in ciuda primelor foarte promitatoare sc", "ondanks veelbelovende scans van deze sec", "az eleinte ígéretes letapogatási eredmén", "nonostante la prima scansione mostrasse ", "despite the first very promising scans o", "malgré un scan du secteur assez promette", "a pesar los resultados iniciales escaneo", "trotz der ersten vielversprechenden scan", "embora este sector tenha mostrado result", "δυστυχώς παρά τις πολλά υποσχόμενες αρχι", "kesif filon acil durum sinyali yakaladi!", "ну по крайней мере мы теперь знаем что к", "už víme že ty podivné anomálie dokážou z", "teraz wiemy że czerwone anomalie klasy 5", "fajn - prinajmenšom už vieme že červené ", "ei bine acum stim ca acele anomalii rosi", "hoe dan ook we weten nu in ieder geval d", "nos mostmár tudjuk hogy azok a piros 5-ö", "bene ora sappiamo che le anomalie rosse ", "well now we know that those red class 5 ", "bien nous savons désormais que les anoma", "bueno ahora sabemos que esas 5 anomalías", "nun zumindest weiß man jetzt dass rote a", "bem agora sabemos que aquelas anomalias ", "τώρα λοιπόν γνωρίζουμε πως αυτές οι αστρ", "galiba kaptanin dogumgününün bu bilinmed", "вскоре после выхода за пределы солнечной", "expedice pořídila skvělé záběry supernov", "podczas wyprawy zrobiono wspaniałe zdjęc", "expedícii sa podarilo zaznamenať úžasné ", "expeditia ta a facut poze superbe la o s", "je expeditie heeft adembenemende foto`s ", "az expedíciód egy fantasztikus képet kés", "la tua spedizione ha fatto stupende foto", "your expedition took gorgeous pictures o", "votre expédition a fait superbes images ", "tu expedición hizo magníficas fotos una ", "deine expedition hat wunderschöne bilder", "a expedição tirou lindas fotos uma super", "η αποστολή έβγαλε μερικές φαντασμαγορικέ", "kesif filon supernova` nin cok güzel res", "думаю не стоило всё-таки отмечать день р", "expedice po nějakou dobu sledovala podiv", "ekspedycja śledziła dziwny sygnał od jak", "expedičná flotila sledovala zvláštny sig", "din ekspeditionsflåde opsnapper nogle un", "flota ta expeditie a urmarit niste semna", "je expeditievloot heeft korte tijd vreem", "az expedíciós flottád különös jeleket kö", "la tua flotta in esplorazione ha seguito", "your expedition fleet followed odd signa", "votre expédition a suivi la trace signau", "tu flota en expedición siguió señales fu", "deine expeditionsflotte folgte einige ze", "a tua frota expedição seguiu uns sinais ", "η αποστολή σας ακολούθησε περίεργα σήματ", "eh en azindan simdi herkes 5 siniftan ki", "кто-то установил на всех корабельных ком", "až na pár malých zvířátek z neznámé baži", "ekspedicija nije vratila ništa drugo osi", "poza osobliwymi małymi zwierzętami pocho", "okrem zvláštneho domáceho zvieratka z ne", "cu exceptia unor animale mici pe o plane", "behalve een bijzonder vreemd klein dier ", "néhány kicsi furcsa háziállaton kívül am", "eccetto alcuni quaint piccoli animali pr", "besides some quaint small pets from a un", "mis à part quelques petits animaux prove", "además algunos pintorescos pequeños anim", "außer einiger kurioser kleiner tierchen ", "para além uns pequenos e esquisitos anim", "εκτός από μερικά περίεργα μικρά ζώα από ", "bir ara kesif filona garip sinyaller esl", "ваш экспедиционный флот следовал некотор", "あなたの艦隊は沼地の惑星で風変わりなペットを見つけた以外に何も収穫がありませんで", "tvá expedice málem nabourala do neutrono", "vaša ekspedicija je naletjela u gravitac", "ktoś zainstalował w komputerach statku s", "ekspeditionsflåden kom tæt på gravitatio", "expeditia ta aproape că a nimerit in câm", "je expeditie kwam bijna in een zwaartekr", "az expedíciód túl közel került egy neutr", "la tua spedizione si è imbattuta nel cam", "your expedition nearly ran into a neutro", "votre flotte d`expédition a eu chaud ell", "tu expedición casi entra en el campo gra", "deine expeditionsflotte geriet gefährlic", "a tua frota entrou no campo gravitaciona", "η αποστολή σας παραλίγο να εγκλωβιστεί σ", "tüm mürettebatın saf enerjiden oluşan bi", "ваш экспедиционный флот попал в опасную ", "chyba nie powinniśmy byli urządzać przyj", "a strange computer virus infected the sh", "the expeditionary fleet followed the str", "tu expedición se ha enfrentado textualme", "aparte unos pintorescos animalitos prove", "a sua frota expedição seguiu uns sinais ", "um estranho vírus computador atacou ao s", "apesar inicialmente esse setor ter mostr", "uma falha no reator da nave principal qu", "a sua expedição entrou no campo gravitac", "a sua expedição descobriu o significado ", "além uns pequenos e esquisitos animais u", "din ekspedition tog fantastiske billeder", "en underlig computervirus angreb navigat", "på grund af en fejl i det centrale compu", "trods det første meget lovende skan af s", "tja nu ved vi at der røde klasse 5 urege", "retkikuntasi otti mahtavia kuvia superno", "retkikuntasi seurasi outoja signaaleja j", "outo tietokonevirus iski navigaatiojärje", "lippulaivan keskustietokoneessa ilmennee", "puhtaasta energiasta tehty elävä pakotti", "huolimatta erittäin lupaavista ensimmäis", "retkikuntajoukkueemme saapui oudolle sii", "no nyt tiedämme että noilla punaisilla l", "vika lippulaivan reaktorin ytimessä lähe", "retkiuntasi lähes ajautui neutronitähden", "retkikuntasi on oppinut paljon avaruuden", "lukuunottamatta joitakin vanhanaikaisia ", "zbog otkazivanja brodskog reaktora jedno", "čudni računalni virus je zarazio brodsku", "sada znamo da te crvene anomalije klase ", "vaša ekspedicija je prikupila nova sazna", "vasa ekspedicija je snimila prelijepe sl", "ekspedicijska flota je pratila čudan sig", "netko je instalirao staru stratešku igru", "živo biće napravljeno od čiste energije ", "旗艦の動力炉の故障によりあなたの艦隊は壊滅しかけました。幸いにも有能な技術者によ", "あなたの艦隊は中性子惑星の引力に曳かれ、脱出するのに幾ばくかの時間がかかりました", "あなたの探索チームは、はるか昔に廃れた妙なコロニーにたどり着いた。\n着陸後、乗組", "我々の本星のある太陽系を脱出するとすぐに、奇妙なコンピューターウイルスがナビゲー", "私たちはクラス５に分類されるエリアは船のナビゲーションシステムを狂わすだけではな", "あなたの艦隊は素晴らしい超新星の写真を撮ることに成功した。探索としては特に何も無", "あなたの艦隊は奇妙なシグナルに近づきました。そしてそれは異星人の調査船の動力炉か", "旗艦のメインコンピューターの故障により探索任務は中止されました。残念ながら艦隊は", "艦隊に接近してきた小型エネルギー生命体によってあなたの艦隊の乗組員は睡眠状態に陥", "とても見込みのある領域を探索したのにも関わらず、残念ながら収穫なしに帰還しました", "tu expedición ha aprendido sobre el exte", "cineva a instalat cu joc strategic vechi", "ett fel i ledarskeppets reaktor förstörd", "din expedition körde nästan in i ett gra", "förutom lite lustiga små djur från en ok", "kaptenens födelsedagsfest skulle kanske ", "ett konstigt datorvirus attackerade navi", "ja nu vet vi iallafall att där röda klas", "din expedition har lärt sig hur tom rymd", "din expedition tog läckra bilder utav en", "din expeditionsflotta följde några märkl", "någon installerade ett gammalt strategis", "en varelse gjord utav enbart energi gjor", "trots den första mycket lovande satellit", "zaradi odpovedi enega od ladijskih motor", "tvoja ekspedicija je naletela na gravita", "ekspedicija je prinesla nazaj samo neka ", "naša ekspedicija je naletela na čudno ko", "čuden virus je napadel navigacijski sist", "zdaj vsaj vemo da te rdeče razreda 5 ano", "tvoja ekspedicija je pridobila znanje o ", "tvoji ekspediciji je uspelo narediti vel", "ekspedicijska flota je nekaj časa spreml", "prišlo je do napake na računalniških sis", "naleteli smo na vesoljsko bitje ki je om", "čeprav so bile prve raziskave področja o", "porucha na reaktore veliteľskej lodi tak", "tvoja expedícia takmer skončila v gravit", "členovia výpravy sa počas expedície dôkl", "nejaký dobrák nainštaloval do palubného ", "napriek pozitívnym očakávaniam vychádzaj", "旗艦的反應器失常差點導致整個遠征探險艦隊覆滅值得慶幸的是技術專家們表現極為出色避", "您的遠征探險隊誤入了一顆中子星的引力場需要一段時間來掙脫該力場由於在掙脫時幾乎將", "在那個不知名的沼澤行星上除了一些新奇有趣的小動物這次遠征探險艦隊在旅程中一無所獲", "我們的遠征探險隊途徑一個已經廢棄很久的怪異殖民星降落以後我們的船員感染了一種外星", "就在剛離開我們母星太陽系宙域不久後一種怪異的電腦病毒入侵了導航系統這使得遠征探險", "現在我們已經知道了那些紅色5級異象不僅對艦船的導航系統帶來混亂干擾同時也使船員產", "您的遠征探險隊已對帝國遼闊的空域瞭如指掌這裡毫無新意甚至連一顆小行星或輻射源乃至", "您的遠征探險隊拍攝了超新星華麗的照片雖然遠征探險隊沒有帶回來任何新的發現但是最起", "您的遠征探險隊斷斷續續追蹤到一些奇怪的訊號後來才知道原來那些訊號是從古老的間諜衛", "由於旗艦的中央電腦系統發生錯誤探險任務不得不終止另外由於電腦故障的原因我們的艦隊", "一個散發著高純能量的生物悄然來到甲板上用意念控制所有的探險隊員令他們凝視著電腦熒", "儘管我們是第一個來到這個非常有希望的區域很不幸我們空手而歸\n\n通訊官日誌記錄似乎", "儘管我們是第一個來到這個非常有希望的區域很不幸我們空手而歸\n\n通訊官日誌記錄作為", "your expedition has learned about the ex"}
					for _, s := range nothing {
						if strings.Contains(content, s) {
							msg.ExpeditionResult = ogame.ExpeditionResultNothing
							break
						}
					}
				}
				if msg.ExpeditionResult == ogame.ExpeditionResultUnknown {
					early := [...]string{"tvá expedice nenalezla nic zajímavého v ", "twoja ekspedycja nie wykryła żadnych ano", "din ekspedition har ikke rapporteret nog", "je expeditie heeft geen rariteiten gevon", "az expedícióid nem jelentett semmilyen r", "la tua spedizione non riporta alcuna ano", "your expeditions doesn`t report any anom", "votre expédition ne signale aucune parti", "tu expedición no informa ninguna anomalí", "deine expedition meldet keine besonderhe", "a tua expedição não reportou qualquer an", "η αποστολή σας δεν αναφέρει ανωμαλίες στ", "tüm mürettebatın saf enerjiden oluşan bi", "неожиданное замыкание в энергетических к", "nečekané výboje v skladištích energie mo", "niespodziewane sprzężenie zwrotne w zwoj", "den nye og vovede kommandør har med succ", "o neasteptata explozie la motoarele din ", "een onverwachte terugkoppeling in energi", "váratlanul meghibásodtak a hajtóművek a ", "un componente inserito nel generatore di", "an unexpected back coupling in the energ", "un petit défaut dans les réacteurs votre", "un inesperado acoplamiento energía en lo", "eine unvorhergesehene rückkopplung in de", "um problema no reactor da nave principal", "μια απρόσμενη ανάστροφη σύζευξη στα ενερ", "kesfedilen bölgede olagandisi bir veriye", "отважный новый командир использовал нест", "nový a odvážný velitel letky úspěšně pro", "młody odważny dowódca pomyślnie przedost", "en uforudset tilbagekobling i energispol", "noul si putin indraznetul comandant a tr", "je nieuwe en ongeremde vlootcommandant h", "az új és kicsit merész parancsnok sikere", "il nuovo e audace comandante ha usato co", "the new and daring commander successfull", "votre nouveau commandant bord étant asse", "¡el nuevo y un poco atrevido comandante ", "der etwas wagemutige neue kommandant nut", "um comandante novo e destemido conseguiu", "ο νέος και τολμηρός διοικητής ταξίδευσε ", "motor takımlarının enerji halkalarında y", "ваша экспедиция не сообщает ничего необы", "keşif filon bir şekilde nötron yıldızlar", "geminin yeni komutani oldukca cesur cikt", "a sua expedição não reportou nenhuma ano", "retikuntasi ei raportoi mitään poikkeami", "el nuevo comandante que es bastante osad", "o novo e destemido comandante conseguiu ", "uusi ja rohkea komentaja ohjasi laivueen", "novi i odvažni komander je uspješno puto", "izvještaji ekspedicije ne javljaju nikak", "新任の勇敢な指揮官は不安定なワームホールを通って帰還を早めることに成功しました。", "あなたの艦隊は探索した地域で異変を発見することはできませんでした。しかし、艦隊は", "expeditia ta nu raporteaza nici o anorma", "den nya och lite orädda befälhavaren lyc", "dina expeditioner rapporterar inga avvik", "novi poveljnik je uspešno potoval skozi ", "poročila ekspedicije ne poročajo o nikak", "nový mimoriadne ctižiadostivý veliteľ ús", "naša výprava nehlási žiadne anomálie v s", "年輕而膽識過人的指揮官成功穿越了一個不穩定的蟲洞減少了返回的飛行時間!而然這支遠", "您的遠征探險艦隊報告稱在遠征探險的宇宙空域內並沒有找到什麼異象正當他們返回之時艦", "um problema inesperado no campo energéti", "odottamaton takaisinkytkentä moottorin e", "neke anomalije u motorima ekspedicijskih", "en oväntad omvänd koppling i energispola", "anomalije v motorjih ekspedicijskih ladi", "neočakávané spätné výboje v pohonných je", "在引擎的能源軸線上發生了一個未能預期耦合逆轉狀況導致艦隊加速了返回時間遠征探險艦"}
					for _, s := range early {
						if strings.Contains(content, s) {
							msg.ExpeditionResult = ogame.ExpeditionResultEarly
							break
						}
					}
				}
				if msg.ExpeditionResult == ogame.ExpeditionResultUnknown {
					late := [...]string{"debido a motivos desconocidos el salto l", "tu expedición entró en un sector asolado", "z neznámých důvodů se expedice vynořila ", "z nieznanych powodów ekspedycja nieomal ", "z neznámych dôvodov sa expedícii podaril", "på grund af en ukendt fejl gik ekspediti", "datorita motivelor necunoscute saltul ex", "door onbekende oorzaak is expeditiespron", "ismeretlen okok miatt az expedíciós ugrá", "per ragioni sconosciute il salto nell`ip", "for unknown reasons the expeditions jump", "pour une raison inconnue le saut spatial", "a causa razones desconocidas el salto la", "aus bisher unbekannten gründen ging der ", "devido a razões ainda desconhecidas o sa", "για άγνωστη αιτία το άλμα της αποστολής ", "kırmızı devin yildiz rüzgari yüzünden ke", "халтурно собранный навигатор неправильно", "tvůj navigátor se dopustil vážné chyby v", "główny nawigator miał zły dzień co spowo", "navigátor výpravy sa dopustil hrubej chy", "din navigationschef havde en dårlig dag ", "de navigatieleider had een slechte dag w", "a navigációs vezetőnek rossz napja volt ", "una svista commessa dal capo-navigazione", "your navigator made a grave error in his", "une erreur calcul votre officier navigat", "el líder la navegación tuvo un mal día y", "ein böser patzer des navigators führte z", "um erro no sistema navegação fez com que", "ο κυβερνήτης της αποστολής δε κοιμήθηκε ", "sebebi bilinmeyen bir arizadan dolayi gö", "новый навигационный модуль всё-таки имее", "nový navigační modul má pořád nějaké mou", "ekspeditionens moderskib havde et sammen", "noul modul navigare inca se lupta cu une", "de nieuwe navigatiemodule heeft nog een ", "kesif filon tanecik fırtınası yaşanan bi", "звёздный ветер от красного гиганта исказ", "nowy system nawigacyjny nadal nie jest w", "das neue navigationsmodul hat wohl doch ", "o novo módulo navegação ainda tem alguns", "a navigációs modul még hibákkal kűzd az ", "il nuovo modulo di navigazione sta ancor", "the new navigation module is still buggy", "votre module navigation semble avoir que", "el nuevo módulo navegación está aún luch", "το νέο σύστημα ναυσιπλοΐας αντιμετωπίζει", "kesif merkez gemin hicbir uyarida bulunm", "по пока неустановленным причинам прыжок ", "sluneční vítr rudého obra překazil skok ", "gwiezdny wiatr wiejący ze strony czerwon", "din navigationschef havde en dårlig dag ", "vantul unei stele ale unui gigant rosu a", "de zonnewind van een rode reus verstoord", "egy vörös óriás csillagszele tönktretett", "il vento-stellare di una gigante rossa h", "the solar wind of a red giant ruined the", "le vent solaire causé par une supernova ", "el viento una estrella gigante roja arru", "der sternwind eines roten riesen verfäls", "o vento solar uma gigante vermelha fez c", "ο αστρικός άνεμος ενός κόκκινου γίγαντα ", "kesif filon navigasyon sistemindeki önem", "ваша экспедиция попала в сектор с усилен", "tvoje expedice se dostala do sektoru pln", "ekspedicija je završila u sektoru sa olu", "twoja ekspedycja osiągnęła sektor pełen ", "din ekspedition er kommet ind i en parti", "expediția a ajuns în mijlocul unei furtu", "je expeditie komt terecht in een sector ", "az expedíciód egy részecskeviharral teli", "la tua spedizione è andata in un settore", "your expedition went into a sector full ", "votre expédition a dû faire face à plusi", "tu expedición entró en un sector lleno t", "deine expedition geriet in einen sektor ", "a missão entrou num sector com tempestad", "η αποστολή σας βρέθηκε σε τομέα σωματιδι", "yeni yönbulma ünitesi hala sorunlarla sa", "ведущий корабль вашего экспедиционного ф", "velitelská loď expedice se při výstupu z", "główny statek ekspedycyjny zderzył się z", "nava mama a expeditiei a facut o coliziu", "het moederschip van expeditie is in bots", "az expedició fő hajója ütközött egy ideg", "la nave ammiraglia della tua spedizione ", "the expedition`s flagship collided with ", "un vos vaisseaux est entré en collision ", "la nave principal la expedición colision", "das führungsschiff deiner expeditionsflo", "a nave principal colidiu com uma nave es", "η ναυαρχίδα της αποστολής συγκρούστηκε μ", "el viento estelar una gigante roja ha ar", "el nuevo módulo navegación aún está lidi", "un pequeño fallo del navegador provocó u", "el nuevo módulo navegación está aún llen", "a missão entrou num setor com tempestade", "o vento uma estrela vermelha gigante fez", "stjernevinden i en gigantisk rød stjerne", "retkikuntasi päätyi hiukkasmyrskyn täytt", "tuntemattomista syistä lentorata oli tot", "punaisen jättiläisen aurinkotuuli pilasi", "navigaatiomoduuli on silti buginen retki", "navigaattori teki vakavan virheen laskel", "a nave principal da expedição colidiu co", "retkikuntasi lippulaiva vieraaseen aluks", "ekspedicijska flota se susrela sa neprij", "navigator glavnog broda je imao loš dan ", "zbog nepoznatih razloga ekspedicijski sk", "gravitacija crvenog diva uništila je sko", "novi navigacijski modul još uvijek ima n", "異星の宇宙船があなたの旗艦に衝突しました。その際相手の宇宙船が爆発し、あなたの旗", "あなたの航海士は艦隊の航路を決めるにあたって深刻な計算ミスを犯しました。その結果", "あなたの艦隊は宇宙嵐に巻き込まれました。結果動力炉は壊れ、宇宙船の基幹システムは", "原因不明の事態により艦隊は予定とは違う場所にたどり着きました。危うく太陽に不時着", "恒星からの太陽風で航路は大きくずれました。\nその空間には全くなにもありませんでし", "新しいナビゲーションシステムにはまだバグが残っていました。それは艦隊を間違った座", "el líder en la navegación tuvo un mal dí", "el viento una estrella gigante roja ha a", "liderul navigatiei a avut o zi proasta s", "флагманский корабль вашего экспедиционно", "när expeditionen avslutade hyperrymdshop", "navigationsledaren hade en dålig dag och", "din expedition hamnade i en sektor fylld", "på grund av okända orsaker så blev exped", "stjärnvinden hos en röd jätte förstörde ", "den nya navigationsmodulen jobbar fortfa", "ekspedicijska flota je naletela na nezna", "poveljnik ekspedicije je imel slab dan i", "tvoja ekspedicijska flota je zašla v nev", "zaradi neznanih razlogov je naša ekspedi", "solarni veter od zvezde velikanke je pov", "novi navigacijski sistem ima še vedno na", "veliteľská loď sa dostala do kolízie s c", "expedícia sa dostala do oblasti postihnu", "slnečný vietor červeného obra kompletne ", "nový navigačný modul má stále vážne nedo", "一艘外來艦船突然跳躍到遠征探險艦隊中心與我們遠征探險隊旗艦相撞了外來艦船繼而爆炸", "您的電腦導航系統發生一個嚴重錯誤導致遠征探險隊空間跳躍失敗不但造成艦隊完全失去目", "您的遠征探險艦隊誤闖入了一個粒子風暴區域這使得能源供給出現超負荷現象並且大部分的", "由於不明原因遠征探險艦隊的空間跳躍總是頻頻出錯這次更離譜竟然跳到一顆恒星的心臟地", "一顆紅巨星的太陽風破壞了遠征探索艦隊的空間跳躍並令到艦隊不得不花費更多的時間來重", "新導航系統組件仍然有問題遠征探索艦隊的空間跳躍錯誤不僅使得他們去錯目的地更使得艦"}
					for _, s := range late {
						if strings.Contains(content, s) {
							msg.ExpeditionResult = ogame.ExpeditionResultLate
							break
						}
					}
				}
				if msg.ExpeditionResult == ogame.ExpeditionResultUnknown {
					alien := [...]string{"exoticky vypadající lodě neznámého původ", "kilka egzotycznie wyglądających statków ", "ekspeditionens moderskib havde et sammen", "niste nave aparent exotice au atacat flo", "onbekende exotisch ogende schepen vallen", "egzotikus megjelenésű hajók támadták meg", "alcune navi straniere hanno attaccato la", "some exotic looking ships attacked the e", "des vaisseaux inconnus ont attaqué la fl", "¡algunas naves apariencia exótica atacar", "einige fremdartig anmutende schiffe habe", "algumas naves exóticas atacaram a frota ", "μερικά σκάφη με εντυπωσιακή εμφάνιση επι", "kücük bir grup bilinmeyen gemi tarafinda", "на вашу экспедицию напал вражеский флот ", "何隻かの見慣れない宇宙船が警告無しに攻撃してきました。", "tvá expedice provedla ne-úplně-přátelské", "twoja ekspedycja napotkała niezbyt przyj", "din ekspeditionsflåde havde ikke en venl", "je expeditievloot heeft een onvriendelij", "a felderítő expedíciód elsőre nem túl ba", "la tua flotta in esplorazione non ha avu", "your expedition fleet had an unfriendly ", "la flotte d`expédition a eu une rencontr", "tu expedición no hizo un primer contacto", "deine expeditionsflotte hatte einen nich", "a tua frota exploração teve um primeiro ", "ο εξερευνητικός στόλος σας ήρθε σε όχι κ", "kesif ekibimiz bilinmeyen bir tür ile hi", "огромная армада хрустальных кораблей неи", "naše expedice byla přepadena malou skupi", "vores ekspedition blev angrebet af en mi", "expeditia noastra a fost atacata un grup", "onze expeditie is aangevallen door een k", "неизвестная раса атакует наш экспедицион", "nasza ekspedycja została zaatakowana prz", "unsere expedition wurde von einer kleine", "az expedíciónkat egy kisebb csapat ismer", "la nostra spedizione è stata attaccata d", "our expedition was attacked by a small g", "notre expédition a été attaquée par un p", "¡nuestra expedición fue atacada por un p", "a nossa missão foi atacada por um pequen", "η αποστολή δέχτηκε επίθεση από ένα μικρό", "ваш экспедиционный флот по всей видимост", "neznámí vetřelci zaútočili na naši exped", "nieznani obcy atakują twoją ekspedycję!", "expeditia noastra a fost atacata un grup", "een onbekende levensvorm valt onze exped", "egy ismeretlen faj megtámadta az expedíc", "una specie sconosciuta sta attaccando la", "an unknown species is attacking our expe", "une espèce inconnue attaque notre expédi", "¡una especie desconocida ataca nuestra e", "eine unbekannte spezies greift unsere ex", "uma espécie desconhecida atacou a nossa ", "связь с нашим экспедиционным флотом прер", "spojení s expediční letkou bylo přerušen", "kontakt z naszą ekspedycją został przerw", "de verbinding met onze expeditievloot we", "a kapcsolat az expedíciós flottával nemr", "il collegamento con la nostra spedizione", "the connection to our expedition fleet w", "nous avons perdu temporairement le conta", "el contacto con nuestra expedición fue i", "die verbindung zu unserer expeditionsflo", "a ligação com nossa frota exploratória f", "ваш экспедиционный флот испытал не особо", "tvá expedice narazila na území ovládané ", "úgy néz ki hogy a felfedező flottád elle", "la tua flotta in spedizione sembra aver ", "your expedition fleet seems to have flow", "votre flotte d`expédition a manifestemen", "tu expedición parece haber entrado en un", "deine expeditionsflotte hat anscheinend ", "tudo indica que a tua frota entrou em te", "ο στόλος της αποστολής εισχώρησε σε μια ", "какие-то корабли неизвестного происхожде", "mieliśmy trudności z wymówieniem dialekt", "je expeditie is het territorium van onbe", "volt egy kis nehézségünk az idegen faj n", "abbiamo avuto difficoltà a pronunciare c", "we had a bit of difficulty pronouncing t", "nous avons rencontré quelques difficulté", "tuvimos dificultades para pronunciar cor", "wir hatten mühe den korrekten dialekt ei", "encontrámos algumas dificuldades em pron", "на нашу экспедицию напала небольшая груп", "een grote onbekende vloot van kristallij", "una grande formazione di navi cristallin", "a large armada of crystalline ships of u", "une flotte vaisseaux cristallins va entr", "una gran formación naves cristalinas ori", "ein großer verband kristalliner schiffe ", "uma grande frota naves cristalinas orige", "язык этой расы труден в произношении сов", "tvá expedice narazila na mimozemskou inv", "twoja flota natrafiła na silną flotę obc", "az expedíciód egy idegen invázióba flott", "your expedition ran into an alien invasi", "votre mission d`expédition a rencontré u", "tu expedición encontró una flota alien i", "deine expedition ist in eine alien-invas", "a tua frota exploração foi atacada por u", "tu flota expedición no tuvo un primer co", "a sua frota expedição teve um primeiro c", "retkikuntalaivueesi loi vihamielisen ens", "¡unas naves exótico aspecto atacaron la ", "et fremmedartet skib angriber din eksped", "eksoottisen näköisiä aluksia hyökkäsi re", "neki brodovi egzotičnog izgleda su napal", "tvoja ekspedicijska flota nije napravila", "あなたの艦隊は友好的ではない正体不明の種族と接触しました。\n\n通信主任の報告書：", "flota ta expeditie a avut un prim contac", "några skepp med exotiskt utseende attack", "din expeditionsflotta har för första gån", "ladje eksotičnega izgleda so napadle naš", "tvoja ekspedicijska flota je imela nepri", "exoticky vyzerajúce lode bez výstrahy za", "naša expedícia má za sebou nie príliš pr", "egzotik görünüslü tarafimizca bilinmeyen", "一批奇形怪狀的外星艦船在事先毫無警告之下襲擊了我們的遠征探險艦隊!\n\n通訊官日誌", "您的遠征探險艦隊與一未知種族的外星人發生了首場衝突接觸\n\n通訊官日誌記錄作為第一", "your expedition fleet made some unfriend"}
					for _, s := range alien {
						if strings.Contains(content, s) {
							msg.ExpeditionResult = ogame.ExpeditionResultAliens
							break
						}
					}
				}
				if msg.ExpeditionResult == ogame.ExpeditionResultUnknown {
					pirate := [...]string{"interceptamos comunicaciones unos pirata", "museli jsme bojovat s vesmírnými piráty ", "musieliśmy walczyć z piratami na szczęśc", "am fost nevoiti sa ne luptam cu niste pi", "we moesten ons verdedigen tegen enkele p", "szükségünk van harcra néhány kalózzal sz", "abbiamo dovuto combattere alcuni pirati ", "we needed to fight some pirates which we", "nous avons dû nous défendre contre des p", "tuvimos que luchar contra algunos pirata", "wir mussten uns gegen einige piraten weh", "tivemos combater com uns piratas que por", "έπρεπε να αντιμετωπίσουμε μερικούς πειρα", "bazi ilkel barbarlar bize uzaygemisi ola", "пойманные сигналы исходили не от иноплан", "zachytili jsme radiovou zprávu od nějaký", "odebraliśmy sygnał radiowy od jakichś pi", "zachytili a dekódovali sme správu ožraté", "vi har sporet nogle berusede pirater der", "am prins un mesaj radio la niste pirati ", "we vingen een radiobericht op van enkele", "elfogtunk egy rádió üzenetet ami ittas k", "abbiamo intercettato messaggi di alcuni ", "we caught some radio transmissions from ", "nous avons capté des messages provenant ", "capturamos algunos mensajes radio alguno", "wir haben ein paar funksprüche sehr betr", "apanhamos umas mensagens via rádio e est", "υποκλέψαμε κάποια ραδιοσήματα από κάποιο", "karsimiza cikan uzay korsanlari neyseki ", "экспедиционный флот сообщает о жестоких ", "nějací primitivní barbaři na nás útočí z", "jacyś prymitywni barbarzyńcy atakują nas", "nogle primitive barbarer angriber os med", "niste pirati ne ataca cu nave inferior t", "enkele primitieve barbaren vallen ons aa", "néhány primitív barbár támadt ránk olyan", "alcuni barbari primitivi ci stanno attac", "some primitive barbarians are attacking ", "des barbares primitifs nous attaquent av", "algunos bárbaros primitivos están atacán", "einige primitive barbaren greifen uns mi", "uns bárbaros primitivos estão nos a atac", "μία πρωτόγονη φυλή πειρατών μας επιτίθετ", "kücük bir grup bilinmeyen gemi tarafinda", "ваш экспедиционный флот пережил неприятн", "nějací naprosto zoufalí vesmírní piráti ", "jacyś bardzo zdesperowani piraci próbowa", "niekoľko zúfalých vesmírnych pirátov sa ", "cativa pirati ai spatiului foarte disper", "een paar wanhopige piraten hebben geprob", "néhány űr-kalóz megpróbálta elfoglalni a", "alcuni pirati dello spazio decisamente d", "some really desperate space pirates trie", "quelques pirates apparemment complètemen", "algunos piratas realmente desesperados i", "ein paar anscheinend sehr verzweifelte w", "alguns piratas desesperados tentaram cap", "μερικοί πραγματικά απελπισμένοι πειρατές", "bazı umutsuz uzay korsanları keşif filom", "мы попались в лапы звёздным пиратам! бой", "vletěli jsme přímo do pasti připravé hvě", "sygnał alarmowy wykryty przez ekspedycję", "we liepen in een hinderlaag van een stel", "belefutottunk egy csillag-kalóz támadásb", "siamo incappati in un`imboscata tesa da ", "we ran straight into an ambush set by so", "nous sommes tombés dans un piège tendu p", "¡caimos en una emboscada organizada por ", "wir sind in den hinterhalt einiger stern", "nós fomos directos para uma emboscada ef", "yildiz korsanlarinin kurdugu tuzagin tam", "сигнал о помощи на который последовала э", "nouzový signál který expedice následoval", "wpadliśmy prosto w pułapkę zastawioną pr", "semnalul urgenta pe care l-a urmat exped", "het noodsignaal dat expeditie volgde ble", "a segélykérő jelet amit követett az expe", "la richiesta di aiuto a cui la spedizion", "that emergency signal that the expeditio", "le message secours était en fait un guet", "la señal emergencia que la expedición si", "der hilferuf dem die expedition folgte s", "o sinal emergência que a expedição receb", "το σήμα κινδύνου που ακολουθήσαμε ήταν δ", "sarhos uzay korsanlarindan bazi telsiz m", "пара отчаянных космических пиратов попыт", "zware gevechten tegen piratenschepen wor", "la spedizione riporta feroci scontri con", "the expedition reports tough battles aga", "votre flotte d`expédition nous signale l", "¡tu expedición informa duras batallas co", "die expeditionsflotte meldet schwere käm", "o relatório expedição relata batalhas ép", "нам пришлось обороняться от пиратов кото", "expedice měla nepříjemné setkání s vesmí", "din ekspeditionsflåde havde et ufint sam", "expeditia ta a avut o intalnire neplacut", "az expedíciódnak elégedetlen találkozása", "la tua spedizione ha avuto uno spiacevol", "your expedition had an unpleasant rendez", "votre flotte d`expédition a fait une ren", "tu expedición tuvo un desagradable encue", "eine expeditionsflotte hatte ein unschön", "a tua expedição deparou-se com uma não m", "мы перехватили переговоры пьяных пиратов", "zarejestrowane sygnały nie pochodziły od", "het noodsignaal dat expeditie volgde ble", "i segnali registrati non provenivano da ", "the recorded signals didn`t come from a ", "les signaux que nous ne pouvions identif", "¡las señales no provenían un extranjero ", "die aufgefangenen signale stammten nicht", "os sinais gravados não foram emitidos po", "нас атакуют какие-то варвары и хотя их п", "votre expédition est tombée sur des pira", "unos piratas realmente desesperados inte", "tuvimos que luchar contra unos piratas q", "unos bárbaros primitivos están atacándon", "necesitamos luchar con algunos piratas q", "apanhamos algumas mensagens rádio alguns", "alguns piratas espaciais desesperados te", "nós tivemos combater com alguns piratas ", "alguns bárbaros primitivos estão nos ata", "nogle øjensynligt fortvivlede pirater ha", "under ekspeditionen blev vi nødt til at ", "nappasimme joitakin radiolähetyksiä juop", "todella epätoivoiset avaruuspiraatit yri", "meidän täytyi taistella piraatteja onnek", "jotkin alkukantaiset barbaarit hyökkäävä", "neki opaki pirati su pokušali zarobiti v", "primili smo radio poruku od nekog pijano", "flota se morala boriti protiv nekoliko p", "neki primitivni svemirski barbari nas na", "いくつかの海賊は捨て身であなたの艦隊を乗っ取ろうとします。\n\n通信主任の報告書：", "あなたの艦隊は泥酔した海賊から通信を受けました。それによるとあなたの艦隊はまもな", "あなたの艦隊はいくつかの海賊と戦う必要がありますが、幸いにもそれはほんの少しだけ", "あなたの艦隊を旧式の宇宙船で攻撃してきた野蛮人の中には海賊とは言えない者もいまし", "algunos piratas espaciales que al parece", "нас атакуют какие-то варвары их примитив", "några riktigt desperata rymdpirater förs", "vi hörde ett radiomeddelande från några ", "vi behövde slåss emot några pirater som ", "några primitiva vildar anfaller oss med ", "obupani pirati so se trudili da bi zajel", "zasegli smo sporočilo od piratov zgleda ", "boriti smo se morali proti piratom kater", "primitivni vesoljski barbari nas napadaj", "musíme sa vysporiadať s pirátskou zberbo", "akási primitívna skupina barbarov sa nás", "一些亡命的宇宙海盜嘗試洗劫我們的遠征探險艦隊\n\n通訊官日誌記錄作為第一批到此未被", "我們從一幫張狂的海盜處收到一些挑釁的無線電訊號看來我們即將遭受攻擊\n\n通訊官日誌", "我們不得不與那裡的海盜進行戰鬥慶幸的是對方艦船數不多\n\n通訊官日誌記錄作為第一批", "一群原始野蠻人正利用太空船向我們的遠征探險艦隊發起攻擊我們甚至連他們叫什麼名都全"}
					for _, s := range pirate {
						if strings.Contains(content, s) {
							msg.ExpeditionResult = ogame.ExpeditionResultPirates
							break
						}
					}
				}
				if msg.ExpeditionResult == ogame.ExpeditionResultUnknown {
					trader := [...]string{"tvá expedice se setkala s přátelskou ras", "flota ekspedycyjna nawiązała kontakt z p", "din ekspeditionsflåde har opnået kontakt", "onze expeditievloot heeft contact gemaak", "az expedíciós flottád kapcsolatba lépett", "la tua spedizione ha avuto contatto con ", "your expedition fleet made contact with ", "votre expédition a eu un bref contact av", "tu flota en expedición tuvo un corto con", "deine expeditionsflotte hatte kurzen kon", "a tua frota contactou com uma raça alien", "ο στόλος της αποστολής σας ήρθε σε επαφή", "kesif filon biraz utangac bir alien irki", "ваш экспедиционный флот вышел на контакт", "tvá expedice zachytila nouzový signál ve", "ekspedycja odebrała sygnał alarmowy ogro", "din ekspeditionsflåde opsnappede et nøds", "een noodoproep bereikte je expeditie een", "az expedíciód vészjelzést fogott egy meg", "la tua spedizione ha ricevuto un segnale", "your expedition picked up an emergency s", "votre flotte d`expédition a recueilli un", "tu expedición captura un grito ayuda era", "deine expeditionsflotte hatte ein notsig", "a tua expedição recebeu um sinal emergên", "η αποστολή σας έλαβε σήμα κινδύνου ένα μ", "ваш экспедиционный флот поймал сигнал по", "a sua frota expedição fez contato com um", "retkikuntasi oli yhteydessä ystävällisee", "a sua expedição recebeu um sinal emergên", "tu expedición captó un grito ayuda era u", "retkikuntasi poimi hätäsignaalin kesken ", "vaša ekspedicija je primila signal za hi", "vaša ekspedicijska flota je uspostavila ", "あなたの艦隊は任務中に救難信号を受信しました。救難信号は巨大な輸送艦から出されて", "あなたの艦隊は友好的な種族の異星人と接触しました。彼らはあなたにとって有益な資源", "expeditia ta a primit un semnal urgenta ", "flota ta expeditia a facut contactul cu ", "din expedition fick en alarmsignal en en", "din expeditionsflotta fick kontakt med e", "ekspedicija je zaznala klice na pomoč og", "ekspedicija je vzpostavila kontakt s sra", "naša expedícia zachytila núdzový signál ", "naša expedícia nadviazala kontakt s mier", "您的遠征探險艦隊在任務中發出一則緊急的訊號一艘巨型貨運船被一顆小行星的萬有引力力", "您的遠征探險艦隊與一友善的外星人種族進行了聯絡他們宣布他們將派遣一名代表與您的帝"}
					for _, s := range trader {
						if strings.Contains(content, s) {
							msg.ExpeditionResult = ogame.ExpeditionResultTrader
							break
						}
					}
				}
				if msg.ExpeditionResult == ogame.ExpeditionResultUnknown {
					blackhole := [...]string{"jediná věc která po expedici zbyla je ná", "po naszej ekspedycji pozostała jedynie t", "singurul lucru ramas la expeditie a fost", "het enige dat is overgebleven van expedi", "az egyetlen dolog ami a küldetésből megm", "l`unica cosa che rimane dalla spedizione", "the only thing left from the expedition ", "voici le dernier signe vie l`expédition ", "lo único que quedó la expedición fue el ", "von der expedition ist nur noch folgende", "o último contacto que tivemos da frota e", "το μόνο που απέμεινε από την αποστολή εξ", "lider geminin ana reaktöründeki bir kayn", "от экспедиции осталось только следующее ", "roztavení jádra v hlavní lodi expedice v", "roztopienie rdzenia głównego statku powo", "o supra alimentare a miezului navei mama", "een ontploffing van hyperruimtemotor ver", "a vezető hajó magjának felmelegedése egy", "una rottura nel nucleo della nave ammira", "a core meltdown of the lead ship leads t", "un incident dans le noyau atomique d`un ", "una fusión del núcleo la nave insignia p", "ein kernbruch des führungsschiffes führt", "uma falha no núcleo do motor da nave-mãe", "ένα πρόβλημα στο σύστημα ψύξης του αντιδ", "раздробление ядра ведущего корабля вызва", "poslední informace od expedice byla velm", "jedina stvar koja je ostala od cijele ek", "ostatnią zdobyczą ekspedycji było napraw", "ultimul lucru pe care il avem la expedit", "het laatste bericht dat we ontvingen was", "az utolsó dolog amit az expedícióról kap", "l`ultima cosa che ci è stata inviata dal", "the last transmission we received from t", "l`expédition nous a envoyé des clichés e", "la última transmisión que obtuvimos la f", "das letzte was von dieser expedition noc", "as últimas imagens que obtivemos da frot", "последнее что удалось получить от экспед", "naše expedice se nevrátila zpět vědci st", "en kernenedsmeltning i moderskibet førte", "de expeditievloot kon niet terugvliegen ", "az expedíciós flotta nem ugrott vissza a", "notre flotte d`expédition a disparu aprè", "экспедиционный флот не вернулся из прыжк", "ekspedycja nie wykonała skoku powrotnego", "die expeditionsflotte ist nicht mehr aus", "la spedizione non è ritornata dal salto ", "contact with the expedition fleet was su", "el contacto con la flota expedición ha s", "a frota em missão não conseguiu voltar d", "ο στόλος εξερεύνησης δεν επέστρεψε ποτέ ", "la flota expedición no ha retornado al e", "la última transmisión que recibimos la f", "la última cosa que obtuvimos la expedici", "la flota en expedición no saltó vuelta a", "as últimas imagens que tivemos da frota ", "a frota em expedição não conseguiu volta", "den sidste radiotransmission vi modtog f", "ekspeditionsflåden kom ikke tilbage vore", "viimeinen lähetys retkikuntalaivueelta o", "retkikuntalaivue ei ikinä palannut lähis", "η τελευταία εικόνα που λήφθηκε από το στ", "ekspedicijska flota se nije vratila na p", "あなたの探索艦隊からの最後の通信はブラックホールが形成されていく壮大な画像でした", "あなたの艦隊は帰還しませんでした。科学者たちは原因を究明していますが、艦隊は永遠", "flota expeditie nu a sarit inapoi in car", "det sista vi fick ifrån expeditionen var", "expeditionsflottan kom aldrig tillbaka t", "zadnje sporočilo ki smo ga dobili je sli", "kontakt z ekspedicijsko floto je bil nen", "poslednou vecou ktorú sme obdržali od ex", "kontakt s expedičnou flotilou bol náhle ", "kesif filosundan alabildigimiz son bilgi", "kesif filosu  ulastigi bölgeden geri don", "我們從遠征探險艦隊收到了最後傳來的影像那是一個大得嚇人的黑洞", "與遠征探險艦隊的聯繫突然間中斷了我們的科學家們還在努力嘗試重新建立聯繫不過似乎艦"}
					for _, s := range blackhole {
						if strings.Contains(content, s) {
							msg.ExpeditionResult = ogame.ExpeditionResultBlackHole
							break
						}
					}
				}

				msgs = append(msgs, msg)
			}
		}
	})
	return msgs, 0, nil
}

func extractLfBonusesFromDoc(doc *goquery.Document) (ogame.LfBonuses, error) {
	b := ogame.LfBonuses{}

	l := map[string]bool{
		"categoryResources":   true,
		"categoryShips":       true,
		"categoryCostAndTime": true,
		"categoryMisc":        false,
	}
	b.Ships = make(map[ogame.ID]ogame.ShipLfBonus)
	// extract resources bonus and ships cargo bonus for expeditions
	doc.Find("bonus-item-content[data-toggable-target^=category]").Each(func(_ int, s *goquery.Selection) {
		category, _ := s.Attr("data-toggable-target")
		if v, e := l[category]; e {
			if !v {
				return
			}
		} else {
			return
		}
		s.Find("inner-bonus-item-heading[data-toggable^=subcategory]").Each(func(_ int, g *goquery.Selection) {
			category, subcategory := extractCategories(g, category)
			if len(category) > 0 && len(subcategory) > 0 {
				b = assignBonusValue(g, b, category, subcategory)
			}
		})
	})

	return b, nil
}

// assign bonus value directly to LfBonuses struct
func assignBonusValue(g *goquery.Selection, b ogame.LfBonuses, category string, subcategory string) ogame.LfBonuses {
	switch category {
	case "Resources":
		switch subcategory {
		case "0":
			b.Production.Metal = extractBonusFromStringPercentage(extractRawResourcesBonusValue(g))
		case "1":
			b.Production.Crystal = extractBonusFromStringPercentage(extractRawResourcesBonusValue(g))
		case "2":
			b.Production.Deuterium = extractBonusFromStringPercentage(extractRawResourcesBonusValue(g))
		case "Expedition":
			b.Expeditions.Resources = extractBonusFromStringPercentage(extractRawResourcesBonusValue(g))
		case "ExpeditionShipsFound":
			b.Expeditions.Ships = extractBonusFromStringPercentage(extractRawResourcesBonusValue(g))
		case "ExpeditionSpeed":
			b.Expeditions.Speed = extractBonusFromStringPercentage(extractRawResourcesBonusValue(g))
		}
	case "Ships":
		b = extractShipStatBonusNew(g, b, subcategory)
	case "CostAndTime":
		b = extractCostReductionBonus(g, b)
		b = extractTimeReductionBonus(g, b)
	}
	return b
}

// extract raw bonus value for resources category
func extractRawResourcesBonusValue(g *goquery.Selection) string {
	return g.Find(".subCategoryBonus").Text()
}

// extract subcategories from attribute
func extractCategories(g *goquery.Selection, category string) (string, string) {
	c := strings.Replace(category, "category", "", 1)
	s, _ := g.Attr("data-toggable")
	v := "sub" + category
	return c, strings.Replace(s, v, "", 1)
}

// Extracts ogame id from a bonus item
func extractBonusID(g *goquery.Selection) ogame.LfBonusID {
	s, e := g.Attr("data-category")
	if !e {
		return 0
	}
	s = strings.Replace(s, "bonus-", "", 1)
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return ogame.LfBonusID(i)
}

// Extracts time reduction from all subitems
func extractTimeReductionBonus(s *goquery.Selection, l ogame.LfBonuses) ogame.LfBonuses {
	extractAllSubitems(s, func(id ogame.ID, bonuses *goquery.Selection) {
		if id.IsBuilding() {
			var tmp ogame.BaseLfBonus
			_, e := l.Buildings[id]
			if e {
				tmp = l.Buildings[id]
			}
			tmp.Duration = utils.RoundThousandth(tmp.Duration + extractBonusFromRow(bonuses.Eq(0).Text()))
			l.Buildings[id] = tmp
		} else if id.IsLfBuilding() {
			var tmp ogame.BaseLfBonus
			_, e := l.LfBuildings[id]
			if e {
				tmp = l.LfBuildings[id]
			}
			tmp.Duration = utils.RoundThousandth(tmp.Duration + extractBonusFromRow(bonuses.Eq(0).Text()))
			l.LfBuildings[id] = tmp
		} else if id.IsTech() {
			var tmp ogame.BaseLfBonus
			_, e := l.Researches[id]
			if e {
				tmp = l.Researches[id]
			}
			tmp.Duration = utils.RoundThousandth(tmp.Duration + extractBonusFromRow(bonuses.Eq(0).Text()))
			l.Researches[id] = tmp
		} else if id.IsLfTech() {
			var tmp ogame.BaseLfBonus
			_, e := l.LfResearches[id]
			if e {
				tmp = l.LfResearches[id]
			}
			tmp.Duration = utils.RoundThousandth(tmp.Duration + extractBonusFromRow(bonuses.Eq(0).Text()))
			l.LfResearches[id] = tmp
		} else if id.IsShip() {
			var tmp ogame.ShipLfBonus
			_, e := l.Ships[id]
			if e {
				tmp = l.Ships[id]
			}
			tmp.Duration = utils.RoundThousandth(tmp.Duration + extractBonusFromRow(bonuses.Eq(0).Text()))
			l.Ships[id] = tmp
		}
	})
	return l
}

// Extract bonus value from a string with percentage sign [ex: 1.056%]
func extractBonusFromStringPercentage(s string) float64 {
	v := strings.Replace(s, "%", "", 1)
	return extractBonusFromString(v)
}

// Extract bonus value from a string [ex: 1.056]
func extractBonusFromString(s string) float64 {
	v := strings.TrimSpace(s)
	v = strings.Replace(v, ",", ".", 1)
	b, _ := strconv.ParseFloat(v, 64)
	return utils.RoundThousandth(b)
}

// Extract bonus value from a row [ex: 1.056 / 30]
func extractBonusFromRow(r string) float64 {
	if strings.Contains(r, "/") {
		s := strings.Split(r, "/")
		return extractBonusFromString(s[0])
	}
	return 0
}

// Get all subitems in a category
func extractAllSubitems(s *goquery.Selection, clb func(ogame.ID, *goquery.Selection)) {
	s.Next().Find(".subItemContent").Each(func(_ int, g *goquery.Selection) {
		data, e := g.Find(".technology > button.details").Attr("data-target")
		if !e {
			return
		}
		reg := regexp.MustCompile(`technologyId=([0-9]+)`)
		raw := reg.FindStringSubmatch(data)
		if len(raw) != 2 {
			return
		}
		i, err := strconv.Atoi(raw[1])
		if err != nil {
			return
		}
		bonuses := g.Find(".innerSubItemHolder .innerSubItem > div:last-of-type")

		id := ogame.ID(i)
		if !id.IsValid() {
			return
		}

		clb(id, bonuses)
	})
}

// Extracts cost reduction from all subitems
func extractCostReductionBonus(s *goquery.Selection, l ogame.LfBonuses) ogame.LfBonuses {
	extractAllSubitems(s, func(id ogame.ID, bonuses *goquery.Selection) {
		if id.IsBuilding() {
			var tmp ogame.BaseLfBonus
			_, e := l.Buildings[id]
			if e {
				tmp = l.Buildings[id]
			}
			tmp.Cost = utils.RoundThousandth(tmp.Cost + extractBonusFromRow(bonuses.Eq(0).Text()))
			l.Buildings[id] = tmp
		} else if id.IsLfBuilding() {
			var tmp ogame.BaseLfBonus
			_, e := l.LfBuildings[id]
			if e {
				tmp = l.LfBuildings[id]
			}
			tmp.Cost = utils.RoundThousandth(tmp.Cost + extractBonusFromRow(bonuses.Eq(0).Text()))
			l.LfBuildings[id] = tmp
		} else if id.IsTech() {
			var tmp ogame.BaseLfBonus
			_, e := l.Researches[id]
			if e {
				tmp = l.Researches[id]
			}
			tmp.Cost = utils.RoundThousandth(tmp.Cost + extractBonusFromRow(bonuses.Eq(0).Text()))
			l.Researches[id] = tmp
		} else if id.IsLfTech() {
			var tmp ogame.BaseLfBonus
			_, e := l.LfResearches[id]
			if e {
				tmp = l.LfResearches[id]
			}
			tmp.Cost = utils.RoundThousandth(tmp.Cost + extractBonusFromRow(bonuses.Eq(0).Text()))
			l.LfResearches[id] = tmp
		}
	})
	return l
}

// Extracts ships consumption reduction
func extractShipConsumptionBonus(s *goquery.Selection, l ogame.LfBonuses) ogame.LfBonuses {
	extractAllSubitems(s, func(id ogame.ID, bonuses *goquery.Selection) {
		if id.IsShip() {
			_, e := l.Ships[id]
			if !e {
				l.Ships[id] = ogame.ShipLfBonus{}
			}
			tmp := l.Ships[id]
			tmp.Consumption = utils.RoundThousandth(tmp.Consumption + extractBonusFromRow(bonuses.Eq(0).Text()))
			l.Ships[id] = tmp
		}
	})
	return l
}

// Extracts ships stats fixed
func extractShipStatBonusNew(s *goquery.Selection, l ogame.LfBonuses, subcategory string) ogame.LfBonuses {
	i, err := strconv.Atoi(subcategory)
	if err != nil {
		return l
	}
	id := ogame.ID(i)
	if !id.IsValid() {
		return l
	}
	extractAllSubitemsNew(s, func(bonuses *goquery.Selection) {
		if id.IsShip() {
			_, e := l.Ships[id]
			if !e {
				l.Ships[id] = ogame.ShipLfBonus{}
			}
			tmp := l.Ships[id]
			tmp.Armour = tmp.Armour + extractBonusFromStringPercentage(bonuses.Children().Eq(0).Text())
			tmp.Shield = tmp.Shield + extractBonusFromStringPercentage(bonuses.Children().Eq(1).Text())
			tmp.Weapon = tmp.Weapon + extractBonusFromStringPercentage(bonuses.Children().Eq(2).Text())
			tmp.Speed = tmp.Speed + extractBonusFromStringPercentage(bonuses.Children().Eq(3).Text())
			tmp.Cargo = tmp.Cargo + extractBonusFromStringPercentage(bonuses.Children().Eq(4).Text())
			tmp.Consumption = tmp.Consumption + extractBonusFromStringPercentage(bonuses.Eq(5).Text())
			l.Ships[id] = tmp
		}
	})
	return l
}

// Get all bonuses from single ship
func extractAllSubitemsNew(s *goquery.Selection, clb func(*goquery.Selection)) {
	s.Find("bonus-items").Each(func(_ int, g *goquery.Selection) {
		bonuses := g
		clb(bonuses)
	})
}

// Extracts ships stats
func extractShipStatsBonus(s *goquery.Selection, l ogame.LfBonuses) ogame.LfBonuses {
	extractAllSubitems(s, func(id ogame.ID, bonuses *goquery.Selection) {
		if id.IsShip() {
			_, e := l.Ships[id]
			if !e {
				l.Ships[id] = ogame.ShipLfBonus{}
			}
			tmp := l.Ships[id]
			tmp.Armour = utils.RoundThousandth(tmp.Armour + extractBonusFromString(bonuses.Eq(0).Text()))
			tmp.Shield = utils.RoundThousandth(tmp.Shield + extractBonusFromString(bonuses.Eq(1).Text()))
			tmp.Weapon = utils.RoundThousandth(tmp.Weapon + extractBonusFromString(bonuses.Eq(2).Text()))
			tmp.Cargo = utils.RoundThousandth(tmp.Cargo + extractBonusFromString(bonuses.Eq(3).Text()))
			tmp.Speed = utils.RoundThousandth(tmp.Speed + extractBonusFromString(bonuses.Eq(4).Text()))
			l.Ships[id] = tmp
		} else if id.IsDefense() {
			_, e := l.Defenses[id]
			if !e {
				l.Defenses[id] = ogame.ShipLfBonus{}
			}
			tmp := l.Defenses[id]
			tmp.Armour = utils.RoundThousandth(tmp.Armour + extractBonusFromString(bonuses.Eq(0).Text()))
			tmp.Shield = utils.RoundThousandth(tmp.Shield + extractBonusFromString(bonuses.Eq(1).Text()))
			tmp.Weapon = utils.RoundThousandth(tmp.Weapon + extractBonusFromString(bonuses.Eq(2).Text()))
			l.Defenses[id] = tmp
		}
	})
	return l
}

func extractAllianceClassFromDoc(doc *goquery.Document) ogame.AllianceClass {
	allianceClass := ogame.NoAllianceClass
	el := doc.Find("div.allianceclass").First()
	if el.HasClass("warrior") {
		allianceClass = ogame.Warrior
	} else if el.HasClass("trader") {
		allianceClass = ogame.Trader
	} else if el.HasClass("explorer") {
		allianceClass = ogame.Researcher
	}
	return allianceClass
}
