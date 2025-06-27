package handler

import (
	"context"
	connect "github.com/bufbuild/connect-go"
	v1 "github.com/ssipflow/coupon-issuance/gen/proto/coupon/v1"
	"github.com/ssipflow/coupon-issuance/gen/proto/coupon/v1/couponv1connect"
	"github.com/ssipflow/coupon-issuance/internal/service"
	"time"
)

type CouponHandler struct {
	service *service.CouponService
}

func NewCouponHandler(service *service.CouponService) couponv1connect.CouponServiceHandler {
	return &CouponHandler{service: service}
}

func (c *CouponHandler) CreateCampaign(ctx context.Context, req *connect.Request[v1.CreateCampaignRequest]) (*connect.Response[v1.CreateCampaignResponse], error) {
	startTime, err := time.Parse(time.RFC3339, req.Msg.GetStartTime())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	id, err := c.service.CreateCampaign(ctx, req.Msg.GetName(), req.Msg.GetTotalCount(), startTime)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.CreateCampaignResponse{
		CampaignId: id,
	}), nil
}

func (c *CouponHandler) GetCampaign(ctx context.Context, req *connect.Request[v1.GetCampaignRequest]) (*connect.Response[v1.GetCampaignResponse], error) {
	return nil, nil
}

func (c *CouponHandler) IssueCoupon(ctx context.Context, req *connect.Request[v1.IssueCouponRequest]) (*connect.Response[v1.IssueCouponResponse], error) {
	return nil, nil
}
