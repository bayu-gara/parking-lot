package common

import (
	"github.com/bayu-gara/parking-lot/internal/domain/model"
)

func IsValidVehicleType(vehicleType model.VehicleType) bool {
	switch vehicleType {
	case model.Motorcycle, model.Automobile, model.Bicycle:
		return true
	}

	return false
}
