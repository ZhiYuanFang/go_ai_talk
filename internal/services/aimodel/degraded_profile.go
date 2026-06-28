package aimodel

// DegradedPolishProfile 润笔额度用尽时的降速 profile（zhipu/glm-4.6v-flash，不计入月度额度）。
func DegradedPolishProfile() Profile {
	return Profile{
		Lane:        LanePolish,
		Provider:    ProviderZhipu,
		Model:       "glm-4.6v-flash",
		MaxInFlight: 1,
		MaxWaiters:  15,
		TimeoutSec:  60,
	}
}

// DegradedClinicProfile 胖宝诊疗额度用尽时的降速 profile（zhipu/glm-4.1v-thinking-flash，不计入月度额度）。
func DegradedClinicProfile() Profile {
	return Profile{
		Lane:        LaneClinic,
		Provider:    ProviderZhipu,
		Model:       "glm-4.1v-thinking-flash",
		MaxInFlight: 1,
		MaxWaiters:  10,
		TimeoutSec:  120,
	}
}
