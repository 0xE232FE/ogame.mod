package ogame

import (
	"math"
	"time"
)

// LazyLfResearches ...
type LazyLfResearches func() LfResearches

type LfResearches struct {
	IntergalacticEnvoys               int64 // Humans techs
	HighPerformanceExtractors         int64
	FusionDrives                      int64
	StealthFieldGenerator             int64
	OrbitalDen                        int64
	ResearchAI                        int64
	HighPerformanceTerraformer        int64
	EnhancedProductionTechnologies    int64
	LightFighterMkII                  int64
	CruiserMkII                       int64
	ImprovedLabTechnology             int64
	PlasmaTerraformer                 int64
	LowTemperatureDrives              int64
	BomberMkII                        int64
	DestroyerMkII                     int64
	BattlecruiserMkII                 int64
	RobotAssistants                   int64
	Supercomputer                     int64
	VolcanicBatteries                 int64 // Rocktal techs
	AcousticScanning                  int64
	HighEnergyPumpSystems             int64
	CargoHoldExpansionCivilianShips   int64
	MagmaPoweredProduction            int64
	GeothermalPowerPlants             int64
	DepthSounding                     int64
	IonCrystalEnhancementHeavyFighter int64
	ImprovedStellarator               int64
	HardenedDiamondDrillHeads         int64
	SeismicMiningTechnology           int64
	MagmaPoweredPumpSystems           int64
	IonCrystalModules                 int64
	OptimisedSiloConstructionMethod   int64
	DiamondEnergyTransmitter          int64
	ObsidianShieldReinforcement       int64
	RuneShields                       int64
	RocktalCollectorEnhancement       int64
	CatalyserTechnology               int64 // Mechas techs
	PlasmaDrive                       int64
	EfficiencyModule                  int64
	DepotAI                           int64
	GeneralOverhaulLightFighter       int64
	AutomatedTransportLines           int64
	ImprovedDroneAI                   int64
	ExperimentalRecyclingTechnology   int64
	GeneralOverhaulCruiser            int64
	SlingshotAutopilot                int64
	HighTemperatureSuperconductors    int64
	GeneralOverhaulBattleship         int64
	ArtificialSwarmIntelligence       int64
	GeneralOverhaulBattlecruiser      int64
	GeneralOverhaulBomber             int64
	GeneralOverhaulDestroyer          int64
	ExperimentalWeaponsTechnology     int64
	MechanGeneralEnhancement          int64
	HeatRecovery                      int64 // Kaelesh techs
	SulphideProcess                   int64
	PsionicNetwork                    int64
	TelekineticTractorBeam            int64
	EnhancedSensorTechnology          int64
	NeuromodalCompressor              int64
	NeuroInterface                    int64
	InterplanetaryAnalysisNetwork     int64
	OverclockingHeavyFighter          int64
	TelekineticDrive                  int64
	SixthSense                        int64
	Psychoharmoniser                  int64
	EfficientSwarmIntelligence        int64
	OverclockingLargeCargo            int64
	GravitationSensors                int64
	OverclockingBattleship            int64
	PsionicShieldMatrix               int64
	KaeleshDiscovererEnhancement      int64
}

func (b LfResearches) Lazy() LazyLfResearches {
	return func() LfResearches { return b }
}

