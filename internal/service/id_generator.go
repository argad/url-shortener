package service

import "github.com/google/uuid"

// UUIDGenerator is a simple ID generator that uses UUIDs.
type UUIDGenerator struct{}

// GenerateID generates a new UUID string.
func (g *UUIDGenerator) GenerateID() string {
	return uuid.New().String()
}
