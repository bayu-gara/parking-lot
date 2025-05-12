package rest

import (
	"encoding/csv"
	"log"
	"net/http"

	//usecase
	parkinglotuc "github.com/bayu-gara/parking-lot/internal/usecase/parking-lot"

	//domain
	model "github.com/bayu-gara/parking-lot/internal/domain/model"

	//external
	jsoniter "github.com/json-iterator/go"
)

func setParkingLotMap(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var (
		err error
	)

	err = r.ParseMultipartForm(200 << 20) // 200MB max memory
	if err != nil {
		http.Error(w, "Could not parse multipart form", http.StatusBadRequest)
		return
	}

	// Get the file from the form field "file"
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Could not get uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	data, err := csvReader.ReadAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = parkinglotuc.UsecaseObj.SetParkingLotMap(ctx, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeSuccess(w, nil)
}

func parkVehicle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var (
		req ParkVehicleRequest
		err error
	)

	decoder := jsoniter.NewDecoder(r.Body)
	err = decoder.Decode(&req)
	if err != nil {
		http.Error(w, "Invalid Parameter", http.StatusBadRequest)
	}

	spotID, err := parkinglotuc.UsecaseObj.ParkVehicle(ctx, model.VehicleType(req.VehicleType), req.VehicleNumber)
	if err != nil {
		log.Printf("There is an error : %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := ParkVehicleResponse{
		SpotID: spotID,
	}

	writeSuccess(w, response)
}

func unparkVehicle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var (
		err error
	)

	queryParams := r.URL.Query()
	spotID := queryParams.Get("spot_id")
	if spotID == "" {
		http.Error(w, "Spot id should not be empty", http.StatusBadRequest)
		return
	}

	vehicleNumber := queryParams.Get("vehicle_number")
	if vehicleNumber == "" {
		http.Error(w, "Vehicle number should not be empty", http.StatusBadRequest)
		return
	}

	err = parkinglotuc.UsecaseObj.UnparkVehicle(ctx, spotID, vehicleNumber)
	if err != nil {
		log.Printf("There is an error : %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeSuccess(w, nil)
}

func availableSpot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var (
		err error
	)

	queryParams := r.URL.Query()
	vehicleType := queryParams.Get("vehicle_type")
	if vehicleType == "" {
		http.Error(w, "Vehicle type should not be empty", http.StatusBadRequest)
		return
	}

	spotIDs, err := parkinglotuc.UsecaseObj.AvailableSpot(ctx, model.VehicleType(vehicleType))
	if err != nil {
		log.Printf("There is an error : %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := AvailableSpotResponse{
		SpotIDs: spotIDs,
	}

	writeSuccess(w, response)
}

func searchVehicle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var (
		err error
	)

	queryParams := r.URL.Query()
	vehicleNumber := queryParams.Get("vehicle_number")
	if vehicleNumber == "" {
		http.Error(w, "Vehicle number should not be empty", http.StatusBadRequest)
		return
	}

	spotID, err := parkinglotuc.UsecaseObj.SearchVehicle(ctx, vehicleNumber)
	if err != nil {
		log.Printf("There is an error : %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := SearchVehicleResponse{
		SpotID: spotID,
	}

	writeSuccess(w, response)
}
