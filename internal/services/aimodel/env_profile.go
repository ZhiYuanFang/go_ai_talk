package aimodel

import (
	"os"
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
}

// LaneEnvKeysFor 返回 lane 对应的 env 键；不支持时 ok=false。
func LaneEnvKeysFor(lane Lane) (LaneEnvKeys, bool) {
	switch lane {
	case LaneVoiceUnderstanding:
		return LaneEnvKeys{
			ProviderKey: "VOICE_LLM_VOICE_UNDERSTANDING_PROVIDER",
			ModelKey:    "VOICE_LLM_VOICE_UNDERSTANDING_MODEL",
		}, true
	case LaneClinic:
		return LaneEnvKeys{
			ProviderKey: "VOICE_LLM_CLINIC_PROVIDER",
			ModelKey:    "VOICE_LLM_CLINIC_MODEL",
		}, true
	case LanePolish:
		return LaneEnvKeys{"UCG_AI_PROVIDER", "UCG_AI_VISION_MODEL"}, true
	case LaneSimText:
		return LaneEnvKeys{"SIM_LLM_SIMTEXT_PROVIDER", "SIM_LLM_SIMTEXT_MODEL"}, true
	case LaneSimVision:
		return LaneEnvKeys{"SIM_LLM_SIMVISION_PROVIDER", "SIM_LLM_SIMVISION_MODEL"}, true
	case LaneSimImageGen:
		return LaneEnvKeys{"SIM_LLM_SIMIMAGEGEN_PROVIDER", "SIM_LLM_SIMIMAGEGEN_MODEL"}, true
	case LaneSimVideoGen:
		return LaneEnvKeys{"SIM_LLM_SIMVIDEOGEN_PROVIDER", "SIM_LLM_SIMVIDEOGEN_MODEL"}, true
	default:
		return LaneEnvKeys{}, false
	}
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
	return p, true
}

// MergeColdStartProfile 冷启动合并：env > yaml > 代码种子。
func MergeColdStartProfile(lane Lane, yamlProfile Profile, yamlOK bool) Profile {
	if p, ok := ProfileFromEnv(lane); ok {
		return p
	}
	if yamlOK && strings.TrimSpace(yamlProfile.Model) != "" {
		yamlProfile.Lane = lane
		if yamlProfile.Provider == "" {
			yamlProfile.Provider = DefaultSeedProfile(lane).Provider
		}
		if yamlProfile.MaxInFlight <= 0 {
			yamlProfile.MaxInFlight = DefaultSeedProfile(lane).MaxInFlight
		}
		if yamlProfile.MaxWaiters < 0 {
			yamlProfile.MaxWaiters = DefaultSeedProfile(lane).MaxWaiters
		}
		if yamlProfile.TimeoutSec <= 0 {
			yamlProfile.TimeoutSec = DefaultSeedProfile(lane).TimeoutSec
		}
		return yamlProfile
	}
	return DefaultSeedProfile(lane)
}
