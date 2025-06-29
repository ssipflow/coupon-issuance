package handler

import (
	"context"
	"github.com/bufbuild/connect-go"
	v1 "github.com/ssipflow/coupon-issuance/gen/proto/coupon/v1"
	"github.com/ssipflow/coupon-issuance/gen/proto/coupon/v1/couponv1connect"
	"github.com/ssipflow/coupon-issuance/internal/errors"
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

	id, err := c.service.CreateCampaign(ctx, req.Msg.GetName(), req.Msg.GetCouponLimit(), startTime)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.CreateCampaignResponse{
		CampaignId: id,
	}), nil
}

func (c *CouponHandler) GetCampaign(ctx context.Context, req *connect.Request[v1.GetCampaignRequest]) (*connect.Response[v1.GetCampaignResponse], error) {
	campaignID := req.Msg.CampaignId
	campaign, err := c.service.GetCampaign(ctx, campaignID)
	if err != nil {
		if err.Error() == "RECORD_NOT_FOUND" {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.GetCampaignResponse{
		Id:            campaign.GetID(),
		Name:          campaign.GetName(),
		CouponLimit:   campaign.GetCouponLimit(),
		CurrentCoupon: campaign.GetCurrentCoupon(),
		StartTime:     campaign.GetStartTime().Format(time.RFC3339),
		CreatedAt:     campaign.GetCreatedAt().Format(time.RFC3339),
		UpdatedAt:     campaign.GetUpdatedAt().Format(time.RFC3339),
		IssuedCoupons: campaign.GetCouponCodes(),
	}), nil
}

func (c *CouponHandler) IssueCoupon(ctx context.Context, req *connect.Request[v1.IssueCouponRequest]) (*connect.Response[v1.IssueCouponResponse], error) {
	campaignID := req.Msg.GetCampaignId()
	userId := req.Msg.GetUserId()

	err := c.service.IssueCoupon(ctx, campaignID, userId)
	if err != nil {
		switch err.Error() {
		case errors.ERR_RECORD_NOT_FOUND:
			return nil, connect.NewError(connect.CodeNotFound, err)
		case errors.ERR_CAMPAIGN_NOT_STARTED:
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		case errors.ERR_COUPON_ALREADY_CLAIMED:
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		case errors.ERR_COUPON_SOLD_OUT:
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	return connect.NewResponse(&v1.IssueCouponResponse{Message: "OK"}), nil
}
