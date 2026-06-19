package aimodel

import "errors"

// CodeLLMQueueFull LLM 按-model 等待队列已满时的业务码。
const CodeLLMQueueFull = 50301

// ErrQueueFull 等待队列已满，调用方应映射为 50301 且不得调用上游。
var ErrQueueFull = errors.New("当前队列已满，请稍后重试")

// QueueFullError 带业务码的队列满错误，便于 HTTP/WS 层映射。
type QueueFullError struct {
	Message string
}

func (e *QueueFullError) Error() string {
	if e == nil || e.Message == "" {
		return ErrQueueFull.Error()
	}
	return e.Message
}

// Code 返回 WS/HTTP 业务码。
func (e *QueueFullError) Code() int {
	return CodeLLMQueueFull
}

// NewQueueFullError 构造队列满错误。
func NewQueueFullError() error {
	return &QueueFullError{Message: ErrQueueFull.Error()}
}

// IsQueueFull 判断是否为队列满错误。
func IsQueueFull(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrQueueFull) {
		return true
	}
	var qe *QueueFullError
	return errors.As(err, &qe)
}
