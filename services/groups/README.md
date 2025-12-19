# Groups Service

## Возможности

- Создание группы владельцем и генерация приглашения
- Присоединение по ссылке/коду приглашения.
- Отправка предложений треков участниками (до 10 активных предложений на человека в каждой группе).
- Владелец может принять предложение и отправить трек в очередь или отклонить с кулдауном (по умолчанию 30 минут) до повторной отправки того же трека.
- Интеграция с внешним playback-queue сервисом для ведения очереди треков группы.

## Переменные окружения

| Переменная                    | Описание                                                                         | Значение по умолчанию                                                                   |
| ----------------------------- | -------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------- |
| `DATABASE_URL`                | строка подключения к Postgres                                                    | `postgres://postgres:password@postgres-groups:5432/music_social_groups?sslmode=disable` |
| `GRPC_PORT`                   | gRPC порт (`groups.v1.GroupsService`)                                            | `9095`                                                                                  |
| `SUGGESTION_LIMIT`            | максимально разрешённое количество активных предложений от пользователя в группе | `10`                                                                                    |
| `SUGGESTION_COOLDOWN_SECONDS` | длительность кулдауна после отклонения                                           | `1800`                                                                                  |
| `GROUP_LINK_BASE`             | базовый URL для генерации invite-ссылки                                          | `https://music.local/groups`                                                            |
| `QUEUE_SERVICE_ADDR`          | адрес gRPC сервиса playback-queue                                                | `playback-queue-service:50056`                                                          |

## API

Сервис предоставляет только gRPC-интерфейс `groups.v1.GroupsService` (описан в `services/groups/api/groups.proto`). Ниже перечислены основные RPC и примеры `grpcurl`.

### CreateGroup

Создаёт группу и сразу возвращает invite-ссылку.

```
grpcurl -plaintext \
  -d '{"owner_id":"<uuid>","name":"Friday Night"}' \
  localhost:9095 groups.v1.GroupsService/CreateGroup
```

### JoinGroup

Присоединяет участника по invite-коду.

```
grpcurl -plaintext \
  -d '{"user_id":"<uuid>","invite_code":"abc123xyz"}' \
  localhost:9095 groups.v1.GroupsService/JoinGroup
```

### LeaveGroup

Участник (кроме владельца) покидает группу.

```
grpcurl -plaintext \
  -d '{"group_id":"<uuid>","user_id":"<uuid>"}' \
  localhost:9095 groups.v1.GroupsService/LeaveGroup
```

### DeleteGroup

Удаление группы владельцем (каскадно очищает данные группы в своей БД; очередь лежит во внешнем playback-queue сервисе).

```
grpcurl -plaintext \
  -d '{"group_id":"<uuid>","owner_id":"<owner uuid>"}' \
  localhost:9095 groups.v1.GroupsService/DeleteGroup
```

### GetGroup

Возвращает информацию о группе и актуальную invite-ссылку.

```
grpcurl -plaintext -d '{"group_id":"<uuid>"}' localhost:9095 groups.v1.GroupsService/GetGroup
```

### SubmitSuggestion

Отправляет предложение трека от участника (учитываются лимит и кулдаун).

```
grpcurl -plaintext \
  -d '{"group_id":"<uuid>","user_id":"<uuid>","track_id":"<track uuid>"}' \
  localhost:9095 groups.v1.GroupsService/SubmitSuggestion
```

### ListSuggestions

Доступно владельцу: возвращает предложения по статусу с пагинацией.

```
grpcurl -plaintext \
  -d '{"group_id":"<uuid>","owner_id":"<owner uuid>","status":"pending","limit":20}' \
  localhost:9095 groups.v1.GroupsService/ListSuggestions
```

### AcceptSuggestion / RejectSuggestion

Владелец подтверждает или отклоняет предложение; при подтверждении трек отправляется во внешний playback-queue сервис.

```
grpcurl -plaintext \
  -d '{"group_id":"<uuid>","owner_id":"<owner uuid>","suggestion_id":"<uuid>"}' \
  localhost:9095 groups.v1.GroupsService/AcceptSuggestion

grpcurl -plaintext \
  -d '{"group_id":"<uuid>","owner_id":"<owner uuid>","suggestion_id":"<uuid>","reason":"duplicate"}' \
  localhost:9095 groups.v1.GroupsService/RejectSuggestion
```

### ListQueue

Возвращает актуальную очередь группы.

```
grpcurl -plaintext -d '{"group_id":"<uuid>","limit":25}' localhost:9095 groups.v1.GroupsService/ListQueue
```

### Health Check

Стандартная gRPC Health служба доступна по `grpc.health.v1.Health/Check`.
