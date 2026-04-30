package cachekit

import "errors"

var (
	ErrUnavailable   = errors.New("cache unavailable")
	ErrInvalidKey    = errors.New("invalid cache key")
	ErrInvalidTTL    = errors.New("invalid cache ttl")
	ErrEmptyValue    = errors.New("empty cache value")
	ErrInvalidIncrBy = errors.New("invalid incr delta")
)

