package services

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func IsAdminUsername(username string) (bool, error) {
	cleanUsername := strings.TrimSpace(username)
	if cleanUsername == "" {
		return false, fmt.Errorf("username is required")
	}

	db, err := dbConn()
	if err != nil {
		return false, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var isAdmin bool
	if err := db.QueryRowContext(
		ctx,
		`SELECT EXISTS(
			SELECT 1
			FROM admin_users a
			INNER JOIN users u ON u.id = a.user_id
			WHERE LOWER(u.username) = LOWER(?)
		)`,
		cleanUsername,
	).Scan(&isAdmin); err != nil {
		return false, err
	}

	return isAdmin, nil
}
