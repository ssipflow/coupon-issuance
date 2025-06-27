package handler

import (
	"context"
	connectgo "github.com/bufbuild/connect-go"
	"github.com/redis/go-redis/v9"
	v1 "github.com/ssipflow/coupon-issuance/gen/proto/coupon/v1"
	"github.com/ssipflow/coupon-issuance/gen/proto/coupon/v1/couponv1connect"
	"gorm.io/gorm"
)

type CouponHandler struct {
	redis *redis.Client
	db    *gorm.DB
}

func NewCouponHandler(redis *redis.Client, db *gorm.DB) couponv1connect.CouponServiceHandler {
	return &CouponHandler{redis, db}
}

func (c *CouponHandler) CreateCampaign(context.Context, *connectgo.Request[v1.CreateCampaignRequest]) (*connectgo.Response[v1.CreateCampaignResponse], error) {
	return nil, nil
}

func (c *CouponHandler) GetCampaign(context.Context, *connectgo.Request[v1.GetCampaignRequest]) (*connectgo.Response[v1.GetCampaignResponse], error) {
	return nil, nil
}

func (c *CouponHandler) IssueCoupon(context.Context, *connectgo.Request[v1.IssueCouponRequest]) (*connectgo.Response[v1.IssueCouponResponse], error) {
	return nil, nil
}
