package entity

import "time"

type Campaign struct {
	ID         int32     `gorm:"primaryKey"`
	Name       string    `gorm:"type:varchar(100);not null"`
	TotalCount int       `gorm:"not null"`
	StartTime  time.Time `gorm:"not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (c *Campaign) GetTableName() string {
	return "campaigns"
}
