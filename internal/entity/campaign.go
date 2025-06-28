package entity

import "time"

type Campaign struct {
	ID            int32     `gorm:"primaryKey"`
	Name          string    `gorm:"type:varchar(100);not null"`
	CouponLimit   int64     `gorm:"not null"`
	CurrentCoupon int64     `gorm:"not null;default:0"`
	StartTime     time.Time `gorm:"not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (c *Campaign) GetTableName() string {
	return "campaigns"
}
