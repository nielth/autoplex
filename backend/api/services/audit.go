package services

import (
	"api/models"
	"context"
	"fmt"
	"strings"
	"time"
)

func UpsertUserFromAuth(user models.User) (uint64, error) {
	username := strings.TrimSpace(user.Username)
	if username == "" {
		return 0, fmt.Errorf("username is required")
	}

	db, err := dbConn()
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := db.ExecContext(
		ctx,
		`INSERT INTO users (username, email, plex_user_id, joined_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			id = LAST_INSERT_ID(id),
			email = VALUES(email),
			plex_user_id = VALUES(plex_user_id),
			updated_at = NOW()`,
		username,
		nullableString(user.Email),
		nullableString(user.ID),
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint64(id), nil
}

func ensureUserByUsername(username string) (uint64, error) {
	clean := strings.TrimSpace(username)
	if clean == "" {
		return 0, fmt.Errorf("username is required")
	}

	db, err := dbConn()
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := db.ExecContext(
		ctx,
		`INSERT INTO users (username, joined_at, updated_at)
		VALUES (?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			id = LAST_INSERT_ID(id),
			updated_at = NOW()`,
		clean,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint64(id), nil
}

func LogLoginEvent(user *models.User, success bool, failureReason string, ipAddress string, userAgent string) error {
	db, err := dbConn()
	if err != nil {
		return err
	}

	var userID any
	var username any

	if user != nil && strings.TrimSpace(user.Username) != "" {
		userIDValue, upsertErr := UpsertUserFromAuth(*user)
		if upsertErr != nil {
			return upsertErr
		}
		userID = userIDValue
		username = strings.TrimSpace(user.Username)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.ExecContext(
		ctx,
		`INSERT INTO login_events (user_id, username, success, failure_reason, ip_address, user_agent)
		VALUES (?, ?, ?, ?, ?, ?)`,
		userID,
		username,
		success,
		nullableString(failureReason),
		nullableString(ipAddress),
		nullableString(userAgent),
	)
	return err
}

func LogSearchEvent(username string, query string, page string, success bool, errorMessage string, ipAddress string, userAgent string) error {
	cleanUsername := strings.TrimSpace(username)
	if cleanUsername == "" {
		return fmt.Errorf("username is required")
	}

	userID, err := ensureUserByUsername(cleanUsername)
	if err != nil {
		return err
	}

	db, err := dbConn()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.ExecContext(
		ctx,
		`INSERT INTO search_events (user_id, username, query_text, page, success, error_message, ip_address, user_agent)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userID,
		cleanUsername,
		strings.TrimSpace(query),
		nullableString(page),
		success,
		nullableString(errorMessage),
		nullableString(ipAddress),
		nullableString(userAgent),
	)
	return err
}

func LogDownloadEvent(username string, data models.DownloadData, qbtHash string, success bool, errorMessage string, ipAddress string, userAgent string) error {
	cleanUsername := strings.TrimSpace(username)
	if cleanUsername == "" {
		return fmt.Errorf("username is required")
	}

	userID, err := ensureUserByUsername(cleanUsername)
	if err != nil {
		return err
	}

	db, err := dbConn()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.ExecContext(
		ctx,
		`INSERT INTO download_events (user_id, username, fid, filename, category_id, torrent_size, is_freeleech, qbt_hash, success, error_message, ip_address, user_agent)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID,
		cleanUsername,
		nullableString(data.Fid),
		nullableString(data.Filename),
		data.CategoryID,
		nullableUint64(data.Size),
		data.IsFreeleech,
		nullableString(qbtHash),
		success,
		nullableString(errorMessage),
		nullableString(ipAddress),
		nullableString(userAgent),
	)
	return err
}

func nullableString(value string) any {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return nil
	}
	return clean
}

func nullableUint64(value uint64) any {
	if value == 0 {
		return nil
	}
	return value
}
