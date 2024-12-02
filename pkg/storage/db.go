package storage

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	logger "github.com/thalq/gopher_mart/internal/middleware"
)

var db *sql.DB

func InitDB(connectionString string) {
	var err error
	db, err = sql.Open("pgx", connectionString)
	if err != nil {
		logger.Sugar.Fatalf("Error open db: %s", err)
	}
	if err := db.Ping(); err != nil {
		logger.Sugar.Fatalf("Error ping db: %s", err)
	}

	createTables := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        username VARCHAR(255) UNIQUE,
        password VARCHAR(255)
    );
    CREATE TABLE IF NOT EXISTS orders (
        user_id INT REFERENCES users(id),
        order_id VARCHAR(255),
        status VARCHAR(10) DEFAULT 'NEW' CHECK (status IN ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED')),
        upload_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        withdrawal FLOAT DEFAULT 0.0,
		accrual FLOAT DEFAULT 0.0
    );
    CREATE TABLE IF NOT EXISTS user_balance (
        user_id INT UNIQUE REFERENCES users(id),
        current_balance FLOAT DEFAULT 0.0
    );
    `

	if _, err := db.Exec(createTables); err != nil {
		logger.Sugar.Fatalf("Error create tables: %s", err)
	}

	logger.Sugar.Info("DB connected")
}

func GetDB() *sql.DB {
	if db == nil {
		logger.Sugar.Fatalf("Database connection is not initialized")
	}
	return db
}
