package dto

import "time"

type Campaign struct {
	ID          int32
	Name        string
	TotalCount  int64
	StartTime   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CouponCodes []string
}
