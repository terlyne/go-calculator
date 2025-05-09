package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/terlyne/go-calculator/api"
	"github.com/terlyne/go-calculator/internal/auth"
	"github.com/terlyne/go-calculator/internal/database"
	"github.com/terlyne/go-calculator/internal/models"
	"github.com/terlyne/go-calculator/pkg/calculator"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedCalculatorServer
	db   *database.Database
	auth *auth.Auth
}

func (s *server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	hashedPassword, err := s.auth.HashPassword(req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	user := &models.User{
		Login:    req.Login,
		Password: hashedPassword,
	}

	if err := s.db.CreateUser(user); err != nil {
		return nil, status.Error(codes.AlreadyExists, "user already exists")
	}

	return &pb.RegisterResponse{Success: true}, nil
}

func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := s.db.GetUserByLogin(req.Login)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	if !s.auth.CheckPasswordHash(req.Password, user.Password) {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	token, err := s.auth.GenerateToken(user.ID, user.Login)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	return &pb.LoginResponse{Token: token}, nil
}

func (s *server) Calculate(ctx context.Context, req *pb.CalculateRequest) (*pb.CalculateResponse, error) {
	claims, err := s.auth.ValidateToken(req.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	result, err := calculator.Calc(req.Expression)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	expr := &models.Expression{
		UserID:     claims.UserID,
		Expression: req.Expression,
		Result:     result,
		Status:     "completed",
	}

	if err := s.db.SaveExpression(expr); err != nil {
		return nil, status.Error(codes.Internal, "failed to save expression")
	}

	return &pb.CalculateResponse{Result: result}, nil
}

func (s *server) GetExpressions(ctx context.Context, req *pb.GetExpressionsRequest) (*pb.GetExpressionsResponse, error) {
	claims, err := s.auth.ValidateToken(req.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	expressions, err := s.db.GetUserExpressions(claims.UserID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get expressions")
	}

	var response []*pb.Expression
	for _, expr := range expressions {
		response = append(response, &pb.Expression{
			Id:         expr.ID,
			Expression: expr.Expression,
			Result:     expr.Result,
			Status:     expr.Status,
			CreatedAt:  expr.CreatedAt.Format(time.RFC3339),
			UpdatedAt:  expr.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &pb.GetExpressionsResponse{Expressions: response}, nil
}

func main() {
	// Initialize database
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = ":memory:" // Используем in-memory базу данных по умолчанию
		log.Println("Using in-memory database for testing")
	}

	db, err := database.NewDatabase(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize auth with a secure secret key
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		secretKey = "your-secret-key" // В продакшене обязательно использовать переменную окружения
		log.Println("Warning: Using default JWT secret key. Set JWT_SECRET_KEY environment variable in production.")
	}
	auth := auth.NewAuth(secretKey)

	// Create server instance
	server := &server{
		db:   db,
		auth: auth,
	}

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		s := grpc.NewServer()
		pb.RegisterCalculatorServer(s, server)

		log.Println("Starting gRPC server on :50051")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Start HTTP server
	go func() {
		r := mux.NewRouter()
		r.HandleFunc("/api/v1/calculate", server.calculateHandler).Methods("POST")

		log.Println("Starting HTTP server on :8080")
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Fatalf("Failed to serve HTTP: %v", err)
		}
	}()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")
}

// HTTP handler for calculation
func (s *server) calculateHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Expression string `json:"expression"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := calculator.Calc(req.Expression)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	response := struct {
		Result string `json:"result"`
	}{
		Result: result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
