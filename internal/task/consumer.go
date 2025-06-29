package task

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/ssipflow/coupon-issuance/internal/entity"
	"github.com/ssipflow/coupon-issuance/internal/infra"
	"github.com/ssipflow/coupon-issuance/internal/repo"
	"github.com/ssipflow/coupon-issuance/pkg/util"
	"log"
)

type Consumer struct {
	couponRepository *repo.CouponRepository
	redisClient      *infra.RedisClient
}

func NewConsumer(redisClient *infra.RedisClient, mySqlRepository *repo.CouponRepository) *Consumer {
	return &Consumer{
		couponRepository: mySqlRepository,
		redisClient:      redisClient,
	}
}

func (c *Consumer) IssueCouponProcessor() asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload struct {
			CampaignID int32 `json:"campaign_id"`
			UserID     int32 `json:"user_id"`
		}
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return err
		}

		lockKey := fmt.Sprintf("lock:coupon:campaign:%d:user:%d", payload.CampaignID, payload.UserID)

		tx := c.couponRepository.GetDB().Begin()
		if tx.Error != nil {
			log.Printf("IssueCouponProcessor.GetDB.Begin.err: %v", tx.Error)
			_, _ = c.redisClient.Del(ctx, lockKey)
			return tx.Error
		}

		var needRollback = true
		defer func() {
			if needRollback {
				_ = tx.Rollback()
			} else {
				if err := tx.Commit().Error; err != nil {
					log.Printf("IssueCouponProcessor.Commit.err: %v", err)
				}
			}
			_, _ = c.redisClient.Del(ctx, lockKey)
		}()

		code := util.GenerateCouponCode()
		coupon := &entity.Coupon{
			CampaignID: payload.CampaignID,
			UserID:     payload.UserID,
			Code:       code,
		}

		if err := c.couponRepository.IncrementCampaignCurrentCoupon(tx, ctx, payload.CampaignID); err != nil {
			log.Printf("IssueCoupon.UpdateCampaignCurrentCoupon.err: %v", err)
			return err
		}

		if err := c.couponRepository.CreateCoupon(tx, ctx, coupon); err != nil {
			log.Printf("IssueCouponProcessor.CreateCoupon.err: %v", err)
			return err
		}

		cacheKey := "cache:coupon:issued"
		issuedCouponCacheKey := fmt.Sprintf("campaign:%d:user:%d", payload.CampaignID, payload.UserID)
		_, _ = c.redisClient.HSet(ctx, cacheKey, map[string]interface{}{issuedCouponCacheKey: code})

		needRollback = false
		log.Printf("[SUCCESS] Issued coupon to user %d in campaign %d", payload.UserID, payload.CampaignID)
		return nil
	}
}
