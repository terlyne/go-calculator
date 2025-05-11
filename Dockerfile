# Используем официальный образ Go
FROM golang:1.21

# Устанавливаем необходимые зависимости
RUN apt-get update && apt-get install -y \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Обновляем зависимости и собираем приложение
RUN go mod tidy && \
    CGO_ENABLED=1 GOOS=linux go build -o main ./cmd/calc_service

# Создаем директорию для хранения базы данных
RUN mkdir -p /app/storage

# Запускаем приложение
CMD ["./main"] 