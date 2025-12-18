# План внедрения сервиса социальных сетей для Music Social

## Обзор

Данный документ содержит пошаговый план внедрения социального функционала в платформу Music Social. Основная концепция - создание музыкально-ориентированной социальной сети, где общение строится вокруг музыки и контекстов, без традиционного фида постов.

## Текущая архитектура системы

### Существующие компоненты

1. **Users Service** (Go, gRPC, PostgreSQL)
   - Аутентификация и управление пользователями
   - JWT токены (access + refresh)
   - Профили пользователей

2. **Tracks Service** (Go, gRPC/HTTP, PostgreSQL)
   - Управление треками
   - Метаданные треков
   - Связь треков с артистами

3. **Playback Queue Service** (Go, gRPC, PostgreSQL)
   - Управление очередями воспроизведения
   - Контекстные очереди (user, group, session)

4. **Session Service** (Kotlin, Spring Boot, MongoDB, Redis)
   - Совместное прослушивание
   - Синхронизация состояний

5. **WebSocket Gateway** (Python, FastAPI)
   - Real-time коммуникация
   - Интеграция с Kafka/Redpanda

6. **API Gateway** (Go, HTTP REST)
   - Единая точка входа
   - JWT аутентификация
   - Маршрутизация запросов

7. **Инфраструктура**
   - PostgreSQL (несколько экземпляров)
   - MongoDB (сессии)
   - Redis (кеширование)
   - Redpanda (Kafka-совместимый брокер)
   - MinIO (S3-совместимое хранилище)

## Требования к социальному функционалу

1. **Личные текстовые сообщения между пользователями**
   - Отправка и получение сообщений
   - История переписки
   - Уведомления о новых сообщениях
   - Статусы прочтения

2. **Отправка треков через чат**
   - Вложение треков в сообщения
   - Предпросмотр треков в чате
   - Прямое воспроизведение из чата

3. **Общение вокруг музыки и контекстов**
   - Связь сообщений с музыкальными контекстами (треки, плейлисты, сессии)
   - Обсуждение треков в контексте
   - Музыкальные рекомендации в чатах

## Пошаговый план внедрения

---

## Этап 1: Проектирование и подготовка инфраструктуры

### 1.1. Проектирование схемы базы данных

**Задача**: Спроектировать схему БД для хранения сообщений, чатов и связей с музыкальными контекстами.

**Действия**:
1. Создать миграции для таблиц:
   - `conversations` - диалоги между пользователями
   - `messages` - текстовые сообщения
   - `message_tracks` - связь сообщений с треками
   - `message_contexts` - связь сообщений с контекстами (плейлисты, сессии)
   - `message_read_status` - статусы прочтения сообщений
   - `user_conversations` - связь пользователей с диалогами

2. Индексы для оптимизации:
   - По `conversation_id` и `created_at` для быстрого получения истории
   - По `sender_id` и `recipient_id` для поиска диалогов
   - По `track_id` для поиска сообщений с треками
   - По `context_type` и `context_id` для контекстных сообщений

**Оценка**: 2-3 дня

### 1.2. Создание нового сервиса Messaging Service

**Задача**: Создать новый микросервис для управления сообщениями.

**Технологии**: Go, gRPC, PostgreSQL

**Структура сервиса**:
```
services/messaging/
├── cmd/
│   └── main.go
├── internal/
│   ├── config/
│   ├── domain/
│   ├── handler/
│   ├── repository/
│   └── service/
├── migrations/
│   └── 001_initial_schema.up.sql
├── proto/
│   └── messaging/
│       └── v1/
│           └── messaging.proto
├── Dockerfile
├── go.mod
└── README.md
```

**Оценка**: 3-4 дня

### 1.3. Определение gRPC протокола

**Задача**: Создать proto файлы для Messaging Service.

