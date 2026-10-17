package dao

import "hello/internal/dao/internal"

type internalSurveyCampaignDao = *internal.SurveyCampaignDao

type surveyCampaignDao struct {
	internalSurveyCampaignDao
}

var (
	// SurveyCampaign 满意度问卷活动。
	SurveyCampaign = surveyCampaignDao{internal.NewSurveyCampaignDao()}
)
