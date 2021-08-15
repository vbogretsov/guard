package main

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var dialects = map[string]func(string) gorm.Dialector{
	"postgres": postgres.Open,
	"mysql":    mysql.Open,
	"sqlite":   sqlite.Open,
}

func dbconnect(cfg Conf) (*gorm.DB, error) {
	dialect, ok := dialects[cfg.DBDriver]
	if !ok {
		return nil, fmt.Errorf("unsupported database driver: %v", cfg.DBDriver)
	}

	db, err := gorm.Open(dialect(cfg.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		return nil, err
	}

	sqldb, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqldb.SetMaxOpenConns(cfg.DBMaxOpenConn)
	sqldb.SetMaxIdleConns(cfg.DBMaxIddleConn)
	sqldb.SetConnMaxLifetime(cfg.DBConnMaxLifetime)
	sqldb.SetConnMaxIdleTime(cfg.DBConnMaxIddleTime)

	return db, err
}
