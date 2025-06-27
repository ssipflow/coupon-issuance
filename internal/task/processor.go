package task

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/ssipflow/coupon-issuance/pkg/util"
	"gorm.io/gorm"
)

func IssueCouponProcessor(db *gorm.DB, redisClient *redis.Client) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload struct {
			CampaignID string `json:"campaign_id"`
			UserID     string `json:"user_id"`
		}
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return err
		}

		code := util.GenerateCouponCode()

		coupon := model.Coupon{
			CampaignID: payload.CampaignID,
			UserID:     payload.UserID,
			Code:       code,
		}
		if err := db.Create(&coupon).Error; err != nil {
			return err
		}

		codeKey := fmt.Sprintf("coupon:codes:%s", payload.CampaignID)
		redisClient.SAdd(ctx, codeKey, code)

		return nil
	}
}