**Основные RPC методы**:
- `CreateConversation` - создание диалога
- `SendMessage` - отправка текстового сообщения
- `SendTrackMessage` - отправка сообщения с треком
- `GetConversations` - получение списка диалогов
- `GetMessages` - получение истории сообщений
- `MarkAsRead` - отметка сообщений как прочитанных
- `GetUnreadCount` - количество непрочитанных сообщений
- `SearchMessages` - поиск по сообщениям

**Оценка**: 1 день

---

## Этап 2: Реализация базового функционала сообщений

### 2.1. Реализация репозитория

**Задача**: Реализовать слой работы с БД для сообщений.

**Действия**:
1. Создать модели данных:
   - `Conversation`
   - `Message`
   - `MessageTrack`
   - `MessageContext`
   - `ReadStatus`

2. Реализовать методы репозитория:
   - `CreateConversation(user1_id, user2_id)`
   - `GetConversation(conversation_id)`
   - `GetUserConversations(user_id)`
   - `CreateMessage(message)`
   - `GetMessages(conversation_id, limit, offset)`
   - `MarkAsRead(conversation_id, user_id, message_id)`
   - `GetUnreadCount(user_id, conversation_id)`

**Оценка**: 3-4 дня

### 2.2. Реализация бизнес-логики

**Задача**: Реализовать сервисный слой с бизнес-логикой.

**Действия**:
1. Валидация сообщений (макс. длина, запрещенные символы)
2. Автоматическое создание диалогов при первом сообщении
3. Проверка прав доступа (только участники диалога могут видеть сообщения)
4. Обработка отправки треков (валидация track_id через Tracks Service)
5. Логика статусов прочтения

**Оценка**: 4-5 дней

### 2.3. Реализация gRPC хэндлеров

**Задача**: Реализовать обработчики gRPC запросов.

**Действия**:
1. Реализовать все методы из proto файла
2. Обработка ошибок и валидация входных данных
3. Интеграция с Users Service для проверки пользователей
4. Интеграция с Tracks Service для валидации треков

**Оценка**: 3-4 дня

### 2.4. Интеграция с API Gateway

**Задача**: Добавить REST эндпоинты в API Gateway.

**Действия**:
1. Добавить подключение к Messaging Service в Gateway
2. Создать REST эндпоинты:
   - `POST /api/v1/conversations` - создание диалога
   - `GET /api/v1/conversations` - список диалогов
   - `GET /api/v1/conversations/{id}` - получение диалога
   - `POST /api/v1/conversations/{id}/messages` - отправка сообщения
   - `GET /api/v1/conversations/{id}/messages` - история сообщений
   - `POST /api/v1/conversations/{id}/messages/{messageId}/read` - отметка прочитанным
   - `GET /api/v1/conversations/{id}/unread` - количество непрочитанных

3. Добавить JWT middleware для всех эндпоинтов

**Оценка**: 2-3 дня

---

## Этап 3: Real-time коммуникация через WebSocket

### 3.1. Расширение WebSocket Gateway

**Задача**: Добавить поддержку сообщений в WebSocket Gateway.

**Действия**:
1. Добавить новые типы сообщений в `models/messages.py`:
   - `MESSAGE_SENT` - новое сообщение
   - `MESSAGE_READ` - сообщение прочитано
   - `TRACK_MESSAGE` - сообщение с треком

2. Расширить `ConnectionManager`:
   - Поддержка подписок на диалоги (не только комнаты)
   - Маршрутизация сообщений по `conversation_id`

3. Добавить Kafka топики:
   - `messaging-events` - события отправки сообщений
   - `messaging-sync` - синхронизация для рассылки

**Оценка**: 3-4 дня

### 3.2. Интеграция Messaging Service с Kafka

**Задача**: Добавить отправку событий в Kafka при создании сообщений.

**Действия**:
1. Добавить Kafka producer в Messaging Service
2. Отправлять события при:
   - Создании нового сообщения
   - Изменении статуса прочтения
