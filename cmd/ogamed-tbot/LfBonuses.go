package main

import "github.com/alaingilbert/ogame/pkg/ogame"

type LfBonuses struct {
	LfResourceBonuses
	LfShipBonuses   LfShipBonuses
	CostTimeBonuses CostTimeBonuses

	// Following lifeform buildings decreases the costs and duration for researching new technologies.
	// Humans ResearchCentre / Rocktal RuneTechnologium / Mechas RoboticsResearchCentre / Kaelesh VortexChamber
	PlanetLfResearchCostTimeBonus CostTimeBonus
}

func NewLfBonuses() *LfBonuses {
	return &LfBonuses{
		LfShipBonuses:   make(LfShipBonuses),
		CostTimeBonuses: make(CostTimeBonuses),
	}
}

type CostTimeBonuses map[ogame.ID]CostTimeBonus

type CostTimeBonus struct {
	Cost     float64
	Duration float64
}

type LfShipBonuses map[ogame.ID]LfShipBonus

type LfShipBonus struct {
	ID                  ogame.ID
	StructuralIntegrity float64
	ShieldPower         float64
	WeaponPower         float64
	Speed               float64
	CargoCapacity       float64
	FuelConsumption     float64
}

type LfResourceBonuses struct {
}
