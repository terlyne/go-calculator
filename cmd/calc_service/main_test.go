package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/terlyne/go-calculator/api"
	"github.com/terlyne/go-calculator/internal/auth"
	"github.com/terlyne/go-calculator/internal/database"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// CalculateRequest represents the request body for calculation
type CalculateRequest struct {
	Expression string `json:"expression"`
}

// CalculateResponse represents the response body for calculation
type CalculateResponse struct {
	Result string `json:"result"`
	Error  string `json:"error,omitempty"`
}

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

func TestServer_Register(t *testing.T) {
	db, err := database.NewDatabase(":memory:")
	assert.NoError(t, err)
	defer db.Close()

	auth := auth.NewAuth("test-secret")
	server := &server{
		db:   db,
		auth: auth,
	}

	tests := []struct {
		name    string
		req     *api.RegisterRequest
		wantErr bool
		errCode int
	}{
		{
			name: "successful registration",
			req: &api.RegisterRequest{
				Login:    "testuser",
				Password: "testpass",
			},
			wantErr: false,
		},
		{
			name: "duplicate user",
			req: &api.RegisterRequest{
				Login:    "testuser",
				Password: "testpass",
			},
			wantErr: true,
			errCode: 13, // AlreadyExists
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.Register(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.True(t, resp.Success)
			}
		})
	}
}

func TestServer_Login(t *testing.T) {
	db, err := database.NewDatabase(":memory:")
	assert.NoError(t, err)
	defer db.Close()

	auth := auth.NewAuth("test-secret")
	server := &server{
		db:   db,
		auth: auth,
	}

	// Register a test user first
	_, err = server.Register(context.Background(), &api.RegisterRequest{
		Login:    "testuser",
		Password: "testpass",
	})
	assert.NoError(t, err)

	tests := []struct {
		name    string
		req     *api.LoginRequest
		wantErr bool
		errCode int
	}{
		{
			name: "successful login",
			req: &api.LoginRequest{
				Login:    "testuser",
				Password: "testpass",
			},
			wantErr: false,
		},
		{
			name: "wrong password",
			req: &api.LoginRequest{
				Login:    "testuser",
				Password: "wrongpass",
			},
			wantErr: true,
			errCode: 16, // Unauthenticated
		},
		{
			name: "user not found",
			req: &api.LoginRequest{
				Login:    "nonexistent",
				Password: "testpass",
			},
			wantErr: true,
			errCode: 5, // NotFound
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.Login(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.Token)
			}
		})
	}
}

func TestServer_Calculate(t *testing.T) {
	db, err := database.NewDatabase(":memory:")
	assert.NoError(t, err)
	defer db.Close()

	auth := auth.NewAuth("test-secret")
	server := &server{
		db:   db,
		auth: auth,
	}

	// Register and login to get a token
	_, err = server.Register(context.Background(), &api.RegisterRequest{
		Login:    "testuser",
		Password: "testpass",
	})
	assert.NoError(t, err)

	loginResp, err := server.Login(context.Background(), &api.LoginRequest{
		Login:    "testuser",
		Password: "testpass",
	})
	assert.NoError(t, err)
	token := loginResp.Token

	tests := []struct {
		name    string
		req     *api.CalculateRequest
		wantErr bool
		errCode int
	}{
		{
			name: "valid expression",
			req: &api.CalculateRequest{
				Expression: "2+2*2",
				Token:      token,
			},
			wantErr: false,
		},
		{
			name: "invalid expression",
			req: &api.CalculateRequest{
				Expression: "2++2",
				Token:      token,
			},
			wantErr: true,
			errCode: 3, // InvalidArgument
		},
		{
			name: "invalid token",
			req: &api.CalculateRequest{
				Expression: "2+2",
				Token:      "invalid-token",
			},
			wantErr: true,
			errCode: 16, // Unauthenticated
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.Calculate(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotZero(t, resp.Result)
			}
		})
	}
}

func setupTestServer(t *testing.T) (*grpc.Server, *server) {
	db, err := database.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	auth := auth.NewAuth("test-secret-key")
	s := grpc.NewServer()
	server := &server{
		db:   db,
		auth: auth,
	}
	pb.RegisterCalculatorServer(s, server)

	return s, server
}

