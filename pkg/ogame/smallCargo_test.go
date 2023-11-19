package ogame

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSmallCargoConstructionTime(t *testing.T) {
	sc := newSmallCargo()
	assert.Equal(t, 164*time.Second, sc.ConstructionTime(1, 7, Facilities{Shipyard: 4}, false, NoClass))
	assert.Equal(t, 328*time.Second, sc.ConstructionTime(2, 7, Facilities{Shipyard: 4}, false, NoClass))
}

func TestSmallCargoSpeed(t *testing.T) {
	sc := newSmallCargo()
	lf := newLfBonuses()
	assert.Equal(t, int64(6000), sc.GetSpeed(Researches{CombustionDrive: 2}, NoClass, lf))
	assert.Equal(t, int64(8000), sc.GetSpeed(Researches{CombustionDrive: 6}, NoClass, lf))
	assert.Equal(t, int64(8000), sc.GetSpeed(Researches{CombustionDrive: 6, ImpulseDrive: 4}, NoClass, lf))
	assert.Equal(t, int64(20000), sc.GetSpeed(Researches{CombustionDrive: 6, ImpulseDrive: 5}, NoClass, lf))
	assert.Equal(t, int64(22000), sc.GetSpeed(Researches{CombustionDrive: 10, ImpulseDrive: 6}, NoClass, lf))
}

func TestSmallCargoFuelConsumption(t *testing.T) {
	sc := newSmallCargo()
	lf := newLfBonuses()
	assert.Equal(t, int64(10), sc.GetFuelConsumption(Researches{}, 1, Collector, lf))
	assert.Equal(t, int64(20), sc.GetFuelConsumption(Researches{ImpulseDrive: 5}, 1, Collector, lf))
}
