package repo

import (
	"github.com/ssipflow/coupon-issuance/internal/entity"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
)

type MySqlRepository struct {
	db *gorm.DB
}

func NewRepository() *MySqlRepository {
	dsn := os.Getenv("MYSQL_DSN")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	return &MySqlRepository{
		db: db,
	}
}

func (r *MySqlRepository) CreateCoupon(coupon *entity.Coupon) error {
	return r.db.Create(coupon).Error
}
