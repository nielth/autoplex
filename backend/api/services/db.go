package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

var mysqlDB *sql.DB

func InitMySQL() error {
	host, err := requiredEnv("MYSQL_HOST")
	if err != nil {
		return err
	}
	dbName, err := requiredEnv("MYSQL_DATABASE")
	if err != nil {
		return err
	}
	user, err := requiredEnv("MYSQL_USER")
	if err != nil {
		return err
	}
	password, err := requiredEnv("MYSQL_PASSWORD")
	if err != nil {
		return err
	}

	port := strings.TrimSpace(os.Getenv("MYSQL_PORT"))
	if port == "" {
		port = "3306"
	}

	cfg := mysql.Config{
		User:      user,
		Passwd:    password,
		Net:       "tcp",
		Addr:      fmt.Sprintf("%s:%s", host, port),
		DBName:    dbName,
		ParseTime: true,
		Collation: "utf8mb4_unicode_ci",
		Params: map[string]string{
			"charset": "utf8mb4",
		},
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return fmt.Errorf("open mysql connection: %w", err)
	}

	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	retries := 10
	if retriesRaw := strings.TrimSpace(os.Getenv("MYSQL_CONNECT_RETRIES")); retriesRaw != "" {
		retriesParsed, parseErr := strconv.Atoi(retriesRaw)
		if parseErr != nil || retriesParsed < 1 {
			return fmt.Errorf("MYSQL_CONNECT_RETRIES must be a positive integer")
		}
		retries = retriesParsed
	}

	pingErr := pingWithRetry(db, retries)
	if pingErr != nil {
		return pingErr
	}

	if err := migrateAuditSchema(db); err != nil {
		return err
	}

	mysqlDB = db
	log.Println("mysql connected and audit schema is ready")
	return nil
}

func pingWithRetry(db *sql.DB, retries int) error {
	var lastErr error
	for i := 1; i <= retries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		lastErr = db.PingContext(ctx)
		cancel()
		if lastErr == nil {
			return nil
		}

		if i < retries {
			time.Sleep(2 * time.Second)
		}
	}

	return fmt.Errorf("could not connect to mysql after %d attempts: %w", retries, lastErr)
}

