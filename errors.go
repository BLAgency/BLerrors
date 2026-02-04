package blerrors

import (
	"fmt"
	"runtime"
	"time"
)

// Error реализует интерфейс error
func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewAppError создает новую ошибку приложения
func NewAppError(code ErrorCode, message, priority, errorCode, errorType, userID string) *AppError {
	now := time.Now()
	return &AppError{
		Code:              code,
		Message:           message,
		Timestamp:         now.Unix(),
		HumanReadableTime: now.Format("2006-01-02 15:04:05"),
		Trace:             getStackTrace(),
		Module:            getCurrentModule(),
		Priority:          priority,
		ErrorCode:         errorCode,
		ErrorType:         errorType,
		UserID:            userID,
	}
}

// WithModule устанавливает модуль ошибки
func (e *AppError) WithModule(module string) *AppError {
	e.Module = module
	return e
}

// WithDetails добавляет детали к ошибке
func (e *AppError) WithDetails(details interface{}) *AppError {
	e.Details = details
	return e
}

// WithRequestID добавляет ID запроса
func (e *AppError) WithRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

// WithoutTrace возвращает ошибку без стека вызовов (для продакшена)
func (e *AppError) WithoutTrace() *AppError {
	e.Trace = nil
	return e
}

// getStackTrace получает стек вызовов
func getStackTrace() []string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(3, pc) // Пропускаем 3 фрейма (NewAppError, With*, getStackTrace)
	frames := runtime.CallersFrames(pc[:n])

	var trace []string
	for {
		frame, more := frames.Next()
		// Фильтруем системные вызовы
		if !isSystemFrame(frame.Function) {
			trace = append(trace, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
		}
		if !more || len(trace) >= 10 { // Ограничиваем глубину стека
			break
		}
	}
	return trace
}

// getCurrentModule пытается определить текущий модуль
func getCurrentModule() string {
	pc := make([]uintptr, 5)
	n := runtime.Callers(3, pc)
	if n > 0 {
		frames := runtime.CallersFrames(pc[:n])
		if frame, more := frames.Next(); more {
			// Извлекаем модуль из пути (например, github.com/org/project/internal/handler -> handler)
			return extractModuleName(frame.Function)
		}
	}
	return "unknown"
}

// isSystemFrame проверяет, является ли фрейм системным
func isSystemFrame(function string) bool {
	systemPrefixes := []string{
		"runtime.",
		"reflect.",
		"blerrors.",
		"net/http.",
		"github.com/gorilla/mux.",
	}
	for _, prefix := range systemPrefixes {
		if len(function) > len(prefix) && function[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// extractModuleName извлекает имя модуля из полного пути функции
func extractModuleName(function string) string {
	parts := []string{}
	current := ""
	for _, r := range function {
		if r == '/' || r == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}

	// Ищем последний значимый компонент (обычно это имя пакета)
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if len(part) > 0 && part != "internal" && part != "pkg" {
			return part
		}
	}
	return "unknown"
}
