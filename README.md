<h1 align="center"> Calculator Service </h1>  

**Calculator Service** — веб-сервис для вычисления арифметических выражений.  

## Возможности  
- Выполнение базовых арифметических операций: `+`, `-`, `*`, `/`.  
- Работа со скобками и вещественными числами.  
- Обработка ошибок (некорректные выражения, деление на ноль и др.).  

---

## Установка  

1. Клонируйте проект:  
   ```bash
   git clone https://github.com/terlyne/go-calculator.git
   cd go-calculator

2. Установите зависимости:
    ```bash
    go mod tidy

3. Запустите сервис:
    ```bash
    go run ./cmd/go-calculator

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
```
### Формат ответа
Успешный запрос:
```json
{
    "result": 6
}
```
Ошибка 422 (некорректное выражение):
```json
{
  "error": "Expression is not valid"
}
```
Ошибка 500 (внутренняя ошибка):
```json
{
  "error": "Internal server error"
}
```

