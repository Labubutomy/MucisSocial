package audio

import (
	"bytes"
	"fmt"
	"io"
	"math"

	"github.com/dhowden/tag"
)

func GetDuration(data []byte) (int64, error) {
	reader := bytes.NewReader(data)

	if _, err := tag.ReadFrom(reader); err != nil {
		return 0, fmt.Errorf("failed to read audio metadata: %w", err)
	}

	// Приблизительный расчет длительности.
	// Предполагаем средний битрейт 128 kbps.
	fileSize := float64(len(data))
	durationSeconds := int64(math.Round((fileSize * 8) / (128 * 1000)))
	if durationSeconds < 0 {
		durationSeconds = 0
	}
	return durationSeconds, nil
}

func GetDurationFromReader(reader io.Reader) (int64, []byte, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read audio data: %w", err)
	}

	duration, err := GetDuration(data)
	if err != nil {
		return 0, data, err
	}

	return duration, data, nil
}
