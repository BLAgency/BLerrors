package blerrors

// ErrorCode - уникальный код ошибки
type ErrorCode string

// Предопределенные коды ошибок
const (
	ErrCodeNotFound           ErrorCode = "NOT_FOUND"
	ErrCodeValidation         ErrorCode = "VALIDATION_ERROR"
	ErrCodeInternal           ErrorCode = "INTERNAL_ERROR"
	ErrCodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden          ErrorCode = "FORBIDDEN"
	ErrCodeBadRequest         ErrorCode = "BAD_REQUEST"
	ErrCodeConflict           ErrorCode = "CONFLICT"
	ErrCodeTooManyRequests    ErrorCode = "TOO_MANY_REQUESTS"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
)

// AppError - унифицированная структура ошибки
type AppError struct {
	Code      ErrorCode   `json:"code"`                 // Уникальный код ошибки
	Message   string      `json:"message"`              // Текстовое описание
	Timestamp int64       `json:"timestamp"`            // Время в Unix timestamp
	Module    string      `json:"module,omitempty"`     // Модуль/сервис где произошла ошибка
	Trace     []string    `json:"trace,omitempty"`      // Стек вызовов
	Details   interface{} `json:"details,omitempty"`    // Дополнительные детали
	RequestID string      `json:"request_id,omitempty"` // ID запроса для трекинга
}

// ErrorResponse представляет структуру ответа с ошибкой
type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   *AppError `json:"error"`
}

// SuccessResponse представляет структуру успешного ответа
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}
