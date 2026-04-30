package eventkit

import "errors"

var (
	ErrUnavailable      = errors.New("event bus unavailable")
	ErrEmptyRoutingKey  = errors.New("empty routing key")
	ErrInvalidRoutingKey = errors.New("invalid routing key")
	ErrEmptyExchange    = errors.New("empty exchange")
	ErrEmptyAPIBase     = errors.New("empty mq api base")
	ErrInvalidAuth      = errors.New("invalid mq auth")
	ErrPublishRejected  = errors.New("event publish rejected")
	ErrDependencyFailed = errors.New("event dependency check failed")
)

