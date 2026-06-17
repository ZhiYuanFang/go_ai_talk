package voice

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"hello/internal/services/device"

	"github.com/gogf/gf/v2/os/glog"
)

// clinicBabyProfile 胖宝诊疗 LLM 注入的宝宝画像（A2 单行 JSON 字段源）。
type clinicBabyProfile struct {
	Birthday  string
	Gender    string
	AgeMonths int
}

// loadClinicBabyProfile 经 DeviceProfile HTTP 契约拉取画像；失败时降级为未设置/女/0 并继续问诊。
func loadClinicBabyProfile(ctx context.Context, deviceNo string) clinicBabyProfile {
	fallback := clinicBabyProfile{
		Birthday:  "未设置",
		Gender:    "女",
		AgeMonths: 0,
	}
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return fallback
	}
	profile, err := DeviceProfile().GetProfile(ctx, deviceNo)
	if err != nil {
		glog.Warningf(ctx, "clinic baby profile degraded: deviceNo=%s err=%v", deviceNo, err)
		return fallback
	}
	return clinicBabyProfileFromDevice(profile)
}

// clinicBabyProfileFromDevice 将 device 域画像格式化为 clinic LLM 字段（与语音球 loadDeviceProfile 口径一致）。
func clinicBabyProfileFromDevice(profile device.DeviceProfileInfo) clinicBabyProfile {
	out := clinicBabyProfile{
		Birthday:  "未设置",
		Gender:    "女",
		AgeMonths: 0,
	}
	if profile.Birthday > 0 {
		out.Birthday = time.Unix(profile.Birthday, 0).In(time.Local).Format("2006-01-02")
		out.AgeMonths = clinicAgeMonths(profile.Birthday)
	}
	if profile.Sex > 0 {
		out.Gender = "男"
	}
	return out
}

// clinicAgeMonths 计算整月月龄：年差×12+月差，当前日小于生日日则减 1。
func clinicAgeMonths(birthdayUnix int64) int {
	if birthdayUnix <= 0 {
		return 0
	}
	birth := time.Unix(birthdayUnix, 0).In(time.Local)
	now := time.Now().In(time.Local)
	months := (now.Year()-birth.Year())*12 + int(now.Month()-birth.Month())
	if now.Day() < birth.Day() {
		months--
	}
	if months < 0 {
		return 0
	}
	return months
}

// promptLine 返回注入 system prompt 的 A2 单行 JSON 前缀块。
func (p clinicBabyProfile) promptLine() string {
	raw, _ := json.Marshal(map[string]interface{}{
		"birthday":   p.Birthday,
		"gender":     p.Gender,
		"age_months": p.AgeMonths,
	})
	return "宝宝信息（JSON）：" + string(raw)
}
