package models

import "time"

type TelematicsData struct {
	VehicleID int
	Timestamp time.Time
	Speed     int
	Latitude  float64
	Longitude float64
}
