# ozon-comments-graphql

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
