package controller

import (
	"context"
	"errors"

	v1 "hello/api/v1"
	"hello/internal/services/contracts"
	voice "hello/internal/services/voice"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// VoiceAIQuotaInternalCtrl voice 域 AI 额度 internal API。
type VoiceAIQuotaInternalCtrl struct{}

func NewVoiceAIQuotaInternalCtrl() *VoiceAIQuotaInternalCtrl {
	return &VoiceAIQuotaInternalCtrl{}
}

func parseVoiceQuotaFeature(s string) (contracts.AIQuotaFeature, error) {
	switch contracts.AIQuotaFeature(s) {
	case contracts.AIQuotaVoiceAI, contracts.AIQuotaClinicAI:
		return contracts.AIQuotaFeature(s), nil
	default:
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "feature 无效")
	}
}

func (c *VoiceAIQuotaInternalCtrl) Check(ctx context.Context, req *v1.VoiceInternalAIQuotaCheckReq) (res *v1.VoiceInternalAIQuotaCheckRes, err error) {
	_ = c
	feature, err := parseVoiceQuotaFeature(req.Feature)
	if err != nil {
		return nil, err
	}
	snap, err := voice.CheckVoiceAIQuotaStore(ctx, req.WxId, feature)
	if err != nil {
		return nil, mapAIQuotaErr(err)
	}
	return &v1.VoiceInternalAIQuotaCheckRes{
		Allowed:  snap.Allowed,
		Used:     snap.Used,
		Limit:    snap.Limit,
		Degraded: snap.Degraded,
	}, nil
}

func (c *VoiceAIQuotaInternalCtrl) Consume(ctx context.Context, req *v1.VoiceInternalAIQuotaConsumeReq) (res *v1.VoiceInternalAIQuotaConsumeRes, err error) {
	_ = c
	feature, err := parseVoiceQuotaFeature(req.Feature)
	if err != nil {
		return nil, err
	}
	snap, err := voice.ConsumeVoiceAIQuotaStore(ctx, req.WxId, feature)
	if err != nil {
		return nil, mapAIQuotaErr(err)
	}
	return &v1.VoiceInternalAIQuotaConsumeRes{Used: snap.Used, Limit: snap.Limit}, nil
}

func mapAIQuotaErr(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, contracts.ErrAINotLoggedIn):
		return gerror.NewCode(contracts.GCodeAINotLoggedIn(), err.Error())
	case errors.Is(err, contracts.ErrAIQuotaExhausted):
		return gerror.NewCode(contracts.GCodeAIQuotaExhausted(), err.Error())
	default:
		return err
	}
}
