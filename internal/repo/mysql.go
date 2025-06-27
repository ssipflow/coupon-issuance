package repo

import (
	"context"
	"errors"
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

func (r *MySqlRepository) CreateCampaign(ctx context.Context, campaign *entity.Campaign) error {
	return r.db.WithContext(ctx).Create(campaign).Error
}

func (r *MySqlRepository) CreateCoupon(ctx context.Context, coupon *entity.Coupon) error {
	return r.db.WithContext(ctx).Create(coupon).Error
}

func (r *MySqlRepository) GetCampaignById(id int32) (*entity.Campaign, error) {
	var campaign entity.Campaign
	if err := r.db.
		Where("id = ?", id).
		First(&campaign).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println("CAMPAIGN_NOT_FOUND")
		}
		return nil, err
	}
	return &campaign, nil
}
