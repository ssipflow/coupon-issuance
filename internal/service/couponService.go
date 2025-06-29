package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/ssipflow/coupon-issuance/internal/dto"
	"github.com/ssipflow/coupon-issuance/internal/entity"
	"github.com/ssipflow/coupon-issuance/internal/errors"
	"github.com/ssipflow/coupon-issuance/internal/infra"
	"github.com/ssipflow/coupon-issuance/internal/repo"
	"github.com/ssipflow/coupon-issuance/internal/task"
	"log"
	"time"
)

type CouponService struct {
	redisClient      *infra.RedisClient
	couponRepository *repo.CouponRepository
}

func NewCouponService(redis *infra.RedisClient, db *repo.CouponRepository) *CouponService {
	return &CouponService{redis, db}
}

func (c *CouponService) CreateCampaign(ctx context.Context, name string, limit int64, startAt time.Time) (int32, error) {
	campaign := &entity.Campaign{
		Name:        name,
		CouponLimit: limit,
		StartTime:   startAt,
	}

	if err := c.couponRepository.CreateCampaign(ctx, campaign); err != nil {
		log.Printf("CreateCampaign.CreateCampaign.err: %v\n", err)
		return 0, errors.NewError(errors.ERR_INTERNAL_SERVER_ERROR)
	}

	return campaign.ID, nil
}

func (c *CouponService) GetCampaign(ctx context.Context, id int32) (*dto.Campaign, error) {
	campaign, err := c.couponRepository.GetCampaignWithCouponsById(ctx, id)
	if err != nil {
		log.Printf("GetCampaign.GetCampaignWithCouponsById.err: %v\n", err.Error())
		return nil, errors.NewError(errors.ERR_INTERNAL_SERVER_ERROR)
	}
	if campaign == nil {
		return nil, errors.NewError(errors.ERR_RECORD_NOT_FOUND)
	}

	return campaign, nil
}

func (c *CouponService) IssueCoupon(ctx context.Context, campaignId int32, userId int32) error {
	campaign, err := c.couponRepository.GetCampaignById(campaignId)
	if err != nil {
		log.Printf("IssueCoupon.GetCampaignById.err: %v\n", err.Error())
		return errors.NewError(errors.ERR_INTERNAL_SERVER_ERROR)
	}
	if campaign == nil {
		return errors.NewError(errors.ERR_RECORD_NOT_FOUND)
	}

	if time.Now().Before(campaign.GetStartTime()) {
		return errors.NewError(errors.ERR_CAMPAIGN_NOT_STARTED)
	}

	cacheKey := "cache:coupon:issued"
	issuedCouponCacheKey := fmt.Sprintf("campaign:%d:user:%d", campaignId, userId)
	isCouponIssued, err := c.redisClient.HExists(ctx, cacheKey, issuedCouponCacheKey)
	if err != nil {
		log.Printf("IssueCoupon.HExists.err: %v\n", err.Error())
		return errors.NewError(errors.ERR_INTERNAL_SERVER_ERROR)
	}
	if isCouponIssued {
		return errors.NewError(errors.ERR_COUPON_ALREADY_CLAIMED)
	}

	issuedCouponCountKey := fmt.Sprintf("coupon:issued:campaign:%d", campaignId)
	count, err := c.redisClient.Incr(ctx, issuedCouponCountKey)
	if err != nil {
		log.Printf("IssueCoupon.Incr.err: %v\n", err.Error())
		return errors.NewError(errors.ERR_INTERNAL_SERVER_ERROR)
	}
	if count == campaign.GetCouponLimit() {
		return errors.NewError(errors.ERR_COUPON_SOLD_OUT)
	}

	lockKey := fmt.Sprintf("lock:coupon:campaign:%d:user:%d", campaignId, userId)
	ok, err := c.redisClient.SetNx(ctx, lockKey, 1, 24*time.Hour)
	if err != nil {
		log.Printf("IssueCoupon.SetNx.err: %v\n", err.Error())
		return errors.NewError(errors.ERR_INTERNAL_SERVER_ERROR)
	}
	if !ok {
		return errors.NewError(errors.ERR_COUPON_ALREADY_CLAIMED)
	}

	payload, _ := json.Marshal(map[string]int32{
		"campaign_id": campaignId,
		"user_id":     userId,
	})

	asynqOpts := []asynq.Option{
		asynq.MaxRetry(3),
		asynq.Timeout(5 * time.Second),
		asynq.Queue("default"),
	}
	if err := task.ProduceTask("task:coupon:issue", payload, asynqOpts...); err != nil {
		log.Printf("IssueCoupon.JobProduce.err: %v\n", err.Error())
		return errors.NewError(errors.ERR_INTERNAL_SERVER_ERROR)
	}

	return nil
}
