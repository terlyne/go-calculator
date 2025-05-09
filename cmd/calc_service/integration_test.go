package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terlyne/go-calculator/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestCalculatorService_Integration(t *testing.T) {
	// Start the server in a goroutine
	go func() {
		main()
	}()

	// Wait for the server to start
	time.Sleep(time.Second)

	// Connect to the server
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := api.NewCalculatorClient(conn)
	ctx := context.Background()

	// Test registration
	t.Run("Registration Flow", func(t *testing.T) {
		// Register a new user
		registerResp, err := client.Register(ctx, &api.RegisterRequest{
			Login:    "testuser",
			Password: "testpass",
		})
		require.NoError(t, err)
		assert.True(t, registerResp.Success)

		// Try to register the same user again
		_, err = client.Register(ctx, &api.RegisterRequest{
			Login:    "testuser",
			Password: "testpass",
		})
		assert.Error(t, err)
	})

	// Test login
	t.Run("Login Flow", func(t *testing.T) {
		// Login with correct credentials
		loginResp, err := client.Login(ctx, &api.LoginRequest{
			Login:    "testuser",
			Password: "testpass",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, loginResp.Token)

		// Login with wrong password
		_, err = client.Login(ctx, &api.LoginRequest{
			Login:    "testuser",
			Password: "wrongpass",
		})
		assert.Error(t, err)
	})

	// Test calculation
	t.Run("Calculation Flow", func(t *testing.T) {
		// Get a valid token
		loginResp, err := client.Login(ctx, &api.LoginRequest{
			Login:    "testuser",
			Password: "testpass",
		})
		require.NoError(t, err)
		token := loginResp.Token

		// Calculate a valid expression
		calcResp, err := client.Calculate(ctx, &api.CalculateRequest{
			Expression: "2+2*2",
			Token:     token,
		})
		require.NoError(t, err)
		assert.Equal(t, 6.0, calcResp.Result)

		// Calculate an invalid expression
		_, err = client.Calculate(ctx, &api.CalculateRequest{
			Expression: "2++2",
			Token:     token,
		})
		assert.Error(t, err)

		// Calculate with invalid token
		_, err = client.Calculate(ctx, &api.CalculateRequest{
			Expression: "2+2",
			Token:     "invalid-token",
		})
		assert.Error(t, err)
	})

	// Test expression history
	t.Run("Expression History Flow", func(t *testing.T) {
		// Get a valid token
		loginResp, err := client.Login(ctx, &api.LoginRequest{
			Login:    "testuser",
			Password: "testpass",
		})
		require.NoError(t, err)
		token := loginResp.Token

		// Calculate some expressions
		_, err = client.Calculate(ctx, &api.CalculateRequest{
			Expression: "1+1",
			Token:     token,
		})
		require.NoError(t, err)

		_, err = client.Calculate(ctx, &api.CalculateRequest{
			Expression: "2*3",
			Token:     token,
		})
		require.NoError(t, err)

		// Get expression history
		historyResp, err := client.GetExpressions(ctx, &api.GetExpressionsRequest{
			Token: token,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(historyResp.Expressions), 2)

		// Get history with invalid token
		_, err = client.GetExpressions(ctx, &api.GetExpressionsRequest{
			Token: "invalid-token",
		})
		assert.Error(t, err)
	})
} 