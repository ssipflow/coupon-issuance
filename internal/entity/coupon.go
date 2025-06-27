package entity

import "time"

type Coupon struct {
	ID         uint   `gorm:"primaryKey"`
	CampaignID uint   `gorm:"index;not null"`
	UserID     string `gorm:"type:varchar(64);not null"`
	Code       string `gorm:"type:varchar(20);unique;not null"`
	CreatedAt  time.Time
}
