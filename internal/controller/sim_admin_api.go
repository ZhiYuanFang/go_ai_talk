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
	cfg, err := simuser.GetFullConfig(ctx)
	if err != nil {
		return nil, err
	}
	iv := simuser.RuntimeConfigToAPIIntervals(cfg.Runtime)
	return &v1.SimAdminConfigGetRes{Config: v1.SimAdminConfigDTO{
		Enabled: cfg.Enabled, MaxSimUsers: cfg.MaxSimUsers,
		UpdatedAt: cfg.UpdatedAt, UpdatedBy: cfg.UpdatedBy,
		TaskSwitches: v1.SimAdminRuntimeTaskSwitchesDTO{
			Register: cfg.Runtime.TaskRegister, Comment: cfg.Runtime.TaskComment,
			PostImage: cfg.Runtime.TaskPostImage, PostVideo: cfg.Runtime.TaskPostVideo,
			PostDebate: cfg.Runtime.TaskPostDebate, DebateComment: cfg.Runtime.TaskDebateComment,
			Chat: cfg.Runtime.TaskChat, Follow: cfg.Runtime.TaskFollow,
		},
		Intervals: v1.SimAdminRuntimeIntervalsDTO{
			Register: iv["register"], Comment: iv["comment"],
			PostImage: iv["postImage"], PostVideo: iv["postVideo"],
			PostDebate: iv["postDebate"], DebateComment: iv["debateComment"],
			Chat: iv["chat"], Follow: iv["follow"],
			PostVideoPollInterval: iv["postVideoPollInterval"],
			PostVideoPollMaxWait:  iv["postVideoPollMaxWait"],
			StartupStaggerMax:     iv["startupStaggerMax"],
		},
		RateLimitRps: cfg.Runtime.RateLimitRps, RateLimitBurst: cfg.Runtime.RateLimitBurst,
	}}, nil
}

// ConfigPut PUT /sim/admin/api/config
func (c *SimAdminCtrl) ConfigPut(ctx context.Context, req *v1.SimAdminConfigPutReq) (res *v1.SimAdminConfigPutRes, err error) {
	_ = c
	if err = verifySimAdmin(ghttp.RequestFromCtx(ctx)); err != nil {
		return nil, err
	}
	current, err := simuser.GetFullConfig(ctx)
	if err != nil {
		return nil, err
	}
	var rt simuser.RuntimeConfigDB
	if isMinimalSimConfigPut(req) {
		rt = current.Runtime
	} else {
		rt, err = simuser.BuildRuntimeConfigFromAPI(simuser.RuntimeAPIInput{
		TaskRegister: req.TaskSwitches.Register, TaskComment: req.TaskSwitches.Comment,
		TaskPostImage: req.TaskSwitches.PostImage, TaskPostVideo: req.TaskSwitches.PostVideo,
		TaskPostDebate: req.TaskSwitches.PostDebate, TaskDebateComment: req.TaskSwitches.DebateComment,
		TaskChat: req.TaskSwitches.Chat, TaskFollow: req.TaskSwitches.Follow,
		IntervalRegister: req.Intervals.Register, IntervalComment: req.Intervals.Comment,
		IntervalPostImage: req.Intervals.PostImage, IntervalPostVideo: req.Intervals.PostVideo,
		IntervalPostDebate: req.Intervals.PostDebate, IntervalDebateComment: req.Intervals.DebateComment,
		IntervalChat: req.Intervals.Chat, IntervalFollow: req.Intervals.Follow,
		IntervalPostVideoPoll: req.Intervals.PostVideoPollInterval,
		IntervalPostVideoPollMaxWait: req.Intervals.PostVideoPollMaxWait,
		StartupStaggerMax: req.Intervals.StartupStaggerMax,
		RateLimitRps: req.RateLimitRps, RateLimitBurst: req.RateLimitBurst,
		}, current.Runtime)
		if err != nil {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
		}
	}
	putRes, err := simuser.UpdateConfigAdmin(ctx, simuser.ConfigAdminPutDTO{
		Enabled: req.Enabled, MaxSimUsers: req.MaxSimUsers, Runtime: rt,
	}, "admin")
	if err != nil {
		return nil, err
	}
	return mapConfigPutRes(putRes), nil
}

// isMinimalSimConfigPut 兼容仅提交 enabled/maxSimUsers 的旧客户端。
func isMinimalSimConfigPut(req *v1.SimAdminConfigPutReq) bool {
	iv := req.Intervals
	return iv.Register == "" && iv.Comment == "" && iv.PostImage == "" && iv.PostVideo == "" &&
		iv.PostDebate == "" && iv.DebateComment == "" &&
		iv.Chat == "" && iv.Follow == "" &&
		iv.PostVideoPollInterval == "" && iv.PostVideoPollMaxWait == "" &&
		iv.StartupStaggerMax == "" &&
		req.RateLimitRps == 0 && req.RateLimitBurst == 0
}

