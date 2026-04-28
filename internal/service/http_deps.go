package service

// HTTPDeps 描述 HTTP 控制器装配所需的服务依赖。
// 通过集中装配降低 register 层的跨领域耦合。
type HTTPDeps struct {
	History DeviceHistoryContract
	Voice   VoiceContract
	Admin   DeviceAdminContract
}

// NewHTTPDeps 返回默认 HTTP 依赖集合。
func NewHTTPDeps() HTTPDeps {
	return HTTPDeps{
		History: DeviceHistory(),
		Voice:   Voice(),
		Admin:   DeviceAdmin(),
	}
}