func TestRegister(t *testing.T) {
	s, _ := setupTestServer(t)
	defer s.Stop()

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewCalculatorClient(conn)

	tests := []struct {
		name    string
		req     *pb.RegisterRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "valid registration",
			req: &pb.RegisterRequest{
				Login:    "testuser",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "duplicate registration",
			req: &pb.RegisterRequest{
				Login:    "testuser",
				Password: "password123",
			},
			wantErr: true,
			errCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Register(context.Background(), tt.req)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.errCode {
						t.Errorf("Expected error code %v, got %v", tt.errCode, st.Code())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !resp.Success {
					t.Error("Expected success, got false")
				}
			}
		})
	}
}

func TestLogin(t *testing.T) {
	s, _ := setupTestServer(t)
	defer s.Stop()

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewCalculatorClient(conn)

	// Register a test user first
	_, err = client.Register(context.Background(), &pb.RegisterRequest{
		Login:    "testuser",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	tests := []struct {
		name    string
		req     *pb.LoginRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "valid login",
			req: &pb.LoginRequest{
				Login:    "testuser",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "invalid password",
			req: &pb.LoginRequest{
				Login:    "testuser",
				Password: "wrongpassword",
			},
			wantErr: true,
			errCode: codes.Unauthenticated,
		},
		{
			name: "non-existent user",
			req: &pb.LoginRequest{
				Login:    "nonexistent",
				Password: "password123",
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Login(context.Background(), tt.req)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.errCode {
						t.Errorf("Expected error code %v, got %v", tt.errCode, st.Code())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp.Token == "" {
					t.Error("Expected token, got empty string")
				}
			}
		})
	}
}

func TestCalculate(t *testing.T) {
	s, _ := setupTestServer(t)
	defer s.Stop()

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewCalculatorClient(conn)

	// Register and login to get a token
	_, err = client.Register(context.Background(), &pb.RegisterRequest{
		Login:    "testuser",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	loginResp, err := client.Login(context.Background(), &pb.LoginRequest{
		Login:    "testuser",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	token := loginResp.Token

	tests := []struct {
		name    string
		req     *pb.CalculateRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "valid calculation",
			req: &pb.CalculateRequest{
				Token:      token,
				Expression: "2 + 2",
			},
			wantErr: false,
		},
		{
			name: "invalid expression",
			req: &pb.CalculateRequest{
				Token:      token,
				Expression: "2 + ",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid token",
			req: &pb.CalculateRequest{
				Token:      "invalid-token",
				Expression: "2 + 2",
			},
			wantErr: true,
			errCode: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Calculate(context.Background(), tt.req)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.errCode {
						t.Errorf("Expected error code %v, got %v", tt.errCode, st.Code())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp.Result == 0 {
					t.Error("Expected non-zero result")
				}
			}
		})
	}
}

func TestGetExpressions(t *testing.T) {
	s, _ := setupTestServer(t)
	defer s.Stop()

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewCalculatorClient(conn)

	// Register and login to get a token
	_, err = client.Register(context.Background(), &pb.RegisterRequest{
		Login:    "testuser",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	loginResp, err := client.Login(context.Background(), &pb.LoginRequest{
		Login:    "testuser",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	token := loginResp.Token

	// Calculate some expressions
	expressions := []string{"2 + 2", "3 * 4", "10 - 5"}
	for _, expr := range expressions {
		_, err := client.Calculate(context.Background(), &pb.CalculateRequest{
			Token:      token,
			Expression: expr,
		})
		if err != nil {
			t.Fatalf("Failed to calculate expression %s: %v", expr, err)
		}
	}

	tests := []struct {
		name    string
		req     *pb.GetExpressionsRequest
		wantErr bool
		errCode codes.Code
		wantLen int
	}{
		{
			name: "get expressions with valid token",
			req: &pb.GetExpressionsRequest{
				Token: token,
			},
			wantErr: false,
			wantLen: len(expressions),
		},
		{
			name: "get expressions with invalid token",
			req: &pb.GetExpressionsRequest{
				Token: "invalid-token",
			},
			wantErr: true,
			errCode: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.GetExpressions(context.Background(), tt.req)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.errCode {
						t.Errorf("Expected error code %v, got %v", tt.errCode, st.Code())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(resp.Expressions) != tt.wantLen {
					t.Errorf("Expected %d expressions, got %d", tt.wantLen, len(resp.Expressions))
				}
			}
		})
	}
}
