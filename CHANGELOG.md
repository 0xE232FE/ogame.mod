# CHANGELOG

## v13-fix.3

### OGame v13 compatibility

#### Lifeform bonuses
- Refactored v13 lifeform bonus parsing into a dedicated `v13_0_0` extractor (`lfBonuses.go`, `trader.go`, `movement.go`).
- The extractor is now selected automatically from the server version (`serverData.xml` -> `getExtractorFor`): `>= 13.0.0` -> `v13_0_0`, else `v12_0_0`, and so on down the extractor chain.
- The v13 page reworked bonus categories (`data-toggable="ships"` instead of `categoryShips`) - ships/researches are now keyed directly by their ID.

#### Fleet recall (fleetsave / sleep mode)
- v13 changed the recall mechanism: the recall button now uses `data-recall-action` (`action=recallFleet&fleetId=...&asJson=1`) plus a CSRF token from the page script.
- `cancelFleet` now issues the correct v13 request (legacy v12 path preserved).
- `CancelFleetHandler` now returns a real error (500) when the recall fails, instead of masking it as a success.

#### Buy Offer of the Day
- v13 moved Import/Export to `component=trader&action=importexport` (JSON envelope `content.trader`).
- Uses the v13 actions `importExportTrade` / `importExportTakeItem` (instead of `trade` / `takeItem`).
- Trade token taken from the response envelope (`newAjaxToken`).

#### Hostile fleet / attack detection (Defender)
- v13 loads the event list via `component=eventlist&action=catchEvents` (JSON envelope `content.eventlist`).
- `getAttacks` now parses the v13 event list - the Defender detects hostile fleets again.

#### Fleet return time (Expeditions timer)
- v13 renders the return countdown as `getElementById("timerNext_..."), N` instead of `getElementByIdWithCache("timerNext_...", N)`.
- `BackIn` is now extracted correctly - the Expedition worker waits for the fleet return instead of re-checking slots every ~2 minutes.
