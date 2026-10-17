package dao

import "hello/internal/dao/internal"

type internalSurveyOptionDao = *internal.SurveyOptionDao

type surveyOptionDao struct {
	internalSurveyOptionDao
}

var (
	// SurveyOption 满意度问卷选项行。
	SurveyOption = surveyOptionDao{internal.NewSurveyOptionDao()}
)
