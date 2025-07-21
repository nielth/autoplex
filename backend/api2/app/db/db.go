package _db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"main/models"
)

func AddUser(db *sql.DB, username string, userInfo models.User) (int64, error) {
	jsonData, err := json.Marshal(userInfo)
	if err != nil {
		return 0, fmt.Errorf("failed to insert userInfo json: %w", err)
	}

	query := `INSERT IGNORE INTO Users (username, user_info) VALUES (?, ?)`
	result, err := db.Exec(query, username, jsonData)
	if err != nil {
		return 0, fmt.Errorf("failed to insert user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get inserted ID: %w", err)
	}
	return id, nil
}

func UserLogin(db *sql.DB, username string, tokenString string) error {
	var id int
	query := `SELECT id FROM Users WHERE username=?`
	err := db.QueryRow(query, username).Scan(&id)
	if err != nil {
		return fmt.Errorf("Failed to retrieve user: %w", err)
	}

	query = `INSERT INTO Login (uid, token) VALUES (?, ?)`
	result, err := db.Exec(query, id, tokenString)
	if err != nil {
		return fmt.Errorf("failed to insert user into Login: %w", err)
	}

	_, err = result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get inserted ID: %w", err)
	}
	return nil
}

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:example@tcp(localhost:3306)/autoplex")
	if err != nil {
		panic(err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS Users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(100),
		user_info JSON,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE (username)
	);
	`

	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}

	query = `
	CREATE TABLE IF NOT EXISTS Login (
		id INT AUTO_INCREMENT PRIMARY KEY,
		uid INT,
		token VARCHAR(255),
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (uid) REFERENCES Users(id)
	);
	`

	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}

	query = `
	CREATE TABLE IF NOT EXISTS Torrents (
		id INT AUTO_INCREMENT PRIMARY KEY,
		uid INT,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (uid) REFERENCES Users(id)
	);
	`

	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}

	return db, nil
}