func migrateAuditSchema(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			username VARCHAR(255) NOT NULL,
			email VARCHAR(255) NULL,
			plex_user_id VARCHAR(128) NULL,
			joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY uq_users_username (username),
			KEY idx_users_plex_user_id (plex_user_id),
			KEY idx_users_joined_at (joined_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS admin_users (
			user_id BIGINT UNSIGNED NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id),
			CONSTRAINT fk_admin_users_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS login_events (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			user_id BIGINT UNSIGNED NULL,
			username VARCHAR(255) NULL,
			success TINYINT(1) NOT NULL,
			failure_reason VARCHAR(255) NULL,
			ip_address VARCHAR(45) NULL,
			user_agent VARCHAR(512) NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY idx_login_events_user_id (user_id),
			KEY idx_login_events_created_at (created_at),
			CONSTRAINT fk_login_events_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS search_events (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			user_id BIGINT UNSIGNED NULL,
			username VARCHAR(255) NULL,
			query_text VARCHAR(1024) NOT NULL,
			page VARCHAR(32) NULL,
			success TINYINT(1) NOT NULL,
			error_message TEXT NULL,
			ip_address VARCHAR(45) NULL,
			user_agent VARCHAR(512) NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY idx_search_events_user_id (user_id),
			KEY idx_search_events_created_at (created_at),
			CONSTRAINT fk_search_events_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS download_events (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			user_id BIGINT UNSIGNED NULL,
			username VARCHAR(255) NULL,
			fid VARCHAR(128) NULL,
			filename VARCHAR(512) NULL,
			tvmaze_id VARCHAR(128) NULL,
			tvmaze_episode_id VARCHAR(128) NULL,
			category_id INT NULL,
			torrent_size BIGINT UNSIGNED NULL,
			is_freeleech TINYINT(1) NOT NULL DEFAULT 0,
			qbt_hash VARCHAR(128) NULL,
			success TINYINT(1) NOT NULL,
			error_message TEXT NULL,
			ip_address VARCHAR(45) NULL,
			user_agent VARCHAR(512) NULL,
			deleted_at DATETIME NULL,
			deleted_by_user_id BIGINT UNSIGNED NULL,
			deleted_by_username VARCHAR(255) NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY idx_download_events_user_id (user_id),
			KEY idx_download_events_created_at (created_at),
			KEY idx_download_events_qbt_hash (qbt_hash),
			KEY idx_download_events_deleted_at (deleted_at),
			KEY idx_download_events_tvmaze_id (tvmaze_id),
			KEY idx_download_events_tvmaze_episode_id (tvmaze_episode_id),
			CONSTRAINT fk_download_events_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
			CONSTRAINT fk_download_events_deleted_by_user FOREIGN KEY (deleted_by_user_id) REFERENCES users(id) ON DELETE SET NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS download_delete_requests (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			download_event_id BIGINT UNSIGNED NOT NULL,
			requested_by_user_id BIGINT UNSIGNED NULL,
			requested_by_username VARCHAR(255) NOT NULL,
			status ENUM('pending', 'approved', 'rejected', 'hit_and_run') NOT NULL DEFAULT 'pending',
			request_note TEXT NULL,
			approved_by_user_id BIGINT UNSIGNED NULL,
			approved_by_username VARCHAR(255) NULL,
			approved_at DATETIME NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY idx_download_delete_requests_download_event_id (download_event_id),
			KEY idx_download_delete_requests_status (status),
			KEY idx_download_delete_requests_created_at (created_at),
			CONSTRAINT fk_download_delete_requests_download_event FOREIGN KEY (download_event_id) REFERENCES download_events(id) ON DELETE CASCADE,
			CONSTRAINT fk_download_delete_requests_requested_by_user FOREIGN KEY (requested_by_user_id) REFERENCES users(id) ON DELETE SET NULL,
			CONSTRAINT fk_download_delete_requests_approved_by_user FOREIGN KEY (approved_by_user_id) REFERENCES users(id) ON DELETE SET NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS tv_show_subscriptions (
				id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
				user_id BIGINT UNSIGNED NOT NULL,
			tvmaze_show_id BIGINT UNSIGNED NOT NULL,
			show_name VARCHAR(255) NULL,
			enabled TINYINT(1) NOT NULL DEFAULT 1,
			last_synced_at DATETIME NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY uq_tv_show_subscriptions_user_show (user_id, tvmaze_show_id),
			KEY idx_tv_show_subscriptions_show_id (tvmaze_show_id),
				KEY idx_tv_show_subscriptions_enabled (enabled),
				CONSTRAINT fk_tv_show_subscriptions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS tv_show_auto_install_qualities (
				id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
				user_id BIGINT UNSIGNED NOT NULL,
				tvmaze_show_id BIGINT UNSIGNED NOT NULL,
				preferred_quality ENUM('1080', '2160') NOT NULL DEFAULT '1080',
				dynamic_range ENUM('any', 'dv', 'hdr') NOT NULL DEFAULT 'any',
				enabled TINYINT(1) NOT NULL DEFAULT 0,
				created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				PRIMARY KEY (id),
				UNIQUE KEY uq_tv_show_auto_install_qualities_user_show_quality (user_id, tvmaze_show_id, preferred_quality),
				KEY idx_tv_show_auto_install_qualities_show_id (tvmaze_show_id),
				KEY idx_tv_show_auto_install_qualities_enabled (enabled),
				CONSTRAINT fk_tv_show_auto_install_qualities_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS tv_episode_jobs (
				id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
				subscription_id BIGINT UNSIGNED NULL,
			user_id BIGINT UNSIGNED NOT NULL,
			tvmaze_show_id BIGINT UNSIGNED NOT NULL,
			tvmaze_episode_id BIGINT UNSIGNED NOT NULL,
			episode_name VARCHAR(255) NULL,
			season_number INT NULL,
			episode_number INT NULL,
			airstamp DATETIME NULL,
			airtime_known TINYINT(1) NOT NULL DEFAULT 1,
			preferred_quality ENUM('1080', '2160') NOT NULL DEFAULT '1080',
			dynamic_range ENUM('any', 'dv', 'hdr') NOT NULL DEFAULT 'any',
			status ENUM('pending', 'searching', 'downloaded', 'failed') NOT NULL DEFAULT 'pending',
			attempt_count INT UNSIGNED NOT NULL DEFAULT 0,
			next_check_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			last_checked_at DATETIME NULL,
			last_error TEXT NULL,
			downloaded_download_event_id BIGINT UNSIGNED NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY uq_tv_episode_jobs_user_episode_quality (user_id, tvmaze_episode_id, preferred_quality),
			KEY idx_tv_episode_jobs_due (status, next_check_at),
			KEY idx_tv_episode_jobs_show_id (tvmaze_show_id),
			KEY idx_tv_episode_jobs_subscription_id (subscription_id),
			CONSTRAINT fk_tv_episode_jobs_subscription FOREIGN KEY (subscription_id) REFERENCES tv_show_subscriptions(id) ON DELETE SET NULL,
			CONSTRAINT fk_tv_episode_jobs_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			CONSTRAINT fk_tv_episode_jobs_download_event FOREIGN KEY (downloaded_download_event_id) REFERENCES download_events(id) ON DELETE SET NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}

	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("schema migration failed: %w", err)
		}
	}

	if err := ensureColumnExists(db, "download_events", "tvmaze_id", "VARCHAR(128) NULL AFTER `filename`"); err != nil {
		return fmt.Errorf("schema migration failed adding download_events.tvmaze_id: %w", err)
	}
	if err := ensureColumnExists(db, "download_events", "tvmaze_episode_id", "VARCHAR(128) NULL AFTER `tvmaze_id`"); err != nil {
		return fmt.Errorf("schema migration failed adding download_events.tvmaze_episode_id: %w", err)
	}
	if err := ensureIndexExists(db, "download_events", "idx_download_events_tvmaze_id", "`tvmaze_id`"); err != nil {
		return fmt.Errorf("schema migration failed adding index download_events.idx_download_events_tvmaze_id: %w", err)
	}
	if err := ensureIndexExists(db, "download_events", "idx_download_events_tvmaze_episode_id", "`tvmaze_episode_id`"); err != nil {
		return fmt.Errorf("schema migration failed adding index download_events.idx_download_events_tvmaze_episode_id: %w", err)
	}
	if err := ensureColumnExists(db, "tv_episode_jobs", "airtime_known", "TINYINT(1) NOT NULL DEFAULT 1 AFTER `airstamp`"); err != nil {
		return fmt.Errorf("schema migration failed adding tv_episode_jobs.airtime_known: %w", err)
	}
	if err := ensureColumnExists(db, "tv_show_auto_install_qualities", "dynamic_range", "ENUM('any', 'dv', 'hdr') NOT NULL DEFAULT 'any' AFTER `preferred_quality`"); err != nil {
		return fmt.Errorf("schema migration failed adding tv_show_auto_install_qualities.dynamic_range: %w", err)
	}
	if err := ensureColumnExists(db, "tv_episode_jobs", "dynamic_range", "ENUM('any', 'dv', 'hdr') NOT NULL DEFAULT 'any' AFTER `preferred_quality`"); err != nil {
		return fmt.Errorf("schema migration failed adding tv_episode_jobs.dynamic_range: %w", err)
	}
	if err := ensureColumnExists(db, "download_delete_requests", "safe_to_delete_at", "DATETIME NULL AFTER `approved_at`"); err != nil {
		return fmt.Errorf("schema migration failed adding download_delete_requests.safe_to_delete_at: %w", err)
	}
	if err := ensureColumnExists(db, "download_delete_requests", "auto_delete_at", "DATETIME NULL AFTER `safe_to_delete_at`"); err != nil {
		return fmt.Errorf("schema migration failed adding download_delete_requests.auto_delete_at: %w", err)
	}
	if _, err := db.Exec(
		"ALTER TABLE `download_delete_requests` MODIFY COLUMN `status` ENUM('pending', 'approved', 'rejected', 'hit_and_run') NOT NULL DEFAULT 'pending'",
	); err != nil {
		return fmt.Errorf("schema migration failed widening download_delete_requests.status enum: %w", err)
	}
	if err := ensureIndexExists(db, "download_delete_requests", "idx_download_delete_requests_auto_delete_at", "`auto_delete_at`"); err != nil {
		return fmt.Errorf("schema migration failed adding index download_delete_requests.idx_download_delete_requests_auto_delete_at: %w", err)
	}
	return nil
}

func ensureColumnExists(db *sql.DB, tableName string, columnName string, definition string) error {
	var count int
	if err := db.QueryRow(
		`SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = ?
		  AND COLUMN_NAME = ?`,
		tableName,
		columnName,
	).Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	statement := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s", tableName, columnName, definition)
	if _, err := db.Exec(statement); err != nil {
		return err
	}

	return nil
}

func ensureIndexExists(db *sql.DB, tableName string, indexName string, columns string) error {
	var count int
	if err := db.QueryRow(
		`SELECT COUNT(*)
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = ?
		  AND INDEX_NAME = ?`,
		tableName,
		indexName,
	).Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	statement := fmt.Sprintf("ALTER TABLE `%s` ADD INDEX `%s` (%s)", tableName, indexName, columns)
	if _, err := db.Exec(statement); err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1061 {
			return nil
		}
		return err
	}

	return nil
}

func dbConn() (*sql.DB, error) {
	if mysqlDB == nil {
		return nil, fmt.Errorf("mysql connection is not initialized")
	}

	return mysqlDB, nil
}