3. Формат событий:
   ```json
   {
     "type": "message_sent",
     "conversation_id": "...",
     "message_id": "...",
     "sender_id": "...",
     "recipient_id": "...",
     "content": "...",
     "track_id": "...", // опционально
     "timestamp": "..."
   }
   ```

**Оценка**: 2-3 дня

### 3.3. Реализация уведомлений

**Задача**: Реализовать систему уведомлений о новых сообщениях.

**Действия**:
1. При создании сообщения отправлять уведомление получателю через WebSocket
2. Если получатель не онлайн, сохранять уведомление в Redis
3. При подключении пользователя отправлять накопленные уведомления
4. Реализовать счетчик непрочитанных сообщений

**Оценка**: 3-4 дня

---

## Этап 4: Интеграция с музыкальными контекстами

### 4.1. Связь сообщений с треками

**Задача**: Реализовать отправку треков через чат.

**Действия**:
1. Расширить модель `Message` для поддержки `track_id`
2. При отправке сообщения с треком:
   - Валидировать `track_id` через Tracks Service
   - Сохранять связь в таблице `message_tracks`
   - Получать метаданные трека (название, артист, обложка)
3. В API ответах включать полную информацию о треке

**Оценка**: 2-3 дня

### 4.2. Связь сообщений с контекстами

**Задача**: Позволить привязывать сообщения к музыкальным контекстам.

**Действия**:
1. Расширить модель для поддержки контекстов:
   - `context_type` (playlist, session, track, route)
   - `context_id`
2. При создании сообщения с контекстом:
   - Валидировать существование контекста
   - Сохранять связь в `message_contexts`
3. API для получения сообщений по контексту:
   - `GET /api/v1/contexts/{type}/{id}/messages`

**Оценка**: 3-4 дня

### 4.3. Интеграция с Playback Queue

**Задача**: Позволить добавлять треки из чата в очередь воспроизведения.

**Действия**:
1. Добавить эндпоинт в Messaging Service:
   - `AddTrackToQueue(conversation_id, message_id, context_type, context_id)`
2. Интеграция с Playback Queue Service:
   - При клике "Добавить в очередь" вызывать `EnqueueTrack`
3. Поддержка разных контекстов:
   - Личная очередь пользователя
   - Очередь сессии совместного прослушивания

**Оценка**: 2-3 дня

---

## Этап 5: Фронтенд интеграция

### 5.1. Создание UI компонентов

**Задача**: Создать UI для чата и сообщений.

**Действия**:
1. Создать компоненты:
   - `ConversationList` - список диалогов
   - `ConversationView` - окно чата
   - `MessageBubble` - сообщение
   - `TrackMessage` - сообщение с треком
   - `MessageInput` - поле ввода с поддержкой треков

2. Интеграция с API:
   - Создать API клиент для Messaging Service
   - Реализовать загрузку истории сообщений
   - Реализовать отправку сообщений

**Оценка**: 5-7 дней

### 5.2. WebSocket интеграция

**Задача**: Интегрировать WebSocket для real-time сообщений.

**Действия**:
1. Расширить существующий WebSocket клиент:
   - Поддержка подписки на диалоги
   - Обработка событий `MESSAGE_SENT`, `MESSAGE_READ`
2. Реализовать:
   - Автоматическое обновление списка диалогов
   - Real-time отображение новых сообщений
   - Индикаторы непрочитанных сообщений

**Оценка**: 3-4 дня

### 5.3. Интеграция с плеером

**Задача**: Позволить воспроизводить треки из чата.

**Действия**:
1. При клике на трек в сообщении:
   - Открывать плеер
   - Загружать трек через Streaming Service
   - Добавлять в очередь воспроизведения
2. Визуальное отображение:
   - Карточка трека в сообщении
   - Кнопка "Добавить в очередь"
   - Кнопка "Воспроизвести"

**Оценка**: 2-3 дня

---

## Этап 6: Дополнительные функции

