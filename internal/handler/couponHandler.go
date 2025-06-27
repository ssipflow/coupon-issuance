package handler

import (
	"context"
	connectgo "github.com/bufbuild/connect-go"
	v1 "github.com/ssipflow/coupon-issuance/gen/proto/coupon/v1"
	"github.com/ssipflow/coupon-issuance/gen/proto/coupon/v1/couponv1connect"
	"github.com/ssipflow/coupon-issuance/internal/service"
)

type CouponHandler struct {
	service *service.CouponService
}

func NewCouponHandler(service *service.CouponService) couponv1connect.CouponServiceHandler {
	return &CouponHandler{service: service}
}

func (c *CouponHandler) CreateCampaign(ctx context.Context, req *connectgo.Request[v1.CreateCampaignRequest]) (*connectgo.Response[v1.CreateCampaignResponse], error) {
	return nil, nil
}

func (c *CouponHandler) GetCampaign(ctx context.Context, req *connectgo.Request[v1.GetCampaignRequest]) (*connectgo.Response[v1.GetCampaignResponse], error) {
	return nil, nil
}

func (c *CouponHandler) IssueCoupon(ctx context.Context, req *connectgo.Request[v1.IssueCouponRequest]) (*connectgo.Response[v1.IssueCouponResponse], error) {
	return nil, nil
}
