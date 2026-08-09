package aimodel

import "errors"

// ErrProfileStoreUnset 进程未注册 ProfileStore。
var ErrProfileStoreUnset = errors.New("aimodel ProfileStore 未注册")

// ErrProviderKeyMissing 对应 provider 的 API Key 环境变量未配置。
var ErrProviderKeyMissing = errors.New("LLM provider API Key 未配置")

// DefaultSeedProfile 返回方案 A 智谱种子配置（DB 无行时使用）。
func DefaultSeedProfile(lane Lane) Profile {
	switch lane {
	case LaneVoiceUnderstanding:
		return Profile{
			Lane: lane, Provider: ProviderZhipu, Model: "glm-4.7-flash",
			MaxInFlight: 1, MaxWaiters: 20, TimeoutSec: 20,
		}
	case LaneClinic:
		return Profile{
			Lane: lane, Provider: ProviderZhipu, Model: "glm-4.1v-thinking-flash",
			MaxInFlight: 1, MaxWaiters: 10, TimeoutSec: 120,
		}
	case LaneCareAlert:
		// 护理留意默认正式模与 clinic 同族闸门；VIP∪额度选模后由 Admin 覆盖。
		return Profile{
			Lane: lane, Provider: ProviderZhipu, Model: "glm-4.7-flash",
			MaxInFlight: 1, MaxWaiters: 10, TimeoutSec: 90,
		}
	case LanePolish:
		return Profile{
			Lane: lane, Provider: ProviderZhipu, Model: "glm-4.6v-flash",
			MaxInFlight: 1, MaxWaiters: 15, TimeoutSec: 60,
		}
	case LaneSimText:
		return Profile{
			Lane: lane, Provider: ProviderZhipu, Model: "glm-4.7-flash",
			MaxInFlight: 1, MaxWaiters: 20, TimeoutSec: 60,
		}
	case LaneSimVision:
		return Profile{
			Lane: lane, Provider: ProviderZhipu, Model: "glm-4.6v-flash",
			MaxInFlight: 1, MaxWaiters: 15, TimeoutSec: 60,
		}
	case LaneSimImageGen:
		return Profile{
			Lane: lane, Provider: ProviderZhipu, Model: "cogview-3-flash",
			MaxInFlight: 1, MaxWaiters: 10, TimeoutSec: 120,
		}
	case LaneSimVideoGen:
		return Profile{
			Lane: lane, Provider: ProviderZhipu, Model: "cogvideox-flash",
			MaxInFlight: 1, MaxWaiters: 5, TimeoutSec: 300,
		}
	default:
		return Profile{Lane: lane, MaxInFlight: 1}
	}
}