func mapConfigPutRes(in simuser.ConfigPutResult) *v1.SimAdminConfigPutRes {
	out := &v1.SimAdminConfigPutRes{
		ScheduleReloaded: in.ScheduleReloaded,
		Effects:          make([]v1.SimAdminConfigEffectDTO, 0, len(in.Effects)),
		TaskSchedule:     make([]v1.SimAdminTaskScheduleDTO, 0, len(in.TaskSchedule)),
	}
	for _, e := range in.Effects {
		out.Effects = append(out.Effects, v1.SimAdminConfigEffectDTO{Kind: e.Kind, Task: e.Task, Message: e.Message})
	}
	for _, t := range in.TaskSchedule {
		out.TaskSchedule = append(out.TaskSchedule, v1.SimAdminTaskScheduleDTO{
			Name: t.Name, Label: t.Label,
			ConfigEnabled: t.ConfigEnabled, Enabled: t.Enabled,
			IntervalSec: t.IntervalSec, LastRunAt: t.LastRunAt, NextRunHint: t.NextRunHint,
		})
	}
	return out
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
	laneProfiles := simuser.LoadAllLaneProfiles()
	out.TaskAiModels = mapTaskAiModels(simuser.BuildTaskAIModelCatalog(laneProfiles))
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
			PostDebate: rt.TaskSwitches.PostDebate,
			DebateComment: rt.TaskSwitches.DebateComment,
			Chat:      rt.TaskSwitches.Chat,
			Follow:    rt.TaskSwitches.Follow,
		},
		Intervals: v1.SimAdminRuntimeIntervalsDTO{
			Register:              rt.Intervals.Register,
			Comment:               rt.Intervals.Comment,
			PostImage:             rt.Intervals.PostImage,
			PostVideo:             rt.Intervals.PostVideo,
			PostDebate:            rt.Intervals.PostDebate,
			DebateComment:         rt.Intervals.DebateComment,
			Chat:                  rt.Intervals.Chat,
			Follow:                rt.Intervals.Follow,
			PostVideoPollInterval: rt.Intervals.PostVideoPollInterval,
			PostVideoPollMaxWait:  rt.Intervals.PostVideoPollMaxWait,
			StartupStaggerMax:     rt.Intervals.StartupStaggerMax,
		},
		RateLimitRps:   rt.RateLimitRps,
		RateLimitBurst: rt.RateLimitBurst,
		LLMLanes:       mapRuntimeLLMLanes(rt.LLMLanes),
		TaskAiModels:   mapTaskAiModels(rt.TaskAiModels),
	}}, nil
}

func mapTaskAiModels(in map[string][]simuser.TaskAIModelItem) map[string][]v1.SimAdminTaskAIModelDTO {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string][]v1.SimAdminTaskAIModelDTO, len(in))
	for taskName, items := range in {
		row := make([]v1.SimAdminTaskAIModelDTO, 0, len(items))
		for _, it := range items {
			row = append(row, v1.SimAdminTaskAIModelDTO{
				LaneKey: it.LaneKey, Usage: it.Usage,
				Provider: it.Provider, Model: it.Model,
			})
		}
		out[taskName] = row
	}
	return out
}

func mapRuntimeLLMLanes(in map[string]simuser.LLMLaneSnapshotDTO) map[string]v1.SimAdminLLMLaneDTO {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]v1.SimAdminLLMLaneDTO, len(in))
	for k, v := range in {
		out[k] = v1.SimAdminLLMLaneDTO{Provider: v.Provider, Model: v.Model}
	}
	return out
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
	if name == "post_video_submit" && simuser.IsVideoPostInFlight() {
		return nil, gerror.NewCode(gcode.CodeInvalidOperation, "任务正在执行中")
	}
	if !simuser.StartManualRunAsync(name, flags) {
		return nil, gerror.NewCode(gcode.CodeInvalidOperation, "任务正在执行中")
	}
	return &v1.SimAdminTaskRunPostRes{
		Accepted: true,
		TaskName: name,
		Message:  "已提交后台执行，请稍后刷新状态查看结果",
	}, nil
}

// UsersGet GET /sim/admin/api/users
func (c *SimAdminCtrl) UsersGet(ctx context.Context, req *v1.SimAdminUsersGetReq) (res *v1.SimAdminUsersGetRes, err error) {
	_ = c
	if err = verifySimAdmin(ghttp.RequestFromCtx(ctx)); err != nil {
		return nil, err
	}
	result, err := simuser.ListSimUsersForAdmin(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	out := &v1.SimAdminUsersGetRes{
		Total: result.Total, Page: result.Page, PageSize: result.PageSize,
		List: make([]v1.SimAdminUserListItem, 0, len(result.List)),
	}
	for _, row := range result.List {
		out.List = append(out.List, v1.SimAdminUserListItem{
			WxId: row.WxId, Account: row.Account, Nickname: row.Nickname,
			AvatarUrl: row.AvatarUrl, AvatarThumbnailUrl: row.AvatarThumbnailUrl,
			CreatedAt: row.CreatedAt, PasswordPlain: row.PasswordPlain,
			PasswordPlainLegacy: row.PasswordPlainLegacy,
		})
	}
	return out, nil
}

// UserDeactivatePost POST /sim/admin/api/users/{wxId}/deactivate
func (c *SimAdminCtrl) UserDeactivatePost(ctx context.Context, req *v1.SimAdminUserDeactivatePostReq) (res *v1.SimAdminUserDeactivatePostRes, err error) {
	_ = c
	if err = verifySimAdmin(ghttp.RequestFromCtx(ctx)); err != nil {
		return nil, err
	}
	if err = simuser.DeactivateSimUserForAdmin(ctx, req.WxId); err != nil {
		msg := err.Error()
		if strings.Contains(msg, "不存在") || strings.Contains(msg, "已注销") {
			return nil, gerror.NewCode(gcode.CodeNotFound, msg)
		}
		if strings.Contains(msg, "非模拟用户") {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, msg)
		}
		if strings.Contains(msg, "无效") {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, msg)
		}
		return nil, err
	}
	return &v1.SimAdminUserDeactivatePostRes{}, nil
}
