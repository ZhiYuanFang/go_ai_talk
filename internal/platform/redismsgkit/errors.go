package redismsgkit

import "errors"

var ErrUnavailable = errors.New("redis messaging unavailable")
var ErrInvalidChannel = errors.New("invalid redis channel")
var ErrEmptyPayload = errors.New("empty redis message payload")