### 6.1. Поиск и фильтрация

**Задача**: Реализовать поиск по сообщениям и фильтрацию.

**Действия**:
1. Поиск по тексту сообщений:
   - `GET /api/v1/conversations/{id}/messages/search?q=...`
2. Фильтрация по типу:
   - Только текстовые
   - Только с треками
   - По дате
3. Индексация для полнотекстового поиска (PostgreSQL FTS)

**Оценка**: 3-4 дня

### 6.2. Статистика и аналитика

**Задача**: Добавить базовую статистику по сообщениям.

**Действия**:
1. Метрики:
   - Количество сообщений в диалоге
   - Самые отправляемые треки
   - Активность пользователей
2. Эндпоинты:
   - `GET /api/v1/conversations/{id}/stats`

**Оценка**: 2-3 дня

### 6.3. Безопасность и модерация

**Задача**: Добавить базовую модерацию и безопасность.

**Действия**:
1. Валидация контента:
   - Фильтрация запрещенных слов
   - Ограничение длины сообщений
   - Rate limiting на отправку сообщений
2. Блокировка пользователей:
   - Возможность заблокировать пользователя
   - Автоматическое отклонение сообщений от заблокированных
3. Жалобы:
   - Система жалоб на сообщения
   - Логирование для модерации

**Оценка**: 4-5 дней

---

## Этап 7: Тестирование и оптимизация

### 7.1. Unit тесты

**Задача**: Написать unit тесты для всех компонентов.

**Действия**:
1. Тесты для репозитория
2. Тесты для сервисного слоя
3. Тесты для gRPC хэндлеров
4. Покрытие кода > 70%

**Оценка**: 5-7 дней

### 7.2. Integration тесты

**Задача**: Написать integration тесты.

**Действия**:
1. Тесты API эндпоинтов
2. Тесты WebSocket коммуникации
3. Тесты интеграции с другими сервисами

**Оценка**: 3-4 дня

### 7.3. Нагрузочное тестирование

**Задача**: Протестировать систему под нагрузкой.

**Действия**:
1. Тестирование отправки сообщений (1000+ сообщений/сек)
2. Тестирование WebSocket соединений (1000+ одновременных)
3. Оптимизация запросов к БД
4. Настройка индексов

**Оценка**: 3-4 дня

### 7.4. Оптимизация производительности

**Задача**: Оптимизировать производительность системы.

**Действия**:
1. Кеширование часто запрашиваемых данных (Redis)
2. Пагинация для больших списков
3. Ленивая загрузка истории сообщений
4. Оптимизация SQL запросов

**Оценка**: 3-4 дня

---

## Схема базы данных

### Таблица: conversations

```sql
CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_message_at TIMESTAMP,
    last_message_id UUID,
    UNIQUE(id)
);

CREATE INDEX idx_conversations_updated_at ON conversations(updated_at DESC);
```

### Таблица: user_conversations

```sql
CREATE TABLE user_conversations (
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    unread_count INT DEFAULT 0,
    last_read_message_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (conversation_id, user_id)
);

CREATE INDEX idx_user_conversations_user_id ON user_conversations(user_id);
CREATE INDEX idx_user_conversations_unread ON user_conversations(user_id, unread_count) WHERE unread_count > 0;
```

### Таблица: messages

```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_messages_conversation_created ON messages(conversation_id, created_at DESC);
CREATE INDEX idx_messages_sender ON messages(sender_id);
CREATE INDEX idx_messages_created_at ON messages(created_at DESC);
```

### Таблица: message_tracks

```sql
CREATE TABLE message_tracks (
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    track_id UUID NOT NULL,
    PRIMARY KEY (message_id, track_id)
);

CREATE INDEX idx_message_tracks_track_id ON message_tracks(track_id);
```

### Таблица: message_contexts