// ByID gets the research level by lfResearch id
func (b LfResearches) ByID(id ID) int64 {
	switch id {
	case IntergalacticEnvoysID:
		return b.IntergalacticEnvoys
	case HighPerformanceExtractorsID:
		return b.HighPerformanceExtractors
	case FusionDrivesID:
		return b.FusionDrives
	case StealthFieldGeneratorID:
		return b.StealthFieldGenerator
	case OrbitalDenID:
		return b.OrbitalDen
	case ResearchAIID:
		return b.ResearchAI
	case HighPerformanceTerraformerID:
		return b.HighPerformanceTerraformer
	case EnhancedProductionTechnologiesID:
		return b.EnhancedProductionTechnologies
	case LightFighterMkIIID:
		return b.LightFighterMkII
	case CruiserMkIIID:
		return b.CruiserMkII
	case ImprovedLabTechnologyID:
		return b.ImprovedLabTechnology
	case PlasmaTerraformerID:
		return b.PlasmaTerraformer
	case LowTemperatureDrivesID:
		return b.LowTemperatureDrives
	case BomberMkIIID:
		return b.BomberMkII
	case DestroyerMkIIID:
		return b.DestroyerMkII
	case BattlecruiserMkIIID:
		return b.BattlecruiserMkII
	case RobotAssistantsID:
		return b.RobotAssistants
	case SupercomputerID:
		return b.Supercomputer
	case VolcanicBatteriesID:
		return b.VolcanicBatteries
	case AcousticScanningID:
		return b.AcousticScanning
	case HighEnergyPumpSystemsID:
		return b.HighEnergyPumpSystems
	case CargoHoldExpansionCivilianShipsID:
		return b.CargoHoldExpansionCivilianShips
	case MagmaPoweredProductionID:
		return b.MagmaPoweredProduction
	case GeothermalPowerPlantsID:
		return b.GeothermalPowerPlants
	case DepthSoundingID:
		return b.DepthSounding
	case IonCrystalEnhancementHeavyFighterID:
		return b.IonCrystalEnhancementHeavyFighter
	case ImprovedStellaratorID:
		return b.ImprovedStellarator
	case HardenedDiamondDrillHeadsID:
		return b.HardenedDiamondDrillHeads
	case SeismicMiningTechnologyID:
		return b.SeismicMiningTechnology
	case MagmaPoweredPumpSystemsID:
		return b.MagmaPoweredPumpSystems
	case IonCrystalModulesID:
		return b.IonCrystalModules
	case OptimisedSiloConstructionMethodID:
		return b.OptimisedSiloConstructionMethod
	case DiamondEnergyTransmitterID:
		return b.DiamondEnergyTransmitter
	case ObsidianShieldReinforcementID:
		return b.ObsidianShieldReinforcement
	case RuneShieldsID:
		return b.RuneShields
	case RocktalCollectorEnhancementID:
		return b.RocktalCollectorEnhancement
	case CatalyserTechnologyID:
		return b.CatalyserTechnology
	case PlasmaDriveID:
		return b.PlasmaDrive
	case EfficiencyModuleID:
		return b.EfficiencyModule
	case DepotAIID:
		return b.DepotAI
	case GeneralOverhaulLightFighterID:
		return b.GeneralOverhaulLightFighter
	case AutomatedTransportLinesID:
		return b.AutomatedTransportLines
	case ImprovedDroneAIID:
		return b.ImprovedDroneAI
	case ExperimentalRecyclingTechnologyID:
		return b.ExperimentalRecyclingTechnology
	case GeneralOverhaulCruiserID:
		return b.GeneralOverhaulCruiser
	case SlingshotAutopilotID:
		return b.SlingshotAutopilot
	case HighTemperatureSuperconductorsID:
		return b.HighTemperatureSuperconductors
	case GeneralOverhaulBattleshipID:
		return b.GeneralOverhaulBattleship
	case ArtificialSwarmIntelligenceID:
		return b.ArtificialSwarmIntelligence
	case GeneralOverhaulBattlecruiserID:
		return b.GeneralOverhaulBattlecruiser
	case GeneralOverhaulBomberID:
		return b.GeneralOverhaulBomber
	case GeneralOverhaulDestroyerID:
		return b.GeneralOverhaulDestroyer
	case ExperimentalWeaponsTechnologyID:
		return b.ExperimentalWeaponsTechnology
	case MechanGeneralEnhancementID:
		return b.MechanGeneralEnhancement
	case HeatRecoveryID:
		return b.HeatRecovery
	case SulphideProcessID:
		return b.SulphideProcess
	case PsionicNetworkID:
		return b.PsionicNetwork
	case TelekineticTractorBeamID:
		return b.TelekineticTractorBeam
	case EnhancedSensorTechnologyID:
		return b.EnhancedSensorTechnology
	case NeuromodalCompressorID:
		return b.NeuromodalCompressor
	case NeuroInterfaceID:
		return b.NeuroInterface
	case InterplanetaryAnalysisNetworkID:
		return b.InterplanetaryAnalysisNetwork
	case OverclockingHeavyFighterID:
		return b.OverclockingHeavyFighter
	case TelekineticDriveID:
		return b.TelekineticDrive
	case SixthSenseID:
		return b.SixthSense
	case PsychoharmoniserID:
		return b.Psychoharmoniser
	case EfficientSwarmIntelligenceID:
		return b.EfficientSwarmIntelligence
	case OverclockingLargeCargoID:
		return b.OverclockingLargeCargo
	case GravitationSensorsID:
		return b.GravitationSensors
	case OverclockingBattleshipID:
		return b.OverclockingBattleship
	case PsionicShieldMatrixID:
		return b.PsionicShieldMatrix
	case KaeleshDiscovererEnhancementID:
		return b.KaeleshDiscovererEnhancement
	}
	return 0
}

// BaseLfResearch base struct for Lifeform techs
type BaseLfResearch struct {
	BaseTechnology
}

// ConstructionTimeWithBonuses returns duration with LfBonuses applied
func (b BaseLfResearch) ConstructionTimeWithBonuses(level, universeSpeed int64, facilities BuildAccelerators, hasTechnocrat bool, class CharacterClass, bonuses LfBonuses) time.Duration {
	duration := b.ConstructionTime(level, universeSpeed, facilities, hasTechnocrat, class)
	bonus := bonuses.ByLfTechID(b.ID).Duration
	return time.Duration(float64(duration) - float64(duration)*bonus)
}

// GetPrice returns the price to build the given level
func (b BaseLfResearch) GetPrice(level int64) Resources {
	tmp := func(baseCost int64, increaseFactor float64, level int64) int64 {
		return int64(float64(baseCost) * math.Pow(increaseFactor, float64(level-1)) * float64(level))
	}
	return Resources{
		Metal:     tmp(b.BaseCost.Metal, b.IncreaseFactor, level),
		Crystal:   tmp(b.BaseCost.Crystal, b.IncreaseFactor, level),
		Deuterium: tmp(b.BaseCost.Deuterium, b.IncreaseFactor, level),
	}
}

