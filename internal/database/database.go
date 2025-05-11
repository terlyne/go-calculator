package database

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/terlyne/go-calculator/internal/models"
)

// Database представляет интерфейс для работы с базой данных
type Database struct {
	db *sql.DB
}

// NewDatabase создает новое подключение к базе данных
func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Создаем таблицы, если они не существуют
	if err := createTables(db); err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

// Close закрывает соединение с базой данных
func (d *Database) Close() error {
	return d.db.Close()
}

// CreateUser создает нового пользователя в базе данных
func (d *Database) CreateUser(user *models.User) error {
	query := `INSERT INTO users (login, password, created_at) VALUES (?, ?, ?)`
	_, err := d.db.Exec(query, user.Login, user.Password, time.Now())
	return err
}

// GetUserByLogin получает пользователя по логину
func (d *Database) GetUserByLogin(login string) (*models.User, error) {
	query := `SELECT id, login, password, created_at FROM users WHERE login = ?`
	user := &models.User{}
	err := d.db.QueryRow(query, login).Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// SaveExpression сохраняет выражение в базе данных
func (d *Database) SaveExpression(expr *models.Expression) error {
	query := `INSERT INTO expressions (user_id, expression, result, status, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?)`
	now := time.Now()
	_, err := d.db.Exec(query, expr.UserID, expr.Expression, expr.Result, expr.Status, now, now)
	return err
}

// GetUserExpressions получает все выражения пользователя
func (d *Database) GetUserExpressions(userID int64) ([]*models.Expression, error) {
	query := `SELECT id, user_id, expression, result, status, created_at, updated_at 
		FROM expressions WHERE user_id = ? ORDER BY created_at DESC`
	rows, err := d.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expressions []*models.Expression
	for rows.Next() {
		expr := &models.Expression{}
		err := rows.Scan(&expr.ID, &expr.UserID, &expr.Expression, &expr.Result, &expr.Status, &expr.CreatedAt, &expr.UpdatedAt)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, expr)
	}
	return expressions, nil
}

// createTables создает необходимые таблицы в базе данных
func createTables(db *sql.DB) error {
	// Создаем таблицу пользователей
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			login TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at DATETIME NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Создаем таблицу выражений
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS expressions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			expression TEXT NOT NULL,
			result REAL NOT NULL,
			status TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users (id)
		)
	`)
	return err
} 