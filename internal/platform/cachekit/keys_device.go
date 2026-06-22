package cachekit

import "strconv"

// WxIDToUnionKey wxId → unionId 映射缓存；TTL 120s，写路径失效。
func WxIDToUnionKey(wxID int64) string {
	return "dev:wx:id2union:" + strconv.FormatInt(wxID, 10)
}

// WxUnionToDeviceKey unionId → deviceNo 映射缓存；TTL 120s。
func WxUnionToDeviceKey(unionID string) string {
	return "dev:wx:union2dev:" + unionID
}

// WxIDToDeviceKey wxId → deviceNo 映射缓存；TTL 120s。
func WxIDToDeviceKey(wxID int64) string {
	return "dev:wx:id2dev:" + strconv.FormatInt(wxID, 10)
}
