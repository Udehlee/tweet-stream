package utils

import (
	"math/rand"
	"strings"

	"github.com/google/uuid"
)

// Airline returns any of the User listed
func GenerateUser() string {
	Users := []string{"Baba-K", "Uncle-lee", "Obekwu", "Naija"}
	index := rand.Intn(len(Users))
	return Users[index]
}

// generateID generates unique ids
func GenerateID() string {
	id := uuid.New()
	firstSix := id.String()[:6]
	return strings.ToUpper(firstSix)
}
