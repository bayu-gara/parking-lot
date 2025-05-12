package model

type VehicleType string

const (
	Bicycle    VehicleType = "B"
	Motorcycle VehicleType = "M"
	Automobile VehicleType = "A"
)

type Spot struct {
	ID            int64       `json:"id"`
	Type          VehicleType `json:"type"`
	Occupied      bool        `json:"occupied"`
	VehicleNumber string      `json:"vehicle_number"`
	IDString      string      `json:"id_string"`
}
