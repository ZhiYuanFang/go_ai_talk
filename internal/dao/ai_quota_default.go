package dao

import "hello/internal/dao/internal"

type aiQuotaDefaultDao struct {
	*internal.AiQuotaDefaultDao
}

var AiQuotaDefault = aiQuotaDefaultDao{internal.NewAiQuotaDefaultDao()}
