package wrapper

import "github.com/alaingilbert/ogame/pkg/ogame"

// BuyItem ...
func (b *Prioritize) BuyItem(ref string, celestialID ogame.CelestialID) error {
	b.begin("BuyItem")
	defer b.done()
	return b.bot.buyItem(ref, celestialID)
}

// NinjaSendFleet ...
func (b *Prioritize) NinjaSendFleet(celestialID ogame.CelestialID, ships []ogame.Quantifiable, speed ogame.Speed, where ogame.Coordinate,
	mission ogame.MissionID, resources ogame.Resources, holdingTime, unionID int64, ensure bool) (ogame.Fleet, error) {
	b.begin("NinjaSendFleet")
	defer b.done()
	return b.bot.ninjaSendFleet(celestialID, ships, speed, where, mission, resources, holdingTime, unionID, ensure)
}

// NjaCancelFleet ...
func (b *Prioritize) NjaCancelFleet(fleetID ogame.FleetID) error {
	b.begin("NjaCancelFleet")
	defer b.done()
	return b.bot.njaCancelFleet(fleetID)
}

// TradeScraper ...
func (b *Prioritize) TradeScraper(ships ogame.ShipsInfos, opts ...Option) error {
	b.begin("TradeScraper")
	defer b.done()
	return b.bot.tradeScraper(ships, opts...)
}

// GetMessages ...
func (b *Prioritize) GetMessages() ([]ogame.Message, error) {
	b.begin("GetMessages")
	defer b.done()
	return b.bot.getMessages()
}

// FlightTime calculate flight time and fuel needed
func (b *Prioritize) FlightTime2(origin, destination ogame.Coordinate, speed ogame.Speed, ships ogame.ShipsInfos, missionID ogame.MissionID, holdingTime int64) (secs, fuel int64) {
	b.begin("FlightTime")
	defer b.done()
	researches := b.bot.getCachedResearch()
	lfbonuses, _ := b.bot.getCachedLfBonuses()
	return CalcFlightTime2(origin, destination, b.bot.serverData.Galaxies, b.bot.serverData.Systems,
		b.bot.serverData.DonutGalaxy, b.bot.serverData.DonutSystem, b.bot.serverData.GlobalDeuteriumSaveFactor,
		float64(speed)/10, GetFleetSpeedForMission(b.bot.serverData, missionID), ships, researches, lfbonuses, b.bot.characterClass, holdingTime)
}
