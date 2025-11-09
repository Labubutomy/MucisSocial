# API Examples - Music Social Gateway

## Примеры HTTP запросов для тестирования API

### 1. Health Check

```bash
curl -X GET http://localhost:8080/__health
```

**Ответ:**
```json
{"status":"ok"}
```

### 2. Регистрация пользователя

```bash
curl -X POST http://localhost:8080/api/v1/auth/sign-up \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepassword123",
    "username": "johndoe",
    "displayName": "John Doe"
  }'
```

**Ответ при успехе:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid-here",
    "username": "johndoe",
    "email": "john@example.com",
    "created_at": {...},
    "updated_at": {...}
  }
}
```

**Ответ при ошибке:**
```json
{
  "error": "registration failed: email already exists",
  "code": 400,
  "message": "registration failed: email already exists"
}
```

### 3. Вход в систему

```bash
curl -X POST http://localhost:8080/api/v1/auth/sign-in \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepassword123"
  }'
```

**Ответ при успехе:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid-here",
    "username": "johndoe",
    "email": "john@example.com",
    "created_at": {...},
    "updated_at": {...}
  }
}
```

### 4. Обновление токена

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "your_refresh_token_here"
  }'
```

**Ответ:**
```json
{
  "access_token": "new_access_token_here",
  "refresh_token": "new_refresh_token_here"
}
```

### 5. Получение профиля пользователя (защищенный endpoint)

```bash
# Замените YOUR_ACCESS_TOKEN на реальный токен из ответа sign-in/sign-up
curl -X GET http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Ответ:**
```json
{
  "id": "uuid-here",
  "username": "johndoe", 
  "email": "john@example.com",
  "displayName": "John Doe",
  "bio": "",
  "avatarUrl": "",
  "isVerified": false,
  "created_at": {...},
  "updated_at": {...}
}
```

### 6. Обновление профиля пользователя

```bash
curl -X PUT http://localhost:8080/api/v1/me \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "displayName": "John Smith",
    "bio": "Music lover and producer",
    "avatarUrl": "https://example.com/avatar.jpg"
  }'
```

**Ответ:**
```json
{
  "id": "uuid-here",
  "username": "johndoe",
  "email": "john@example.com", 
  "displayName": "John Smith",
  "bio": "Music lover and producer",
  "avatarUrl": "https://example.com/avatar.jpg",
  "isVerified": false,
  "created_at": {...},
  "updated_at": {...}
}
```

### 7. Получение истории поиска

```bash
curl -X GET http://localhost:8080/api/v1/me/search-history \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Ответ:**
```json
{
  "searches": [
    {
      "query": "rock music",
      "timestamp": {...}
    },
    {
      "query": "jazz artists", 
      "timestamp": {...}
    }
  ]
}
```

### 8. Добавление записи в историю поиска

```bash
curl -X POST http://localhost:8080/api/v1/me/search-history \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "query": "classical music composers"
  }'
```

**Ответ:**
```json
{
  "message": "Search history updated successfully"
}
```

### 9. Очистка истории поиска

```bash
curl -X DELETE http://localhost:8080/api/v1/me/search-history \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Ответ:**
```json
{
  "message": "Search history cleared successfully"
}
```

## Общие ошибки

### 401 Unauthorized - отсутствует или неверный JWT токен

```json
{
  "error": "Missing Authorization header",
  "code": 401,
  "message": "Missing Authorization header"
}
```

```json
{
  "error": "Invalid token",
  "code": 401, 
  "message": "Invalid token"
}
```

### 400 Bad Request - неверный формат данных

```json
{
  "error": "Invalid request body",
  "code": 400,
  "message": "Invalid request body"
}
```

### 500 Internal Server Error - ошибка сервера

```json
{
  "error": "Internal server error", 
  "code": 500,
  "message": "Internal server error"
}
```

## Использование в разных средах

### Postman
1. Создайте новую коллекцию
2. Добавьте переменную `baseUrl` = `http://localhost:8080`
3. Добавьте переменную `accessToken` для хранения JWT токена
4. Импортируйте запросы из примеров выше

### JavaScript/Node.js
```javascript
const API_BASE = 'http://localhost:8080';

// Регистрация
const signUp = async (userData) => {
  const response = await fetch(`${API_BASE}/api/v1/auth/sign-up`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(userData)
  });
  return response.json();
};

// Получение профиля
const getProfile = async (accessToken) => {
  const response = await fetch(`${API_BASE}/api/v1/me`, {
    headers: {
      'Authorization': `Bearer ${accessToken}`
    }
  });
  return response.json();
};
```

### Python
```python
import requests

API_BASE = 'http://localhost:8080'

def sign_in(email, password):
    response = requests.post(f'{API_BASE}/api/v1/auth/sign-in', json={
        'email': email,
        'password': password
    })
    return response.json()

def get_profile(access_token):
    response = requests.get(f'{API_BASE}/api/v1/me', headers={
        'Authorization': f'Bearer {access_token}'
    })
    return response.json()
```

## Тестирование последовательности

Для полного тестирования API выполните запросы в следующем порядке:

1. **Health Check** - убедитесь что сервис работает
2. **Sign Up** - зарегистрируйте нового пользователя
3. **Sign In** - войдите с теми же данными
4. **Get Profile** - получите профиль с полученным токеном  
5. **Update Profile** - обновите информацию профиля
6. **Add Search History** - добавьте записи в историю поиска
7. **Get Search History** - получите историю поиска
8. **Clear Search History** - очистите историю
9. **Refresh Token** - обновите токен доступа

Это поможет убедиться что вся система работает корректно от начала до конца.