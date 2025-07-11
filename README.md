# ozon-comments-graphql

Простая система постов и комментариев с GraphQL API. Поддерживаются как локальное in-memory хранилище, так и PostgreSQL. Сервис позволяет создавать посты, оставлять комментарии, работать с вложенными комментариями и подписываться на новые в реальном времени.

## Возможности

**Посты**
- Создание постов
- Просмотр списка постов
- Возможность включения или отключения комментариев автором

**Комментарии**
- Вложенные (иерархические) комментарии
- Ограничение длины комментария до 2000 символов
- Пагинация при получении комментариев

**Real-time обновления**
- Поддержка подписок (GraphQL Subscriptions) на новые комментарии

## Технологии

- Go (без фреймворков, чистый код)
- GraphQL (на базе gqlgen)
- PostgreSQL или in-memory хранилище
- Docker для контейнеризации и запуска

## Варианты запуска сервера

### 1. Локальный запуск (in-memory хранилище)

```bash
go run server.go
```

Сервер запускается с in-memory хранилищем (без БД).

---

### 2. Локальный запуск с PostgreSQL

```bash
DATABASE_URL="postgres://user:password@localhost:5432/dbname?sslmode=disable" 
go run server.go
```

Требуется поднятый локальный PostgreSQL и указанный URL в переменной окружения `DATABASE_URL`.

---

### 3. Запуск через Docker (in-memory)

```bash
docker-compose -f docker-compose.memory.yml up --build
```

Сервер запускается с in-memory хранилищем.

---

### 4. Запуск через Docker (с PostgreSQL)

```bash
docker-compose -f docker-compose.postgres.yml up --build
```

Запускается контейнер с сервером и PostgreSQL.

---

Сервер доступен по адресу: `http://localhost:8080/graphql`
