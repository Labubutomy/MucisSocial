# API Gateway Documentation

## Обзор

API Gateway - это HTTP-REST интерфейс для взаимодействия с gRPC микросервисами. Он обеспечивает:
- Трансляцию HTTP запросов в gRPC вызовы
- JWT аутентификацию и авторизацию
- CORS поддержку
- Централизованную обработку ошибок

## Архитектура

```
HTTP Client → API Gateway → gRPC Service
     ↓             ↓             ↓
  REST API    HTTP-to-gRPC   Proto Interface
              Translation
```

## Текущие сервисы

### Users Service
- **gRPC адрес**: `users-service:50051`
- **Proto файлы**: `proto/users/v1/`
- **Endpoints**:
  - `POST /api/v1/auth/sign-up` - Регистрация
  - `POST /api/v1/auth/sign-in` - Авторизация
  - `POST /api/v1/auth/refresh` - Обновление токена
  - `GET /api/v1/me` - Получение профиля (защищенный)
  - `PUT /api/v1/me` - Обновление профиля (защищенный)
  - `GET /api/v1/me/search-history` - История поиска (защищенный)
  - `POST /api/v1/me/search-history` - Добавить в историю (защищенный)
  - `DELETE /api/v1/me/search-history` - Очистить историю (защищенный)

## Порядок развертывания

### 1. Базовая инфраструктура
```bash
# Запуск PostgreSQL
cd /path/to/MucisSocial/services
docker compose up -d postgres

# Ждем готовности БД
docker compose ps postgres
# Статус должен быть "healthy"
```

### 2. gRPC сервисы
```bash
# Запуск всех gRPC сервисов
docker compose up -d users-service

# Проверка готовности
docker compose ps users-service
# Статус должен быть "healthy"

# Проверка gRPC напрямую (опционально)
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext localhost:50051 users.v1.UserService/Ping
```

### 3. API Gateway
```bash
# Запуск API Gateway
docker compose up -d api-gateway

# Проверка готовности
curl http://localhost:8080/__health
# Ответ: {"status":"ok"}

# Проверка всех сервисов
docker compose ps
```