// GetPriceWithBonus return the price with LfBonuses applied
func (b BaseLfResearch) GetPriceWithBonuses(level int64, bonuses LfBonuses) Resources {
	price := b.GetPrice(level)
	bonus := bonuses.ByLfTechID(b.ID).Cost
	return price.SubPercent(bonus)
}

// Humans
type intergalacticEnvoys struct {
	BaseLfResearch
}

func newIntergalacticEnvoys() *intergalacticEnvoys {
	b := new(intergalacticEnvoys)
	b.Name = "IntergalacticEnvoys"
	b.ID = IntergalacticEnvoysID
	b.IncreaseFactor = 1.30
	b.BaseCost = Resources{Metal: 5000, Crystal: 2500, Deuterium: 500}
	b.Requirements = map[ID]int64{}
	return b
}

type highPerformanceExtractors struct {
	BaseLfResearch
}

func newHighPerformanceExtractors() *highPerformanceExtractors {
	b := new(highPerformanceExtractors)
	b.Name = "HighPerformanceExtractors"
	b.ID = HighPerformanceExtractorsID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 7000, Crystal: 10000, Deuterium: 5000}
	b.Requirements = map[ID]int64{}
	return b
}

type fusionDrives struct {
	BaseLfResearch
}

func newFusionDrives() *fusionDrives {
	b := new(fusionDrives)
	b.Name = "FusionDrives"
	b.ID = FusionDrivesID
	b.IncreaseFactor = 1.30
	b.BaseCost = Resources{Metal: 15000, Crystal: 10000, Deuterium: 5000}
	b.Requirements = map[ID]int64{}
	return b
}

type stealthFieldGenerator struct {
	BaseLfResearch
}

func newStealthFieldGenerator() *stealthFieldGenerator {
	b := new(stealthFieldGenerator)
	b.Name = "StealthFieldGenerator"
	b.ID = StealthFieldGeneratorID
	b.IncreaseFactor = 1.30
	b.BaseCost = Resources{Metal: 20000, Crystal: 15000, Deuterium: 7500}
	b.Requirements = map[ID]int64{}
	return b
}

type orbitalDen struct {
	BaseLfResearch
}

func newOrbitalDen() *orbitalDen {
	b := new(orbitalDen)
	b.Name = "OrbitalDen"
	b.ID = OrbitalDenID
	b.IncreaseFactor = 1.40
	b.BaseCost = Resources{Metal: 25000, Crystal: 20000, Deuterium: 10000}
	b.Requirements = map[ID]int64{}
	return b
}

type researchAI struct {
	BaseLfResearch
}

func newResearchAI() *researchAI {
	b := new(researchAI)
	b.Name = "ResearchAI"
	b.ID = ResearchAIID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 35000, Crystal: 25000, Deuterium: 15000}
	b.Requirements = map[ID]int64{}
	return b
}

type highPerformanceTerraformer struct {
	BaseLfResearch
}

func newHighPerformanceTerraformer() *highPerformanceTerraformer {
	b := new(highPerformanceTerraformer)
	b.Name = "HighPerformanceTerraformer"
	b.ID = HighPerformanceTerraformerID
	b.IncreaseFactor = 1.30
	b.BaseCost = Resources{Metal: 70000, Crystal: 40000, Deuterium: 20000}
	b.Requirements = map[ID]int64{}
	return b
}

type enhancedProductionTechnologies struct {
	BaseLfResearch
}

func newEnhancedProductionTechnologies() *enhancedProductionTechnologies {
	b := new(enhancedProductionTechnologies)
	b.Name = "EnhancedProductionTechnologies"
	b.ID = EnhancedProductionTechnologiesID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 80000, Crystal: 50000, Deuterium: 20000}
	b.Requirements = map[ID]int64{}
	return b
}

type lightFighterMkII struct {
	BaseLfResearch
}

func newLightFighterMkII() *lightFighterMkII {
	b := new(lightFighterMkII)
	b.Name = "LightFighterMkII"
	b.ID = LightFighterMkIIID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 320000, Crystal: 240000, Deuterium: 100000}
	b.Requirements = map[ID]int64{}
	return b
}

type cruiserMkII struct {
	BaseLfResearch
}

func newCruiserMkII() *cruiserMkII {
	b := new(cruiserMkII)
	b.Name = "CruiserMkII"
	b.ID = CruiserMkIIID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 320000, Crystal: 240000, Deuterium: 100000}
	b.Requirements = map[ID]int64{}
	return b
}

type improvedLabTechnology struct {
	BaseLfResearch
}

func newImprovedLabTechnology() *improvedLabTechnology {
	b := new(improvedLabTechnology)
	b.Name = "ImprovedLabTechnology"
	b.ID = ImprovedLabTechnologyID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 120000, Crystal: 30000, Deuterium: 25000}
	b.Requirements = map[ID]int64{}
	return b
}

type plasmaTerraformer struct {
	BaseLfResearch
}

