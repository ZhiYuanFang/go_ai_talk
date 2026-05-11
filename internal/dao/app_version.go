package dao

import (
	"hello/internal/dao/internal"
)

type internalAppVersionDao = *internal.AppVersionDao

type appVersionDao struct {
	internalAppVersionDao
}

var AppVersion = appVersionDao{
	internal.NewAppVersionDao(),
}
