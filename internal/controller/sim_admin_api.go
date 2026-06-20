package controller

import (
	"context"
	"os"
	"strings"

	v1 "hello/api/v1"
	simuser "hello/internal/services/simuser"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// SimAdminCtrl 模拟用户管理 Admin API。
type SimAdminCtrl struct{}

func NewSimAdminCtrl() *SimAdminCtrl { return &SimAdminCtrl{} }

func simAdminPassword(ctx context.Context) string {
	if v := strings.TrimSpace(os.Getenv("SIM_ADMIN_PASSWORD")); v != "" {
		return v
	}
	return strings.TrimSpace(g.Cfg().MustGet(ctx, "simUser.adminPassword").String())
}

func verifySimAdmin(r *ghttp.Request) error {
	expected := simAdminPassword(r.Context())
	if expected == "" {
		return gerror.NewCode(gcode.CodeNotAuthorized, "SIM_ADMIN_PASSWORD 未配置")
	}
	if strings.TrimSpace(r.GetHeader("X-Admin-Password")) != expected {
		return gerror.NewCode(gcode.CodeNotAuthorized, "管理口令错误")
	}
	return nil
}

// ConfigGet GET /sim/admin/api/config
func (c *SimAdminCtrl) ConfigGet(ctx context.Context, req *v1.SimAdminConfigGetReq) (res *v1.SimAdminConfigGetRes, err error) {
	_ = c
	_ = req
	if err = verifySimAdmin(ghttp.RequestFromCtx(ctx)); err != nil {
		return nil, err
	}
	cfg, err := simuser.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.SimAdminConfigGetRes{Config: v1.SimAdminConfigDTO{
		Enabled: cfg.Enabled, MaxSimUsers: cfg.MaxSimUsers,
		UpdatedAt: cfg.UpdatedAt, UpdatedBy: cfg.UpdatedBy,
	}}, nil
}

// ConfigPut PUT /sim/admin/api/config
func (c *SimAdminCtrl) ConfigPut(ctx context.Context, req *v1.SimAdminConfigPutReq) (res *v1.SimAdminConfigPutRes, err error) {
	_ = c
	if err = verifySimAdmin(ghttp.RequestFromCtx(ctx)); err != nil {
		return nil, err
	}
	if err = simuser.UpdateConfig(ctx, req.Enabled, req.MaxSimUsers, "admin"); err != nil {
		return nil, err
	}
	return &v1.SimAdminConfigPutRes{}, nil
}

// PromptGet GET /sim/admin/api/prompts/{taskType}
func (c *SimAdminCtrl) PromptGet(ctx context.Context, req *v1.SimAdminPromptGetReq) (res *v1.SimAdminPromptGetRes, err error) {
	_ = c
	if err = verifySimAdmin(ghttp.RequestFromCtx(ctx)); err != nil {
		return nil, err
	}
	p, err := simuser.GetPrompt(ctx, req.TaskType)
	if err != nil {
		return nil, err
	}
	return &v1.SimAdminPromptGetRes{Prompt: v1.SimAdminPromptDTO{
		TaskType: p.TaskType, SystemPrompt: p.SystemPrompt,
		UserPromptTemplate: p.UserPromptTemplate, UpdatedAt: p.UpdatedAt, UpdatedBy: p.UpdatedBy,
	}}, nil
}

// PromptPut PUT /sim/admin/api/prompts/{taskType}
func (c *SimAdminCtrl) PromptPut(ctx context.Context, req *v1.SimAdminPromptPutReq) (res *v1.SimAdminPromptPutRes, err error) {
	_ = c
	if err = verifySimAdmin(ghttp.RequestFromCtx(ctx)); err != nil {
		return nil, err
	}
	if err = simuser.UpdatePrompt(ctx, req.TaskType, req.SystemPrompt, req.UserPromptTemplate, "admin"); err != nil {
		return nil, err
	}
	return &v1.SimAdminPromptPutRes{}, nil
}

// StatusGet GET /sim/admin/api/status
func (c *SimAdminCtrl) StatusGet(ctx context.Context, req *v1.SimAdminStatusGetReq) (res *v1.SimAdminStatusGetRes, err error) {
	_ = c
	_ = req
	if err = verifySimAdmin(ghttp.RequestFromCtx(ctx)); err != nil {
		return nil, err
	}
	st, err := simuser.GetStatus(ctx)
	if err != nil {
		return nil, err
	}
	out := v1.SimAdminStatusDTO{PendingVideoJobs: st.PendingVideoJobs}
	for _, t := range st.Tasks {
		out.Tasks = append(out.Tasks, v1.SimAdminTaskRunDTO{
			TaskName: t.TaskName, LastRunAt: t.LastRunAt,
			SuccessCount: t.SuccessCount, FailCount: t.FailCount, LastError: t.LastError,
		})
	}
	return &v1.SimAdminStatusGetRes{Status: out}, nil
}

// RuntimeGet GET /sim/admin/api/runtime
func (c *SimAdminCtrl) RuntimeGet(ctx context.Context, req *v1.SimAdminRuntimeGetReq) (res *v1.SimAdminRuntimeGetRes, err error) {
	_ = c
	_ = req
	if err = verifySimAdmin(ghttp.RequestFromCtx(ctx)); err != nil {
		return nil, err
	}
	rt, err := simuser.GetRuntimeSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.SimAdminRuntimeGetRes{Runtime: v1.SimAdminRuntimeDTO{
		ServiceEnabled:    rt.ServiceEnabled,
		DbEnabled:         rt.DbEnabled,
		DatabaseName:      rt.DatabaseName,
		SimUserCount:      rt.SimUserCount,
		SimUserCountError: rt.SimUserCountError,
		MaxSimUsers:       rt.MaxSimUsers,
		TaskSwitches: v1.SimAdminRuntimeTaskSwitchesDTO{
			Register:  rt.TaskSwitches.Register,
			Comment:   rt.TaskSwitches.Comment,
			PostImage: rt.TaskSwitches.PostImage,
			PostVideo: rt.TaskSwitches.PostVideo,
			Chat:      rt.TaskSwitches.Chat,
			Follow:    rt.TaskSwitches.Follow,
			VideoPoll: rt.TaskSwitches.VideoPoll,
		},
		Intervals: v1.SimAdminRuntimeIntervalsDTO{
			Register:            rt.Intervals.Register,
			Comment:             rt.Intervals.Comment,
			PostImage:           rt.Intervals.PostImage,
			PostVideo:           rt.Intervals.PostVideo,
			Chat:                rt.Intervals.Chat,
			Follow:              rt.Intervals.Follow,
			VideoPollIdle:       rt.Intervals.VideoPollIdle,
			VideoPollActive:     rt.Intervals.VideoPollActive,
			StartupStaggerMax:   rt.Intervals.StartupStaggerMax,
			EphemeralChatLoop:   rt.Intervals.EphemeralChatLoop,
			EphemeralChatWindow: rt.Intervals.EphemeralChatWindow,
		},
		RateLimitRps:   rt.RateLimitRps,
		RateLimitBurst: rt.RateLimitBurst,
	}}, nil
}

// TaskRunPost POST /sim/admin/api/tasks/{taskName}/run
func (c *SimAdminCtrl) TaskRunPost(ctx context.Context, req *v1.SimAdminTaskRunPostReq) (res *v1.SimAdminTaskRunPostRes, err error) {
	_ = c
	if err = verifySimAdmin(ghttp.RequestFromCtx(ctx)); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.TaskName)
	if simuser.NormalizeRunnableTaskPublic(name) == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "未知任务名")
	}
	flags := simuser.LoadRuntimeFlags(ctx)
	if !simuser.StartManualRunAsync(name, flags) {
		return nil, gerror.NewCode(gcode.CodeInvalidOperation, "任务正在执行中")
	}
	return &v1.SimAdminTaskRunPostRes{
		Accepted: true,
		TaskName: name,
		Message:  "已提交后台执行，请稍后刷新状态查看结果",
	}, nil
}
