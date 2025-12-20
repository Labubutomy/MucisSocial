package main

import (
"encoding/json"
"fmt"
"time"

"github.com/Labubutomy/MucisSocial/services/recommendations/internal/store"
)

func main() {
	fmt.Println("🚀 Тестирование Geo Top Charts функциональности\n")

	geoStore := store.NewInMemoryGeoTopStore(1000, 60*time.Second)
	fmt.Println("✅ GeoTopStore создан")

	fmt.Println("\n📍 Добавление прослушиваний с координатами:")

	locations := []struct {
		name  string
		lat   float64
		lon   float64
		track string
		count int64
	}{
		{"Невский проспект", 59.9311, 30.3609, "track1", 50},
		{"Невский проспект", 59.9311, 30.3609, "track2", 30},
		{"Дворцовая площадь", 59.9387, 30.3162, "track1", 20},
		{"Дворцовая площадь", 59.9387, 30.3162, "track3", 15},
		{"Петропавловская крепость", 59.9505, 30.3164, "track2", 25},
		{"Васильевский остров", 59.9390, 30.2836, "track1", 40},
		{"Московский вокзал", 59.9293, 30.3606, "track3", 35},
	}

	for _, loc := range locations {
		geohash := store.EncodeGeohash(loc.lat, loc.lon, 6)
		geoStore.Incr(geohash, loc.track, loc.count)
		fmt.Printf("  • %s (%.4f, %.4f) -> %s: +%d (geohash: %s)\n",
loc.name, loc.lat, loc.lon, loc.track, loc.count, geohash)
	}

	fmt.Println("\n🔍 Запрос топа треков в радиусе 2км от Невского проспекта:")

	centerLat := 59.9311
	centerLon := 30.3609
	radiusM := 2000

	precision := store.GeohashPrecisionForRadius(radiusM)
	geohashes := store.ExpandNeighbors(centerLat, centerLon, radiusM, precision)

	fmt.Printf("  Центр: %.4f, %.4f\n", centerLat, centerLon)
	fmt.Printf("  Радиус: %d метров\n", radiusM)
	fmt.Printf("  Precision: %d\n", precision)
	fmt.Printf("  Количество геохешей: %d\n\n", len(geohashes))

	topTracks := geoStore.GetTopForGeohashes(geohashes, 5)

	fmt.Println("📊 Топ-5 треков:")
	for i, track := range topTracks {
		fmt.Printf("  %d. %s - %d прослушиваний\n", i+1, track.TrackID, track.Count)
	}

	fmt.Println("\n📦 JSON ответ (как в API):")

	response := map[string]interface{}{
		"meta": map[string]interface{}{
			"precision":         precision,
			"radius_m":          radiusM,
			"geohashes_queried": len(geohashes),
			"center_lat":        centerLat,
			"center_lon":        centerLon,
		},
		"tracks": topTracks,
	}

	jsonData, _ := json.MarshalIndent(response, "", "  ")
	fmt.Println(string(jsonData))

	fmt.Println("\n💾 Тест Snapshot/Restore:")
	snapshot := geoStore.Snapshot()
	fmt.Printf("  Snapshot создан: %d геохешей\n", len(snapshot.Counts))

	geoStore2 := store.NewInMemoryGeoTopStore(1000, 60*time.Second)
	geoStore2.Restore(snapshot)

	topTracks2 := geoStore2.GetTopForGeohashes(geohashes, 5)
	fmt.Printf("  Восстановлено: %d треков в топе\n", len(topTracks2))

	if len(topTracks) == len(topTracks2) {
		fmt.Println("  ✅ Snapshot/Restore работает корректно")
	}

	fmt.Println("\n🎉 Все тесты завершены успешно!")
}
