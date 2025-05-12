package model

type ParkingLotHistory struct {
	ID              int64  `json:"id"`
	VehicleNumber   string `json:"vehicle_number"`
	SpotID          string `json:"spot_id"`
	ParkingDateTime int64  `json:"parking_date_time"`
}
