package internal

import (
	"github.com/google/uuid"
)

// parseUUIDs парсит массив строк UUID в массив uuid.UUID
func parseUUIDs(uuidStrings []string) ([]uuid.UUID, error) {
	uuids := make([]uuid.UUID, 0, len(uuidStrings))
	for _, uuidStr := range uuidStrings {
		id, err := uuid.Parse(uuidStr)
		if err != nil {
			return nil, err
		}
		uuids = append(uuids, id)
	}
	return uuids, nil
}

