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

// validateArtists проверяет, что все артисты найдены
func validateArtists(found []Artist, requested []uuid.UUID) bool {
	if len(found) != len(requested) {
		return false
	}

	// Создаем map для быстрой проверки
	foundMap := make(map[uuid.UUID]bool, len(found))
	for _, artist := range found {
		foundMap[artist.ID] = true
	}

	// Проверяем, что все запрошенные артисты найдены
	for _, id := range requested {
		if !foundMap[id] {
			return false
		}
	}

	return true
}
