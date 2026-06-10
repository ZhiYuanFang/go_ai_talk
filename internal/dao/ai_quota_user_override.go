package dao

import "hello/internal/dao/internal"

type aiQuotaUserOverrideDao struct {
	*internal.AiQuotaUserOverrideDao
}

var AiQuotaUserOverride = aiQuotaUserOverrideDao{internal.NewAiQuotaUserOverrideDao()}
