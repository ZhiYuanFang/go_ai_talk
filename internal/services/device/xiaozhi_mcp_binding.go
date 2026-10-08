package device

// xiaozhi_mcp_binding.go：小智 MCP 音箱绑定领域服务。
//
// 业务：
//   - C 端用户在已绑宝宝账号下登记多个小智 MCP token；
//   - deviceNo 取当前 wx 绑定宝宝，禁止客户端指定他人设备；
//   - mcp_token 全局唯一；列表脱敏不回传全文 token；
//   - 写成功后由调用方（controller）通知 xiaozhi-mcp-service（失败不回滚）。

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const (
	xiaozhiMcpStatusActive   = 1
	xiaozhiMcpAliasMaxRunes  = 64
	xiaozhiMcpTokenMaxRunes  = 512
)

// XiaozhiMcpBindingView App 列表项（脱敏 token）。
type XiaozhiMcpBindingView struct {
	Id        int64  `json:"id"`
	Alias     string `json:"alias"`
	TokenMask string `json:"tokenMask"`
	DeviceNo  string `json:"deviceNo"`
	Status    int    `json:"status"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

// XiaozhiMcpBindingFull 内部全量项（含完整 token，仅内网）。
type XiaozhiMcpBindingFull struct {
	Id       int64  `json:"id"`
	WxId     int64  `json:"wxId"`
	DeviceNo string `json:"deviceNo"`
	McpToken string `json:"mcpToken"`
	Alias    string `json:"alias"`
	Status   int    `json:"status"`
}

// MaskXiaozhiMcpToken 脱敏展示：保留前 6 与后 4（过短则全 ***）。
func MaskXiaozhiMcpToken(token string) string {
	token = strings.TrimSpace(token)
	n := len(token)
	if n <= 10 {
		return "***"
	}
	return token[:6] + "***" + token[n-4:]
}

// ListXiaozhiMcpBindingsForWx 列出当前用户的小智绑定（脱敏）。
// Args: wxID 当前登录 wx。
// Returns: 视图列表；错误。
// Side Effects: 读 xiaozhi_mcp_binding。
func ListXiaozhiMcpBindingsForWx(ctx context.Context, wxID int64) ([]XiaozhiMcpBindingView, error) {
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少登录用户")
	}
	c := dao.XiaozhiMcpBinding.Columns()
	rows, err := dao.XiaozhiMcpBinding.Ctx(ctx).
		Where(c.WxId, wxID).
		OrderDesc(c.Id).
		All()
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "读取小智绑定失败")
	}
	out := make([]XiaozhiMcpBindingView, 0, len(rows))
	for _, r := range rows {
		var row entity.XiaozhiMcpBinding
		if err := r.Struct(&row); err != nil {
			return nil, gerror.WrapCode(gcode.CodeInternalError, err, "解析小智绑定失败")
		}
		out = append(out, XiaozhiMcpBindingView{
			Id:        row.Id,
			Alias:     row.Alias,
			TokenMask: MaskXiaozhiMcpToken(row.McpToken),
			DeviceNo:  row.DeviceNo,
			Status:    row.Status,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return out, nil
}

// AddXiaozhiMcpBinding 添加绑定：deviceNo 取当前 wx 已绑宝宝；token 全局去重。
// Args: wxID、mcpToken、alias。
// Returns: 新行 id 与完整绑定（供写路径 Upsert）；错误。
// Side Effects: 写 xiaozhi_mcp_binding。
func AddXiaozhiMcpBinding(ctx context.Context, wxID int64, mcpToken, alias string) (*XiaozhiMcpBindingFull, error) {
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少登录用户")
	}
	mcpToken = strings.TrimSpace(mcpToken)
	alias = strings.TrimSpace(alias)
	if mcpToken == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "mcpToken 不能为空")
	}
	if utf8.RuneCountInString(mcpToken) > xiaozhiMcpTokenMaxRunes {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "mcpToken 过长")
	}
	if alias == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "备注不能为空")
	}
	if utf8.RuneCountInString(alias) > xiaozhiMcpAliasMaxRunes {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "备注过长")
	}
	deviceNo, err := WxIDBoundDeviceNoOrReject(ctx, wxID)
	if err != nil {
		return nil, err
	}
	// 全局去重：已存在相同 token 则拒绝（含其他 wx）。
	c := dao.XiaozhiMcpBinding.Columns()
	exist, err := dao.XiaozhiMcpBinding.Ctx(ctx).Where(c.McpToken, mcpToken).One()
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "校验小智 token 失败")
	}
	if !exist.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeInvalidOperation, "该小智 token 已被绑定")
	}
	now := time.Now().Unix()
	id, err := dao.XiaozhiMcpBinding.Ctx(ctx).Data(g.Map{
		c.WxId:      wxID,
		c.DeviceNo:  deviceNo,
		c.McpToken:  mcpToken,
		c.Alias:     alias,
		c.Status:    xiaozhiMcpStatusActive,
		c.CreatedAt: now,
		c.UpdatedAt: now,
	}).InsertAndGetId()
	if err != nil {
		// 并发下唯一索引冲突
		if strings.Contains(err.Error(), "Duplicate") || strings.Contains(err.Error(), "uk_mcp_token") {
			return nil, gerror.NewCode(gcode.CodeInvalidOperation, "该小智 token 已被绑定")
		}
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "写入小智绑定失败")
	}
	return &XiaozhiMcpBindingFull{
		Id:       id,
		WxId:     wxID,
		DeviceNo: deviceNo,
		McpToken: mcpToken,
		Alias:    alias,
		Status:   xiaozhiMcpStatusActive,
	}, nil
}

// UpdateXiaozhiMcpBindingAlias 更新属于当前用户的绑定备注。
func UpdateXiaozhiMcpBindingAlias(ctx context.Context, wxID, id int64, alias string) error {
	if wxID <= 0 || id <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "参数无效")
	}
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "备注不能为空")
	}
	if utf8.RuneCountInString(alias) > xiaozhiMcpAliasMaxRunes {
		return gerror.NewCode(gcode.CodeInvalidParameter, "备注过长")
	}
	row, err := getOwnedXiaozhiBinding(ctx, wxID, id)
	if err != nil {
		return err
	}
	c := dao.XiaozhiMcpBinding.Columns()
	_, err = dao.XiaozhiMcpBinding.Ctx(ctx).
		Where(c.Id, row.Id).
		Data(g.Map{
			c.Alias:     alias,
			c.UpdatedAt: time.Now().Unix(),
		}).Update()
	if err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err, "更新小智绑定备注失败")
	}
	return nil
}

// DeleteXiaozhiMcpBinding 删除属于当前用户的绑定；返回被删完整行（供 Remove 通知）。
func DeleteXiaozhiMcpBinding(ctx context.Context, wxID, id int64) (*XiaozhiMcpBindingFull, error) {
	if wxID <= 0 || id <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "参数无效")
	}
	row, err := getOwnedXiaozhiBinding(ctx, wxID, id)
	if err != nil {
		return nil, err
	}
	c := dao.XiaozhiMcpBinding.Columns()
	_, err = dao.XiaozhiMcpBinding.Ctx(ctx).Where(c.Id, row.Id).Delete()
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "删除小智绑定失败")
	}
	return &XiaozhiMcpBindingFull{
		Id:       row.Id,
		WxId:     row.WxId,
		DeviceNo: row.DeviceNo,
		McpToken: row.McpToken,
		Alias:    row.Alias,
		Status:   row.Status,
	}, nil
}

// ListActiveXiaozhiMcpBindingsFull 内部：全部 active 绑定（含完整 token）。
func ListActiveXiaozhiMcpBindingsFull(ctx context.Context) ([]XiaozhiMcpBindingFull, error) {
	c := dao.XiaozhiMcpBinding.Columns()
	rows, err := dao.XiaozhiMcpBinding.Ctx(ctx).
		Where(c.Status, xiaozhiMcpStatusActive).
		OrderAsc(c.Id).
		All()
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "读取小智绑定全量失败")
	}
	out := make([]XiaozhiMcpBindingFull, 0, len(rows))
	for _, r := range rows {
		var row entity.XiaozhiMcpBinding
		if err := r.Struct(&row); err != nil {
			return nil, gerror.WrapCode(gcode.CodeInternalError, err, "解析小智绑定失败")
		}
		out = append(out, XiaozhiMcpBindingFull{
			Id:       row.Id,
			WxId:     row.WxId,
			DeviceNo: row.DeviceNo,
			McpToken: row.McpToken,
			Alias:    row.Alias,
			Status:   row.Status,
		})
	}
	return out, nil
}

// WxIDBoundDeviceNoOrReject 取当前 wx 已绑 deviceNo；空则拒绝。
func WxIDBoundDeviceNoOrReject(ctx context.Context, wxID int64) (string, error) {
	deviceNo, err := WxDeviceNoByWxID(ctx, wxID)
	if err != nil {
		return "", err
	}
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return "", gerror.NewCode(gcode.CodeInvalidOperation, "请先绑定宝宝设备后再添加小智音箱")
	}
	return deviceNo, nil
}

// getOwnedXiaozhiBinding 按 id 取行并校验归属；无行或不属于当前 wx 则错误。
func getOwnedXiaozhiBinding(ctx context.Context, wxID, id int64) (*entity.XiaozhiMcpBinding, error) {
	c := dao.XiaozhiMcpBinding.Columns()
	one, err := dao.XiaozhiMcpBinding.Ctx(ctx).Where(c.Id, id).One()
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "读取小智绑定失败")
	}
	if one.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotFound, "小智绑定不存在")
	}
	var row entity.XiaozhiMcpBinding
	if err := one.Struct(&row); err != nil {
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "解析小智绑定失败")
	}
	if row.WxId != wxID {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "无权操作该小智绑定")
	}
	return &row, nil
}
