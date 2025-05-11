package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

// Структура сервера, реализующая gRPC интерфейс
type server struct {
	pb.UnimplementedCalculatorServer
	db   *database.Database
	auth *auth.Auth
}

// Регистрация нового пользователя
func (s *server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	hashedPassword, err := s.auth.HashPassword(req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, "ошибка хеширования пароля")
	}

	user := &models.User{
		Login:    req.Login,
		Password: hashedPassword,
	}

	if err := s.db.CreateUser(user); err != nil {
		return nil, status.Error(codes.AlreadyExists, "пользователь уже существует")
	}

	return &pb.RegisterResponse{Success: true}, nil
}

// Вход в систему
func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := s.db.GetUserByLogin(req.Login)
	if err != nil {
		return nil, status.Error(codes.NotFound, "пользователь не найден")
	}

	if !s.auth.CheckPasswordHash(req.Password, user.Password) {
		return nil, status.Error(codes.Unauthenticated, "неверные учетные данные")
	}

	token, err := s.auth.GenerateToken(user.ID, user.Login)
	if err != nil {
		return nil, status.Error(codes.Internal, "ошибка генерации токена")
	}

	return &pb.LoginResponse{Token: token}, nil
}

// Вычисление выражения
func (s *server) Calculate(ctx context.Context, req *pb.CalculateRequest) (*pb.CalculateResponse, error) {
	claims, err := s.auth.ValidateToken(req.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "неверный токен")
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
		return nil, status.Error(codes.Internal, "ошибка сохранения выражения")
	}

	return &pb.CalculateResponse{Result: result}, nil
}

// Получение истории вычислений
func (s *server) GetExpressions(ctx context.Context, req *pb.GetExpressionsRequest) (*pb.GetExpressionsResponse, error) {
	claims, err := s.auth.ValidateToken(req.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "неверный токен")
	}

	expressions, err := s.db.GetUserExpressions(claims.UserID)
	if err != nil {
		return nil, status.Error(codes.Internal, "ошибка получения выражений")
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
	// Инициализация базы данных
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = ":memory:" // Используем in-memory базу данных по умолчанию
		log.Println("Используется in-memory база данных для тестирования")
	}

	db, err := database.NewDatabase(dbPath)
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}
	defer db.Close()

	// Инициализация аутентификации с безопасным секретным ключом
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		secretKey = "your-secret-key" // В продакшене обязательно использовать переменную окружения
		log.Println("Предупреждение: Используется стандартный JWT секретный ключ. Установите переменную окружения JWT_SECRET_KEY в продакшене.")
	}
	auth := auth.NewAuth(secretKey)

	// Создание экземпляра сервера
	server := &server{
		db:   db,
		auth: auth,
	}

	// Запуск gRPC сервера
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}

		s := grpc.NewServer()
		pb.RegisterCalculatorServer(s, server)

		log.Println("Запуск gRPC сервера на порту :50051")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Ошибка работы сервера: %v", err)
		}
	}()

	// Запуск HTTP сервера
	go func() {
		r := mux.NewRouter()
		r.HandleFunc("/api/v1/calculate", server.calculateHandler).Methods("POST")

		log.Println("Запуск HTTP сервера на порту :8080")
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Fatalf("Ошибка работы HTTP сервера: %v", err)
		}
	}()

	// Обработка graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Завершение работы серверов...")
}

// HTTP handler for calculation
func (s *server) calculateHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем токен из заголовка Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, `{"error": "Authorization header is required"}`, http.StatusUnauthorized)
		return
	}

	// Проверяем формат токена (Bearer token)
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		http.Error(w, `{"error": "Invalid authorization header format"}`, http.StatusUnauthorized)
		return
	}

	token := tokenParts[1]

	// Проверяем валидность токена
	claims, err := s.auth.ValidateToken(token)
	if err != nil {
		http.Error(w, `{"error": "Invalid token"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	result, err := calculator.Calc(req.Expression)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	expr := &models.Expression{
		UserID:     claims.UserID,
		Expression: req.Expression,
		Result:     result,
		Status:     "completed",
	}

	if err := s.db.SaveExpression(expr); err != nil {
		http.Error(w, `{"error": "Failed to save expression"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"result": result})
}
