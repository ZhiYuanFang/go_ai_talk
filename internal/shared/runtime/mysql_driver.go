package runtime

import (
	// 统一注册 GoFrame MySQL 驱动，供各服务运行时复用。
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	// 统一注册 GoFrame Redis 驱动，避免运行时出现 adapter 未注册。
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
)
