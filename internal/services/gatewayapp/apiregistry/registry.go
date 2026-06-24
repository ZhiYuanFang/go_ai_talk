// Package apiregistry 从 api/v1 路由元数据构建 API 注册表，供使用统计路径归一化与中文 summary 展示。
package apiregistry

import (
	"io/fs"
	"regexp"
	"sort"
	"strings"
	"sync"

	v1 "hello/api/v1"
	v2 "hello/api/v2"

	"github.com/gogf/gf/v2/os/glog"
)

const unregisteredSummary = "未登记"

// Entry 单条 API 路由模板。
type Entry struct {
	Method   string
	Template string // 如 /ucg/app/api/posts/{id}
	Summary  string
}

var (
	loadOnce sync.Once
	routes   []Entry
	byExact  map[string]Entry // key: METHOD + " " + template
)

var metaTagRe = regexp.MustCompile("`([^`]+)`")

// Init 加载 api/v1 与 api/v2 下所有 g.Meta 定义；进程内幂等。
func Init() {
	loadOnce.Do(func() {
		byExact = make(map[string]Entry)
		seen := make(map[string]struct{})
		walkMetaFS(v1.APIMetaFS, seen)
		walkMetaFS(v2.APIMetaFS, seen)
		sort.Slice(routes, func(i, j int) bool {
			if routes[i].Method != routes[j].Method {
				return routes[i].Method < routes[j].Method
			}
			return routes[i].Template < routes[j].Template
		})
		if len(routes) == 0 {
			glog.Warning(nil, "[apiregistry] api embed 未解析到任何 g.Meta 路由，usage 统计 summary 将全部为「未登记」")
		}
	})
}

// RouteCount 返回已加载路由条数（供启动自检）。
func RouteCount() int {
	Init()
	return len(routes)
}

func walkMetaFS(fsys fs.FS, seen map[string]struct{}) {
	_ = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		content, readErr := fs.ReadFile(fsys, path)
		if readErr != nil {
			return nil
		}
		parseMetaContent(string(content), seen)
		return nil
	})
}

func parseMetaContent(content string, seen map[string]struct{}) {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "g.Meta") {
			continue
		}
		m := metaTagRe.FindStringSubmatch(line)
		if len(m) < 2 {
			continue
		}
		tags := parseMetaTags(m[1])
		p := strings.TrimSpace(tags["path"])
		method := strings.ToUpper(strings.TrimSpace(tags["method"]))
		summary := strings.TrimSpace(tags["summary"])
		if p == "" || method == "" {
			continue
		}
		key := method + " " + p
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		e := Entry{Method: method, Template: p, Summary: summary}
		routes = append(routes, e)
		byExact[key] = e
	}
}

func parseMetaTags(raw string) map[string]string {
	out := make(map[string]string)
	for _, part := range strings.Fields(raw) {
		i := strings.Index(part, ":")
		if i <= 0 {
			continue
		}
		k := part[:i]
		v := strings.Trim(part[i+1:], `"`)
		out[k] = v
	}
	return out
}

// Normalize 将原始 METHOD/path 归一化为模板 apiKey，并返回 summary。
func Normalize(method, rawPath string) (apiKey, summary string) {
	Init()
	method = strings.ToUpper(strings.TrimSpace(method))
	rawPath = normalizePath(rawPath)
	apiKey = method + " " + rawPath
	if e, ok := byExact[apiKey]; ok {
		return apiKey, pickSummary(e.Summary)
	}
	best := matchTemplate(method, rawPath)
	if best != nil {
		return method + " " + best.Template, pickSummary(best.Summary)
	}
	return apiKey, unregisteredSummary
}

// SummaryOf 已知 apiKey 的 summary；未知返回「未登记」。
// 支持模板 apiKey 与可匹配的 raw 路径（如 posts/123/like → posts/{id}/like）。
func SummaryOf(apiKey string) string {
	Init()
	apiKey = strings.TrimSpace(apiKey)
	if e, ok := byExact[apiKey]; ok {
		return pickSummary(e.Summary)
	}
	method, path, ok := splitAPIKey(apiKey)
	if !ok {
		return unregisteredSummary
	}
	if best := matchTemplate(method, path); best != nil {
		return pickSummary(best.Summary)
	}
	return unregisteredSummary
}

func splitAPIKey(apiKey string) (method, path string, ok bool) {
	i := strings.IndexByte(apiKey, ' ')
	if i <= 0 || i >= len(apiKey)-1 {
		return "", "", false
	}
	return strings.ToUpper(strings.TrimSpace(apiKey[:i])), normalizePath(apiKey[i+1:]), true
}

func pickSummary(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return unregisteredSummary
	}
	return s
}

func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if len(p) > 1 && strings.HasSuffix(p, "/") {
		p = strings.TrimRight(p, "/")
	}
	return p
}

func matchTemplate(method, rawPath string) *Entry {
	rawSegs := strings.Split(strings.Trim(normalizePath(rawPath), "/"), "/")
	var best *Entry
	bestScore := -1

	for i := range routes {
		e := &routes[i]
		if e.Method != method {
			continue
		}
		tplSegs := strings.Split(strings.Trim(e.Template, "/"), "/")
		if len(tplSegs) != len(rawSegs) {
			continue
		}
		score := 0
		ok := true
		for j := range tplSegs {
			ts, rs := tplSegs[j], rawSegs[j]
			if strings.HasPrefix(ts, "{") && strings.HasSuffix(ts, "}") {
				if rs == "" {
					ok = false
					break
				}
				score++
				continue
			}
			if ts != rs {
				ok = false
				break
			}
			score += 2
		}
		if !ok {
			continue
		}
		if score > bestScore {
			bestScore = score
			best = e
		}
	}
	return best
}
