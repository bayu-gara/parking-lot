package rest

type ParkVehicleRequest struct {
	VehicleType   string `json:"vehicle_type"`
	VehicleNumber string `json:"vehicle_number"`
}

type ParkVehicleResponse struct {
	SpotID string `json:"spot_id"`
}

type AvailableSpotResponse struct {
	SpotIDs []string `json:"spot_ids"`
}

type SearchVehicleResponse struct {
	SpotID string `json:"spot_id"`
}
