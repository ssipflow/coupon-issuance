package limit

import (
	"bytes"
	"encoding/json"
	"github.com/ssipflow/coupon-issuance/test/test_util"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestIssueCoupon_LimitCoupon(t *testing.T) {
	start := time.Now()
	campaignIDs := util.CreateCampaigns(util.CampaignCount, util.CouponLimitForExceed, util.CampaignNameForLimit)
	duration := time.Since(start)
	log.Printf("Created %d campaigns in %v", len(campaignIDs), duration)

	var wg sync.WaitGroup
	var success, failed int64
	dedup := sync.Map{}

	client := &http.Client{Timeout: 5 * time.Second}

	for i := 0; i < util.Concurrency; i++ {
		wg.Add(1)
		go func(uid int) {
			defer wg.Done()

			campaignID := campaignIDs[uid%len(campaignIDs)]

			reqBody, _ := json.Marshal(util.IssueCouponRequest{
				CampaignID: campaignID,
				UserID:     int32(uid),
			})

			resp, err := client.Post(util.BaseIssueURL, "application/json", bytes.NewReader(reqBody))
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
				if resp.StatusCode == http.StatusInternalServerError {
					log.Printf("[ERROR] %v", http.StatusText(resp.StatusCode))
				}
				atomic.AddInt64(&failed, 1)
			}
		}(i + 1)
	}

	wg.Wait()
	log.Printf("Total: %d, Success: %d, Failed: %d", util.Concurrency, success, failed)

	successRate := float64(success) / float64(util.Concurrency) * 100
	log.Printf("Success Rate: %.2f%%", successRate)

	if success+failed != int64(util.Concurrency) {
		t.Errorf("Mismatch in total requests: success + failed != concurrency")
	}

	if successRate < 99.0 {
		t.Errorf("Success rate too low: %.2f%%", successRate)
	}
}
