package device

// xiaozhi_mcp_binding.go：小智 MCP 音箱绑定领域服务。
//
// 业务：
//   - C 端用户在已绑宝宝账号下登记多个小智 MCP token；
//   - 音箱以 speaker_mac 为全局唯一身份；同 MAC 再添加则更新 token；
//   - deviceNo 取当前 wx 绑定宝宝，禁止客户端指定他人设备；
//   - mcp_token 全局唯一；列表脱敏不回传全文 token；
//   - 列表附带 mcp 侧 WebSocket 连接态（内部 HTTP；失败降级为 false）；
//   - 写成功后由调用方（controller）通知 xiaozhi-mcp-service（失败不回滚）。

import (
	"context"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	cashclient "hello/internal/clients/cash"
	xiaozhimcpclient "hello/internal/clients/xiaozhimcp"
	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	xiaozhiMcpStatusActive  = 1
	xiaozhiMcpAliasMaxRunes = 64
	xiaozhiMcpTokenMaxRunes = 512
)

// XiaozhiMcpBindingView App 列表项（脱敏 token）。
type XiaozhiMcpBindingView struct {
	Id         int64  `json:"id"`
	Alias      string `json:"alias"`
	TokenMask  string `json:"tokenMask"`
	SpeakerMac string `json:"speakerMac"`
	DeviceNo   string `json:"deviceNo"`
	Status     int    `json:"status"`
	Connected  bool   `json:"connected"`
	CreatedAt  int64  `json:"createdAt"`
	UpdatedAt  int64  `json:"updatedAt"`
}

// XiaozhiMcpBindingFull 内部全量项（含完整 token，仅内网）。
type XiaozhiMcpBindingFull struct {
	Id         int64  `json:"id"`
	WxId       int64  `json:"wxId"`
	DeviceNo   string `json:"deviceNo"`
	McpToken   string `json:"mcpToken"`
	SpeakerMac string `json:"speakerMac"`
	Alias      string `json:"alias"`
	Status     int    `json:"status"`
}

// XiaozhiMcpAddResult 添加/同 MAC 更新的写结果，供 controller 决定 mcp 通知。
type XiaozhiMcpAddResult struct {
	Binding      *XiaozhiMcpBindingFull // 当前行
	PrevMcpToken string                 // 更新前旧 token；与 Binding.McpToken 不同时须 Remove 旧桥
	TokenChanged bool                   // token 是否变化（仅 alias 变更时可跳过 Upsert）
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

// NormalizeSpeakerMac 将音箱 MAC 规范为小写冒号分隔六段（如 3c:dc:75:fc:7f:c4）。
// 允许输入含冒号/横线/无分隔等形式；非法则业务错误。
//
// Args: raw 用户输入。
// Returns: 规范化 MAC；错误。
func NormalizeSpeakerMac(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "speakerMac 不能为空")
	}
	var hex []byte
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		switch {
		case c == ':' || c == '-' || c == '.' || c == ' ' || c == '_':
			continue
		case c >= '0' && c <= '9':
			hex = append(hex, c)
		case c >= 'a' && c <= 'f':
			hex = append(hex, c)
		case c >= 'A' && c <= 'F':
			hex = append(hex, c+('a'-'A'))
		default:
			return "", gerror.NewCode(gcode.CodeInvalidParameter, "speakerMac 无效")
		}
	}
	if len(hex) != 12 {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "speakerMac 无效")
	}
	parts := make([]string, 6)
	for i := 0; i < 6; i++ {
		parts[i] = string(hex[i*2 : i*2+2])
	}
	return strings.Join(parts, ":"), nil
}