func newPlasmaTerraformer() *plasmaTerraformer {
	b := new(plasmaTerraformer)
	b.Name = "PlasmaTerraformer"
	b.ID = PlasmaTerraformerID
	b.IncreaseFactor = 1.30
	b.BaseCost = Resources{Metal: 100000, Crystal: 40000, Deuterium: 30000}
	b.Requirements = map[ID]int64{}
	return b
}

type lowTemperatureDrives struct {
	BaseLfResearch
}

func newLowTemperatureDrives() *lowTemperatureDrives {
	b := new(lowTemperatureDrives)
	b.Name = "LowTemperatureDrives"
	b.ID = LowTemperatureDrivesID
	b.IncreaseFactor = 1.30
	b.BaseCost = Resources{Metal: 200000, Crystal: 100000, Deuterium: 100000}
	b.Requirements = map[ID]int64{}
	return b
}

type bomberMkII struct {
	BaseLfResearch
}

func newBomberMkII() *bomberMkII {
	b := new(bomberMkII)
	b.Name = "BomberMkII"
	b.ID = BomberMkIIID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 160000, Crystal: 120000, Deuterium: 50000}
	b.Requirements = map[ID]int64{}
	return b
}

type destroyerMkII struct {
	BaseLfResearch
}

func newDestroyerMkII() *destroyerMkII {
	b := new(destroyerMkII)
	b.Name = "DestroyerMkII"
	b.ID = DestroyerMkIIID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 160000, Crystal: 120000, Deuterium: 50000}
	b.Requirements = map[ID]int64{}
	return b
}

type battlecruiserMkII struct {
	BaseLfResearch
}

func newBattlecruiserMkII() *battlecruiserMkII {
	b := new(battlecruiserMkII)
	b.Name = "BattlecruiserMkII"
	b.ID = BattlecruiserMkIIID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 320000, Crystal: 240000, Deuterium: 100000}
	b.Requirements = map[ID]int64{}
	return b
}

type robotAssistants struct {
	BaseLfResearch
}

func newRobotAssistants() *robotAssistants {
	b := new(robotAssistants)
	b.Name = "robotAssistants"
	b.ID = RobotAssistantsID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 300000, Crystal: 180000, Deuterium: 120000}
	b.Requirements = map[ID]int64{}
	return b
}

type supercomputer struct {
	BaseLfResearch
}

func newSupercomputer() *supercomputer {
	b := new(supercomputer)
	b.Name = "Supercomputer"
	b.ID = SupercomputerID
	b.IncreaseFactor = 1.30
	b.BaseCost = Resources{Metal: 500000, Crystal: 300000, Deuterium: 200000}
	b.Requirements = map[ID]int64{}
	return b
}

// Rocktal
type volcanicBatteries struct {
	BaseLfResearch
}

func newVolcanicBatteries() *volcanicBatteries {
	b := new(volcanicBatteries)
	b.Name = "VolcanicBatteries"
	b.ID = VolcanicBatteriesID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 10000, Crystal: 6000, Deuterium: 1000}
	b.Requirements = map[ID]int64{}
	return b
}

type acousticScanning struct {
	BaseLfResearch
}

func newAcousticScanning() *acousticScanning {
	b := new(acousticScanning)
	b.Name = "AcousticScanning"
	b.ID = AcousticScanningID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 7500, Crystal: 12500, Deuterium: 5000}
	b.Requirements = map[ID]int64{}
	return b
}

type highEnergyPumpSystems struct {
	BaseLfResearch
}

func newHighEnergyPumpSystems() *highEnergyPumpSystems {
	b := new(highEnergyPumpSystems)
	b.Name = "HighEnergyPumpSystems"
	b.ID = HighEnergyPumpSystemsID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 15000, Crystal: 10000, Deuterium: 5000}
	b.Requirements = map[ID]int64{}
	return b
}

type cargoHoldExpansionCivilianShips struct {
	BaseLfResearch
}

func newCargoHoldExpansionCivilianShips() *cargoHoldExpansionCivilianShips {
	b := new(cargoHoldExpansionCivilianShips)
	b.Name = "CargoHoldExpansionCivilianShips"
	b.ID = CargoHoldExpansionCivilianShipsID
	b.IncreaseFactor = 1.30
	b.BaseCost = Resources{Metal: 20000, Crystal: 15000, Deuterium: 7500}
	b.Requirements = map[ID]int64{}
	return b
}

type magmaPoweredProduction struct {
	BaseLfResearch
}

func newMagmaPoweredProduction() *magmaPoweredProduction {
	b := new(magmaPoweredProduction)
	b.Name = "MagmaPoweredProduction"
	b.ID = MagmaPoweredProductionID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 25000, Crystal: 20000, Deuterium: 10000}
	b.Requirements = map[ID]int64{}
	return b
}

type geothermalPowerPlants struct {
	BaseLfResearch
}

func newGeothermalPowerPlants() *geothermalPowerPlants {
	b := new(geothermalPowerPlants)
	b.Name = "GeothermalPowerPlants"
	b.ID = GeothermalPowerPlantsID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 50000, Crystal: 50000, Deuterium: 20000}
	b.Requirements = map[ID]int64{}
	return b
}

