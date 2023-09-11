package models


type GPS struct{

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