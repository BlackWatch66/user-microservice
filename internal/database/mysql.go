package database

import (
	"fmt"
	"log"
	"time"

	"github.com/blackwatch66/user-microservice/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

// InitDatabase initializes database connection and performs migration
func InitDatabase(dsn string) (*gorm.DB, error) {
	var err error
	// 配置 GORM，禁用自动创建外键和索引
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Log level can be adjusted as needed
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键约束自动创建
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// Connection pool settings
	sqlDB, err := DB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get generic database object: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection established.")

	// 我们假设表结构已经在init.sql中创建，不需要自动迁移
	// 如果需要验证结构，可以使用HasTable等方法检查

	// 检查users表是否存在
	if !DB.Migrator().HasTable(&model.User{}) {
		log.Println("Users table does not exist, creating...")
		if err := DB.Migrator().CreateTable(&model.User{}); err != nil {
			return nil, fmt.Errorf("failed to create users table: %w", err)
		}
	}

	// 检查addresses表是否存在
	if !DB.Migrator().HasTable(&model.Address{}) {
		log.Println("Addresses table does not exist, creating...")
		if err := DB.Migrator().CreateTable(&model.Address{}); err != nil {
			return nil, fmt.Errorf("failed to create addresses table: %w", err)
		}
	}

	log.Println("Database migration completed.")

	return DB, nil
}

// GetDB returns the initialized database connection instance
func GetDB() *gorm.DB {
	return DB
} 