```sql
CREATE TABLE message_contexts (
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    context_type VARCHAR(50) NOT NULL, -- 'playlist', 'session', 'track', 'route'
    context_id UUID NOT NULL,
    PRIMARY KEY (message_id, context_type, context_id)
);

CREATE INDEX idx_message_contexts_context ON message_contexts(context_type, context_id);
```

### Таблица: message_read_status

```sql
CREATE TABLE message_read_status (
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    read_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (message_id, user_id)
);

CREATE INDEX idx_message_read_status_user ON message_read_status(user_id);
```

---

## Примеры gRPC протоколов

### messaging.proto

```protobuf
syntax = "proto3";

package messaging.v1;

option go_package = "github.com/MusicSocial/messaging-service/proto/messaging/v1";

import "google/protobuf/timestamp.proto";

service MessagingService {
  rpc CreateConversation(CreateConversationRequest) returns (CreateConversationResponse);
  rpc GetConversations(GetConversationsRequest) returns (GetConversationsResponse);
  rpc GetConversation(GetConversationRequest) returns (GetConversationResponse);
  
  rpc SendMessage(SendMessageRequest) returns (SendMessageResponse);
  rpc SendTrackMessage(SendTrackMessageRequest) returns (SendTrackMessageResponse);
  rpc GetMessages(GetMessagesRequest) returns (GetMessagesResponse);
  rpc SearchMessages(SearchMessagesRequest) returns (SearchMessagesResponse);
  
  rpc MarkAsRead(MarkAsReadRequest) returns (MarkAsReadResponse);
  rpc GetUnreadCount(GetUnreadCountRequest) returns (GetUnreadCountResponse);
  
  rpc GetContextMessages(GetContextMessagesRequest) returns (GetContextMessagesResponse);
}

message CreateConversationRequest {
  string user_id = 1; // текущий пользователь
  string other_user_id = 2;
}

message CreateConversationResponse {
  Conversation conversation = 1;
}

message GetConversationsRequest {
  string user_id = 1;
  int32 limit = 2;
  int32 offset = 3;
}

message GetConversationsResponse {
  repeated Conversation conversations = 1;
  int32 total = 2;
}

message Conversation {
  string id = 1;
  string other_user_id = 2;
  string other_username = 3;
  string other_avatar_url = 4;
  Message last_message = 5;
  int32 unread_count = 6;
  google.protobuf.Timestamp updated_at = 7;
}

message SendMessageRequest {
  string conversation_id = 1;
  string sender_id = 2;
  string content = 3;
}

message SendMessageResponse {
  Message message = 1;
}

message SendTrackMessageRequest {
  string conversation_id = 1;
  string sender_id = 2;
  string content = 3; // опционально
  string track_id = 4;
}

message SendTrackMessageResponse {
  Message message = 1;
}

message GetMessagesRequest {
  string conversation_id = 1;
  int32 limit = 2;
  string before_message_id = 3; // для пагинации
}

message GetMessagesResponse {
  repeated Message messages = 1;
  bool has_more = 2;
}

message Message {
  string id = 1;
  string conversation_id = 2;
  string sender_id = 3;
  string sender_username = 4;
  string sender_avatar_url = 5;
  string content = 6;
  TrackInfo track = 7; // если есть трек
  ContextInfo context = 8; // если есть контекст
  bool is_read = 9;
  google.protobuf.Timestamp created_at = 10;
}

message TrackInfo {
  string id = 1;
  string title = 2;
  repeated string artist_names = 3;
  string cover_url = 4;
  int32 duration_seconds = 5;
}

message ContextInfo {
  string type = 1; // 'playlist', 'session', 'track', 'route'
  string id = 2;
  string name = 3;
}

message MarkAsReadRequest {
  string conversation_id = 1;
  string user_id = 2;
  string message_id = 3; // опционально, если не указано - все до последнего
}

message MarkAsReadResponse {
  bool success = 1;
}

message GetUnreadCountRequest {
  string user_id = 1;
  string conversation_id = 2; // опционально
}

message GetUnreadCountResponse {
  int32 count = 1;
}

message GetContextMessagesRequest {
  string context_type = 1;
  string context_id = 2;
  int32 limit = 3;
  int32 offset = 4;
}

message GetContextMessagesResponse {
  repeated Message messages = 1;
  int32 total = 2;
}

message SearchMessagesRequest {
  string conversation_id = 1;
  string query = 2;
  int32 limit = 3;
}

message SearchMessagesResponse {
  repeated Message messages = 1;
}
```

