package rest

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	//external
	"github.com/gorilla/mux"
	jsoniter "github.com/json-iterator/go"
)

type RESTServer struct {
	Port int
}

func (rs RESTServer) Serve() error {
	log.Println("HTTP server listen at port: ", rs.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", rs.Port), getHandler())
	if errors.Is(err, http.ErrServerClosed) {
		log.Println("server closed")
	}

	return err
}

func getHandler() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/v1/vehicle/parking-map", setParkingLotMap).Methods("POST")
	router.HandleFunc("/v1/vehicle/park", parkVehicle).Methods("POST")
	router.HandleFunc("/v1/vehicle/park", unparkVehicle).Methods("DELETE")
	router.HandleFunc("/v1/vehicle/available-spot", availableSpot).Methods("GET")
	router.HandleFunc("/v1/vehicle/search", searchVehicle).Methods("GET")

	return router
}

func writeSuccess(w http.ResponseWriter, data interface{}) {
	jsonRes, err := jsoniter.Marshal(data)
	if err != nil {
		http.Error(w, "Failed to encode data to json", http.StatusInternalServerError)
		return
	}

	// Set content type header
	w.Header().Set("Content-Type", "application/json")

	// Write JSON response
	w.WriteHeader(http.StatusOK)
	w.Write(jsonRes)
}
