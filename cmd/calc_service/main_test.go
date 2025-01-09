package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestCalculateHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    CalculateRequest
		expectedStatus int
		expectedResult string
		expectedError  string
	}{
		{
			name:           "Valid Expression",
			requestBody:    CalculateRequest{Expression: "3 + 5"},
			expectedStatus: http.StatusOK,
			expectedResult: "8",
		},
		{
			name:           "Division by Zero",
			requestBody:    CalculateRequest{Expression: "3 / 0"},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "Деление на ноль",
		},
		{
			name:           "Mismatched Parentheses",
			requestBody:    CalculateRequest{Expression: "3 + (5 * (2 - 4)"},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "Несовпадение скобок",
		},
		{
			name:           "Invalid Character",
			requestBody:    CalculateRequest{Expression: "3 + 5 * a"},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "Недопустимый символ в выражении",
		},
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/calculate", calculateHandler).Methods("POST")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, rr.Code)
			}

			var response CalculateResponse
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Fatalf("could not decode response: %v", err)
			}

			if response.Result != tt.expectedResult {
				t.Errorf("expected result %v, got %v", tt.expectedResult, response.Result)
			}

			if response.Error != tt.expectedError {
				t.Errorf("expected error %v, got %v", tt.expectedError, response.Error)
			}
		})
	}
}