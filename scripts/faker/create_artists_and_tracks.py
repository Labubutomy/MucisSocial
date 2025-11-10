#!/usr/bin/env python3
"""
Скрипт для создания исполнителей через gateway и загрузки их треков.

Использование:
    python create_artists_and_tracks.py [--gateway-url GATEWAY_URL] [--email EMAIL] [--password PASSWORD]

Параметры:
    --gateway-url: URL gateway (по умолчанию: http://localhost:8080)
    --email: Email для регистрации/входа (по умолчанию: faker@example.com)
    --password: Пароль для регистрации/входа (по умолчанию: faker123)
"""

import argparse
import json
import os
import re
import sys
from pathlib import Path
from typing import Dict, List, Optional, Tuple

import requests


class GatewayClient:
    """Клиент для работы с Gateway API."""

    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url.rstrip("/")
        self.access_token: Optional[str] = None

    def sign_up(self, email: str, password: str, username: str) -> Dict:
        """Регистрация нового пользователя."""
        url = f"{self.base_url}/api/v1/auth/sign-up"
        payload = {
            "email": email,
            "password": password,
            "username": username,
        }
        response = requests.post(url, json=payload)
        response.raise_for_status()
        data = response.json()
        self.access_token = data.get("access_token")
        return data

    def sign_in(self, email: str, password: str) -> Dict:
        """Вход в систему."""
        url = f"{self.base_url}/api/v1/auth/sign-in"
        payload = {
            "email": email,
            "password": password,
        }
        response = requests.post(url, json=payload)
        response.raise_for_status()
        data = response.json()
        self.access_token = data.get("access_token")
        return data

    def create_artist(
        self, name: str, avatar_url: str, genres: List[str]
    ) -> Dict:
        """Создание исполнителя."""
        if not self.access_token:
            raise ValueError("Необходимо сначала войти в систему")

        url = f"{self.base_url}/api/v1/artists"
        headers = {
            "Authorization": f"Bearer {self.access_token}",
            "Content-Type": "application/json",
        }
        payload = {
            "name": name,
            "avatar_url": avatar_url,
            "genres": genres,
        }
        response = requests.post(url, json=payload, headers=headers)
        response.raise_for_status()
        return response.json()

    def upload_track(
        self,
        file_path: str,
        track_name: str,
        artist_ids: List[str],
        genre: str,
    ) -> Dict:
        """Загрузка трека."""
        url = f"{self.base_url}/api/v1/upload/track"

        with open(file_path, "rb") as f:
            files = {"file": (os.path.basename(file_path), f)}
            # Используем правильный формат для multipart/form-data с массивом
            form_data = []
            form_data.append(("track_name", track_name))
            form_data.append(("genre", genre))
            for artist_id in artist_ids:
                form_data.append(("artist_ids", artist_id))

            response = requests.post(
                url, files=files, data=form_data
            )
        response.raise_for_status()
        return response.json()


def normalize_name(name: str) -> str:
    """Нормализация имени для сопоставления."""
    # Убираем специальные символы, приводим к верхнему регистру
    name = re.sub(r"[^\w\s]", "", name.upper())
    # Заменяем множественные пробелы на один
    name = re.sub(r"\s+", "_", name)
    return name


