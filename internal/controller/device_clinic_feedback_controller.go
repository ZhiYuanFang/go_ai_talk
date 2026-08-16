package controller

import (
	"context"

	"hello/api/v1"
)

// DeviceClinicFeedbackController 诊疗反馈控制器（已退役）。
//
// 业务说明：Clinic 显式点赞已删除；飞轮仅 Gateway Clinic agent 内隐式采纳。
// 本控制器保留仅防旧客户端误调，返回业务拒绝，不调用 Python。
type DeviceClinicFeedbackController struct{}

// ClinicFeedback 拒绝显式反馈（无飞轮转发）。
func (c *DeviceClinicFeedbackController) ClinicFeedback(ctx context.Context, req *v1.DeviceClinicFeedbackReq) (res *v1.DeviceClinicFeedbackRes, err error) {
	_ = ctx
	_ = req
	return &v1.DeviceClinicFeedbackRes{
		Code:    410,
		Message: "clinic 显式反馈已下线，请使用 Gateway 内隐式采纳",
	}, nil
}
