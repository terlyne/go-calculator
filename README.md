# Калькулятор на Go

Сервис для вычисления математических выражений с поддержкой аутентификации и хранения истории вычислений.

## Возможности

- Регистрация и аутентификация пользователей с использованием JWT
- Вычисление математических выражений с поддержкой базовых арифметических операций
- Отслеживание истории вычислений для каждого пользователя
- gRPC API для эффективной коммуникации
- SQLite база данных для хранения данных
- Поддержка in-memory базы данных для тестирования

## Требования

- Go 1.21 или выше
- Docker и Docker Compose (для запуска через Docker)
- SQLite3 (для локальной разработки)

## Запуск через Docker

1. Соберите и запустите контейнер:
```bash
docker-compose up --build
```

Сервис будет доступен на следующих портах:
- gRPC: 50051
- HTTP: 8080

## Локальная разработка

1. Установите переменные окружения:
```bash
# Windows PowerShell
$env:CGO_ENABLED=1
$env:JWT_SECRET_KEY="your-secure-secret-key"
$env:DB_PATH="storage/storage.db"  # Опционально, по умолчанию используется in-memory база

# Windows CMD
set CGO_ENABLED=1
set JWT_SECRET_KEY=your-secure-secret-key
set DB_PATH=storage/storage.db
```

2. Запустите сервис:
```bash
go run ./cmd/calc_service/main.go
```

## Использование API

### Регистрация нового пользователя
```bash
grpcurl -plaintext -d '{"login": "user1", "password": "password123"}' \
    localhost:50051 calculator.Calculator/Register
```

### Вход в систему
```bash
grpcurl -plaintext -d '{"login": "user1", "password": "password123"}' \
    localhost:50051 calculator.Calculator/Login
```

### Вычисление выражения (HTTP API)
```bash
curl -X POST http://localhost:8080/api/v1/calculate \
    -H "Authorization: Bearer YOUR_JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"expression": "2+2*2"}'
```

### Вычисление выражения (gRPC API)
```bash
grpcurl -plaintext -d '{"expression": "2+2*2", "token": "YOUR_JWT_TOKEN"}' \
    localhost:50051 calculator.Calculator/Calculate
```

### Получение истории вычислений
```bash
grpcurl -plaintext -d '{"token": "YOUR_JWT_TOKEN"}' \
    localhost:50051 calculator.Calculator/GetExpressions
```

## Структура проекта

```
.
├── api/
│   └── calculator.proto        # Определение gRPC сервиса
├── cmd/
│   └── calc_service/
│       ├── main.go            # Основной файл сервиса
│       ├── main_test.go       # Модульные тесты
│       └── integration_test.go # Интеграционные тесты
├── internal/
│   ├── auth/
│   │   └── auth.go           # Аутентификация и JWT
│   ├── database/
│   │   └── database.go       # Работа с базой данных
│   └── models/
│       └── models.go         # Модели данных
├── pkg/
│   └── calculator/
│       └── calculator.go     # Логика вычислений
├── Dockerfile               # Конфигурация Docker
├── docker-compose.yml      # Конфигурация Docker Compose
└── README.md              # Документация
```

## Переменные окружения

- `CGO_ENABLED` - Включение поддержки CGO (требуется для SQLite)
- `JWT_SECRET_KEY` - Секретный ключ для JWT токенов
- `DB_PATH` - Путь к файлу базы данных SQLite (по умолчанию используется in-memory база)

## Тестирование

Запуск модульных тестов:
```bash
go test ./...
```

Запуск интеграционных тестов:
```bash
go test ./cmd/calc_service/integration_test.go
```

## Обработка ошибок

Сервис возвращает соответствующие gRPC коды статуса и сообщения об ошибках для различных сценариев:

- Неверные выражения
- Ошибки аутентификации
- Ошибки базы данных
- Неверные токены

## Безопасность

- Пароли хешируются с использованием bcrypt
- JWT токены используются для аутентификации
- Все чувствительные данные хранятся безопасно в базе данных
- В продакшене обязательно использовать переменную окружения JWT_SECRET_KEY

