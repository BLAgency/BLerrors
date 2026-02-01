# BL Errors

Универсальный модуль для обработки ошибок в Go микросервисах.

## Установка

```bash
go get github.com/blackboxai/blerrors
```

## Быстрый старт

### 1. Импорт модуля

```go
import (
    "github.com/blackboxai/blerrors"
)
```

### 2. Использование в HTTP handlers

```go
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    userID := mux.Vars(r)["id"]

    user, err := h.userService.GetUser(userID)
    if err != nil {
        // Создаем ошибку с кодом и сообщением
        appErr := blerrors.NewAppError(
            blerrors.ErrCodeNotFound,
            "Пользователь не найден",
        ).WithModule("user-service")

        // Отправляем JSON ответ с ошибкой
        blerrors.WriteErrorResponse(w, appErr)
        return
    }

    // Отправляем успешный ответ
    blerrors.WriteSuccessResponse(w, user)
}
```

### 3. Использование middleware

```go
func main() {
    r := mux.NewRouter()

    // Добавляем middleware для обработки ошибок
    r.Use(blerrors.ErrorRecoveryMiddleware)
    r.Use(blerrors.RequestIDMiddleware)

    // Ваши маршруты...
    r.HandleFunc("/users/{id}", GetUserHandler).Methods("GET")

    http.ListenAndServe(":8080", r)
}
```

## Структура ошибки

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Пользователь не найден",
    "timestamp": 1703123456,
    "module": "user-service",
    "trace": [
      "path/to/file.go:42 main.Handler.GetUser",
      "path/to/file.go:28 main.main"
    ],
    "request_id": "1703123456789123456"
  }
}
```

## API

### Создание ошибок

```go
// Базовая ошибка
err := blerrors.NewAppError(blerrors.ErrCodeValidation, "Неверные данные")

// С дополнительными полями
err := blerrors.NewAppError(blerrors.ErrCodeNotFound, "Пользователь не найден").
    WithModule("user-service").
    WithDetails(map[string]interface{}{"user_id": 123}).
    WithRequestID("req-123")
```

### Предопределенные коды ошибок

- `ErrCodeNotFound` - ресурс не найден (404)
- `ErrCodeValidation` - ошибка валидации (400)
- `ErrCodeInternal` - внутренняя ошибка сервера (500)
- `ErrCodeUnauthorized` - не авторизован (401)
- `ErrCodeForbidden` - доступ запрещен (403)
- `ErrCodeBadRequest` - неверный запрос (400)
- `ErrCodeConflict` - конфликт данных (409)
- `ErrCodeTooManyRequests` - слишком много запросов (429)
- `ErrCodeServiceUnavailable` - сервис недоступен (503)

### HTTP ответы

```go
// Ошибка
blerrors.WriteErrorResponse(w, appErr)

// Успех
blerrors.WriteSuccessResponse(w, data)
```

### Middleware

```go
// Обработка паник
r.Use(blerrors.ErrorRecoveryMiddleware)

// Добавление request ID
r.Use(blerrors.RequestIDMiddleware)
```

## Интеграция в существующий проект

### 1. Добавьте зависимость

```bash
go get github.com/blackboxai/blerrors
```

### 2. Замените текущую обработку ошибок

**Было:**
```go
if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}
```

**Стало:**
```go
if err != nil {
    appErr := blerrors.NewAppError(blerrors.ErrCodeInternal, "Внутренняя ошибка").
        WithModule("your-service")
    blerrors.WriteErrorResponse(w, appErr)
    return
}
```

### 3. Добавьте middleware в main.go

```go
r := mux.NewRouter()
r.Use(blerrors.ErrorRecoveryMiddleware)
r.Use(blerrors.RequestIDMiddleware)
```

## Особенности

- **Уникальные коды ошибок** для стандартизации
- **Автоматический стек вызовов** для отладки
- **JSON формат** для API
- **Middleware** для централизованной обработки
- **Безопасность** - можно отключать trace в продакшене
- **Расширяемость** - легко добавить новые коды ошибок

## Конфигурация

### Отключение стека в продакшене

```go
appErr := blerrors.NewAppError(blerrors.ErrCodeInternal, "Ошибка").
    WithoutTrace()
```

### Кастомные коды ошибок

```go
const (
    ErrCodeCustom ErrorCode = "CUSTOM_ERROR"
)
```

## Лицензия

MIT License
