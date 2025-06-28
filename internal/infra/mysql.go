package infra

import (
	"github.com/ssipflow/coupon-issuance/internal/entity"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strconv"
	"time"
)

type MySql struct {
	db *gorm.DB
}

func NewMySql() *MySql {
	dsn := os.Getenv("MYSQL_DSN")

	var gormDB *gorm.DB
	var err error

	maxRetry := 10
	for i := 0; i < maxRetry; i++ {
		gormDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			}),
		})
		if err == nil {
			sqlDB, pingErr := gormDB.DB()
			if pingErr == nil && sqlDB.Ping() == nil {
				log.Printf("Connected to MySQL after %d attempt(s)\n", i+1)
				break
			}
		}

		log.Printf("Waiting for MySQL... (%d/%d) err=%v\n", i+1, maxRetry, err)
		time.Sleep(2 * time.Second)
	}

	if gormDB == nil || err != nil {
		log.Printf("failed to connect to MySQL after %d retries: %v\n", maxRetry, err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB from gorm: %v", err)
	}

	// generate os.Getenv("MYSQL_MAX_CONN") to an integer
	maxConn, _ := strconv.Atoi(os.Getenv("MYSQL_MAX_CONN"))
	maxIdleConn, _ := strconv.Atoi(os.Getenv("MYSQL_MAX_IDLE_CONN"))
	maxLifetime, _ := strconv.Atoi(os.Getenv("MYSQL_MAX_LIFETIME"))
	maxIdleTime, _ := strconv.Atoi(os.Getenv("MYSQL_MAX_IDLE_TIME"))

	sqlDB.SetMaxOpenConns(maxConn)
	sqlDB.SetMaxIdleConns(maxIdleConn)
	sqlDB.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(maxIdleTime) * time.Minute)

	return &MySql{
		db: gormDB,
	}
}

func (r *MySql) GetDB() *gorm.DB {
	return r.db
}

func (r *MySql) AutoMigrate() error {
	if err := r.db.Migrator().AutoMigrate(&entity.Coupon{}, &entity.Campaign{}); err != nil {
		log.Fatalf("failed to auto migrate: %v", err)
		return err
	}
	return nil
}

func (r *MySql) DropTables() error {
	log.Println("Dropping tables: Coupon and Campaign")
	if err := r.db.Migrator().DropTable(&entity.Coupon{}, &entity.Campaign{}); err != nil {
		log.Fatalf("failed to drop tables: %v", err)
		return err
	}
	return nil
}
