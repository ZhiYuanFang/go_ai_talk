package dao

import (
	"hello/internal/dao/internal"
)

type internalFeedbackDao = *internal.FeedbackDao

type feedbackDao struct {
	internalFeedbackDao
}

var (
	Feedback = feedbackDao{
		internal.NewFeedbackDao(),
	}
)
