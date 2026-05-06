package provider

import (
	"fmt"

	"github.com/google/uuid"
)

func uuidFromString(s string) (uuid.UUID, error) {
	if s == "" {
		return uuid.Nil, fmt.Errorf("empty uuid")
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}