def find_matching_file(artist_name: str, input_dir: Path, debug: bool = False) -> Optional[Path]:
    """Находит файл, соответствующий имени исполнителя.
    
    Теперь файлы находятся в подпапках по имени исполнителя.
    """
    # Полный маппинг имен исполнителей из artists.json в имена папок
    # Это нужно, так как имена в JSON могут быть на кириллице, а папки на латинице
    artist_to_folder = {
        "Anna Asti": "ANNA_ASTI",
        "AP$ENT": "APENT",
        "APENT": "APENT",
        "Artik & Asti": "ARTIK_ASTI",
        "Артур Пирожков": "ARTUR_PIROZHKOV",
        "GAYAZOV$ BROTHERS$": "GAYAZOV_BROTHER",
        "GAYAZOV_BROTHERS": "GAYAZOV_BROTHER",
        "Кравц": "KRAVC",
        "Гио Пика": "GIO_PIKA",
        "Король и Шут": "KOROL_I_SHUT",
        "Планета 9": "PLANETA_9",
    }

    if debug:
        print(f"    [DEBUG] Ищем файл для исполнителя: {artist_name}")

    # Сначала проверяем точное совпадение
    folder_name = artist_to_folder.get(artist_name)
    
    # Нормализуем имя для дальнейшего использования
    normalized_artist = normalize_name(artist_name)
    
    if debug:
        print(f"    [DEBUG] Имя папки из маппинга: {folder_name}")
        print(f"    [DEBUG] Нормализованное имя: {normalized_artist}")

    # Если точного совпадения нет, пробуем найти через нормализацию
    if not folder_name:
        # Проверяем нормализованное имя в маппинге
        for artist, folder in artist_to_folder.items():
            if normalize_name(artist) == normalized_artist:
                folder_name = folder
                if debug:
                    print(f"    [DEBUG] Найдено через нормализацию: {folder_name}")
                break
        
        # Если все еще не найдено, используем нормализованное имя как есть
        if not folder_name:
            folder_name = normalized_artist
            if debug:
                print(f"    [DEBUG] Используем нормализованное имя как папку: {folder_name}")
    
    # Ищем папку с таким именем
    artist_folder = input_dir / folder_name
    if not artist_folder.exists() or not artist_folder.is_dir():
        if debug:
            print(f"    [DEBUG] Папка {folder_name} не найдена, ищем похожие...")
            print(f"    [DEBUG] Доступные папки: {[f.name for f in input_dir.iterdir() if f.is_dir()]}")
        
        # Пробуем найти папку, которая содержит нормализованное имя
        found = False
        for folder in input_dir.iterdir():
            if folder.is_dir():
                folder_name_normalized = normalize_name(folder.name)
                if debug:
                    print(f"    [DEBUG] Проверяем папку: {folder.name} -> {folder_name_normalized}")
                
                if (
                    folder_name_normalized == normalized_artist
                    or normalized_artist in folder_name_normalized
                    or folder_name_normalized in normalized_artist
                ):
                    artist_folder = folder
                    found = True
                    if debug:
                        print(f"    [DEBUG] Найдена папка: {folder.name}")
                    break
        
        if not found:
            if debug:
                print(f"    [DEBUG] Папка не найдена для {artist_name}")
            return None

    # Ищем первый .wav файл в папке
    wav_files = list(artist_folder.glob("*.wav"))
    if debug:
        print(f"    [DEBUG] Найдено .wav файлов в папке {artist_folder.name}: {len(wav_files)}")
    
    if wav_files:
        # Если несколько файлов, берем первый
        return wav_files[0]

    if debug:
        print(f"    [DEBUG] .wav файлы не найдены в папке {artist_folder}")
    return None


def extract_track_name_from_file(file_path: Path) -> str:
    """Извлекает название трека из имени файла.
    
    Теперь файлы имеют нормальные названия без ID, просто берем имя файла.
    """
    # Просто возвращаем имя файла без расширения
    return file_path.stem


