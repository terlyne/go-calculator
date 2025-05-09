# Distributed Calculator

Распределенный калькулятор с поддержкой аутентификации пользователей, вычислением выражений и отслеживанием истории.

## Возможности

- Регистрация и аутентификация пользователей с использованием JWT
- Вычисление математических выражений с поддержкой базовых арифметических операций
- Отслеживание истории вычислений для каждого пользователя
- gRPC API для эффективной коммуникации
- SQLite база данных для хранения данных
- Поддержка in-memory базы данных для тестирования

## Требования

- Go 1.23 или выше
- Protocol Buffers compiler (protoc)
- Go плагины для protoc
- GCC (для работы с SQLite)

## Установка

1. Клонируйте репозиторий:
```bash
git clone https://github.com/yourusername/go-calculator.git
cd go-calculator
```

2. Установите зависимости:
```bash
go mod download
```

3. Установите Protocol Buffers compiler:
```bash
# Windows (Chocolatey)
choco install protoc

# Или скачайте вручную с https://github.com/protocolbuffers/protobuf/releases
```

4. Установите Go плагины для protoc:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

5. Сгенерируйте gRPC код:
```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/calculator.proto
```

## Запуск сервиса

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

Сервис запустится на порту 50051.

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

### Вычисление выражения
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
│   │   └── database.go       # Работа с SQLite
│   └── models/
│       └── user.go           # Модели данных
├── pkg/
│   └── calculator/
│       └── calculator.go      # Логика вычислений
├── storage/
│   └── storage.db            # База данных SQLite
├── go.mod
├── go.sum
└── README.md
```

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

## Вклад в проект

1. Форкните репозиторий
2. Создайте ветку для вашей функциональности
3. Зафиксируйте изменения
4. Отправьте изменения в ветку
5. Создайте Pull Request