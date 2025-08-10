package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"golang-rest-api-template/pkg/config"
	"golang-rest-api-template/pkg/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database interface {
	Offset(offset int) *gorm.DB
	Limit(limit int) *gorm.DB
	Find(interface{}, ...interface{}) *gorm.DB
	Create(value interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) Database
	Delete(interface{}, ...interface{}) *gorm.DB
	Model(model interface{}) *gorm.DB
	First(dest interface{}, conds ...interface{}) Database
	Updates(interface{}) *gorm.DB
	Order(value interface{}) *gorm.DB
	Select(query interface{}, args ...interface{}) Database
	Save(value interface{}) *gorm.DB
	WithContext(ctx context.Context) Database
	Error() error
}

type GormDatabase struct {
	*gorm.DB
}

func (db *GormDatabase) Where(query interface{}, args ...interface{}) Database {
	return &GormDatabase{db.DB.Where(query, args...)}
}

func (db *GormDatabase) First(dest interface{}, conds ...interface{}) Database {
	return &GormDatabase{db.DB.First(dest, conds...)}
}

func (db *GormDatabase) Select(query interface{}, args ...interface{}) Database {
	return &GormDatabase{db.DB.Select(query, args...)}
}

func (db *GormDatabase) WithContext(ctx context.Context) Database {
	return &GormDatabase{db.DB.WithContext(ctx)}
}

func (db *GormDatabase) Error() error {
	return db.DB.Error
}

func NewDatabase(cfg *config.Config) *gorm.DB {
	var database *gorm.DB
	var err error

	// 构造 DSN，包含 statement_timeout
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s options='-c statement_timeout=%dms'",
		cfg.PostgresHost,
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresDB,
		cfg.PostgresPort,
		cfg.PostgresSSLMode,
		cfg.DBQueryTimeoutMs,
	)

	// 指数退避重试
	backoff := time.Duration(cfg.DBConnectInitialBackoffMs) * time.Millisecond
	if backoff <= 0 {
		backoff = 500 * time.Millisecond
	}
	maxRetries := cfg.DBConnectMaxRetries
	if maxRetries <= 0 {
		maxRetries = 5
	}

	for i := 0; i < maxRetries; i++ {
		database, err = gorm.Open(postgres.New(postgres.Config{DSN: dsn}), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("DB connect attempt %d failed: %v", i+1, err)
		time.Sleep(backoff)
		backoff *= 2
	}
	if err != nil {
		log.Fatalf("failed to initialize database after retries: %v", err)
	}

	// 连接池设置
	sqlDB, err := database.DB()
	if err == nil {
		if cfg.PostgresMaxOpenConns > 0 {
			sqlDB.SetMaxOpenConns(cfg.PostgresMaxOpenConns)
		}
		if cfg.PostgresMaxIdleConns > 0 {
			sqlDB.SetMaxIdleConns(cfg.PostgresMaxIdleConns)
		}
		if cfg.PostgresConnMaxIdleTimeSec > 0 {
			sqlDB.SetConnMaxIdleTime(time.Duration(cfg.PostgresConnMaxIdleTimeSec) * time.Second)
		}
		if cfg.PostgresConnMaxLifetimeSec > 0 {
			sqlDB.SetConnMaxLifetime(time.Duration(cfg.PostgresConnMaxLifetimeSec) * time.Second)
		}
	} else {
		log.Printf("warning: failed to get sql DB: %v", err)
	}

	// 自动迁移（注意：生产环境建议改用迁移工具）
	_ = database.AutoMigrate(&models.Book{})
	_ = database.AutoMigrate(&models.User{})

	return database
}
