package audio

import (
	"bytes"
	"fmt"
)

func DetectExtension(data []byte) (string, error) {
	if len(data) < 12 {
		return "", fmt.Errorf("file too small to detect type")
	}

	// WAV: RIFF....WAVE
	if bytes.HasPrefix(data, []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WAVE")) {
		return ".wav", nil
	}

	// MP3: ID3 or 0xFF 0xFB (MPEG frame sync)
	if bytes.HasPrefix(data, []byte("ID3")) {
		return ".mp3", nil
	}
	if len(data) >= 2 && data[0] == 0xFF && (data[1]&0xE0) == 0xE0 {
		return ".mp3", nil
	}

	// FLAC: fLaC
	if bytes.HasPrefix(data, []byte("fLaC")) {
		return ".flac", nil
	}

	// OGG: OggS
	if bytes.HasPrefix(data, []byte("OggS")) {
		return ".ogg", nil
	}

	// M4A/AAC: ftyp
	if len(data) >= 8 && bytes.Equal(data[4:8], []byte("ftyp")) {
		return ".m4a", nil
	}

	return "", fmt.Errorf("unsupported audio format")
}
