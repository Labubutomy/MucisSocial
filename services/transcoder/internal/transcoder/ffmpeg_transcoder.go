package transcoder

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/MucisSocial/transcoder/internal/storage"
	"github.com/MucisSocial/transcoder/internal/tracks"
)

type FFmpegTranscoder struct {
	storage     *storage.MinIO
	trackClient tracks.Client
	bucketName  string
	workDir     string
	logger      *log.Logger
	ffmpegPath  string
	ffprobePath string
}

func NewFFmpegTranscoder(storage *storage.MinIO, trackClient tracks.Client, workDir string, logger *log.Logger) *FFmpegTranscoder {
	if workDir == "" {
		workDir = os.TempDir()
	}
	return &FFmpegTranscoder{
		storage:     storage,
		trackClient: trackClient,
		bucketName:  storage.Bucket(),
		workDir:     workDir,
		logger:      logger,
		ffmpegPath:  getEnvOrDefault("FFMPEG_PATH", "ffmpeg"),
		ffprobePath: getEnvOrDefault("FFPROBE_PATH", "ffprobe"),
	}
}

func (t *FFmpegTranscoder) Transcode(ctx context.Context, task Task) error {
	jobDir, err := os.MkdirTemp(t.workDir, fmt.Sprintf("transcode-%s-%s-", task.ArtistID, shortID()))
	if err != nil {
		return fmt.Errorf("failed to create job workspace: %w", err)
	}
	defer os.RemoveAll(jobDir)

	bucket, objectKey, baseURL, err := parseTrackURL(task.TrackURL)
	if err != nil {
		return err
	}
	if bucket == "" {
		bucket = t.bucketName
	}

	if baseURL == "" {
		baseURL = defaultBaseURL(t.storage.Endpoint())
	}

	sourceFile := filepath.Join(jobDir, filepath.Base(objectKey))
	if err := t.storage.DownloadToFile(ctx, bucket, objectKey, sourceFile); err != nil {
		return fmt.Errorf("failed to download source audio: %w", err)
	}

	t.logger.Printf("downloaded source audio to %s", sourceFile)

	techMeta, err := t.extractTechMetadata(ctx, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to extract metadata: %w", err)
	}

	loudness, err := t.measureLoudness(ctx, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to measure loudness: %w", err)
	}

	transcodedDir := filepath.Join(jobDir, "transcoded")
	if err := os.MkdirAll(transcodedDir, 0o755); err != nil {
		return fmt.Errorf("failed to create transcoded directory: %w", err)
	}

	if err := t.generateHLS(ctx, sourceFile, transcodedDir); err != nil {
		return fmt.Errorf("failed to generate HLS outputs: %w", err)
	}

	metadataPrefix := path.Join(task.ArtistID, task.TrackID, "metadata")
	if err := t.storage.UploadJSON(ctx, bucket, path.Join(metadataPrefix, "tech_meta.json"), techMeta); err != nil {
		return fmt.Errorf("failed to upload tech_meta.json: %w", err)
	}
	if err := t.storage.UploadJSON(ctx, bucket, path.Join(metadataPrefix, "loudness.json"), loudness); err != nil {
		return fmt.Errorf("failed to upload loudness.json: %w", err)
	}

	transcodedPrefix := path.Join(task.ArtistID, task.TrackID, "transcoded")
	if err := t.storage.UploadDirectory(ctx, bucket, transcodedPrefix, transcodedDir); err != nil {
		return fmt.Errorf("failed to upload transcoded assets: %w", err)
	}

	if t.trackClient != nil {
		masterKey := path.Join(transcodedPrefix, "master.m3u8")
		masterURL := t.buildObjectURL(baseURL, bucket, masterKey)

		rounded := int64(math.Round(techMeta.DurationSec))
		var duration32 int32
		switch {
		case rounded < 0:
			duration32 = 0
		case rounded > math.MaxInt32:
			duration32 = math.MaxInt32
		default:
			duration32 = int32(rounded)
		}

		if err := t.trackClient.UpdateTrackInfo(ctx, task.TrackID, masterURL, duration32, ""); err != nil {
			return fmt.Errorf("failed to update track info: %w", err)
		}
	}

	t.logger.Printf("successfully processed track_id=%s artist_id=%s", task.TrackID, task.ArtistID)
	return nil
}

type TechMetadata struct {
	DurationSec     float64 `json:"duration_sec"`
	SampleRate      int     `json:"sample_rate"`
	Channels        int     `json:"channels"`
	OriginalCodec   string  `json:"original_codec"`
	BitDepth        *int    `json:"bit_depth,omitempty"`
	OriginalBitrate int     `json:"original_bitrate,omitempty"`
	FileSize        int64   `json:"file_size"`
	ChannelLayout   string  `json:"channel_layout,omitempty"`
}

