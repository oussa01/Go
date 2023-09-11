package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type LocationData struct {
	VehiculeID     string  `json:"VehiculeID"`
	Lat            float64 `json:"lat"`
	Lang           float64 `json:"lang"`
	Alt            float64 `json:"alt"`
	Speed          float64 `json:"speed"`
	Bearing        float64 `json:"bearing"`
	Acc            float64 `json:"acc"`
	Addr           string  `json:"addr"`
	RunningTime    string  `json:"runningTime"`
	VersionAndroid string  `json:"versionandroid"`
}


func receiveLocation(w http.ResponseWriter, r *http.Request) {
    // Parse the JSON data from the request body
    var location LocationData
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&location); err != nil {
        http.Error(w, "Failed to decode JSON data", http.StatusBadRequest)
        return
    }

    fmt.Printf("Received location data: %+v\n", location)
    response := "Location data received successfully"
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(response))
	
}

func main() {
	http.HandleFunc("/", receiveLocation) // Use an empty string for the path pattern
	port := ":5600" // Choose a port number

	fmt.Println("Server listening on port", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Println("Server error:", err)
	}
}

