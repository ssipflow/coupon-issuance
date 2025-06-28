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
	"strconv"
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
		issuedKey := fmt.Sprintf("coupon:issued:campaign:%d", payload.CampaignID)

		tx := c.couponRepository.GetDB().Begin()
		if tx.Error != nil {
			log.Printf("[ERROR] Transaction begin failed: %v", tx.Error)
			_, _ = c.redisClient.Del(ctx, lockKey)
			return tx.Error
		}

		var needRollback = true
		defer func() {
			if needRollback {
				_ = tx.Rollback()
				_, _ = c.redisClient.Decr(ctx, issuedKey)
			} else {
				if err := tx.Commit().Error; err != nil {
					log.Printf("[ERROR] Commit failed: %v", err)
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

		if err := c.couponRepository.CreateCoupon(tx, ctx, coupon); err != nil {
			log.Printf("CreateCoupon failed: %v", err)
			return err
		}

		val, _ := c.redisClient.Get(ctx, issuedKey)
		issuedCount, _ := strconv.Atoi(val)
		if err := c.couponRepository.UpdateCampaignCurrentCoupon(tx, ctx, payload.CampaignID, int64(issuedCount)); err != nil {
			log.Printf("UpdateCampaignCurrentCoupon failed: %v", err)
			return err
		}

		needRollback = false
		log.Printf("[SUCCESS] Issued coupon to user %d in campaign %d", payload.UserID, payload.CampaignID)
		return nil
	}
}