// NormalizeXiaozhiMcpToken 将用户粘贴的接入点字符串规范为裸 MCP token。
// 业务：小智后台常复制整段 wss://api.xiaozhi.me/mcp/?token=xxx（等号后可能有空格/引号）；
// 入库与拨号只应保存 query 中的 token 值。已是裸 token 时原样 trim。
//
// Args: raw 用户输入。
// Returns: 规范化后的 token；无法解析时返回业务错误。
func NormalizeXiaozhiMcpToken(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "mcpToken 不能为空")
	}
	// 优先标准 URL 解析（无异常空格时最干净）。
	if u, err := url.Parse(raw); err == nil && u.Scheme != "" && u.RawQuery != "" {
		if t := strings.TrimSpace(u.Query().Get("token")); t != "" {
			t = strings.Trim(t, "\"'")
			if t != "" && !looksLikeXiaozhiMcpURL(t) {
				return t, nil
			}
		}
	}
	// 兼容 token= 后带空格、整段非严格 URL：手工截取。
	lower := strings.ToLower(raw)
	const marker = "token="
	if idx := strings.Index(lower, marker); idx >= 0 {
		rest := strings.TrimSpace(raw[idx+len(marker):])
		rest = strings.Trim(rest, "\"'")
		if amp := strings.Index(rest, "&"); amp >= 0 {
			rest = rest[:amp]
		}
		rest = strings.TrimSpace(strings.Trim(rest, "\"'"))
		if rest == "" {
			return "", gerror.NewCode(gcode.CodeInvalidParameter, "mcpToken 无效：未解析到 token 值")
		}
		if looksLikeXiaozhiMcpURL(rest) {
			return "", gerror.NewCode(gcode.CodeInvalidParameter, "mcpToken 无效：请粘贴含 token= 的接入点或裸 token")
		}
		return rest, nil
	}
	// 无 token=：若仍像 URL 则拒绝，避免把整段 wss:// 当 token 入库。
	if looksLikeXiaozhiMcpURL(raw) {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "mcpToken 无效：请粘贴含 token= 的接入点或裸 token")
	}
	return raw, nil
}

// looksLikeXiaozhiMcpURL 粗判是否仍为接入点 URL（而非裸 token）。
func looksLikeXiaozhiMcpURL(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return strings.Contains(s, "://") ||
		strings.HasPrefix(s, "wss:") ||
		strings.HasPrefix(s, "ws:") ||
		strings.HasPrefix(s, "http:") ||
		strings.HasPrefix(s, "https:")
}

// ListXiaozhiMcpBindingsForWx 列出当前用户的小智绑定（脱敏 + 连接态）。
// Args: wxID 当前登录 wx。
// Returns: 视图列表；错误。
// Side Effects: 读 xiaozhi_mcp_binding；调用 xiaozhi-mcp 内部连接态接口（失败降级红灯）。
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
	type rowPack struct {
		view  XiaozhiMcpBindingView
		token string
	}
	packs := make([]rowPack, 0, len(rows))
	tokens := make([]string, 0, len(rows))
	for _, r := range rows {
		var row entity.XiaozhiMcpBinding
		if err := r.Struct(&row); err != nil {
			return nil, gerror.WrapCode(gcode.CodeInternalError, err, "解析小智绑定失败")
		}
		packs = append(packs, rowPack{
			view: XiaozhiMcpBindingView{
				Id:         row.Id,
				Alias:      row.Alias,
				TokenMask:  MaskXiaozhiMcpToken(row.McpToken),
				SpeakerMac: row.SpeakerMac,
				DeviceNo:   row.DeviceNo,
				Status:     row.Status,
				Connected:  false,
				CreatedAt:  row.CreatedAt,
				UpdatedAt:  row.UpdatedAt,
			},
			token: row.McpToken,
		})
		if t := strings.TrimSpace(row.McpToken); t != "" {
			tokens = append(tokens, t)
		}
	}
	// 批量查连接态；失败则全部保持 false，列表仍成功。
	statusMap, statusErr := xiaozhimcpclient.ConnectionStatus(ctx, tokens)
	if statusErr != nil {
		glog.Warningf(ctx, "[xiaozhi-mcp-binding] connection-status 失败，列表降级红灯 err=%v", statusErr)
	}
	out := make([]XiaozhiMcpBindingView, 0, len(packs))
	for _, p := range packs {
		if statusErr == nil && statusMap != nil {
			p.view.Connected = statusMap[p.token]
		}
		out = append(out, p.view)
	}
	return out, nil
}

