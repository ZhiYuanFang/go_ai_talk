package ucg

import (
	"math"
	"strconv"

	"hello/internal/model/entity"
)

// ViewerCoords 请求侧有效坐标（WGS84 十进制度）。
type ViewerCoords struct {
	Lat float64
	Lng float64
}

// ValidViewerCoords 判断 lat/lng 是否为可用 viewer 坐标。
func ValidViewerCoords(lat, lng *float64) (ViewerCoords, bool) {
	if lat == nil || lng == nil {
		return ViewerCoords{}, false
	}
	if !validCoord(*lat, *lng) {
		return ViewerCoords{}, false
	}
	return ViewerCoords{Lat: *lat, Lng: *lng}, true
}

func validCoord(lat, lng float64) bool {
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return false
	}
	if lat == 0 && lng == 0 {
		return false
	}
	return true
}

func postEntityHasCoords(post entity.UcgPost) bool {
	if post.Lat == nil || post.Lng == nil {
		return false
	}
	return validCoord(*post.Lat, *post.Lng)
}

// haversineKm 计算两点球面距离（km）。
func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusKm = 6371.0
	rad := math.Pi / 180
	dLat := (lat2 - lat1) * rad
	dLng := (lng2 - lng1) * rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*rad)*math.Cos(lat2*rad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

// computeDistanceTerm 距离加性项；无 viewer 或帖坐标时返回 0。
func computeDistanceTerm(cfg FeedConfig, viewer ViewerCoords, postLat, postLng float64, distKm float64) float64 {
	if !validCoord(postLat, postLng) {
		return 0
	}
	if distKm <= 0 {
		distKm = haversineKm(viewer.Lat, viewer.Lng, postLat, postLng)
	}
	if cfg.DistDecayKm <= 0 {
		return 0
	}
	return cfg.WDist * math.Exp(-distKm/cfg.DistDecayKm)
}

// computeFinalScore baseScore + distanceTerm。
func computeFinalScore(cfg FeedConfig, baseScore float64, viewer ViewerCoords, hasViewer bool, postLat, postLng float64, distKm float64) float64 {
	if !hasViewer {
		return baseScore
	}
	return baseScore + computeDistanceTerm(cfg, viewer, postLat, postLng, distKm)
}

// formatDistanceMeters 返回 API 用米数字符串；无效时返回空。
func formatDistanceMeters(distKm float64) string {
	if distKm < 0 {
		return ""
	}
	meters := int64(math.Round(distKm * 1000))
	if meters < 0 {
		return ""
	}
	return strconv.FormatInt(meters, 10)
}