---

## Интеграция с существующими сервисами

### Users Service

- **Проверка существования пользователей** при создании диалога
- **Получение информации о пользователях** для отображения в диалогах
- **Валидация JWT токенов** (через API Gateway)

### Tracks Service

- **Валидация track_id** при отправке треков
- **Получение метаданных треков** для отображения в сообщениях
- **Проверка доступности треков** для пользователя

### Playback Queue Service

- **Добавление треков в очередь** из сообщений
- **Создание контекстных очередей** для диалогов (опционально)

### WebSocket Gateway

- **Real-time доставка сообщений**
- **Уведомления о новых сообщениях**
- **Синхронизация статусов прочтения**

---

## Оценка времени и ресурсов

### Общая оценка

| Этап | Время | Приоритет |
|------|-------|-----------|
| Этап 1: Проектирование | 6-8 дней | Высокий |
| Этап 2: Базовый функционал | 12-16 дней | Высокий |
| Этап 3: Real-time | 8-11 дней | Высокий |
| Этап 4: Музыкальные контексты | 7-10 дней | Средний |
| Этап 5: Фронтенд | 10-14 дней | Высокий |
| Этап 6: Дополнительные функции | 9-12 дней | Низкий |
| Этап 7: Тестирование | 14-19 дней | Высокий |
| **ИТОГО** | **66-90 дней** | |

### Команда

- **Backend разработчик (Go)**: 1-2 человека
- **Backend разработчик (Python)**: 1 человек (для WebSocket Gateway)
- **Frontend разработчик (React)**: 1 человек
- **DevOps**: 0.5 человека (для настройки инфраструктуры)

---

## Риски и митигация

### Технические риски

1. **Производительность при большом количестве сообщений**
   - Митигация: Пагинация, индексы, кеширование, архивация старых сообщений

2. **Масштабируемость WebSocket соединений**
   - Митигация: Горизонтальное масштабирование, Redis для синхронизации

3. **Согласованность данных между сервисами**
   - Митигация: Event-driven архитектура, retry механизмы

### Бизнес-риски

1. **Сложность интеграции с существующей системой**
   - Митигация: Поэтапное внедрение, тщательное тестирование

2. **Изменение требований в процессе разработки**
   - Митигация: Гибкая архитектура, модульный дизайн

---

## Следующие шаги после MVP

1. **Групповые чаты** - расширение для общения нескольких пользователей
2. **Голосовые сообщения** - интеграция с аудио стримингом
3. **Реакции на сообщения** - эмодзи реакции
4. **Пересылка сообщений** - возможность переслать сообщение другому пользователю
5. **Закрепленные сообщения** - возможность закрепить важные сообщения
6. **Музыкальные рекомендации** - AI-рекомендации треков на основе общения
7. **Интеграция с социальными сетями** - импорт контактов, шаринг

---

## Заключение

Данный план представляет собой комплексный подход к внедрению социального функционала в платформу Music Social. Ключевые особенности:

- **Модульная архитектура** - новый сервис легко интегрируется с существующей системой
- **Real-time коммуникация** - использование существующей WebSocket инфраструктуры
- **Музыкальная ориентация** - глубокое интегрирование с музыкальными контекстами
- **Масштабируемость** - готовность к росту нагрузки

Рекомендуется начать с MVP (Этапы 1-5), протестировать на реальных пользователях, и затем добавить дополнительные функции из Этапа 6.

