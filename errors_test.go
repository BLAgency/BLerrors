package blerrors

import (
	"strings"
	"testing"
	"time"
)

func TestNewAppError(t *testing.T) {
	message := "Test error message"
	err := NewAppError(ErrCodeNotFound, message).WithErrorCode("FS001").IsUserError().WithUserID("user123")

	if err.Code != ErrCodeNotFound {
		t.Errorf("Expected code %s, got %s", ErrCodeNotFound, err.Code)
	}

	if err.Message != message {
		t.Errorf("Expected message '%s', got '%s'", message, err.Message)
	}

	if err.Timestamp == 0 {
		t.Error("Timestamp should not be zero")
	}

	if len(err.Trace) == 0 {
		t.Error("Trace should not be empty")
	}

	if err.Module == "" {
		t.Error("Module should not be empty")
	}
}

func TestAppError_Error(t *testing.T) {
	err := NewAppError(ErrCodeValidation, "Validation failed").WithErrorCode("FS001").IsUserError().WithUserID("user123")

	expected := "[VALIDATION_ERROR] Validation failed"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestAppError_WithModule(t *testing.T) {
	err := NewAppError(ErrCodeInternal, "Internal error").WithErrorCode("FS001").IsUserError().WithUserID("user123").WithModule("test-service")

	if err.Module != "test-service" {
		t.Errorf("Expected module 'test-service', got '%s'", err.Module)
	}
}

func TestAppError_WithDetails(t *testing.T) {
	details := map[string]interface{}{"user_id": 123, "action": "login"}
	err := NewAppError(ErrCodeUnauthorized, "Unauthorized").WithErrorCode("FS001").IsUserError().WithUserID("user123").WithDetails(details)

	if err.Details == nil {
		t.Fatal("Details should not be nil")
	}

	detailsMap, ok := err.Details.(map[string]interface{})
	if !ok {
		t.Fatal("Details should be a map")
	}

	if detailsMap["user_id"] != 123 {
		t.Errorf("Expected user_id 123, got %v", detailsMap["user_id"])
	}
}

func TestAppError_WithRequestID(t *testing.T) {
	requestID := "req-12345"
	err := NewAppError(ErrCodeBadRequest, "Bad request").WithErrorCode("FS001").IsUserError().WithUserID("user123").WithRequestID(requestID)

	if err.RequestID != requestID {
		t.Errorf("Expected request ID '%s', got '%s'", requestID, err.RequestID)
	}
}

func TestAppError_WithoutTrace(t *testing.T) {
	err := NewAppError(ErrCodeInternal, "Internal error").WithErrorCode("FS001").IsUserError().WithUserID("user123").WithoutTrace()

	if err.Trace != nil {
		t.Error("Trace should be nil after WithoutTrace()")
	}
}

func TestAppError_Chaining(t *testing.T) {
	err := NewAppError(ErrCodeForbidden, "Access denied").WithErrorCode("FS001").IsUserError().WithUserID("user123").
		WithModule("auth-service").
		WithDetails("Insufficient permissions").
		WithRequestID("req-abc")

	if err.Code != ErrCodeForbidden {
		t.Errorf("Expected code %s, got %s", ErrCodeForbidden, err.Code)
	}

	if err.Module != "auth-service" {
		t.Errorf("Expected module 'auth-service', got '%s'", err.Module)
	}

	if err.Details != "Insufficient permissions" {
		t.Errorf("Expected details 'Insufficient permissions', got '%v'", err.Details)
	}

	if err.RequestID != "req-abc" {
		t.Errorf("Expected request ID 'req-abc', got '%s'", err.RequestID)
	}
}

func TestGetStackTrace(t *testing.T) {
	// This is an internal function, but we can test it indirectly
	err := NewAppError(ErrCodeInternal, "Test").WithErrorCode("FS001").IsUserError().WithUserID("user123")

	if len(err.Trace) == 0 {
		t.Error("Stack trace should not be empty")
	}

	// Check that trace contains file information
	found := false
	for _, trace := range err.Trace {
		if strings.Contains(trace, ".go:") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Stack trace should contain file information")
	}
}

func TestGetCurrentModule(t *testing.T) {
	// This is an internal function, but we can test it indirectly
	err := NewAppError(ErrCodeInternal, "Test").WithErrorCode("FS001").IsUserError().WithUserID("user123")

	if err.Module == "" {
		t.Error("Module should not be empty")
	}

	// Module should be a reasonable name
	if err.Module == "unknown" {
		t.Log("Module detection returned 'unknown', which is acceptable")
	}
}

func TestErrorCodes(t *testing.T) {
	testCases := []struct {
		code     ErrorCode
		expected string
	}{
		{ErrCodeNotFound, "NOT_FOUND"},
		{ErrCodeValidation, "VALIDATION_ERROR"},
		{ErrCodeInternal, "INTERNAL_ERROR"},
		{ErrCodeUnauthorized, "UNAUTHORIZED"},
		{ErrCodeForbidden, "FORBIDDEN"},
		{ErrCodeBadRequest, "BAD_REQUEST"},
		{ErrCodeConflict, "CONFLICT"},
		{ErrCodeTooManyRequests, "TOO_MANY_REQUESTS"},
		{ErrCodeServiceUnavailable, "SERVICE_UNAVAILABLE"},
	}

	for _, tc := range testCases {
		if string(tc.code) != tc.expected {
			t.Errorf("Expected %s, got %s", tc.expected, string(tc.code))
		}
	}
}

func TestTimestampIsRecent(t *testing.T) {
	before := time.Now().Unix()
	err := NewAppError(ErrCodeInternal, "Test").WithErrorCode("FS001").IsUserError().WithUserID("user123")
	after := time.Now().Unix()

	if err.Timestamp < before || err.Timestamp > after {
		t.Errorf("Timestamp %d is not within expected range [%d, %d]", err.Timestamp, before, after)
	}
}
