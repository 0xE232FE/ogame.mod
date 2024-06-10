package ogame

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLargeCargo_GetSpeed(t *testing.T) {
	lc := newLargeCargo()
	assert.Equal(t, int64(12000), lc.GetSpeed(Researches{CombustionDrive: 6}, LfBonuses{}, NoClass))
	assert.Equal(t, int64(19500), lc.GetSpeed(Researches{CombustionDrive: 6}, LfBonuses{}, Collector))
}

func TestLargeCargo_GetCargoCapacity(t *testing.T) {
	lc := newLargeCargo()
	assert.Equal(t, int64(35000), lc.GetCargoCapacity(Researches{HyperspaceTechnology: 8}, LfBonuses{}, NoClass, 0.05, false))
	assert.Equal(t, int64(37500), lc.GetCargoCapacity(Researches{HyperspaceTechnology: 10}, LfBonuses{}, NoClass, 0.05, false))
	assert.Equal(t, int64(43750), lc.GetCargoCapacity(Researches{HyperspaceTechnology: 10}, LfBonuses{}, Collector, 0.05, false))
}

func TestLargeCargo_GetFuelConsumption(t *testing.T) {
	lc := newLargeCargo()
	lfBonuses := LfBonuses{LfShipBonuses: make(LfShipBonuses)}
	lfBonuses.LfShipBonuses[LargeCargoID] = LfShipBonus{Speed: 0.16, FuelConsumption: 0.019}
	assert.Equal(t, int64(24), lc.GetFuelConsumption(Researches{}, lfBonuses, NoClass, 0.5))
}
