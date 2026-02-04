package blerrors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WriteErrorResponse отправляет JSON ответ с ошибкой
func WriteErrorResponse(w http.ResponseWriter, err *AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(getHTTPStatusCode(err.Code))

	response := ErrorResponse{
		Success: false,
		Error:   err,
	}

	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		// Fallback в случае ошибки кодирования
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// WriteSuccessResponse отправляет JSON ответ с успешными данными
func WriteSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := SuccessResponse{
		Success: true,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

// getHTTPStatusCode возвращает HTTP статус код для кода ошибки
func getHTTPStatusCode(code ErrorCode) int {
	switch code {
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeValidation, ErrCodeBadRequest:
		return http.StatusBadRequest
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeConflict:
		return http.StatusConflict
	case ErrCodeTooManyRequests:
		return http.StatusTooManyRequests
	case ErrCodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrCodeInternal:
		fallthrough
	default:
		return http.StatusInternalServerError
	}
}

// ErrorRecoveryMiddleware middleware для восстановления от паник
func ErrorRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Создаем ошибку для паники
				appErr := NewAppError(ErrCodeInternal, "Internal server error").IsCritical().WithErrorCode("FS001").IsSystemError().WithUserID("system").
					WithDetails(fmt.Sprintf("Panic recovered: %v", err)).
					WithModule("middleware")

				WriteErrorResponse(w, appErr)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RequestIDMiddleware добавляет request ID к контексту
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Генерируем простой request ID (в продакшене используйте UUID)
		requestID := fmt.Sprintf("%d", time.Now().UnixNano())

		// Добавляем в контекст
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// GetRequestID извлекает request ID из контекста
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}