type depthSounding struct {
	BaseLfResearch
}

func newDepthSounding() *depthSounding {
	b := new(depthSounding)
	b.Name = "DepthSounding"
	b.ID = DepthSoundingID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 70000, Crystal: 40000, Deuterium: 20000}
	b.Requirements = map[ID]int64{}
	return b
}

type ionCrystalEnhancementHeavyFighter struct {
	BaseLfResearch
}

func newIonCrystalEnhancementHeavyFighter() *ionCrystalEnhancementHeavyFighter {
	b := new(ionCrystalEnhancementHeavyFighter)
	b.Name = "IonCrystalEnhancementHeavyFighter"
	b.ID = IonCrystalEnhancementHeavyFighterID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 160000, Crystal: 120000, Deuterium: 50000}
	b.Requirements = map[ID]int64{}
	return b
}

type improvedStellarator struct {
	BaseLfResearch
}

func newImprovedStellarator() *improvedStellarator {
	b := new(improvedStellarator)
	b.Name = "ImprovedStellarator"
	b.ID = ImprovedStellaratorID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 75000, Crystal: 55000, Deuterium: 25000}
	b.Requirements = map[ID]int64{}
	return b
}

type hardenedDiamondDrillHeads struct {
	BaseLfResearch
}

func newHardenedDiamondDrillHeads() *hardenedDiamondDrillHeads {
	b := new(hardenedDiamondDrillHeads)
	b.Name = "HardenedDiamondDrillHeads"
	b.ID = HardenedDiamondDrillHeadsID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 85000, Crystal: 40000, Deuterium: 35000}
	b.Requirements = map[ID]int64{}
	return b
}

type seismicMiningTechnology struct {
	BaseLfResearch
}

func newSeismicMiningTechnology() *seismicMiningTechnology {
	b := new(seismicMiningTechnology)
	b.Name = "SeismicMiningTechnology"
	b.ID = SeismicMiningTechnologyID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 120000, Crystal: 30000, Deuterium: 25000}
	b.Requirements = map[ID]int64{}
	return b
}

type magmaPoweredPumpSystems struct {
	BaseLfResearch
}

func newMagmaPoweredPumpSystems() *magmaPoweredPumpSystems {
	b := new(magmaPoweredPumpSystems)
	b.Name = "MagmaPoweredPumpSystems"
	b.ID = MagmaPoweredPumpSystemsID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 100000, Crystal: 40000, Deuterium: 30000}
	b.Requirements = map[ID]int64{}
	return b
}

type ionCrystalModules struct {
	BaseLfResearch
}

func newIonCrystalModules() *ionCrystalModules {
	b := new(ionCrystalModules)
	b.Name = "IonCrystalModules"
	b.ID = IonCrystalModulesID
	b.IncreaseFactor = 1.20
	b.BaseCost = Resources{Metal: 200000, Crystal: 100000, Deuterium: 100000}
	b.Requirements = map[ID]int64{}
	return b
}

type optimisedSiloConstructionMethod struct {
	BaseLfResearch
}

func newOptimisedSiloConstructionMethod() *optimisedSiloConstructionMethod {
	b := new(optimisedSiloConstructionMethod)
	b.Name = "OptimisedSiloConstructionMethod"
	b.ID = OptimisedSiloConstructionMethodID
	b.IncreaseFactor = 1.30
	b.BaseCost = Resources{Metal: 220000, Crystal: 110000, Deuterium: 110000}
	b.Requirements = map[ID]int64{}
	return b
}

type diamondEnergyTransmitter struct {
	BaseLfResearch
}

func newDiamondEnergyTransmitter() *diamondEnergyTransmitter {
	b := new(diamondEnergyTransmitter)
	b.Name = "DiamondEnergyTransmitter"
	b.ID = DiamondEnergyTransmitterID
	b.IncreaseFactor = 1.30
	b.BaseCost = Resources{Metal: 240000, Crystal: 120000, Deuterium: 120000}
	b.Requirements = map[ID]int64{}
	return b
}

type obsidianShieldReinforcement struct {
	BaseLfResearch
}

func newObsidianShieldReinforcement() *obsidianShieldReinforcement {
	b := new(obsidianShieldReinforcement)
	b.Name = "ObsidianShieldReinforcement"
	b.ID = ObsidianShieldReinforcementID
	b.IncreaseFactor = 1.40
	b.BaseCost = Resources{Metal: 250000, Crystal: 250000, Deuterium: 250000}
	b.Requirements = map[ID]int64{}
	return b
}

type runeShields struct {
	BaseLfResearch
}

func newRuneShields() *runeShields {
	b := new(runeShields)
	b.Name = "RuneShields"
	b.ID = RuneShieldsID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 500000, Crystal: 300000, Deuterium: 200000}
	b.Requirements = map[ID]int64{}
	return b
}

type rocktalCollectorEnhancement struct {
	BaseLfResearch
}

func newRocktalCollectorEnhancement() *rocktalCollectorEnhancement {
	b := new(rocktalCollectorEnhancement)
	b.Name = "RocktalCollectorEnhancement"
	b.ID = RocktalCollectorEnhancementID
	b.IncreaseFactor = 1.70
	b.BaseCost = Resources{Metal: 300000, Crystal: 180000, Deuterium: 120000}
	b.Requirements = map[ID]int64{}
	return b
}

