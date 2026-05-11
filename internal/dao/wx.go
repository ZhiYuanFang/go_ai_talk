package dao

import (
	"hello/internal/dao/internal"
)

type internalWxDao = *internal.WxDao

type wxDao struct {
	internalWxDao
}

var Wx = wxDao{
	internal.NewWxDao(),
}
