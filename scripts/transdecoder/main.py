import os
import json
import subprocess
from minio import Minio

# =========================
# Конфигурация MinIO
# =========================
MINIO_ENDPOINT = "localhost:9000"
MINIO_ACCESS_KEY = "minioadmin"
MINIO_SECRET_KEY = "minioadmin"
MINIO_BUCKET = "tracks"

# =========================
# Входные параметры
# =========================
INPUT_FILE = "input.mp3"  # или input.mp3
ARTIST_ID = "1"
TRACK_ID = "1"
OUTPUT_DIR = f"tmp/{TRACK_ID}"

BITRATES = [256, 160, 96]  # в kbps

# =========================
# Создаем MinIO клиент
# =========================
client = Minio(
    MINIO_ENDPOINT,
    access_key=MINIO_ACCESS_KEY,
    secret_key=MINIO_SECRET_KEY,
    secure=False
)

if not client.bucket_exists(MINIO_BUCKET):
    client.make_bucket(MINIO_BUCKET)

# =========================
# Шаг 0: Подготовка директорий
# =========================
dirs = [
    os.path.join(OUTPUT_DIR, "original"),
    os.path.join(OUTPUT_DIR, "metadata"),
    os.path.join(OUTPUT_DIR, "transcoded")
]
for d in dirs:
    os.makedirs(d, exist_ok=True)

# Копируем оригинальный файл
original_path = os.path.join(OUTPUT_DIR, "original", "original.wav")
subprocess.run(["ffmpeg", "-i", INPUT_FILE, original_path], check=True)

# =========================
# Шаг 1: Генерация метаданных (tech_meta + loudness)
# =========================
tech_meta = {}
loudness = {}

# Получаем duration, sample_rate, channels через ffprobe
cmd_probe = [
    "ffprobe", "-v", "error",
    "-select_streams", "a:0",
    "-show_entries", "stream=duration,sample_rate,channels",
    "-of", "json",
    original_path
]
res = subprocess.run(cmd_probe, capture_output=True, text=True)
info = json.loads(res.stdout)
stream = info["streams"][0]
tech_meta = {
    "duration": float(stream.get("duration", 0)),
    "sample_rate": int(stream.get("sample_rate", 44100)),
    "channels": int(stream.get("channels", 2)),
    "codec": "wav" if INPUT_FILE.endswith(".wav") else "mp3"
}

# Loudness placeholder
loudness = {
    "integrated_loudness": -23.0,
    "true_peak": -1.0
}

# Сохраняем json
with open(os.path.join(OUTPUT_DIR, "metadata", "tech_meta.json"), "w") as f:
    json.dump(tech_meta, f, indent=2)
with open(os.path.join(OUTPUT_DIR, "metadata", "loudness.json"), "w") as f:
    json.dump(loudness, f, indent=2)

# =========================
# Шаг 2: Транскодирование в несколько битрейтов и нарезка на fMP4 сегменты
# =========================
transcoded_dir = os.path.join(OUTPUT_DIR, "transcoded")
for br in BITRATES:
    br_dir = os.path.join(transcoded_dir, f"aac_{br}")
    os.makedirs(br_dir, exist_ok=True)

    playlist_path = os.path.join(br_dir, "index.m3u8")
    chunk_path = os.path.join(br_dir, "chunk_%03d.m4s")
    init_path = os.path.join(br_dir, "init.mp4")

    ffmpeg_cmd = [
        "ffmpeg",
        "-i", original_path,
        "-c:a", "aac",
        "-b:a", f"{br}k",
        "-f", "hls",
        "-hls_time", "2",
        "-hls_segment_type", "fmp4",
        "-hls_playlist_type", "vod",
        "-hls_segment_filename", os.path.join(br_dir, "chunk_%03d.m4s"),
        os.path.join(br_dir, "index.m3u8")
    ]

    subprocess.run(ffmpeg_cmd, check=True)

# =========================
# Шаг 3: Создаем master playlist
# =========================
master_playlist_path = os.path.join(transcoded_dir, "master.m3u8")
with open(master_playlist_path, "w") as f:
    f.write("#EXTM3U\n")
    for br in BITRATES:
        f.write(f"#EXT-X-STREAM-INF:BANDWIDTH={br*1000}\n")
        f.write(f"aac_{br}/index.m3u8\n")

# =========================
# Шаг 4: Загрузка всех файлов в MinIO
# =========================
for root, dirs, files in os.walk(OUTPUT_DIR):
    for file in files:
        local_path = os.path.join(root, file)
        relative_path = os.path.relpath(local_path, OUTPUT_DIR)
        minio_path = f"tracks/{ARTIST_ID}/{TRACK_ID}/{relative_path.replace(os.sep,'/')}"
        print(f"Загружаем {local_path} → {minio_path}")
        client.fput_object(MINIO_BUCKET, minio_path, local_path)

print("Все файлы загружены в MinIO!")
