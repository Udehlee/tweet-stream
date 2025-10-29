package generator

import "time"

// SimulatedFlightData holds the current flight data of a single aircraft
type SimulatedFlightData struct {
	FlightID      int
	FlightNumber  string
	Airline       string
	Latitude      float64
	Longitude     float64
	Altitude      float64
	FuelFlow      float64
	FuelRemaining float64
	Timestamps    time.Time
	UpdatedAt     time.Time
	CreatedAt     time.Time
}
