package ucg

import (
	"context"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

var (
	ipSearcherOnce sync.Once
	ipSearcher     *xdb.Searcher
	ipSearcherErr  error
)

func ip2regionSearcher() (*xdb.Searcher, error) {
	ipSearcherOnce.Do(func() {
		path := strings.TrimSpace(os.Getenv("IP2REGION_XDB_PATH"))
		if path == "" {
			path = strings.TrimSpace(g.Cfg().MustGet(context.Background(), "ucg.ip2regionXdbPath").String())
		}
		if path == "" {
			ipSearcherErr = nil
			return
		}
		ipSearcher, ipSearcherErr = xdb.NewWithFileOnly(xdb.IPv4, path)
	})
	return ipSearcher, ipSearcherErr
}

// ResolveIPLocation 将客户端 IP 解析为展示用属地（省份/城市）；无库或私网 IP 返回空字符串。
func ResolveIPLocation(ctx context.Context, clientIP string) string {
	clientIP = strings.TrimSpace(clientIP)
	if clientIP == "" {
		return ""
	}
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return ""
	}
	if ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return ""
	}
	searcher, err := ip2regionSearcher()
	if err != nil || searcher == nil {
		return ""
	}
	region, err := searcher.Search(clientIP)
	if err != nil || strings.TrimSpace(region) == "" {
		return ""
	}
	return formatIp2regionRegion(region)
}

// formatIp2regionRegion 将 ip2region v4 xdb 管道格式转为展示文案（省/市；不含 ISP）。
// 字段顺序：国家|省份|城市|ISP|国家代码；parts[3] 为运营商（如移动/电信），属地展示须忽略。
func formatIp2regionRegion(region string) string {
	parts := strings.Split(region, "|")
	if len(parts) < 3 {
		return ""
	}
	country := strings.TrimSpace(parts[0])
	province := strings.TrimSpace(parts[1])
	city := strings.TrimSpace(parts[2])
	if country != "" && country != "0" && country != "中国" {
		if province != "" && province != "0" {
			return province
		}
		return country
	}
	if province == "" || province == "0" {
		return ""
	}
	if city != "" && city != "0" && city != province {
		return province + " " + city
	}
	return province
}
