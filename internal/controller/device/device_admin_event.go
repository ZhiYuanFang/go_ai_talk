package devicectrl

import (
	"context"
	"errors"
	"net/http"
	"strings"

	device "hello/internal/services/device"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// deviceAdminEventAdd 管理端新增事件（multipart：name/eventType/extraNames/color，可选 logo 文件）。
func AdminEventAdd(r *ghttp.Request, c *AdminCtrl) {
	if r.Method != http.MethodPost {
		r.Response.WriteStatusExit(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	if err := c.requireAdmin(ctx); err != nil {
		writeDeviceAdminError(r, err)
		return
	}
	maxBytes := device.EventImageMaxBytes(ctx) + (4 << 20)
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		writeDeviceAdminFail(r, gcode.CodeInvalidParameter, "multipart 解析失败")
		return
	}
	name, eventType, extraNames, color, unit, err := parseEventMultipartFields(r)
	if err != nil {
		writeDeviceAdminFail(r, gcode.CodeInvalidParameter, err.Error())
		return
	}
	parentID := r.GetForm("parentId").Int64()
	if parentID < 0 {
		parentID = 0
	}
	eventID, err := c.Admin.AddEvent(ctx, name, eventType, extraNames, color, unit, "", parentID)
	if err != nil {
		writeDeviceAdminEventErr(r, err)
		return
	}
	logoPath, err := saveEventLogoFromRequest(ctx, r)
	if err != nil {
		writeDeviceAdminFail(r, gcode.CodeInvalidParameter, err.Error())
		return
	}
	if logoPath != "" {
		if err := c.Admin.UpdateEvent(ctx, eventID, name, eventType, extraNames, color, unit, logoPath, nil); err != nil {
			writeDeviceAdminEventErr(r, err)
			return
		}
	}
	writeDeviceAdminOK(r, nil)
}

// deviceAdminEventUpdate 管理端更新事件（multipart，未传 logo 文件则保留原 logo）。
func AdminEventUpdate(r *ghttp.Request, c *AdminCtrl) {
	if r.Method != http.MethodPost {
		r.Response.WriteStatusExit(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	if err := c.requireAdmin(ctx); err != nil {
		writeDeviceAdminError(r, err)
		return
	}
	maxBytes := device.EventImageMaxBytes(ctx) + (4 << 20)
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		writeDeviceAdminFail(r, gcode.CodeInvalidParameter, "multipart 解析失败")
		return
	}
	id := r.GetForm("id").Int64()
	if id <= 0 {
		writeDeviceAdminFail(r, gcode.CodeInvalidParameter, "事件ID无效")
		return
	}
	name, eventType, extraNames, color, unit, err := parseEventMultipartFields(r)
	if err != nil {
		writeDeviceAdminFail(r, gcode.CodeInvalidParameter, err.Error())
		return
	}
	logoPath, err := saveEventLogoFromRequest(ctx, r)
	if err != nil {
		writeDeviceAdminFail(r, gcode.CodeInvalidParameter, err.Error())
		return
	}
	var parentPtr *int64
	if parentID, ok := eventParentIDFromMultipart(r); ok {
		parentPtr = &parentID
	}
	if err := c.Admin.UpdateEvent(ctx, id, name, eventType, extraNames, color, unit, logoPath, parentPtr); err != nil {
		writeDeviceAdminEventErr(r, err)
		return
	}
	writeDeviceAdminOK(r, nil)
}

// eventParentIDFromMultipart 解析表单 parentId；第二返回值表示客户端是否显式提交该字段。
func eventParentIDFromMultipart(r *ghttp.Request) (int64, bool) {
	if r.Request.MultipartForm == nil {
		return 0, false
	}
	if _, ok := r.Request.MultipartForm.Value["parentId"]; !ok {
		return 0, false
	}
	return device.NormalizeEventParentIDForAPI(r.GetForm("parentId").Int64()), true
}

func parseEventMultipartFields(r *ghttp.Request) (name string, eventType string, extraNames, color, unit string, err error) {
	name = strings.TrimSpace(r.GetForm("name").String())
	if name == "" {
		return "", "", "", "", "", gerror.New("事件名称不能为空")
	}
	eventType = device.NormalizeEventType(r.GetForm("eventType").String())
	if err := device.ValidateEventType(eventType); err != nil {
		return "", "", "", "", "", err
	}
	extraNames = r.GetForm("extraNames").String()
	color = strings.TrimSpace(r.GetForm("color").String())
	if err := device.ValidateEventColor(color); err != nil {
		return "", "", "", "", "", err
	}
	unit = strings.TrimSpace(r.GetForm("unit").String())
	return name, eventType, extraNames, color, unit, nil
}

func saveEventLogoFromRequest(ctx context.Context, r *ghttp.Request) (string, error) {
	file, hdr, err := r.Request.FormFile("logo")
	if err != nil {
		return "", nil
	}
	defer func() { _ = file.Close() }()
	return device.UploadEventLogo(ctx, hdr.Filename, file, hdr.Size)
}

func writeDeviceAdminOK(r *ghttp.Request, data interface{}) {
	r.Response.WriteJson(g.Map{"code": 0, "message": "OK", "data": data})
}

func writeDeviceAdminFail(r *ghttp.Request, code gcode.Code, msg string) {
	r.Response.WriteJson(g.Map{"code": code.Code(), "message": msg})
}

func writeDeviceAdminError(r *ghttp.Request, err error) {
	if e, ok := err.(*gerror.Error); ok {
		writeDeviceAdminFail(r, e.Code(), e.Error())
		return
	}
	writeDeviceAdminFail(r, gcode.CodeInternalError, err.Error())
}

func writeDeviceAdminEventErr(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, device.ErrEventExists):
		writeDeviceAdminFail(r, gcode.CodeInvalidOperation, err.Error())
	case errors.Is(err, device.ErrEventNotFound):
		writeDeviceAdminFail(r, gcode.CodeNotFound, err.Error())
	case errors.Is(err, device.ErrEventHasChildren):
		writeDeviceAdminFail(r, gcode.CodeInvalidOperation, err.Error())
	case errors.Is(err, device.ErrEventParentInvalid):
		writeDeviceAdminFail(r, gcode.CodeInvalidParameter, err.Error())
	case errors.Is(err, device.ErrEventParentCycle):
		writeDeviceAdminFail(r, gcode.CodeInvalidOperation, err.Error())
	default:
		if strings.TrimSpace(err.Error()) != "" {
			writeDeviceAdminFail(r, gcode.CodeInvalidParameter, err.Error())
		} else {
			writeDeviceAdminError(r, err)
		}
	}
}
