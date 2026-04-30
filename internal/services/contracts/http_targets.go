package contracts

import (
	"fmt"
	"os"
	"strings"
)

const (
	envHistoryServiceURL = "HISTORY_SERVICE_URL"
	envVoiceServiceURL   = "VOICE_SERVICE_URL"
	envDeviceServiceURL  = "DEVICE_SERVICE_URL"
)

type HTTPTargets struct {
	HistoryBaseURL string
	VoiceBaseURL   string
	DeviceBaseURL  string
}

// ResolveHTTPTargets 从环境变量解析内部服务目标地址。
// 未配置时回退本地默认端口，便于单机联调。
func ResolveHTTPTargets() HTTPTargets {
	return HTTPTargets{
		HistoryBaseURL: normalizeBaseURL(defaultString(os.Getenv(envHistoryServiceURL), "http://127.0.0.1:9801")),
		VoiceBaseURL:   normalizeBaseURL(defaultString(os.Getenv(envVoiceServiceURL), "http://127.0.0.1:9802")),
		DeviceBaseURL:  normalizeBaseURL(defaultString(os.Getenv(envDeviceServiceURL), "http://127.0.0.1:9803")),
	}
}

func (t HTTPTargets) VoiceTextChatPath() string {
	return "/voice/text/chat"
}

func (t HTTPTargets) DeviceAdminRegisterPath() string {
	return "/device/admin/register"
}

func (t HTTPTargets) HistoryListPath() string {
	return "/device/history/api/list"
}

func (t HTTPTargets) HistorySuggestPath() string {
	return "/device/history/api/suggest"
}

func (t HTTPTargets) HistorySuggestDeletePath() string {
	return "/device/history/api/suggest/delete"
}

func (t HTTPTargets) HistoryEventOptionsPath() string {
	return "/device/history/api/event/options"
}

func (t HTTPTargets) HistoryBirthdayPath() string {
	return "/device/history/api/birthday"
}

func (t HTTPTargets) HistoryBirthdaySavePath() string {
	return "/device/history/api/birthday/save"
}

func (t HTTPTargets) HistoryEventAddPath() string {
	return "/device/history/api/event/add"
}

func (t HTTPTargets) HistoryEventUpdatePath() string {
	return "/device/history/api/event/update"
}

func (t HTTPTargets) HistoryEventDeletePath() string {
	return "/device/history/api/event/delete"
}

func (t HTTPTargets) HistoryEventLatestPath() string {
	return "/device/history/api/event/latest"
}

func (t HTTPTargets) HistoryEventEndLatestPath() string {
	return "/device/history/api/event/end-latest"
}

func (t HTTPTargets) DeviceProfileGetPath() string {
	return "/device/profile/api/get"
}

func (t HTTPTargets) VoiceTextChatURL() string {
	// URL 统一通过 base + path 组合，避免调用方自行拼接导致路径不一致。
	return t.VoiceBaseURL + t.VoiceTextChatPath()
}

func (t HTTPTargets) DeviceAdminRegisterURL() string {
	return t.DeviceBaseURL + t.DeviceAdminRegisterPath()
}

func (t HTTPTargets) HistoryListURL() string {
	return t.HistoryBaseURL + t.HistoryListPath()
}

func (t HTTPTargets) DeviceProfileGetURL() string {
	return t.DeviceBaseURL + t.DeviceProfileGetPath()
}

func normalizeBaseURL(raw string) string {
	v := strings.TrimSpace(raw)
	// 去掉尾部斜杠，避免后续拼接 path 时出现双斜杠。
	v = strings.TrimRight(v, "/")
	if v == "" {
		return ""
	}
	return v
}

func defaultString(v, d string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return d
	}
	return v
}

func (t HTTPTargets) Validate() error {
	// 任何一个目标地址缺失都视为契约不完整，应在启动阶段暴露。
	if t.HistoryBaseURL == "" || t.VoiceBaseURL == "" || t.DeviceBaseURL == "" {
		return fmt.Errorf("invalid service base url contracts")
	}
	return nil
}

