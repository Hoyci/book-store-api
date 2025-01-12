package utils

import "github.com/google/uuid"

type UUIDGeneratorUtil struct{}

func (g *UUIDGeneratorUtil) New() string {
	return uuid.New().String()
}