//Mechas

type catalyserTechnology struct {
	BaseLfResearch
}

func newCatalyserTechnology() *catalyserTechnology {
	b := new(catalyserTechnology)
	b.Name = "CatalyserTechnology"
	b.ID = CatalyserTechnologyID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 10000, Crystal: 6000, Deuterium: 1000}
	b.Requirements = map[ID]int64{}
	return b
}

type plasmaDrive struct {
	BaseLfResearch
}

func newPlasmaDrive() *plasmaDrive {
	b := new(plasmaDrive)
	b.Name = "PlasmaDrive"
	b.ID = PlasmaDriveID
	b.IncreaseFactor = 1.30
	b.BaseCost = Resources{Metal: 7500, Crystal: 12500, Deuterium: 5000}
	b.Requirements = map[ID]int64{}
	return b
}

type efficiencyModule struct {
	BaseLfResearch
}

func newEfficiencyModule() *efficiencyModule {
	b := new(efficiencyModule)
	b.Name = "EfficiencyModule"
	b.ID = EfficiencyModuleID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 15000, Crystal: 10000, Deuterium: 5000}
	b.Requirements = map[ID]int64{}
	return b
}

type depotAI struct {
	BaseLfResearch
}

func newDepotAI() *depotAI {
	b := new(depotAI)
	b.Name = "DepotAI"
	b.ID = DepotAIID
	b.IncreaseFactor = 1.30
	b.BaseCost = Resources{Metal: 20000, Crystal: 15000, Deuterium: 7500}
	b.Requirements = map[ID]int64{}
	return b
}

type generalOverhaulLightFighter struct {
	BaseLfResearch
}

func newGeneralOverhaulLightFighter() *generalOverhaulLightFighter {
	b := new(generalOverhaulLightFighter)
	b.Name = "GeneralOverhaulLightFighter"
	b.ID = GeneralOverhaulLightFighterID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 160000, Crystal: 120000, Deuterium: 50000}
	b.Requirements = map[ID]int64{}
	return b
}

type automatedTransportLines struct {
	BaseLfResearch
}

func newAutomatedTransportLines() *automatedTransportLines {
	b := new(automatedTransportLines)
	b.Name = "AutomatedTransportLines"
	b.ID = AutomatedTransportLinesID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 50000, Crystal: 50000, Deuterium: 20000}
	b.Requirements = map[ID]int64{}
	return b
}

type improvedDroneAI struct {
	BaseLfResearch
}

func newImprovedDroneAI() *improvedDroneAI {
	b := new(improvedDroneAI)
	b.Name = "ImprovedDroneAI"
	b.ID = ImprovedDroneAIID
	b.IncreaseFactor = 1.30
	b.BaseCost = Resources{Metal: 70000, Crystal: 40000, Deuterium: 20000}
	b.Requirements = map[ID]int64{}
	return b
}

type experimentalRecyclingTechnology struct {
	BaseLfResearch
}

func newExperimentalRecyclingTechnology() *experimentalRecyclingTechnology {
	b := new(experimentalRecyclingTechnology)
	b.Name = "ExperimentalRecyclingTechnology"
	b.ID = ExperimentalRecyclingTechnologyID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 160000, Crystal: 120000, Deuterium: 50000}
	b.Requirements = map[ID]int64{}
	return b
}

type generalOverhaulCruiser struct {
	BaseLfResearch
}

func newGeneralOverhaulCruiser() *generalOverhaulCruiser {
	b := new(generalOverhaulCruiser)
	b.Name = "GeneralOverhaulCruiser"
	b.ID = GeneralOverhaulCruiserID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 160000, Crystal: 120000, Deuterium: 50000}
	b.Requirements = map[ID]int64{}
	return b
}

type slingshotAutopilot struct {
	BaseLfResearch
}

func newSlingshotAutopilot() *slingshotAutopilot {
	b := new(slingshotAutopilot)
	b.Name = "SlingshotAutopilot"
	b.ID = SlingshotAutopilotID
	b.IncreaseFactor = 1.20
	b.BaseCost = Resources{Metal: 85000, Crystal: 40000, Deuterium: 35000}
	b.Requirements = map[ID]int64{}
	return b
}

type highTemperatureSuperconductors struct {
	BaseLfResearch
}

func newHighTemperatureSuperconductors() *highTemperatureSuperconductors {
	b := new(highTemperatureSuperconductors)
	b.Name = "HighTemperatureSuperconductors"
	b.ID = HighTemperatureSuperconductorsID
	b.IncreaseFactor = 1.30
	b.BaseCost = Resources{Metal: 120000, Crystal: 30000, Deuterium: 25000}
	b.Requirements = map[ID]int64{}
	return b
}

type generalOverhaulBattleship struct {
	BaseLfResearch
}

