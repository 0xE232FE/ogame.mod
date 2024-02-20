package ogame

import (
	"math"
	stdmath "math"
	"time"
)

func (s ShipsInfos) Div(d int64) ShipsInfos {
	var res = ShipsInfos{}
	if d > 0 {
		res.Battlecruiser = int64(math.Floor(float64(s.Battlecruiser) / float64(d)))
		res.Battleship = int64(math.Floor(float64(s.Battleship) / float64(d)))
		res.Bomber = int64(math.Floor(float64(s.Bomber) / float64(d)))
		res.ColonyShip = int64(math.Floor(float64(s.ColonyShip) / float64(d)))
		res.Cruiser = int64(math.Floor(float64(s.Cruiser) / float64(d)))
		res.Deathstar = int64(math.Floor(float64(s.Deathstar) / float64(d)))
		res.Destroyer = int64(math.Floor(float64(s.Destroyer) / float64(d)))
		res.EspionageProbe = int64(math.Floor(float64(s.EspionageProbe) / float64(d)))
		res.HeavyFighter = int64(math.Floor(float64(s.HeavyFighter) / float64(d)))
		res.LargeCargo = int64(math.Floor(float64(s.LargeCargo) / float64(d)))
		res.SmallCargo = int64(math.Floor(float64(s.SmallCargo) / float64(d)))
		res.LightFighter = int64(math.Floor(float64(s.LightFighter) / float64(d)))
		res.Pathfinder = int64(math.Floor(float64(s.Pathfinder) / float64(d)))
		res.Reaper = int64(math.Floor(float64(s.Reaper) / float64(d)))
		res.Recycler = int64(math.Floor(float64(s.Recycler) / float64(d)))

		res.SolarSatellite = int64(math.Floor(float64(s.SolarSatellite) / float64(d)))
		res.Crawler = int64(math.Floor(float64(s.Crawler) / float64(d)))
	}
	return res
}

func (r *ResourcesDetails) AvailableResIn(res Resources) (time.Duration, string, int64, int64) {

	var metalSec time.Duration
	neededMetal := r.Metal.Available - res.Metal
	metalProduction := r.Metal.CurrentProduction
	if neededMetal <= 0 && r.Metal.CurrentProduction != 0 {
		neededMetal = (-1) * neededMetal
		metalSec = time.Duration(stdmath.Ceil((float64(neededMetal)/float64(r.Metal.CurrentProduction))*float64(3600))) * time.Second
	}

	var crystalSec time.Duration
	neededCrystal := r.Crystal.Available - res.Crystal
	crystalProduction := r.Crystal.CurrentProduction
	if neededCrystal <= 0 && r.Crystal.CurrentProduction != 0 {
		neededCrystal = (-1) * neededCrystal
		crystalSec = time.Duration(stdmath.Ceil((float64(neededCrystal)/float64(r.Crystal.CurrentProduction))*float64(3600))) * time.Second
	}

	var deuteriumSec time.Duration
	neededDeuterium := r.Deuterium.Available - res.Deuterium
	deuteriumProduction := r.Deuterium.CurrentProduction
	if neededDeuterium <= 0 && r.Deuterium.CurrentProduction != 0 {
		neededDeuterium = (-1) * neededDeuterium
		deuteriumSec = time.Duration(stdmath.Ceil((float64(neededDeuterium)/float64(r.Deuterium.CurrentProduction))*float64(3600))) * time.Second
	}

	var maxDuration time.Duration
	maxDuration = metalSec
	maxResourcesName := "Metal"
	maxNeededResources := neededMetal
	maxProduction := metalProduction
	if crystalSec.Seconds() > maxDuration.Seconds() {
		maxDuration = crystalSec
		maxResourcesName = "Crystal"
		maxNeededResources = neededCrystal
		maxProduction = crystalProduction
	}
	if deuteriumSec.Seconds() > maxDuration.Seconds() {
		maxDuration = deuteriumSec
		maxResourcesName = "Deuterium"
		maxNeededResources = neededDeuterium
		maxProduction = deuteriumProduction
	}

	return maxDuration, maxResourcesName, maxNeededResources, maxProduction
}

// AvailableIn returns the resources available
func (r ResourcesDetails) AvailableIn(secs float64) Resources {
	var res Resources

	if r.Metal.CurrentProduction != 0 {
		prodSec := float64(r.Metal.CurrentProduction) / 3600
		storageLeft := float64(r.Metal.StorageCapacity - r.Metal.Available)
		var storageFullInSec float64
		if prodSec != 0 {
			storageFullInSec = stdmath.Floor(storageLeft / prodSec)
		}
		if storageFullInSec <= float64(secs) {
			res.Metal = r.Metal.StorageCapacity
		} else {
			res.Metal += r.Metal.Available + int64(prodSec*float64(secs))
		}
	} else {
		// If Production is 0 the Storage is full
		res.Metal = r.Metal.Available
	}

	if r.Crystal.CurrentProduction != 0 {

		prodSec := float64(r.Crystal.CurrentProduction) / 3600
		storageLeft := float64(r.Crystal.StorageCapacity - r.Crystal.Available)
		var storageFullInSec float64
		if prodSec != 0 {
			storageFullInSec = stdmath.Floor(storageLeft / prodSec)
		}
		if storageFullInSec <= float64(secs) {
			res.Crystal = r.Crystal.StorageCapacity
		} else {
			res.Crystal += r.Crystal.Available + int64(prodSec*float64(secs))
		}
	} else {
		// If Production is 0 the Storage is full
		res.Crystal = r.Crystal.Available
	}

	if r.Deuterium.CurrentProduction != 0 {

		prodSec := float64(r.Deuterium.CurrentProduction) / 3600
		storageLeft := float64(r.Deuterium.StorageCapacity - r.Deuterium.Available)
		var storageFullInSec float64
		if prodSec != 0 {
			storageFullInSec = stdmath.Floor(storageLeft / prodSec)
		}
		if storageFullInSec <= float64(secs) {
			res.Deuterium = r.Deuterium.StorageCapacity
		} else {
			res.Deuterium += r.Deuterium.Available + int64(prodSec*float64(secs))
		}
	} else {
		// If Production is 0 the Storage is full
		res.Deuterium = r.Deuterium.Available
	}
	res.Energy = r.Energy.Available
	res.Darkmatter = r.Darkmatter.Available

	return res
}
