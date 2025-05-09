package database

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/terlyne/go-calculator/internal/models"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := createTables(db); err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func createTables(db *sql.DB) error {
	// Create users table
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

	// Create expressions table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS expressions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			expression TEXT NOT NULL,
			result REAL,
			status TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)
	return err
}

func (d *Database) CreateUser(user *models.User) error {
	query := `INSERT INTO users (login, password, created_at) VALUES (?, ?, ?)`
	result, err := d.db.Exec(query, user.Login, user.Password, time.Now())
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = id
	return nil
}

func (d *Database) GetUserByLogin(login string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, login, password, created_at FROM users WHERE login = ?`
	err := d.db.QueryRow(query, login).Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (d *Database) SaveExpression(expr *models.Expression) error {
	query := `INSERT INTO expressions (user_id, expression, result, status, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?)`
	result, err := d.db.Exec(query, expr.UserID, expr.Expression, expr.Result, expr.Status, time.Now(), time.Now())
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	expr.ID = id
	return nil
}

func (d *Database) GetUserExpressions(userID int64) ([]models.Expression, error) {
	query := `SELECT id, user_id, expression, result, status, created_at, updated_at 
			  FROM expressions WHERE user_id = ? ORDER BY created_at DESC`
	rows, err := d.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expressions []models.Expression
	for rows.Next() {
		var expr models.Expression
		err := rows.Scan(&expr.ID, &expr.UserID, &expr.Expression, &expr.Result, &expr.Status, &expr.CreatedAt, &expr.UpdatedAt)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, expr)
	}
	return expressions, nil
}

func (d *Database) Close() error {
	return d.db.Close()
} 