func newGeneralOverhaulBattleship() *generalOverhaulBattleship {
	b := new(generalOverhaulBattleship)
	b.Name = "GeneralOverhaulBattleship"
	b.ID = GeneralOverhaulBattleshipID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 160000, Crystal: 120000, Deuterium: 50000}
	b.Requirements = map[ID]int64{}
	return b
}

type artificialSwarmIntelligence struct {
	BaseLfResearch
}

func newArtificialSwarmIntelligence() *artificialSwarmIntelligence {
	b := new(artificialSwarmIntelligence)
	b.Name = "ArtificialSwarmIntelligence"
	b.ID = ArtificialSwarmIntelligenceID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 200000, Crystal: 100000, Deuterium: 100000}
	b.Requirements = map[ID]int64{}
	return b
}

type generalOverhaulBattlecruiser struct {
	BaseLfResearch
}

func newGeneralOverhaulBattlecruiser() *generalOverhaulBattlecruiser {
	b := new(generalOverhaulBattlecruiser)
	b.Name = "GeneralOverhaulBattlecruiser"
	b.ID = GeneralOverhaulBattlecruiserID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 160000, Crystal: 120000, Deuterium: 50000}
	b.Requirements = map[ID]int64{}
	return b
}

type generalOverhaulBomber struct {
	BaseLfResearch
}

func newGeneralOverhaulBomber() *generalOverhaulBomber {
	b := new(generalOverhaulBomber)
	b.Name = "GeneralOverhaulBomber"
	b.ID = GeneralOverhaulBomberID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 320000, Crystal: 240000, Deuterium: 100000}
	b.Requirements = map[ID]int64{}
	return b
}

type generalOverhaulDestroyer struct {
	BaseLfResearch
}

func newGeneralOverhaulDestroyer() *generalOverhaulDestroyer {
	b := new(generalOverhaulDestroyer)
	b.Name = "GeneralOverhaulDestroyer"
	b.ID = GeneralOverhaulDestroyerID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 320000, Crystal: 240000, Deuterium: 100000}
	b.Requirements = map[ID]int64{}
	return b
}

type experimentalWeaponsTechnology struct {
	BaseLfResearch
}

func newExperimentalWeaponsTechnology() *experimentalWeaponsTechnology {
	b := new(experimentalWeaponsTechnology)
	b.Name = "ExperimentalWeaponsTechnology"
	b.ID = ExperimentalWeaponsTechnologyID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 500000, Crystal: 300000, Deuterium: 200000}
	b.Requirements = map[ID]int64{}
	return b
}

type mechanGeneralEnhancement struct {
	BaseLfResearch
}

func newMechanGeneralEnhancement() *mechanGeneralEnhancement {
	b := new(mechanGeneralEnhancement)
	b.Name = "MechanGeneralEnhancement"
	b.ID = MechanGeneralEnhancementID
	b.IncreaseFactor = 1.70
	b.BaseCost = Resources{Metal: 300000, Crystal: 180000, Deuterium: 120000}
	b.Requirements = map[ID]int64{}
	return b
}

// Kaelesh
type heatRecovery struct {
	BaseLfResearch
}

func newHeatRecovery() *heatRecovery {
	b := new(heatRecovery)
	b.Name = "HeatRecovery"
	b.ID = HeatRecoveryID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 10000, Crystal: 6000, Deuterium: 1000}
	b.Requirements = map[ID]int64{}
	return b
}

type sulphideProcess struct {
	BaseLfResearch
}

func newSulphideProcess() *sulphideProcess {
	b := new(sulphideProcess)
	b.Name = "SulphideProcess"
	b.ID = SulphideProcessID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 7500, Crystal: 12500, Deuterium: 5000}
	b.Requirements = map[ID]int64{}
	return b
}

type psionicNetwork struct {
	BaseLfResearch
}

func newPsionicNetwork() *psionicNetwork {
	b := new(psionicNetwork)
	b.Name = "PsionicNetwork"
	b.ID = PsionicNetworkID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 15000, Crystal: 10000, Deuterium: 5000}
	b.Requirements = map[ID]int64{}
	return b
}

type telekineticTractorBeam struct {
	BaseLfResearch
}

func newTelekineticTractorBeam() *telekineticTractorBeam {
	b := new(telekineticTractorBeam)
	b.Name = "TelekineticTractorBeam"
	b.ID = TelekineticTractorBeamID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 20000, Crystal: 15000, Deuterium: 7500}
	b.Requirements = map[ID]int64{}
	return b
}

type enhancedSensorTechnology struct {
	BaseLfResearch
}

func newEnhancedSensorTechnology() *enhancedSensorTechnology {
	b := new(enhancedSensorTechnology)
	b.Name = "EnhancedSensorTechnology"
	b.ID = EnhancedSensorTechnologyID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 25000, Crystal: 20000, Deuterium: 10000}
	b.Requirements = map[ID]int64{}
	return b
}

type neuromodalCompressor struct {
	BaseLfResearch
}

func newNeuromodalCompressor() *neuromodalCompressor {
	b := new(neuromodalCompressor)
	b.Name = "NeuromodalCompressor"
	b.ID = NeuromodalCompressorID
	b.IncreaseFactor = 1.30
	b.BaseCost = Resources{Metal: 50000, Crystal: 50000, Deuterium: 20000}
	b.Requirements = map[ID]int64{}
	return b
}

