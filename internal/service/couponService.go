package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ssipflow/coupon-issuance/internal/dto"
	"github.com/ssipflow/coupon-issuance/internal/entity"
	"github.com/ssipflow/coupon-issuance/internal/repo"
	"github.com/ssipflow/coupon-issuance/internal/task"
	"gorm.io/gorm"
	"log"
	"time"
)

type CouponService struct {
	redisClient     *repo.RedisClient
	mySqlRepository *repo.MySqlRepository
}

func NewCouponService(redis *repo.RedisClient, db *repo.MySqlRepository) *CouponService {
	return &CouponService{redis, db}
}

func (c *CouponService) CreateCampaign(ctx context.Context, name string, total int64, startAt time.Time) (int32, error) {
	campaign := &entity.Campaign{
		Name:       name,
		TotalCount: total,
		StartTime:  startAt,
	}

	if err := c.mySqlRepository.CreateCampaign(ctx, campaign); err != nil {
		return 0, err
	}

	return campaign.ID, nil
}

func (c *CouponService) GetCampaign(ctx context.Context, id int32) (*dto.Campaign, error) {
	campaign, err := c.mySqlRepository.GetCampaignWithCouponsById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("RECORD_NOT_FOUND")
		}
		log.Println(err.Error())
		return nil, fmt.Errorf("MYSQL: %s", "INTERNAL_SERVER_ERROR")
	}

	return campaign, nil
}

func (c *CouponService) IssueCoupon(ctx context.Context, campaignId int32, userId int32) error {
	campaign, err := c.mySqlRepository.GetCampaignById(campaignId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("RECORD_NOT_FOUND")
		}
		log.Println(err.Error())
		return fmt.Errorf("MYSQL: %s", "INTERNAL_SERVER_ERROR")
	}

	if time.Now().Before(campaign.StartTime) {
		return fmt.Errorf("CAMPAIGN_NOT_STARTED")
	}

	claimedKey := fmt.Sprintf("coupon:claimed:%d:%d", campaignId, userId)
	ok, err := c.redisClient.SetNx(ctx, claimedKey, 1, 24*time.Hour)
	if err != nil {
		log.Println(err.Error())
		return fmt.Errorf("SETNX: %s", "INTERNAL_SERVER_ERROR")
	}
	if !ok {
		return fmt.Errorf("ALREADY_CLAIMED")
	}

	issuedKey := fmt.Sprintf("coupon:issued:%d", campaignId)
	count, err := c.redisClient.Incr(ctx, issuedKey)
	if err != nil {
		log.Println(err.Error())
		return fmt.Errorf("INCR: %s", "INTERNAL_SERVER_ERROR")
	}
	if count > campaign.TotalCount {
		return fmt.Errorf("COUPON_SOLD_OUT")
	}

	payload, _ := json.Marshal(map[string]int32{
		"campaign_id": campaignId,
		"user_id":     userId,
	})
	if err := task.JobProduce("coupon:issue", payload); err != nil {
		log.Println(err.Error())
		return fmt.Errorf("PRODUCER: %s", "INTERNAL_SERVER_ERROR")
	}

	return nil
}
