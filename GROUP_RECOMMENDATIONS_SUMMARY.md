# Group Recommendations Feature - Summary

## Что реализовано

Реализована система группового матчинга рекомендаций для пользователей, слушающих музыку в одной комнате.

## Основные компоненты

### 1. Модели данных ([models.go](services/recommendations/internal/models/models.go))
- **GroupProfile** - агрегированный профиль группы пользователей
- **AggregateProfiles** - функция объединения профилей пользователей
- Методы для анализа группового профиля (GetTopGenres, GetTopArtists, HasListenedByMajority)

### 2. Скоринг ([group_scorer.go](services/recommendations/internal/engine/scorers/group_scorer.go))
- **GroupScorer** - специализированный алгоритм оценки для групп
- Учитывает баланс предпочтений всех участников
- Формула скоринга:
  ```
  Score = GenreAffinity * 3.0 +
          ArtistAffinity * 2.5 +
          GlobalPopularity * 1.5 +
          Freshness * 0.5 +
          Diversity * 2.0
  ```

### 3. Движок рекомендаций ([engine.go](services/recommendations/internal/engine/engine.go))
- **RecommendForGroup** - метод генерации групповых рекомендаций
- Создание синтетического профиля на основе группы
- Фильтрация треков, знакомых большинству

### 4. API ([server.go](services/recommendations/internal/api/server.go))
- **POST /recommendations/group** - endpoint для групповых рекомендаций
- Валидация (макс. 50 пользователей, макс. 100 треков)
- Поддержка фильтров (explicit, genres, fresh_days)

### 5. Тесты ([group_test.go](services/recommendations/internal/models/group_test.go))
- Тесты агрегации профилей
- Тесты определения большинства
- Все тесты проходят успешно

## Пример использования

```bash
# Запрос групповых рекомендаций для 3 пользователей
curl -X POST http://localhost:8080/recommendations/group \
  -H "Content-Type: application/json" \
  -d '{
    "user_ids": ["user1", "user2", "user3"],
    "limit": 20,
    "filters": {
      "exclude_explicit": true
    }
  }'
```

## Как это работает

1. **Агрегация**: Объединяем профили всех пользователей в GroupProfile
2. **Генерация кандидатов**: Используем агрегированный профиль для поиска треков
3. **Фильтрация**: Убираем треки, которые слышало >50% группы
4. **Скоринг**: Оцениваем треки с учетом групповых предпочтений
5. **Ранжирование**: Сортируем и возвращаем топ-N треков

## Преимущества

✅ **Справедливость**: Учитывает вкусы всех участников, а не только одного
✅ **Разнообразие**: Поощряет треки, новые для большинства группы
✅ **Баланс**: Находит компромисс между популярными и нишевыми треками
✅ **Гибкость**: Поддерживает различные фильтры и ограничения
✅ **Масштабируемость**: Работает с группами до 50 человек

## Файлы изменений

- `services/recommendations/internal/models/models.go` - добавлены модели GroupProfile
- `services/recommendations/internal/engine/scorers/group_scorer.go` - новый скорер
- `services/recommendations/internal/engine/engine.go` - метод RecommendForGroup
- `services/recommendations/internal/api/server.go` - новый endpoint
- `services/recommendations/cmd/main.go` - обновлена инициализация engine
- `services/recommendations/internal/models/group_test.go` - тесты
- `services/recommendations/GROUP_RECOMMENDATIONS.md` - полная документация

## Следующие шаги

Для интеграции с сервисом сессий:
1. Добавить вызов `/recommendations/group` при создании/обновлении плейлиста комнаты
2. Передавать список `participants.userId` из комнаты
3. Использовать полученные треки для автоматического пополнения очереди

## Тестирование

```bash
# Тесты моделей
cd services/recommendations
go test ./internal/models/... -v

# Компиляция
go build ./...
```

Все тесты проходят успешно ✅
