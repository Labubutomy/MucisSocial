# Music Requests Service

Сервис для управления заказами треков от пользователей к артистам через QR коды.

## Возможности

- **QR-коды для заказов**: Артисты могут генерировать QR-коды для приема заказов треков
- **Система монет**: Заказ трека стоит 1 виртуальную монету
- **Принятие/отклонение заказов**: Артисты могут принять или отклонить заказ
- **Интеграция с очередью**: Принятые треки автоматически добавляются в очередь воспроизведения артиста
- **Транзакции монет**: При принятии заказа артист получает монету, при отклонении - монета возвращается заказчику

## API Endpoints

### Сессии заказов (Request Sessions)

- `POST /api/v1/sessions` - Создать новую сессию для приема заказов
- `GET /api/v1/sessions/my` - Получить активную сессию артиста
- `DELETE /api/v1/sessions/{session_id}` - Деактивировать сессию
- `GET /api/v1/sessions/{session_code}/qr` - Получить QR-код сессии

### Заказы треков (Track Requests)

- `POST /api/v1/requests` - Создать новый заказ трека (стоит 1 монету)
- `GET /api/v1/requests/incoming` - Получить входящие заказы (для артиста)
- `GET /api/v1/requests/outgoing` - Получить исходящие заказы (для пользователя)
- `PATCH /api/v1/requests/{request_id}` - Принять или отклонить заказ

### Монеты

- `GET /api/v1/users/{user_id}/coins` - Получить баланс монет пользователя

## Конфигурация

Переменные окружения (префикс `MUSIC_REQUESTS_`):

- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - JWT secret key (должен совпадать с API Gateway)
- `GATEWAY_URL` - URL API Gateway для взаимодействия с другими сервисами
- `PLAYBACK_QUEUE_GRPC_URL` - URL gRPC сервиса очереди воспроизведения
- `COIN_COST_PER_REQUEST` - Стоимость заказа в монетах (по умолчанию 1)
- `HOST` - Host для HTTP сервера (по умолчанию 0.0.0.0)
- `PORT` - Port для HTTP сервера (по умолчанию 8080)

## База данных

### Таблицы

1. **request_sessions** - Сессии для приема заказов
   - id, artist_id, session_code, is_active, created_at, updated_at

2. **track_requests** - Заказы треков
   - id, session_id, requester_id, artist_id, track_id, status, message, created_at, updated_at

3. **coin_transactions** - Транзакции монет
   - id, from_user_id, to_user_id, amount, request_id, transaction_type, created_at

## Запуск

### Локально

```bash
poetry install
poetry run uvicorn main:app --reload
```

### Docker

```bash
docker build -t music-requests-service .
docker run -p 8080:8080 music-requests-service
```

## Интеграция

### С Users Service
Сервис взаимодействует с Users Service через API Gateway для:
- Получения баланса монет пользователя
- Обновления баланса монет

### С Playback Queue Service
При принятии заказа трек добавляется в очередь воспроизведения артиста через gRPC.

## Workflow

1. **Артист создает сессию**:
   - POST /api/v1/sessions
   - Получает QR-код

2. **Пользователь сканирует QR-код**:
   - Переходит на страницу заказа трека
   - Выбирает трек и отправляет заказ (тратит 1 монету)

3. **Артист видит заказы**:
   - GET /api/v1/requests/incoming
   - Видит список ожидающих заказов

4. **Артист принимает заказ**:
   - PATCH /api/v1/requests/{request_id} {"action": "accept"}
   - Трек добавляется в очередь
   - Артист получает 1 монету

5. **Или артист отклоняет заказ**:
   - PATCH /api/v1/requests/{request_id} {"action": "decline"}
   - Монета возвращается пользователю