type neuroInterface struct {
	BaseLfResearch
}

func newNeuroInterface() *neuroInterface {
	b := new(neuroInterface)
	b.Name = "NeuroInterface"
	b.ID = NeuroInterfaceID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 70000, Crystal: 40000, Deuterium: 20000}
	b.Requirements = map[ID]int64{}
	return b
}

type interplanetaryAnalysisNetwork struct {
	BaseLfResearch
}

func newInterplanetaryAnalysisNetwork() *interplanetaryAnalysisNetwork {
	b := new(interplanetaryAnalysisNetwork)
	b.Name = "InterplanetaryAnalysisNetwork"
	b.ID = InterplanetaryAnalysisNetworkID
	b.IncreaseFactor = 1.20
	b.BaseCost = Resources{Metal: 80000, Crystal: 50000, Deuterium: 20000}
	b.Requirements = map[ID]int64{}
	return b
}

type overclockingHeavyFighter struct {
	BaseLfResearch
}

func newOverclockingHeavyFighter() *overclockingHeavyFighter {
	b := new(overclockingHeavyFighter)
	b.Name = "OverclockingHeavyFighter"
	b.ID = OverclockingHeavyFighterID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 320000, Crystal: 240000, Deuterium: 100000}
	b.Requirements = map[ID]int64{}
	return b
}

type telekineticDrive struct {
	BaseLfResearch
}

func newTelekineticDrive() *telekineticDrive {
	b := new(telekineticDrive)
	b.Name = "TelekineticDrive"
	b.ID = TelekineticDriveID
	b.IncreaseFactor = 1.20
	b.BaseCost = Resources{Metal: 85000, Crystal: 40000, Deuterium: 35000}
	b.Requirements = map[ID]int64{}
	return b
}

type sixthSense struct {
	BaseLfResearch
}

func newSixthSense() *sixthSense {
	b := new(sixthSense)
	b.Name = "SixthSense"
	b.ID = SixthSenseID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 120000, Crystal: 30000, Deuterium: 25000}
	b.Requirements = map[ID]int64{}
	return b
}

type psychoharmoniser struct {
	BaseLfResearch
}

func newPsychoharmoniser() *psychoharmoniser {
	b := new(psychoharmoniser)
	b.Name = "Psychoharmoniser"
	b.ID = PsychoharmoniserID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 100000, Crystal: 40000, Deuterium: 30000}
	b.Requirements = map[ID]int64{}
	return b
}

type efficientSwarmIntelligence struct {
	BaseLfResearch
}

func newEfficientSwarmIntelligence() *efficientSwarmIntelligence {
	b := new(efficientSwarmIntelligence)
	b.Name = "EfficientSwarmIntelligence"
	b.ID = EfficientSwarmIntelligenceID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 200000, Crystal: 100000, Deuterium: 100000}
	b.Requirements = map[ID]int64{}
	return b
}

type overclockingLargeCargo struct {
	BaseLfResearch
}

func newOverclockingLargeCargo() *overclockingLargeCargo {
	b := new(overclockingLargeCargo)
	b.Name = "OverclockingLargeCargo"
	b.ID = OverclockingLargeCargoID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 160000, Crystal: 120000, Deuterium: 50000}
	b.Requirements = map[ID]int64{}
	return b
}

type gravitationSensors struct {
	BaseLfResearch
}

func newGravitationSensors() *gravitationSensors {
	b := new(gravitationSensors)
	b.Name = "GravitationSensors"
	b.ID = GravitationSensorsID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 240000, Crystal: 120000, Deuterium: 120000}
	b.Requirements = map[ID]int64{}
	return b
}

type overclockingBattleship struct {
	BaseLfResearch
}

func newOverclockingBattleship() *overclockingBattleship {
	b := new(overclockingBattleship)
	b.Name = "OverclockingBattleship"
	b.ID = OverclockingBattleshipID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 320000, Crystal: 240000, Deuterium: 100000}
	b.Requirements = map[ID]int64{}
	return b
}

type psionicShieldMatrix struct {
	BaseLfResearch
}

func newPsionicShieldMatrix() *psionicShieldMatrix {
	b := new(psionicShieldMatrix)
	b.Name = "PsionicShieldMatrix"
	b.ID = PsionicShieldMatrixID
	b.IncreaseFactor = 1.50
	b.BaseCost = Resources{Metal: 500000, Crystal: 300000, Deuterium: 200000}
	b.Requirements = map[ID]int64{}
	return b
}

type kaeleshDiscovererEnhancement struct {
	BaseLfResearch
}

func newKaeleshDiscovererEnhancement() *kaeleshDiscovererEnhancement {
	b := new(kaeleshDiscovererEnhancement)
	b.Name = "KaeleshDiscovererEnhancement"
	b.ID = KaeleshDiscovererEnhancementID
	b.IncreaseFactor = 1.70
	b.BaseCost = Resources{Metal: 300000, Crystal: 180000, Deuterium: 120000}
	b.Requirements = map[ID]int64{}
	return b
}