def main():
    parser = argparse.ArgumentParser(
        description="Создание исполнителей и загрузка треков через gateway"
    )
    parser.add_argument(
        "--gateway-url",
        default="http://localhost:8080",
        help="URL gateway (по умолчанию: http://localhost:8080)",
    )
    parser.add_argument(
        "--email",
        default="faker@example.com",
        help="Email для регистрации/входа (по умолчанию: faker@example.com)",
    )
    parser.add_argument(
        "--password",
        default="faker123",
        help="Пароль для регистрации/входа (по умолчанию: faker123)",
    )
    parser.add_argument(
        "--artists-file",
        default="artists.json",
        help="Путь к файлу с исполнителями (по умолчанию: artists.json)",
    )
    parser.add_argument(
        "--input-dir",
        default="input",
        help="Директория с аудиофайлами (по умолчанию: input)",
    )
    parser.add_argument(
        "--debug",
        action="store_true",
        help="Включить отладочный вывод",
    )

    args = parser.parse_args()

    # Определяем пути
    script_dir = Path(__file__).parent
    artists_file = script_dir / args.artists_file
    input_dir = script_dir / args.input_dir

    if not artists_file.exists():
        print(f"Ошибка: файл {artists_file} не найден")
        sys.exit(1)

    if not input_dir.exists():
        print(f"Ошибка: директория {input_dir} не найдена")
        sys.exit(1)

    # Загружаем данные об исполнителях
    with open(artists_file, "r", encoding="utf-8") as f:
        artists = json.load(f)

    # Создаем клиент
    client = GatewayClient(args.gateway_url)

    # Регистрируемся или входим
    print(f"Попытка входа/регистрации с email: {args.email}")
    try:
        client.sign_in(args.email, args.password)
        print("✓ Успешный вход")
    except requests.exceptions.HTTPError:
        print("Вход не удался, пробуем зарегистрироваться...")
        try:
            username = args.email.split("@")[0]
            client.sign_up(args.email, args.password, username)
            print("✓ Успешная регистрация")
        except requests.exceptions.HTTPError as e:
            print(f"Ошибка регистрации: {e}")
            if e.response is not None:
                print(f"Ответ сервера: {e.response.text}")
            sys.exit(1)

    # Создаем исполнителей и загружаем треки
    created_artists = []
    uploaded_tracks = []

    for i, artist_data in enumerate(artists, 1):
        artist_name = artist_data["name"]
        avatar_url = artist_data.get("avatar_url", "")
        genres = artist_data.get("genres", [])

        print(f"\n[{i}/{len(artists)}] Обработка исполнителя: {artist_name}")

        # Создаем исполнителя
        try:
            artist_response = client.create_artist(
                name=artist_name, avatar_url=avatar_url, genres=genres
            )
            artist_id = artist_response["artist"]["id"]
            created_artists.append({"name": artist_name, "id": artist_id})
            print(f"  ✓ Исполнитель создан: {artist_id}")
        except requests.exceptions.HTTPError as e:
            print(f"  ✗ Ошибка создания исполнителя: {e}")
            if e.response is not None:
                print(f"    Ответ сервера: {e.response.text}")
            continue

        # Ищем соответствующий файл
        matching_file = find_matching_file(artist_name, input_dir, debug=args.debug)
        if not matching_file:
            print(f"  ⚠ Файл для исполнителя {artist_name} не найден")
            continue

        print(f"  ✓ Найден файл: {matching_file.name}")

        # Извлекаем название трека из имени файла
        track_name = extract_track_name_from_file(matching_file)

        # Загружаем трек
        try:
            # Используем первый жанр из списка, если есть
            genre = genres[0] if genres else "Поп"
            upload_response = client.upload_track(
                file_path=str(matching_file),
                track_name=track_name,
                artist_ids=[artist_id],
                genre=genre,
            )
            track_id = upload_response.get("track_id", "unknown")
            uploaded_tracks.append(
                {
                    "artist": artist_name,
                    "track": track_name,
                    "track_id": track_id,
                }
            )
            print(f"  ✓ Трек загружен: {track_name} (ID: {track_id})")
        except requests.exceptions.HTTPError as e:
            print(f"  ✗ Ошибка загрузки трека: {e}")
            if e.response is not None:
                error_text = e.response.text
                print(f"    Ответ сервера: {error_text}")
                # Проверяем, не связана ли ошибка с недоступностью upload-service
                if "upload-service" in error_text.lower() and "no such host" in error_text.lower():
                    print(f"    ⚠ Внимание: upload-service недоступен. Убедитесь, что все сервисы запущены.")

    # Итоговая статистика
    print("\n" + "=" * 60)
    print("ИТОГИ:")
    print(f"  Создано исполнителей: {len(created_artists)}")
    print(f"  Загружено треков: {len(uploaded_tracks)}")
    print("=" * 60)

    if created_artists:
        print("\nСозданные исполнители:")
        for artist in created_artists:
            print(f"  - {artist['name']}: {artist['id']}")

    if uploaded_tracks:
        print("\nЗагруженные треки:")
        for track in uploaded_tracks:
            print(f"  - {track['artist']} - {track['track']}: {track['track_id']}")


if __name__ == "__main__":
    main()

