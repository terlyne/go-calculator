# Calculator Service  

**Calculator Service** — веб-сервис для вычисления арифметических выражений.  

## Возможности  
- Выполнение базовых арифметических операций: `+`, `-`, `*`, `/`.  
- Работа со скобками и вещественными числами.  
- Обработка ошибок (некорректные выражения, деление на ноль и др.).  
- Поддержка распределённой архитектуры с использованием оркестратора и агентов.

---

## Установка  

1. Клонируйте проект:  
   ```bash
   git clone https://github.com/terlyne/go-calculator.git
   cd go-calculator

2. Установите зависимости:
    ```bash
    go mod tidy

3. Запустите оркестратор:
    ```bash
    go run ./cmd/calc_service/main.go

4. Запустите агента в отдельном терминале:
    ```bash
    go run ./cmd/agent/main.go

---

## Использование

Для проверки работы сервиса можно использовать:
1. curl. Пример команды:
    ```bash
    curl --location 'http://localhost:8080/api/v1/calculate' \
    --header 'Content-Type: application/json' \
    --data '{"expression": "2+2*2"}'
2. Postman:
    * Установите метод POST.
    * Введите URL: http://localhost:8080/api/v1/calculate.
    * Укажите тело запроса:
        ```json
        {
            "expression": "2+2*2"
        }
    * Убедитесь, что заголовок запроса содержит:
        ```bash 
        Content-Type: application/json

---

### Формат запроса
    ```json
    {
        "expression": "2+2*2"
    }
### Формат ответа
Успешный запрос:
    ```json
    {
        "id": "expr_1"
    }
### Получение результата
После добавления выражения, вы можете получить его статус и результат, используя следующий запрос:
    ```bash
    curl --location 'http://localhost:8080/api/v1/expressions'
Формат ответа:
    ```json
    {
        "expressions": [
            {
                "id": "expr_1",
                "status": "completed",
                "result": "6"
            }
        ]
    }
### Ошибки
Ошибка 422 (некорректное выражение):
    ```json
    {
        "error": "Expression is not valid"
    }
Ошибка 500 (внутренняя ошибка):
    ```json
    {
        "error": "Internal server error"
    }
## Тестирование
Для запуска тестов используйте команду:
    ```bash
    go test ./...