### 4. Полная проверка системы
```bash
# Тест регистрации
curl -X POST http://localhost:8080/api/v1/auth/sign-up \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "username": "testuser",
    "displayName": "Test User"
  }'

# Тест авторизации
curl -X POST http://localhost:8080/api/v1/auth/sign-in \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

## Добавление нового сервиса

### Шаг 1: Подготовка proto файлов

1. Скопируйте proto файлы нового сервиса:
```bash
mkdir -p proto/service_name/v1/
cp /path/to/service/proto/* proto/service_name/v1/
```

2. Сгенерируйте Go код:
```bash
# В директории gateway
go mod download
protoc --go_out=. --go-grpc_out=. proto/service_name/v1/*.proto
```

### Шаг 2: Обновление Gateway кода

1. **Добавьте import для нового сервиса** в `main.go`:
```go
import (
    // ... существующие импорты
    usersPb "github.com/MucisSocial/api-gateway/proto/users/v1"
    servicenamePb "github.com/MucisSocial/api-gateway/proto/servicename/v1"
)
```

2. **Расширьте структуру Gateway**:
```go
type Gateway struct {
    userClient       usersPb.UserServiceClient
    servicenameClient servicenamePb.ServiceNameClient // Добавить
    jwtSecret        []byte
}
```

3. **Подключение к gRPC сервису** в `main()`:
```go
func main() {
    // ... существующее подключение к users-service
    
    // Подключение к новому сервису
    servicenameConn, err := grpc.Dial(
        "servicename-service:50052", // Убедитесь что порт правильный
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        log.Fatalf("Failed to connect to servicename gRPC service: %v", err)
    }
    defer servicenameConn.Close()

    gateway := &Gateway{
        userClient:        usersPb.NewUserServiceClient(conn),
        servicenameClient: servicenamePb.NewServiceNameClient(servicenameConn), // Добавить
        jwtSecret:         []byte(getEnv("JWT_SECRET", "your-secret")),
    }
    
    // ... остальной код
}
```

4. **Добавьте роуты для нового сервиса**:
```go
// В main() после существующих роутов

// Публичные эндпоинты (если нужны)
r.HandleFunc("/api/v1/servicename/public-endpoint", gateway.serviceNamePublicHandler).Methods("POST", "OPTIONS")

// Защищенные эндпоинты
protected.HandleFunc("/servicename/endpoint", gateway.serviceNameHandler).Methods("GET", "OPTIONS")
```

### Шаг 3: Реализация обработчиков

```go
// Пример обработчика для нового сервиса
func (g *Gateway) serviceNameHandler(w http.ResponseWriter, r *http.Request) {
    // Получение claims из JWT (для защищенных эндпоинтов)
    claims, ok := r.Context().Value("claims").(*jwt.MapClaims)
    if !ok {
        writeError(w, "Failed to get user claims", http.StatusUnauthorized)
        return
    }
    userID := (*claims)["user_id"].(string)

    // Парсинг входных данных (для POST/PUT запросов)
    var req RequestType
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // gRPC вызов
    resp, err := g.servicenameClient.SomeMethod(r.Context(), &servicenamePb.SomeRequest{
        UserId: userID,
        // ... другие поля
    })
    if err != nil {
        handleGRPCError(w, err)
        return
    }

    // Ответ
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}
```

### Шаг 4: Обновление docker-compose.yml

Добавьте новый сервис в `services/docker-compose.yml`:

```yaml
services:
  # ... существующие сервисы
  
  servicename-service:
    build:
      context: ./servicename
      dockerfile: Dockerfile
    ports:
      - "50052:50052"  # Убедитесь что порт уникальный
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=music_social
      - DB_USER=user
      - DB_PASSWORD=password
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - music-network
    healthcheck:
      test: ["CMD", "grpcurl", "-plaintext", "localhost:50052", "servicename.v1.ServiceNameService/Ping"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s

  api-gateway:
    # ... существующая конфигурация
    depends_on:
      # ... существующие зависимости
      servicename-service:
        condition: service_healthy
```

### Шаг 5: Обновление Dockerfile

Если добавили новые зависимости в `go.mod`, пересоберите образ:
```bash
docker compose build api-gateway
```

### Шаг 6: Тестирование

1. **Развертывание**:
```bash
# Остановка старых контейнеров
docker compose down

# Запуск с новым сервисом
docker compose up -d postgres
docker compose up -d users-service servicename-service
docker compose up -d api-gateway

# Проверка статуса
docker compose ps
```

2. **Тестирование endpoints**:
```bash
# Тест нового endpoint'а
curl -X POST http://localhost:8080/api/v1/servicename/endpoint \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"some": "data"}'
```

## Структура проекта

```
gateway/
├── main.go              # Основной файл с роутингом
├── Dockerfile           # Контейнеризация
├── go.mod              # Go зависимости
├── go.sum              # Checksums зависимостей
├── .dockerignore       # Исключения для Docker
├── README.md           # Данная документация
└── proto/              # Proto файлы и сгенерированный код
    ├── users/
    │   └── v1/
    │       ├── user_service.proto
    │       ├── user_service.pb.go
    │       └── user_service_grpc.pb.go
    └── servicename/    # Добавляется при интеграции нового сервиса
        └── v1/
            ├── servicename.proto
            ├── servicename.pb.go
            └── servicename_grpc.pb.go
```

## Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `PORT` | Порт для HTTP сервера | `8080` |
| `JWT_SECRET` | Секретный ключ для JWT | `your-super-secret...` |

## Мониторинг и отладка

### Health Check
```bash
curl http://localhost:8080/__health
```

### Логи сервиса
```bash
docker compose logs -f api-gateway
```

### Проверка gRPC соединений
```bash
# Прямой тест gRPC сервисов
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext localhost:50052 list

# Тест внутри Docker сети
docker exec -it music-social-gateway sh
grpcurl -plaintext users-service:50051 list
```

## Troubleshooting

### Частые проблемы

1. **gRPC соединение не устанавливается**
   - Проверьте что gRPC сервис запущен и healthy
   - Проверьте правильность адреса и порта
   - Убедитесь что сервисы в одной Docker сети

2. **404 на endpoint'ах**
   - Проверьте правильность путей в роутере
   - Убедитесь что HTTP методы совпадают
   - Проверьте middleware на защищенных роутах

3. **JWT ошибки**
   - Проверьте что JWT_SECRET одинаковый везде
   - Убедитесь что токен передается в заголовке Authorization
   - Проверьте формат: `Bearer <token>`

4. **Proto generation ошибки**
   - Убедитесь что установлен protoc
   - Проверьте что все import'ы в proto файлах корректные
   - Запустите `go mod tidy` после генерации

### Логи и диагностика

```bash
# Все сервисы
docker compose logs

# Конкретный сервис
docker compose logs api-gateway

# Следить за логами в реальном времени
docker compose logs -f api-gateway

# Статус контейнеров
docker compose ps

# Статистика ресурсов
docker stats
```