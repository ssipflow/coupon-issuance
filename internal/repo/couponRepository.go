package repo

import (
	"context"
	"errors"
	"github.com/ssipflow/coupon-issuance/internal/dto"
	"github.com/ssipflow/coupon-issuance/internal/entity"
	"gorm.io/gorm"
	"log"
)

type CouponRepository struct {
	db *gorm.DB
}

func NewCouponRepository(db *gorm.DB) *CouponRepository {
	return &CouponRepository{db: db}
}

func (r *CouponRepository) GetDB() *gorm.DB {
	if r.db == nil {
		log.Fatal("CouponRepository: db is nil")
	}
	return r.db
}

func (r *CouponRepository) CreateCampaign(ctx context.Context, campaign *entity.Campaign) error {
	err := r.db.WithContext(ctx).Create(campaign).Error
	if err != nil {
		log.Printf("CreateCampaign.err: %v\n", err.Error())
		return err
	}
	return nil
}

func (r *CouponRepository) GetCampaignWithCouponsById(ctx context.Context, id int32) (*dto.Campaign, error) {
	var campaign entity.Campaign
	if err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&campaign).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Printf("GetCampaignWithCouponsById.err: %v\n", err.Error())
		return nil, err
	}

	coupon := &entity.Coupon{}
	var codes []string
	if err := r.db.WithContext(ctx).
		Table(coupon.GetTableName()).
		Where("campaign_id = ?", id).
		Pluck("code", &codes).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			codes = []string{}
		} else {
			log.Printf("GetCampaignWithCouponsById.pluck.err: %v\n", err.Error())
			return nil, err
		}
	}

	return dto.NewCampaignFromEntity(&campaign).SetCouponCodes(codes), nil
}

func (r *CouponRepository) GetCampaignById(id int32) (*dto.Campaign, error) {
	var campaign entity.Campaign
	if err := r.db.
		Where("id = ?", id).
		First(&campaign).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Printf("GetCampaignById.err: %v\n", err.Error())
		return nil, err
	}

	return dto.NewCampaignFromEntity(&campaign), nil
}

func (r *CouponRepository) CreateCoupon(tx *gorm.DB, ctx context.Context, coupon *entity.Coupon) error {
	db := r.db
	if tx != nil {
		db = tx
	}

	err := db.WithContext(ctx).Create(coupon).Error
	if err != nil {
		log.Printf("CreateCoupon.err: %v\n", err.Error())
		return err
	}
	return nil
}

func (r *CouponRepository) IncrementCampaignCurrentCoupon(tx *gorm.DB, ctx context.Context, campaignId int32) error {
	db := r.db
	if tx != nil {
		db = tx
	}

	err := db.WithContext(ctx).
		Model(&entity.Campaign{}).
		Where("id = ?", campaignId).
		UpdateColumn("current_coupon", gorm.Expr("current_coupon + ?", 1)).Error
	if err != nil {
		log.Printf("UpdateCampaignCurrentCoupon.err: %v\n", err.Error())
		return err
	}
	return nil
}

func (r *CouponRepository) GetCouponByUserAndCampaign(ctx context.Context, userId int32, campaignId int32) (*entity.Coupon, error) {
	var coupon entity.Coupon
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND campaign_id = ?", userId, campaignId).
		First(&coupon).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Printf("GetCouponByUserAndCampaign.err: %v\n", err.Error())
		return nil, err
	}
	return &coupon, nil
}
