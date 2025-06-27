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

func IssueCouponProcessor(db *repo.MySqlRepository, redisClient *repo.RedisClient) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload struct {
			CampaignID uint   `json:"campaign_id"`
			UserID     string `json:"user_id"`
		}
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return err
		}

		code := util.GenerateCouponCode()

		coupon := entity.Coupon{
			CampaignID: payload.CampaignID,
			UserID:     payload.UserID,
			Code:       code,
		}
		if err := db.CreateCoupon(&coupon); err != nil {
			return err
		}

		codeKey := fmt.Sprintf("coupon:codes:%s", payload.CampaignID)
		redisClient.SAdd(ctx, codeKey, code)

		return nil
	}
}
