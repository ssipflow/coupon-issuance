package stress

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	baseIssueURL     = "http://localhost:8080/coupon.v1.CouponService/IssueCoupon"
	baseCreateURL    = "http://localhost:8080/coupon.v1.CouponService/CreateCampaign"
	campaignCount    = 5
	concurrency      = 1000
	totalPerCampaign = 1000
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

func createCampaigns(n int, countPerCampaign int) []int32 {
	campaignIDs := make([]int32, 0, n)

	for i := 0; i < n; i++ {
		body := map[string]interface{}{
			"name":        fmt.Sprintf("campaign-%d", i),
			"couponLimit": countPerCampaign,
			"startTime":   time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
		}
		b, _ := json.Marshal(body)

		resp, err := http.Post(baseCreateURL, "application/json", bytes.NewReader(b))
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

func TestIssueCoupon_Stress(t *testing.T) {
	campaignIDs := createCampaigns(campaignCount, totalPerCampaign)

	var wg sync.WaitGroup
	var success, failed int64
	dedup := sync.Map{}

	client := &http.Client{Timeout: 5 * time.Second}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(uid int) {
			defer wg.Done()

			campaignID := campaignIDs[uid%len(campaignIDs)]

			reqBody, _ := json.Marshal(IssueCouponRequest{
				CampaignID: campaignID,
				UserID:     int32(uid),
			})

			resp, err := client.Post(baseIssueURL, "application/json", bytes.NewReader(reqBody))
			if err != nil {
				log.Printf("[ERROR] User %d: %v\n", uid, err)
				atomic.AddInt64(&failed, 1)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				if _, exists := dedup.LoadOrStore(uid, true); exists {
					log.Printf("[DUPLICATE] User %d received multiple responses!", uid)
					atomic.AddInt64(&failed, 1)
				} else {
					atomic.AddInt64(&success, 1)
				}
			} else {
				log.Printf("[ERROR] %v", http.StatusText(resp.StatusCode))
				atomic.AddInt64(&failed, 1)
			}
		}(i + 1)
	}

	wg.Wait()
	log.Printf("Total: %d, Success: %d, Failed: %d", concurrency, success, failed)

	successRate := float64(success) / float64(concurrency) * 100
	log.Printf("Success Rate: %.2f%%", successRate)

	if success+failed != int64(concurrency) {
		t.Errorf("Mismatch in total requests: success + failed != concurrency")
	}

	if successRate < 99.0 {
		t.Errorf("Success rate too low: %.2f%%", successRate)
	}
}
