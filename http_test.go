package blerrors

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteErrorResponse(t *testing.T) {
	err := NewAppError(ErrCodeNotFound, "Resource not found").WithModule("test-service")

	w := httptest.NewRecorder()
	WriteErrorResponse(w, err)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Success != false {
		t.Error("Success should be false")
	}

	if response.Error.Code != ErrCodeNotFound {
		t.Errorf("Expected error code %s, got %s", ErrCodeNotFound, response.Error.Code)
	}

	if response.Error.Message != "Resource not found" {
		t.Errorf("Expected message 'Resource not found', got '%s'", response.Error.Message)
	}

	if response.Error.Module != "test-service" {
		t.Errorf("Expected module 'test-service', got '%s'", response.Error.Module)
	}
}

func TestWriteSuccessResponse(t *testing.T) {
	data := map[string]interface{}{
		"user_id": 123,
		"name":    "John Doe",
	}

	w := httptest.NewRecorder()
	WriteSuccessResponse(w, data)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}

	var response SuccessResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Success != true {
		t.Error("Success should be true")
	}

	if response.Data == nil {
		t.Fatal("Data should not be nil")
	}

	dataMap, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Data should be a map")
	}

	if dataMap["user_id"] != float64(123) { // JSON unmarshals numbers as float64
		t.Errorf("Expected user_id 123, got %v", dataMap["user_id"])
	}
}

func TestGetHTTPStatusCode(t *testing.T) {
	testCases := []struct {
		code     ErrorCode
		expected int
	}{
		{ErrCodeNotFound, http.StatusNotFound},
		{ErrCodeValidation, http.StatusBadRequest},
		{ErrCodeBadRequest, http.StatusBadRequest},
		{ErrCodeUnauthorized, http.StatusUnauthorized},
		{ErrCodeForbidden, http.StatusForbidden},
		{ErrCodeConflict, http.StatusConflict},
		{ErrCodeTooManyRequests, http.StatusTooManyRequests},
		{ErrCodeServiceUnavailable, http.StatusServiceUnavailable},
		{ErrCodeInternal, http.StatusInternalServerError},
		{ErrorCode("UNKNOWN"), http.StatusInternalServerError}, // default case
	}

	for _, tc := range testCases {
		result := getHTTPStatusCode(tc.code)
		if result != tc.expected {
			t.Errorf("For code %s, expected status %d, got %d", tc.code, tc.expected, result)
		}
	}
}

func TestErrorRecoveryMiddleware(t *testing.T) {
	// Create a handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("Test panic")
	})

	// Wrap with middleware
	middleware := ErrorRecoveryMiddleware(panicHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// This should not panic, but handle the panic gracefully
	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Success != false {
		t.Error("Success should be false")
	}

	if response.Error.Code != ErrCodeInternal {
		t.Errorf("Expected error code %s, got %s", ErrCodeInternal, response.Error.Code)
	}

	if response.Error.Module != "middleware" {
		t.Errorf("Expected module 'middleware', got '%s'", response.Error.Module)
	}

	// Check that panic details are in the error
	if response.Error.Details == nil {
		t.Error("Details should contain panic information")
	}

	detailsStr, ok := response.Error.Details.(string)
	if !ok {
		t.Error("Details should be a string")
	}

	if !strings.Contains(detailsStr, "Test panic") {
		t.Errorf("Details should contain panic message, got: %s", detailsStr)
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	// Create a simple handler that checks for request ID
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		if requestID == "" {
			t.Error("Request ID should not be empty")
		}

		// Check that it's a reasonable timestamp-based ID
		if len(requestID) < 10 {
			t.Errorf("Request ID seems too short: %s", requestID)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with middleware
	middleware := RequestIDMiddleware(testHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetRequestID(t *testing.T) {
	testCases := []struct {
		name       string
		ctx        context.Context
		expectedID string
	}{
		{
			name:       "with request ID",
			ctx:        context.WithValue(context.Background(), "request_id", "req-123"),
			expectedID: "req-123",
		},
		{
			name:       "without request ID",
			ctx:        context.Background(),
			expectedID: "",
		},
		{
			name:       "with wrong type",
			ctx:        context.WithValue(context.Background(), "request_id", 123),
			expectedID: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetRequestID(tc.ctx)
			if result != tc.expectedID {
				t.Errorf("Expected request ID '%s', got '%s'", tc.expectedID, result)
			}
		})
	}
}

func TestJSONEncoding(t *testing.T) {
	// Test that our structs can be properly encoded to JSON
	err := NewAppError(ErrCodeValidation, "Invalid input").
		WithModule("user-service").
		WithDetails(map[string]string{"field": "email"}).
		WithRequestID("req-123")

	response := ErrorResponse{
		Success: false,
		Error:   err,
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(response); err != nil {
		t.Fatalf("Failed to encode response: %v", err)
	}

	// Try to decode it back
	var decoded ErrorResponse
	if err := json.NewDecoder(&buf).Decode(&decoded); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if decoded.Success != false {
		t.Error("Decoded success should be false")
	}

	if decoded.Error.Code != ErrCodeValidation {
		t.Errorf("Expected decoded code %s, got %s", ErrCodeValidation, decoded.Error.Code)
	}

	if decoded.Error.Message != "Invalid input" {
		t.Errorf("Expected decoded message 'Invalid input', got '%s'", decoded.Error.Message)
	}
}