type LoudnessMetrics struct {
	InputI         float64 `json:"input_i"`
	InputTP        float64 `json:"input_tp"`
	InputLRA       float64 `json:"input_lra"`
	InputThreshold float64 `json:"input_thresh"`
	TargetOffset   float64 `json:"target_offset"`
}

func (t *FFmpegTranscoder) extractTechMetadata(ctx context.Context, input string) (*TechMetadata, error) {
	cmd := exec.CommandContext(ctx, t.ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		input,
	)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	var probe ffprobeOutput
	if err := json.Unmarshal(output, &probe); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	stream := probe.primaryAudioStream()
	if stream == nil {
		return nil, errors.New("no audio stream found")
	}

	metadata := &TechMetadata{
		DurationSec:     roundToDecimals(probe.formatDuration(), 1),
		SampleRate:      stream.sampleRate(),
		Channels:        stream.Channels,
		OriginalCodec:   strings.ToUpper(stream.CodecName),
		OriginalBitrate: probe.formatBitrate(),
		FileSize:        probe.formatSize(),
		ChannelLayout:   stream.ChannelLayout,
	}

	if metadata.OriginalBitrate == 0 && stream.BitRate != "" {
		if v, err := strconv.Atoi(stream.BitRate); err == nil {
			metadata.OriginalBitrate = v
		}
	}

	if metadata.OriginalCodec == "" {
		metadata.OriginalCodec = strings.ToUpper(probe.Format.CodecName)
	}

	if depth := stream.bitDepth(); depth > 0 {
		metadata.BitDepth = &depth
	}

	if metadata.DurationSec == 0 && probe.Format.Duration != "" {
		if v, err := strconv.ParseFloat(probe.Format.Duration, 64); err == nil {
			metadata.DurationSec = roundToDecimals(v, 1)
		}
	}

	return metadata, nil
}

func (t *FFmpegTranscoder) measureLoudness(ctx context.Context, input string) (*LoudnessMetrics, error) {
	cmd := exec.CommandContext(ctx, t.ffmpegPath,
		"-hide_banner",
		"-i", input,
		"-af", "loudnorm=I=-14:LRA=7:TP=-2:print_format=json",
		"-f", "null",
		"-",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg loudnorm failed: %w (output=%s)", err, string(output))
	}

	jsonPayload, err := extractJSONBlock(output)
	if err != nil {
		return nil, fmt.Errorf("failed to extract loudness json: %w", err)
	}

	var raw loudnessRaw
	if err := json.Unmarshal(jsonPayload, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse loudness json: %w", err)
	}

	return &LoudnessMetrics{
		InputI:         raw.InputI,
		InputTP:        raw.InputTP,
		InputLRA:       raw.InputLRA,
		InputThreshold: raw.InputThresh,
		TargetOffset:   raw.TargetOffset,
	}, nil
}

func (t *FFmpegTranscoder) generateHLS(ctx context.Context, input string, outputDir string) error {
	variants := []struct {
		Name      string
		BitrateK  int
		Bandwidth int
	}{
		{"aac_256", 256, 256000},
		{"aac_160", 160, 160000},
		{"aac_96", 96, 96000},
	}

	for _, variant := range variants {
		dir := filepath.Join(outputDir, variant.Name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		segmentPattern := filepath.ToSlash(filepath.Join(dir, "chunk_%05d.m4s"))
		indexPath := filepath.Join(dir, "index.m3u8")

		args := []string{
			"-hide_banner",
			"-y",
			"-i", input,
			"-map", "0:a:0",
			"-c:a", "aac",
			"-b:a", fmt.Sprintf("%dk", variant.BitrateK),
			"-ac", "2",
			"-movflags", "+faststart",
			"-f", "hls",
			"-hls_time", "2",
			"-hls_playlist_type", "vod",
			"-hls_segment_type", "fmp4",
			"-hls_fmp4_init_filename", "init.mp4",
			"-hls_segment_filename", segmentPattern,
			indexPath,
		}

		cmd := exec.CommandContext(ctx, t.ffmpegPath, args...)
		var stderr bytes.Buffer
		cmd.Stdout = &stderr
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("ffmpeg failed for variant %s: %w (output=%s)", variant.Name, err, stderr.String())
		}
	}

	return t.writeMasterPlaylist(outputDir, variants)
}

func (t *FFmpegTranscoder) writeMasterPlaylist(outputDir string, variants []struct {
	Name      string
	BitrateK  int
	Bandwidth int
}) error {
	masterPath := filepath.Join(outputDir, "master.m3u8")
	file, err := os.Create(masterPath)
	if err != nil {
		return fmt.Errorf("failed to create master playlist: %w", err)
	}
	defer file.Close()

	var builder strings.Builder
	builder.WriteString("#EXTM3U\n")
	builder.WriteString("#EXT-X-VERSION:7\n")
	builder.WriteString("#EXT-X-INDEPENDENT-SEGMENTS\n")

	for _, variant := range variants {
		builder.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,AVERAGE-BANDWIDTH=%d,CODECS=\"mp4a.40.2\",NAME=\"AAC %d\"\n", variant.Bandwidth, variant.Bandwidth, variant.BitrateK))
		builder.WriteString(fmt.Sprintf("%s/index.m3u8\n", variant.Name))
	}

	if _, err := file.WriteString(builder.String()); err != nil {
		return fmt.Errorf("failed to write master playlist: %w", err)
	}

	return nil
}

