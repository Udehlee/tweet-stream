package generator

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type AircraftGenerator struct {
	aircraft map[string]*SimulatedFlightData
	mu       sync.Mutex
	logger   *zerolog.Logger
}

func NewAircraftGenerator(logger *zerolog.Logger) *AircraftGenerator {
	A := &AircraftGenerator{
		aircraft: make(map[string]*SimulatedFlightData),
		logger:   logger,
	}

	return A
}

// Generate creates flight data
func (a *AircraftGenerator) Generate(lat, lon, alt, fuelFlow, fuelR float64) (*SimulatedFlightData, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	airline := Airline()
	id, flightNum := GenerateID(), FlightNumber(airline)

	flight := &SimulatedFlightData{
		FlightID:      id,
		FlightNumber:  flightNum,
		Airline:       airline,
		Latitude:      lat,
		Longitude:     lon,
		Altitude:      alt,
		FuelFlow:      fuelFlow,
		FuelRemaining: fuelR,
		UpdatedAt:     time.Now(),
		CreatedAt:     time.Now(),
	}

	a.aircraft[id] = flight
	a.logger.Info().Msg(fmt.Sprintf("flight data created id=%s flightNumber=%s lat=%.6f lon=%.6f", flight.FlightID, flight.FlightNumber, flight.Latitude, flight.Longitude))

	return flight, nil
}
