package handler

import (
	"context"
	connectgo "github.com/bufbuild/connect-go"
	v1 "github.com/ssipflow/coupon-issuance/gen/proto/coupon/v1"
	"github.com/ssipflow/coupon-issuance/gen/proto/coupon/v1/couponv1connect"
	"github.com/ssipflow/coupon-issuance/internal/repo"
)

type CouponHandler struct {
	redisClient     *repo.RedisClient
	mySqlRepository *repo.MySqlRepository
}

func NewCouponHandler(redis *repo.RedisClient, db *repo.MySqlRepository) couponv1connect.CouponServiceHandler {
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
