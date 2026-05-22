package controller

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"hello/internal/dao"
	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/glog"
)

// 简易上传节流：同一管理会话两次上传最小间隔（防误触连点）。
const versionAdminUploadMinInterval = 3 * time.Second

var (
	versionAdminUploadMu    sync.Mutex
	versionAdminLastUpload  = map[string]time.Time{} // sessionId -> last upload
	versionAdminLoginBurst  sync.Mutex
	versionAdminLastLoginIP = map[string]time.Time{} // remote IP -> last attempt
)

const versionAdminLoginMinInterval = time.Second

func gatewayAppVersionAdminLogin(r *ghttp.Request) {
	if r.Method != http.MethodPost {
		r.Response.WriteStatusExit(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	remote := clientIP(r)
	versionAdminLoginBurst.Lock()
	if t, ok := versionAdminLastLoginIP[remote]; ok && time.Since(t) < versionAdminLoginMinInterval {
		versionAdminLoginBurst.Unlock()
		r.Response.Status = http.StatusTooManyRequests
		r.Response.WriteJson(g.Map{"code": 429, "message": "请求过快"})
		return
	}
	versionAdminLastLoginIP[remote] = time.Now()
	versionAdminLoginBurst.Unlock()

	pw := strings.TrimSpace(r.Get("password").String())
	if pw == "" {
		ct := strings.ToLower(r.Header.Get("Content-Type"))
		if strings.Contains(ct, "application/json") {
			var body struct {
				Password string `json:"password"`
			}
			if err := json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&body); err == nil {
				pw = strings.TrimSpace(body.Password)
			}
		}
	}
	expected := gatewayapp.VersionAdminPassword(ctx)
	if expected == "" {
		glog.Warningf(ctx, "[gateway-app-version-admin] 未配置管理员口令，拒绝登录")
		r.Response.Status = http.StatusServiceUnavailable
		r.Response.WriteJson(g.Map{"code": 503, "message": "版本管理未启用（未配置口令）"})
		return
	}
	if !constantTimeStringEqual(pw, expected) {
		glog.Warningf(ctx, "[gateway-app-version-admin] 管理员口令错误 remote=%s", remote)
		r.Response.Status = http.StatusUnauthorized
		r.Response.WriteJson(g.Map{"code": 401, "message": "口令错误"})
		return
	}
	sid, err := gatewayapp.NewVersionAdminSessionID()
	if err != nil {
		glog.Warningf(ctx, "[gateway-app-version-admin] 生成会话失败 err=%v", err)
		r.Response.Status = http.StatusInternalServerError
		r.Response.WriteJson(g.Map{"code": 500, "message": "内部错误"})
		return
	}
	ttl := gatewayapp.VersionAdminSessionTTL(ctx)
	sec := int(ttl.Seconds())
	if sec <= 0 {
		sec = 8 * 3600
	}
	key := gatewayapp.RedisSessionKey(sid)
	if _, err := g.Redis().Do(ctx, "SET", key, "1", "EX", sec); err != nil {
		glog.Warningf(ctx, "[gateway-app-version-admin] 写入会话 Redis 失败 err=%v", err)
		r.Response.Status = http.StatusInternalServerError
		r.Response.WriteJson(g.Map{"code": 500, "message": "会话创建失败"})
		return
	}
	setVersionAdminSessionCookie(r, sid, sec)
	r.Response.WriteJson(g.Map{"code": 0, "message": "ok"})
}

func gatewayAppVersionAdminUpload(r *ghttp.Request) {
	if r.Method != http.MethodPost {
		r.Response.WriteStatusExit(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	sid := versionAdminSessionIDFromRequest(r)
	if sid == "" || !versionAdminSessionValid(ctx, sid) {
		r.Response.Status = http.StatusUnauthorized
		r.Response.WriteJson(g.Map{"code": 401, "message": "请先登录管理页"})
		return
	}
	versionAdminUploadMu.Lock()
	if t, ok := versionAdminLastUpload[sid]; ok && time.Since(t) < versionAdminUploadMinInterval {
		versionAdminUploadMu.Unlock()
		r.Response.Status = http.StatusTooManyRequests
		r.Response.WriteJson(g.Map{"code": 429, "message": "上传过于频繁，请稍后再试"})
		return
	}
	versionAdminLastUpload[sid] = time.Now()
	versionAdminUploadMu.Unlock()

	maxBytes := gatewayapp.ApkMaxBytes(ctx)
	if err := r.ParseMultipartForm(maxBytes + (4 << 20)); err != nil { // 略大于单文件上限以容纳表单字段
		glog.Warningf(ctx, "[gateway-app-version-admin] ParseMultipartForm 失败 err=%v", err)
		r.Response.Status = http.StatusBadRequest
		r.Response.WriteJson(g.Map{"code": 400, "message": "multipart 解析失败"})
		return
	}
	latestVer := strings.TrimSpace(r.GetForm("latestVersion").String())
	if latestVer == "" {
		r.Response.Status = http.StatusBadRequest
		r.Response.WriteJson(g.Map{"code": 400, "message": "latestVersion 不能为空"})
		return
	}
	releaseNotes := strings.TrimSpace(r.GetForm("releaseNotes").String())
	force := strings.TrimSpace(r.GetForm("forceUpdate").String()) == "1" || strings.EqualFold(strings.TrimSpace(r.GetForm("forceUpdate").String()), "true")

	file, hdr, err := r.Request.FormFile("apk")
	if err != nil {
		glog.Warningf(ctx, "[gateway-app-version-admin] 缺少上传文件 err=%v", err)
		r.Response.Status = http.StatusBadRequest
		r.Response.WriteJson(g.Map{"code": 400, "message": "请选择 apk 文件（字段名 apk）"})
		return
	}
	defer func() { _ = file.Close() }()
	if hdr.Size > 0 && hdr.Size > maxBytes {
		r.Response.Status = http.StatusRequestEntityTooLarge
		r.Response.WriteJson(g.Map{"code": 413, "message": fmt.Sprintf("文件过大，上限 %d 字节", maxBytes)})
		return
	}
	origName := strings.TrimSpace(hdr.Filename)
	if !strings.HasSuffix(strings.ToLower(origName), ".apk") {
		glog.Warningf(ctx, "[gateway-app-version-admin] 非 apk 扩展名 name=%q", origName)
		r.Response.Status = http.StatusBadRequest
		r.Response.WriteJson(g.Map{"code": 400, "message": "仅支持 .apk 文件"})
		return
	}

	dir := filepath.Clean(gatewayapp.ApkStorageDir(ctx))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		glog.Warningf(ctx, "[gateway-app-version-admin] 创建目录失败 dir=%s err=%v", dir, err)
		r.Response.Status = http.StatusInternalServerError
		r.Response.WriteJson(g.Map{"code": 500, "message": "创建存储目录失败"})
		return
	}
	safeVer := sanitizeVersionForFilename(latestVer)
	serverName := fmt.Sprintf("%s_%d.apk", safeVer, time.Now().UnixNano())
	dest := filepath.Join(dir, serverName)
	if strings.Contains(serverName, "..") || strings.ContainsAny(serverName, `/\`) {
		r.Response.Status = http.StatusBadRequest
		r.Response.WriteJson(g.Map{"code": 400, "message": "非法文件名"})
		return
	}

	out, err := os.Create(dest)
	if err != nil {
		glog.Warningf(ctx, "[gateway-app-version-admin] 创建目标文件失败 path=%s err=%v", dest, err)
		r.Response.Status = http.StatusInternalServerError
		r.Response.WriteJson(g.Map{"code": 500, "message": "保存文件失败"})
		return
	}
	defer func() { _ = out.Close() }()
	lr := io.LimitReader(file, maxBytes+1)
	written, err := io.Copy(out, lr)
	if err != nil {
		_ = os.Remove(dest)
		glog.Warningf(ctx, "[gateway-app-version-admin] 写入 apk 失败 err=%v", err)
		r.Response.Status = http.StatusInternalServerError
		r.Response.WriteJson(g.Map{"code": 500, "message": "写入文件失败"})
		return
	}
	if written > maxBytes {
		_ = os.Remove(dest)
		r.Response.Status = http.StatusRequestEntityTooLarge
		r.Response.WriteJson(g.Map{"code": 413, "message": fmt.Sprintf("文件过大，上限 %d 字节", maxBytes)})
		return
	}

	dlPath := "/device/app/apk/" + serverName

	forceN := 0
	if force {
		forceN = 1
	}
	_, err = dao.AppVersion.Ctx(ctx).Data(g.Map{
		dao.AppVersion.Columns().LatestVersion: latestVer,
		dao.AppVersion.Columns().DownloadUrl:   dlPath,
		dao.AppVersion.Columns().ReleaseNotes:  releaseNotes,
		dao.AppVersion.Columns().ForceUpdate:   forceN,
		dao.AppVersion.Columns().ReleaseDate:   time.Now().Unix(),
	}).Insert()
	if err != nil {
		_ = os.Remove(dest)
		glog.Warningf(ctx, "[gateway-app-version-admin] 写库失败 err=%v", err)
		r.Response.Status = http.StatusInternalServerError
		r.Response.WriteJson(g.Map{"code": 500, "message": "数据库写入失败"})
		return
	}
	gatewayapp.InvalidateAppVersionLatestCache(ctx)

	r.Response.WriteJson(g.Map{
		"code":          0,
		"message":       "ok",
		"downloadUrl":   dlPath,
		"latestVersion": latestVer,
		"savedFile":     serverName,
	})
}

func gatewayAppApkDownload(r *ghttp.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		r.Response.WriteStatusExit(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	name := strings.TrimSpace(r.Get("filename").String())
	if name == "" {
		name = strings.TrimPrefix(r.URL.Path, "/device/app/apk/")
	}
	if name == "" || strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, "..") {
		r.Response.WriteStatusExit(http.StatusBadRequest)
		return
	}
	if !strings.HasSuffix(strings.ToLower(name), ".apk") {
		r.Response.WriteStatusExit(http.StatusBadRequest)
		return
	}
	if !apkFilenameSafe(name) {
		r.Response.WriteStatusExit(http.StatusBadRequest)
		return
	}
	dir := filepath.Clean(gatewayapp.ApkStorageDir(ctx))
	abs := filepath.Join(dir, name)
	dirClean := filepath.Clean(dir)
	absClean := filepath.Clean(abs)
	rel, err := filepath.Rel(dirClean, absClean)
	if err != nil || strings.HasPrefix(rel, "..") {
		r.Response.WriteStatusExit(http.StatusForbidden)
		return
	}
	if !gfile.Exists(abs) {
		r.Response.WriteStatusExit(http.StatusNotFound)
		return
	}
	r.Response.Header().Set("Content-Type", "application/vnd.android.package-archive")
	r.Response.Header().Set("Content-Disposition", `attachment; filename="`+strings.ReplaceAll(name, `"`, ``)+`"`)
	if r.Method == http.MethodHead {
		if fi, err := os.Stat(abs); err == nil {
			r.Response.Header().Set("Content-Length", fmt.Sprintf("%d", fi.Size()))
		}
		return
	}
	r.Response.ServeFile(abs)
}

func apkFilenameSafe(name string) bool {
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '.' || c == '_' || c == '-' {
			continue
		}
		return false
	}
	return true
}

func sanitizeVersionForFilename(v string) string {
	var b strings.Builder
	for _, c := range strings.TrimSpace(v) {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '.' || c == '_' || c == '-' {
			b.WriteRune(c)
			continue
		}
		b.WriteByte('_')
	}
	s := b.String()
	if s == "" {
		return "ver"
	}
	return s
}

func versionAdminSessionIDFromRequest(r *ghttp.Request) string {
	if c, err := r.Request.Cookie(gatewayapp.VersionAdminSessionCookieName); err == nil && c != nil {
		return strings.TrimSpace(c.Value)
	}
	return ""
}

func versionAdminSessionValid(ctx context.Context, sid string) bool {
	if sid == "" {
		return false
	}
	v, err := g.Redis().Do(ctx, "GET", gatewayapp.RedisSessionKey(sid))
	if err != nil || v == nil {
		return false
	}
	return strings.TrimSpace(v.String()) != ""
}

func setVersionAdminSessionCookie(r *ghttp.Request, sid string, maxAgeSec int) {
	http.SetCookie(r.Response.Writer, &http.Cookie{
		Name:     gatewayapp.VersionAdminSessionCookieName,
		Value:    sid,
		Path:     "/device/app",
		MaxAge:   maxAgeSec,
		HttpOnly: true,
		Secure:   gatewayapp.VersionAdminCookieSecure(r.Context()),
		SameSite: http.SameSiteLaxMode,
	})
}

func constantTimeStringEqual(a, b string) bool {
	if len(a) != len(b) {
		// 长度不等时仍做一次等长比较，略降低口令长度侧信道可区分度。
		dummy := strings.Repeat("\x00", 64)
		_ = subtle.ConstantTimeCompare([]byte(dummy), []byte(dummy))
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func clientIP(r *ghttp.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	return strings.TrimSpace(r.RemoteAddr)
}
