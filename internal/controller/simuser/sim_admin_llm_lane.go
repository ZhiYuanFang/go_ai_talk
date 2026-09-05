package simuserctrl

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/services/aimodel"
	simuser "hello/internal/services/simuser"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// SimAdminLLMLanesCtrl sim 域 LLM lane Admin API。
type SimAdminLLMLanesCtrl struct{}

func NewSimAdminLLMLanesCtrl() *SimAdminLLMLanesCtrl { return &SimAdminLLMLanesCtrl{} }

func (c *SimAdminLLMLanesCtrl) Get(ctx context.Context, req *v1.SimAdminLLMLanesGetReq) (res *v1.SimAdminLLMLanesGetRes, err error) {
	_ = c
	_ = req
	if err = verifySimAdmin(ghttp.RequestFromCtx(ctx)); err != nil {
		return nil, err
	}
	dto, err := simuser.GetSimLLMLanesForAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return mapSimLLMLanesRes(dto), nil
}

func (c *SimAdminLLMLanesCtrl) Put(ctx context.Context, req *v1.SimAdminLLMLanesPutReq) (res *v1.SimAdminLLMLanesPutRes, err error) {
	if err = verifySimAdmin(ghttp.RequestFromCtx(ctx)); err != nil {
		return nil, err
	}
	updatedBy := strings.TrimSpace(req.UpdatedBy)
	if updatedBy == "" {
		updatedBy = "admin"
	}
	if err = simuser.UpdateSimLLMLanesForAdmin(ctx,
		req.SimText.ToLaneDTO(), req.SimVision.ToLaneDTO(),
		req.SimImageGen.ToLaneDTO(), req.SimVideoGen.ToLaneDTO(),
		updatedBy); err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
	}
	dto, err := simuser.GetSimLLMLanesForAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return mapSimLLMLanesRes(dto), nil
}

func mapSimLLMLanesRes(dto simuser.SimLLMLanesAdminDTO) *v1.SimAdminLLMLanesGetRes {
	return &v1.SimAdminLLMLanesGetRes{
		SimText:     simLaneItemFromDTO(dto.SimText),
		SimVision:   simLaneItemFromDTO(dto.SimVision),
		SimImageGen: simLaneItemFromDTO(dto.SimImageGen),
		SimVideoGen: simLaneItemFromDTO(dto.SimVideoGen),
		Allowlist:   dto.Allowlist,
	}
}

func simLaneItemFromDTO(d aimodel.LaneProfileDTO) v1.SimAdminLLMLaneItem {
	return v1.SimAdminLLMLaneItem{
		Provider: d.Provider, Model: d.Model,
		MaxInFlight: d.MaxInFlight, MaxWaiters: d.MaxWaiters,
		UpdatedAt: d.UpdatedAt, UpdatedBy: d.UpdatedBy,
	}
}
