package ogame

import (
	"time"
)

// Fleet represent a player fleet information
type Fleet struct {
	Mission        MissionID
	ReturnFlight   bool
	InDeepSpace    bool
	ID             FleetID
	Resources      Resources
	Origin         Coordinate
	Destination    Coordinate
	Ships          ShipsInfos
	StartTime      time.Time
	ArrivalTime    time.Time
	BackTime       time.Time
	ArriveIn       int64
	BackIn         int64
	UnionID        int64
	TargetPlanetID int64
}

// MakeFleet make a new Fleet object
func MakeFleet() Fleet {
	return Fleet{}
}

type PhalanxFleet struct {
	Fleet
	BaseSpeed int64
}
