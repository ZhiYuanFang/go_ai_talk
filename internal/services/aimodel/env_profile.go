package aimodel

import (
	"os"
	"strconv"
	"strings"
)

const seedUpdatedBy = "seed"

// IsSeedUpdatedBy 判断配置行是否仍为冷启动种子（env 可覆盖 provider/model）。
func IsSeedUpdatedBy(updatedBy string) bool {
	return strings.TrimSpace(updatedBy) == seedUpdatedBy
}

// LaneEnvKeys 某 lane 冷启动 env 变量名。
type LaneEnvKeys struct {
	ProviderKey string
	ModelKey    string
	// 闸门参数（可选；未设时沿用 yaml/代码种子）
	MaxInFlightKey string
	MaxWaitersKey  string
	TimeoutSecKey  string
}

func laneEnvPrefix(lane Lane) (prefix string, ok bool) {
	switch lane {
	case LaneVoiceUnderstanding:
		return "VOICE_LLM_VOICE_UNDERSTANDING", true
	case LaneClinic:
		return "VOICE_LLM_CLINIC", true
	case LaneCareAlert:
		return "VOICE_LLM_CARE_ALERT", true
	case LanePolish:
		return "UCG_AI", true
	case LaneSimText:
		return "SIM_LLM_SIMTEXT", true
	case LaneSimVision:
		return "SIM_LLM_SIMVISION", true
	case LaneSimImageGen:
		return "SIM_LLM_SIMIMAGEGEN", true
	case LaneSimVideoGen:
		return "SIM_LLM_SIMVIDEOGEN", true
	default:
		return "", false
	}
}

// LaneEnvKeysFor 返回 lane 对应的 env 键；不支持时 ok=false。
func LaneEnvKeysFor(lane Lane) (LaneEnvKeys, bool) {
	prefix, ok := laneEnvPrefix(lane)
	if !ok {
		return LaneEnvKeys{}, false
	}
	keys := LaneEnvKeys{
		MaxInFlightKey: prefix + "_MAX_INFLIGHT",
		MaxWaitersKey:  prefix + "_MAX_WAITERS",
		TimeoutSecKey:  prefix + "_TIMEOUT_SEC",
	}
	switch lane {
	case LanePolish:
		keys.ProviderKey = "UCG_AI_PROVIDER"
		keys.ModelKey = "UCG_AI_VISION_MODEL"
	default:
		keys.ProviderKey = prefix + "_PROVIDER"
		keys.ModelKey = prefix + "_MODEL"
	}
	return keys, true
}

// ApplyEnvGateOverrides 用 env 覆盖 profile 闸门参数（env > yaml/DB 种子；Admin 非 seed 行由调用方跳过）。
func ApplyEnvGateOverrides(lane Lane, p *Profile) {
	if p == nil {
		return
	}
	keys, ok := LaneEnvKeysFor(lane)
	if !ok {
		return
	}
	if v, ok := envIntPositive(keys.MaxInFlightKey); ok {
		p.MaxInFlight = v
	}
	if v, ok := envIntNonNegative(keys.MaxWaitersKey); ok {
		p.MaxWaiters = v
	}
	if v, ok := envIntPositive(keys.TimeoutSecKey); ok {
		p.TimeoutSec = v
	}
}

func envIntPositive(key string) (int, bool) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return 0, false
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

func envIntNonNegative(key string) (int, bool) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return 0, false
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

// ProfileFromEnv 从 env 读取 lane provider/model；未配置或非法组合时 ok=false。
func ProfileFromEnv(lane Lane) (Profile, bool) {
	keys, ok := LaneEnvKeysFor(lane)
	if !ok {
		return Profile{}, false
	}
	provider := Provider(strings.TrimSpace(os.Getenv(keys.ProviderKey)))
	model := strings.TrimSpace(os.Getenv(keys.ModelKey))
	if provider == "" && model == "" {
		return Profile{}, false
	}
	base := DefaultSeedProfile(lane)
	if provider == "" {
		provider = base.Provider
	}
	if model == "" {
		model = base.Model
	}
	if !IsAllowedModel(provider, model) {
		return Profile{}, false
	}
	p := base
	p.Lane = lane
	p.Provider = provider
	p.Model = model
	ApplyEnvGateOverrides(lane, &p)
	return p, true
}

// MergeColdStartProfile 冷启动合并：env provider/model > yaml > 代码种子；闸门 env > yaml > 代码种子。
func MergeColdStartProfile(lane Lane, yamlProfile Profile, yamlOK bool) Profile {
	if p, ok := ProfileFromEnv(lane); ok {
		return p
	}
	var p Profile
	if yamlOK && strings.TrimSpace(yamlProfile.Model) != "" {
		p = yamlProfile
		p.Lane = lane
		if p.Provider == "" {
			p.Provider = DefaultSeedProfile(lane).Provider
		}
		if p.MaxInFlight <= 0 {
			p.MaxInFlight = DefaultSeedProfile(lane).MaxInFlight
		}
		if p.MaxWaiters < 0 {
			p.MaxWaiters = DefaultSeedProfile(lane).MaxWaiters
		}
		if p.TimeoutSec <= 0 {
			p.TimeoutSec = DefaultSeedProfile(lane).TimeoutSec
		}
	} else {
		p = DefaultSeedProfile(lane)
	}
	ApplyEnvGateOverrides(lane, &p)
	return p
}
