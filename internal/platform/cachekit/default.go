package cachekit

// Default 返回带 LoggingObserver 的全局 Redis 缓存单例，供业务包直接使用。
func Default() Cache {
	return WithObserver(NewRedisCache(), LoggingObserver{})
}
