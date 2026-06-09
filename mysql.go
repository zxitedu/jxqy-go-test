package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

func connectMySQL(cfg databaseConfig, timeout time.Duration) (*sql.DB, string, error) {
	driverConfig, err := mysqlDriver.ParseDSN(cfg.Source)
	if err != nil {
		return nil, "", err
	}
	if driverConfig.Timeout == 0 {
		driverConfig.Timeout = timeout
	}
	if driverConfig.ReadTimeout == 0 {
		driverConfig.ReadTimeout = timeout
	}
	if driverConfig.WriteTimeout == 0 {
		driverConfig.WriteTimeout = timeout
	}

	db, err := sql.Open(cfg.Driver, driverConfig.FormatDSN())
	if err != nil {
		return nil, "", err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, "", err
	}

	dbName, err := currentDatabase(ctx, db)
	if err != nil {
		db.Close()
		return nil, "", err
	}

	return db, dbName, nil
}

func mysqlDatabaseName(source string) (string, error) {
	driverConfig, err := mysqlDriver.ParseDSN(source)
	if err != nil {
		return "", err
	}
	if driverConfig.DBName == "" {
		return "", fmt.Errorf("database name is empty")
	}
	return driverConfig.DBName, nil
}

func currentDatabase(ctx context.Context, db *sql.DB) (string, error) {
	var dbName sql.NullString
	if err := db.QueryRowContext(ctx, "SELECT DATABASE()").Scan(&dbName); err != nil {
		return "", err
	}
	if !dbName.Valid || dbName.String == "" {
		return "", fmt.Errorf("mysql did not select a database")
	}
	return dbName.String, nil
}

func listUsernames(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SELECT username FROM `user` ORDER BY username")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	usernames := make([]string, 0)
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, err
		}
		usernames = append(usernames, username)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return usernames, nil
}