// AddXiaozhiMcpBinding 按 MAC upsert：新 MAC 插入；同宝宝同 MAC 更新 token/alias；他宝宝占用则拒绝。
// Args: wxID、mcpToken、alias、speakerMac。
// Returns: 写结果（含旧 token 供停桥）；错误。
// Side Effects: 写 xiaozhi_mcp_binding。
func AddXiaozhiMcpBinding(ctx context.Context, wxID int64, mcpToken, alias, speakerMac string) (*XiaozhiMcpAddResult, error) {
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少登录用户")
	}
	// Add 前确保有效权益：可触发首次试用 claim；失败 fail-closed。
	if unlockErr := cashclient.RemoteEnsureXiaozhiMcpAccessForAdd(ctx, wxID); unlockErr != nil {
		glog.Warningf(ctx, "[xiaozhi-mcp-binding] ensure access failed wxId=%d err=%v", wxID, unlockErr)
		// 连通/配置类失败：对用户只提示接入点错误，不暴露内网 URL。
		if isXiaozhiMcpEnsureInfraError(unlockErr) {
			return nil, gerror.NewCode(gcode.CodeOperationFailed, "小智MCP接入点错误")
		}
		// 业务拒绝（试用用尽等）：用 cash 文案，去掉多余包装。
		msg := strings.TrimSpace(unlockErr.Error())
		if msg == "" {
			msg = "请先开通或体验小智 MCP 后再添加绑定"
		}
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, msg)
	}
	alias = strings.TrimSpace(alias)
	mac, err := NormalizeSpeakerMac(speakerMac)
	if err != nil {
		return nil, err
	}
	// 粘贴完整接入点 URL 时只保留 token 字段，再参与去重与入库。
	normalized, err := NormalizeXiaozhiMcpToken(mcpToken)
	if err != nil {
		return nil, err
	}
	mcpToken = normalized
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
	c := dao.XiaozhiMcpBinding.Columns()
	// 按 MAC 查是否已有行。
	macOne, err := dao.XiaozhiMcpBinding.Ctx(ctx).Where(c.SpeakerMac, mac).One()
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "校验音箱 MAC 失败")
	}
	now := time.Now().Unix()
	if !macOne.IsEmpty() {
		var exist entity.XiaozhiMcpBinding
		if err := macOne.Struct(&exist); err != nil {
			return nil, gerror.WrapCode(gcode.CodeInternalError, err, "解析小智绑定失败")
		}
		// 他宝宝占用：拒绝。
		if strings.TrimSpace(exist.DeviceNo) != deviceNo {
			return nil, gerror.NewCode(gcode.CodeInvalidOperation, "该音箱已绑定其他宝宝")
		}
		if exist.WxId != wxID {
			return nil, gerror.NewCode(gcode.CodeInvalidOperation, "该音箱已绑定其他宝宝")
		}
		prevToken := exist.McpToken
		tokenChanged := prevToken != mcpToken
		if tokenChanged {
			// token 全局去重：排除自身 id。
			tokOne, err := dao.XiaozhiMcpBinding.Ctx(ctx).
				Where(c.McpToken, mcpToken).
				WhereNot(c.Id, exist.Id).
				One()
			if err != nil {
				return nil, gerror.WrapCode(gcode.CodeInternalError, err, "校验小智 token 失败")
			}
			if !tokOne.IsEmpty() {
				return nil, gerror.NewCode(gcode.CodeInvalidOperation, "该小智 token 已被绑定")
			}
		}
		_, err = dao.XiaozhiMcpBinding.Ctx(ctx).
			Where(c.Id, exist.Id).
			Data(g.Map{
				c.McpToken:  mcpToken,
				c.Alias:     alias,
				c.DeviceNo:  deviceNo,
				c.UpdatedAt: now,
			}).Update()
		if err != nil {
			if strings.Contains(err.Error(), "Duplicate") || strings.Contains(err.Error(), "uk_mcp_token") {
				return nil, gerror.NewCode(gcode.CodeInvalidOperation, "该小智 token 已被绑定")
			}
			return nil, gerror.WrapCode(gcode.CodeInternalError, err, "更新小智绑定失败")
		}
		return &XiaozhiMcpAddResult{
			Binding: &XiaozhiMcpBindingFull{
				Id:         exist.Id,
				WxId:       wxID,
				DeviceNo:   deviceNo,
				McpToken:   mcpToken,
				SpeakerMac: mac,
				Alias:      alias,
				Status:     exist.Status,
			},
			PrevMcpToken: prevToken,
			TokenChanged: tokenChanged,
		}, nil
	}
	// 新 MAC：token 全局去重后插入。
	existTok, err := dao.XiaozhiMcpBinding.Ctx(ctx).Where(c.McpToken, mcpToken).One()
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "校验小智 token 失败")
	}
	if !existTok.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeInvalidOperation, "该小智 token 已被绑定")
	}
	id, err := dao.XiaozhiMcpBinding.Ctx(ctx).Data(g.Map{
		c.WxId:       wxID,
		c.DeviceNo:   deviceNo,
		c.McpToken:   mcpToken,
		c.SpeakerMac: mac,
		c.Alias:      alias,
		c.Status:     xiaozhiMcpStatusActive,
		c.CreatedAt:  now,
		c.UpdatedAt:  now,
	}).InsertAndGetId()
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") || strings.Contains(err.Error(), "uk_mcp_token") {
			return nil, gerror.NewCode(gcode.CodeInvalidOperation, "该小智 token 已被绑定")
		}
		if strings.Contains(err.Error(), "uk_speaker_mac") {
			return nil, gerror.NewCode(gcode.CodeInvalidOperation, "该音箱已绑定其他宝宝")
		}
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "写入小智绑定失败")
	}
	return &XiaozhiMcpAddResult{
		Binding: &XiaozhiMcpBindingFull{
			Id:         id,
			WxId:       wxID,
			DeviceNo:   deviceNo,
			McpToken:   mcpToken,
			SpeakerMac: mac,
			Alias:      alias,
			Status:     xiaozhiMcpStatusActive,
		},
		TokenChanged: true, // 新插入视为需要 Upsert
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
		Id:         row.Id,
		WxId:       row.WxId,
		DeviceNo:   row.DeviceNo,
		McpToken:   row.McpToken,
		SpeakerMac: row.SpeakerMac,
		Alias:      row.Alias,
		Status:     row.Status,
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
			Id:         row.Id,
			WxId:       row.WxId,
			DeviceNo:   row.DeviceNo,
			McpToken:   row.McpToken,
			SpeakerMac: row.SpeakerMac,
			Alias:      row.Alias,
			Status:     row.Status,
		})
	}
	return out, nil
}

// isXiaozhiMcpEnsureInfraError 是否为 cash 连通/配置类失败（应对用户隐藏内网细节）。
func isXiaozhiMcpEnsureInfraError(err error) bool {
	if err == nil {
		return false
	}
	if gerror.Code(err) == gcode.CodeInternalError {
		return true
	}
	s := err.Error()
	switch {
	case strings.Contains(s, "不可达"),
		strings.Contains(s, "未配置"),
		strings.Contains(s, "非 JSON"),
		strings.Contains(s, "ensure HTTP"),
		strings.Contains(s, "connection refused"),
		strings.Contains(s, "dial tcp"):
		return true
	default:
		return false
	}
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
