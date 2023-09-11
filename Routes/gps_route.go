package routes

import (
	controllers "e/Controllers"

	"github.com/gorilla/mux"
)


func Gpsroute(router *mux.Router){
	router.HandleFunc("/track",controllers.HandleWebSocket)
	router.HandleFunc("/",controllers.StoreLocationwithoutCond).Methods("POST")
}

