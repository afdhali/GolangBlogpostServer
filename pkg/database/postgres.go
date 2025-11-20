package database

import (
	"fmt"
	"log"

	"github.com/afdhali/GolangBlogpostServer/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgresDB(cfg *config.Config) (*gorm.DB, error){
	dsn := cfg.Database.DSN()

	logMode := logger.Silent
	if cfg.App.Debug {
		logMode = logger.Info
	}

	db, err:= gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logMode),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w",err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w",err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("Database connected succesfully")
	return db, nil
}