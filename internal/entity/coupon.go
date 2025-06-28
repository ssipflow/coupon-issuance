package entity

import "time"

type Coupon struct {
	ID         int32  `gorm:"primaryKey"`
	CampaignID int32  `gorm:"not null;index:idx_campaign_user,unique;index:idx_campaign_code,unique"`
	UserID     int32  `gorm:"not null;index:idx_campaign_user,unique"`
	Code       string `gorm:"type:varchar(20);not null;uniqueIndex;index:idx_campaign_code,unique"`
	CreatedAt  time.Time
}

func (c *Coupon) GetTableName() string {
	return "coupons"
}
