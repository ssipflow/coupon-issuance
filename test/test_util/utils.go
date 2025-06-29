package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	BaseIssueURL          = "http://localhost:8080/coupon.v1.CouponService/IssueCoupon"
	BaseCreateURL         = "http://localhost:8080/coupon.v1.CouponService/CreateCampaign"
	CampaignCount         = 5
	Concurrency           = 1000
	CouponLimitForStress  = 500
	CouponLimitForExceed  = 100
	CampaignNameForStress = "campaign-stress"
	CampaignNameForLimit  = "campaign-limit"
)

type IssueCouponRequest struct {
	CampaignID int32 `json:"campaignId"`
	UserID     int32 `json:"userId"`
}

type IssueCouponResponse struct {
	CouponCode string `json:"couponCode"`
}

type CreateCampaignResponse struct {
	CampaignID int32 `json:"campaignId"`
}

func CreateCampaigns(count int, limitPerCampaign int, name string) []int32 {
	campaignIDs := make([]int32, 0, count)

	for i := 0; i < count; i++ {
		body := map[string]interface{}{
			"name":        fmt.Sprintf("%s-%d", name, i),
			"couponLimit": limitPerCampaign,
			"startTime":   time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
		}
		b, _ := json.Marshal(body)

		resp, err := http.Post(BaseCreateURL, "application/json", bytes.NewReader(b))
		if err != nil {
			log.Fatalf("create campaign failed: %v", err)
		}
		defer resp.Body.Close()

		var res CreateCampaignResponse
		json.NewDecoder(resp.Body).Decode(&res)
		campaignIDs = append(campaignIDs, res.CampaignID)
	}

	return campaignIDs
}
