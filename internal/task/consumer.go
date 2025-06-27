package task

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/ssipflow/coupon-issuance/internal/entity"
	"github.com/ssipflow/coupon-issuance/internal/repo"
	"github.com/ssipflow/coupon-issuance/pkg/util"
)

type Consumer struct {
	mySqlRepository *repo.MySqlRepository
	redisClient     *repo.RedisClient
}

func NewConsumer(redisClient *repo.RedisClient, mySqlRepository *repo.MySqlRepository) *Consumer {
	return &Consumer{
		mySqlRepository: mySqlRepository,
		redisClient:     redisClient,
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

		code := util.GenerateCouponCode()

		coupon := &entity.Coupon{
			CampaignID: payload.CampaignID,
			UserID:     payload.UserID,
			Code:       code,
		}
		if err := c.mySqlRepository.CreateCoupon(ctx, coupon); err != nil {
			return err
		}

		codeKey := fmt.Sprintf("coupon:codes:%d", payload.CampaignID)
		c.redisClient.SAdd(ctx, codeKey, code)

		return nil
	}
}
