package generator

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/google/uuid"
)

// Airline returns any of the airlines listed
func Airline() string {
	airlines := []string{"Baba-K", "Uncle-lee", "Obekwu", "Naija"}
	index := rand.Intn(len(airlines))
	return airlines[index]
}

// FlightNumber generates a flight number in the format UNC001-UNC888
func FlightNumber(airline string) string {
	prefix := airline
	if len(airline) > 3 {
		prefix = airline[:3]
	}

	prefix = strings.ToUpper(prefix)
	num := fmt.Sprintf("%03d", rand.Intn(888))

	return fmt.Sprintf("%s-%s", prefix, num)
}

// generateID generates unique ids
func GenerateID() string {
	id := uuid.New()
	firstSix := id.String()[:6]
	return strings.ToUpper(firstSix)
}
