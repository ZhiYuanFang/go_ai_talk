package dao

import "hello/internal/dao/internal"

type internalSurveyResponseDao = *internal.SurveyResponseDao

type surveyResponseDao struct {
	internalSurveyResponseDao
}

var (
	// SurveyResponse 满意度问卷提交。
	SurveyResponse = surveyResponseDao{internal.NewSurveyResponseDao()}
)
