package entity

import "time"

type Coupon struct {
	ID         int32  `gorm:"primaryKey"`
	CampaignID int32  `gorm:"index;not null"`
	UserID     int32  `gorm:"not null"`
	Code       string `gorm:"type:varchar(20);unique;not null"`
	CreatedAt  time.Time
}

func (c *Coupon) GetTableName() string {
	return "coupons"
}