type ffprobeOutput struct {
	Streams []ffprobeStream `json:"streams"`
	Format  ffprobeFormat   `json:"format"`
}

type ffprobeStream struct {
	CodecType     string `json:"codec_type"`
	CodecName     string `json:"codec_name"`
	SampleRate    string `json:"sample_rate"`
	Channels      int    `json:"channels"`
	BitsPerSample int    `json:"bits_per_sample"`
	BitsPerRaw    string `json:"bits_per_raw_sample"`
	BitRate       string `json:"bit_rate"`
	ChannelLayout string `json:"channel_layout"`
}

type ffprobeFormat struct {
	Duration  string `json:"duration"`
	BitRate   string `json:"bit_rate"`
	Size      string `json:"size"`
	CodecName string `json:"format_name"`
}

func (o ffprobeOutput) primaryAudioStream() *ffprobeStream {
	for i := range o.Streams {
		if o.Streams[i].CodecType == "audio" {
			return &o.Streams[i]
		}
	}
	return nil
}

func (o ffprobeOutput) formatDuration() float64 {
	if o.Format.Duration == "" {
		return 0
	}
	v, err := strconv.ParseFloat(o.Format.Duration, 64)
	if err != nil {
		return 0
	}
	return v
}

func (o ffprobeOutput) formatBitrate() int {
	if o.Format.BitRate == "" {
		return 0
	}
	v, err := strconv.Atoi(o.Format.BitRate)
	if err != nil {
		return 0
	}
	return v
}

func (o ffprobeOutput) formatSize() int64 {
	if o.Format.Size == "" {
		return 0
	}
	v, err := strconv.ParseInt(o.Format.Size, 10, 64)
	if err != nil {
		return 0
	}
	return v
}

func (s *ffprobeStream) sampleRate() int {
	if s.SampleRate == "" {
		return 0
	}
	v, err := strconv.Atoi(s.SampleRate)
	if err != nil {
		return 0
	}
	return v
}

func (s *ffprobeStream) bitDepth() int {
	if s.BitsPerSample > 0 {
		return s.BitsPerSample
	}
	if s.BitsPerRaw != "" {
		if v, err := strconv.Atoi(s.BitsPerRaw); err == nil {
			return v
		}
	}
	return 0
}

type loudnessRaw struct {
	InputI       float64 `json:"input_i,string"`
	InputTP      float64 `json:"input_tp,string"`
	InputLRA     float64 `json:"input_lra,string"`
	InputThresh  float64 `json:"input_thresh,string"`
	TargetOffset float64 `json:"target_offset,string"`
}

func extractJSONBlock(output []byte) ([]byte, error) {
	start := bytes.IndexByte(output, '{')
	end := bytes.LastIndexByte(output, '}')
	if start == -1 || end == -1 || end <= start {
		return nil, errors.New("json block not found")
	}
	return output[start : end+1], nil
}

func parseTrackURL(raw string) (bucket string, objectKey string, base string, err error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid track url: %w", err)
	}

	if u.Scheme == "" && strings.HasPrefix(raw, "/") {
		raw = "http://placeholder" + raw
		if u, err = url.Parse(raw); err != nil {
			return "", "", "", fmt.Errorf("invalid track url: %w", err)
		}
	}

	pathParts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(pathParts) < 2 {
		return "", "", "", fmt.Errorf("track url missing components: %s", raw)
	}
	bucket = pathParts[0]
	objectKey = strings.Join(pathParts[1:], "/")
	if u.Scheme != "" && u.Host != "" {
		base = strings.TrimRight(u.Scheme+"://"+u.Host, "/")
	}
	return bucket, objectKey, base, nil
}

func roundToDecimals(v float64, decimals int) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	multiplier := math.Pow(10, float64(decimals))
	return math.Round(v*multiplier) / multiplier
}

func getEnvOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func shortID() string {
	return fmt.Sprintf("%06d", rand.New(rand.NewSource(time.Now().UnixNano())).Intn(999999))
}

func (t *FFmpegTranscoder) buildObjectURL(base, bucket, objectKey string) string {
	base = strings.TrimRight(base, "/")
	if base == "" {
		return path.Join(bucket, objectKey)
	}
	return base + "/" + path.Join(bucket, objectKey)
}

func defaultBaseURL(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return ""
	}
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return strings.TrimRight(endpoint, "/")
	}
	return "http://" + strings.TrimRight(endpoint, "/")
}
