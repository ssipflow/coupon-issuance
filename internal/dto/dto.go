package dto

import (
	"github.com/ssipflow/coupon-issuance/internal/entity"
	"time"
)

type Campaign struct {
	id            int32
	name          string
	couponLimit   int64
	currentCoupon int64
	startTime     time.Time
	createdAt     time.Time
	updatedAt     time.Time
	couponCodes   []string
}

func NewCampaign() *Campaign {
	return &Campaign{}
}

func NewCampaignFromEntity(entity *entity.Campaign) *Campaign {
	return &Campaign{
		id:            entity.ID,
		name:          entity.Name,
		couponLimit:   entity.CouponLimit,
		currentCoupon: entity.CurrentCoupon,
		startTime:     entity.StartTime,
		createdAt:     entity.CreatedAt,
		updatedAt:     entity.UpdatedAt,
		couponCodes:   nil,
	}
}

func (c *Campaign) GetID() int32 {
	return c.id
}

func (c *Campaign) GetName() string {
	return c.name
}

func (c *Campaign) GetCouponLimit() int64 {
	return c.couponLimit
}

func (c *Campaign) GetCurrentCoupon() int64 {
	return c.currentCoupon
}

func (c *Campaign) GetStartTime() time.Time {
	return c.startTime
}

func (c *Campaign) GetCreatedAt() time.Time {
	return c.createdAt
}

func (c *Campaign) GetUpdatedAt() time.Time {
	return c.updatedAt
}

func (c *Campaign) GetCouponCodes() []string {
	return c.couponCodes
}

func (c *Campaign) SetID(id int32) *Campaign {
	c.id = id
	return c
}

func (c *Campaign) SetName(name string) *Campaign {
	c.name = name
	return c
}

func (c *Campaign) SetCouponLimit(limit int64) *Campaign {
	c.couponLimit = limit
	return c
}

func (c *Campaign) SetCurrentCoupon(current int64) *Campaign {
	c.currentCoupon = current
	return c
}

func (c *Campaign) SetStartTime(start time.Time) *Campaign {
	c.startTime = start
	return c
}

func (c *Campaign) SetCreatedAt(created time.Time) *Campaign {
	c.createdAt = created
	return c
}

func (c *Campaign) SetUpdatedAt(updated time.Time) *Campaign {
	c.updatedAt = updated
	return c
}

func (c *Campaign) SetCouponCodes(codes []string) *Campaign {
	c.couponCodes = codes
	return c
}

func (c *Campaign) AddCouponCode(code string) *Campaign {
	c.couponCodes = append(c.couponCodes, code)
	return c
}
