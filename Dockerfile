# Используем официальный образ Go
FROM golang:1.21-alpine

# Устанавливаем необходимые зависимости
RUN apk add --no-cache gcc musl-dev protobuf-dev

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Генерируем proto файлы (используем конкретные версии)
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0

# Генерируем proto файлы
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/calculator.proto

# Обновляем зависимости
RUN go mod tidy
RUN go mod download

# Устанавливаем переменные окружения для сборки SQLite
ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE -D_FILE_OFFSET_BITS=64"

# Собираем приложение
RUN go build -o /app/calculator ./cmd/calc_service/main.go

# Открываем порты
EXPOSE 50051 8080

# Запускаем приложение
CMD ["/app/calculator"] 