package controller

import (
	"context"

	"hello/api/v1"
	"hello/internal/services/voice"

	"github.com/gogf/gf/v2/frame/g"
)

// DeviceClinicFeedbackController 诊疗/小贴士反馈控制器。
//
// 业务说明：承接 Flutter clinic/tip 点赞飞轮，将 answerId+feedback 转发至 Python AI；
// 宿主为 voice-service（与 tip generate 同进程），经 gateway /device/api/clinic|tip/* 反代。
// 设计思路：校验 feedback∈{1,-1} 后调用 PythonAIClient，失败返回业务 code 而非抛错，避免网关 500 噪声。
type DeviceClinicFeedbackController struct{}

// ClinicFeedback 提交诊疗回答反馈。
//
// 业务逻辑：校验 feedback 合法后 POST Python /v1/clinic/feedback（JSON Body answer_id+feedback）。
// Args: req.AnswerId 来自 clinic answer_done；req.Feedback 为 1（赞）或 -1（踩）。
// Returns: 信封 {code,message}；非法参数 code=400；Python 失败 code=500。
// Side Effects: 调用外部 Python 反馈服务（可能触发限流 429，透传 code/message）。
func (c *DeviceClinicFeedbackController) ClinicFeedback(ctx context.Context, req *v1.DeviceClinicFeedbackReq) (res *v1.DeviceClinicFeedbackRes, err error) {
	if req.Feedback != 1 && req.Feedback != -1 {
		return &v1.DeviceClinicFeedbackRes{
			Code:    400,
			Message: "反馈值必须为 1（thumbs up）或 -1（thumbs down）",
		}, nil
	}
	pythonClient := voice.PythonAIClientFromCfg()
	pythonResp, pythonErr := pythonClient.ClinicFeedback(ctx, &voice.FeedbackRequest{
		AnswerID: req.AnswerId,
		Feedback: req.Feedback,
	})
	if pythonErr != nil {
		g.Log().Warning(ctx, "[Clinic Feedback] 调用 Python 反馈服务失败: ", pythonErr)
		return &v1.DeviceClinicFeedbackRes{
			Code:    500,
			Message: "反馈提交失败",
		}, nil
	}
	return &v1.DeviceClinicFeedbackRes{
		Code:    pythonResp.Code,
		Message: pythonResp.Message,
	}, nil
}

// TipFeedback 提交小贴士回答反馈。
//
// 业务逻辑：校验 feedback 合法后 POST Python /v1/tip/feedback（JSON Body answer_id+feedback）。
// Args: req.AnswerId 来自 tip SSE done；req.Feedback 为 1（赞）或 -1（踩）。
// Returns: 信封 {code,message}；非法参数 code=400；Python 失败 code=500。
// Side Effects: 调用外部 Python 反馈服务。
func (c *DeviceClinicFeedbackController) TipFeedback(ctx context.Context, req *v1.DeviceTipFeedbackReq) (res *v1.DeviceTipFeedbackRes, err error) {
	if req.Feedback != 1 && req.Feedback != -1 {
		return &v1.DeviceTipFeedbackRes{
			Code:    400,
			Message: "反馈值必须为 1（thumbs up）或 -1（thumbs down）",
		}, nil
	}
	pythonClient := voice.PythonAIClientFromCfg()
	pythonResp, pythonErr := pythonClient.TipFeedback(ctx, &voice.FeedbackRequest{
		AnswerID: req.AnswerId,
		Feedback: req.Feedback,
	})
	if pythonErr != nil {
		g.Log().Warning(ctx, "[Tip Feedback] 调用 Python 反馈服务失败: ", pythonErr)
		return &v1.DeviceTipFeedbackRes{
			Code:    500,
			Message: "反馈提交失败",
		}, nil
	}
	return &v1.DeviceTipFeedbackRes{
		Code:    pythonResp.Code,
		Message: pythonResp.Message,
	}, nil
}
