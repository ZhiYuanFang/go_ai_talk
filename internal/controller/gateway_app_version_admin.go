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
	"hello/internal/model/entity"
	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/glog"
)

// 简易上传节流：同一管理会话两次上传最小间隔（防误触连点）。
const versionAdminUploadMinInterval = 3 * time.Second

const (
	versionAdminListLimitDefault = 50
	versionAdminListLimitMax     = 200
	versionAdminApkURLPrefix     = "/device/app/apk/"
)

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
	remote := gatewayapp.ClientIP(r)
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

// requireVersionAdminSession 校验版本管理已启用且持有有效管理会话；失败时已写 JSON 响应。
func requireVersionAdminSession(r *ghttp.Request) bool {
	ctx := r.Context()
	if gatewayapp.VersionAdminPassword(ctx) == "" {
		glog.Warningf(ctx, "[gateway-app-version-admin] 未配置管理员口令，拒绝管理操作")
		r.Response.Status = http.StatusServiceUnavailable
		r.Response.WriteJson(g.Map{"code": 503, "message": "版本管理未启用（未配置口令）"})
		return false
	}
	sid := versionAdminSessionIDFromRequest(r)
	if sid == "" || !versionAdminSessionValid(ctx, sid) {
		r.Response.Status = http.StatusUnauthorized
		r.Response.WriteJson(g.Map{"code": 401, "message": "请先登录管理页"})
		return false
	}
	return true
}

func gatewayAppVersionAdminUpload(r *ghttp.Request) {
	if r.Method != http.MethodPost {
		r.Response.WriteStatusExit(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	if !requireVersionAdminSession(r) {
		return
	}
	sid := versionAdminSessionIDFromRequest(r)
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

func gatewayAppVersionAdminList(r *ghttp.Request) {
	if r.Method != http.MethodGet {
		r.Response.WriteStatusExit(http.StatusMethodNotAllowed)
		return
	}
	if !requireVersionAdminSession(r) {
		return
	}
	ctx := r.Context()
	limit := r.Get("limit").Int()
	if limit <= 0 {
		limit = versionAdminListLimitDefault
	}
	if limit > versionAdminListLimitMax {
		limit = versionAdminListLimitMax
	}
	offset := r.Get("offset").Int()
	if offset < 0 {
		offset = 0
	}
	total, err := dao.AppVersion.Ctx(ctx).Count()
	if err != nil {
		glog.Warningf(ctx, "[gateway-app-version-admin] 统计版本行失败 err=%v", err)
		r.Response.Status = http.StatusInternalServerError
		r.Response.WriteJson(g.Map{"code": 500, "message": "数据库查询失败"})
		return
	}
	maxID := versionAdminMaxRowID(ctx)
	rows, err := dao.AppVersion.Ctx(ctx).OrderDesc(dao.AppVersion.Columns().Id).Limit(limit).Offset(offset).All()
	if err != nil {
		glog.Warningf(ctx, "[gateway-app-version-admin] 列表查询失败 err=%v", err)
		r.Response.Status = http.StatusInternalServerError
		r.Response.WriteJson(g.Map{"code": 500, "message": "数据库查询失败"})
		return
	}
	items := make([]g.Map, 0, len(rows))
	for _, rec := range rows {
		row, ok := appVersionFromRecord(rec)
		if !ok {
			glog.Warningf(ctx, "[gateway-app-version-admin] 解析版本行失败 id=%v", rec[dao.AppVersion.Columns().Id])
			continue
		}
		items = append(items, appVersionItemJSON(row, maxID > 0 && row.Id == maxID))
	}
	if int(total) > len(items) && len(rows) > 0 {
		glog.Warningf(ctx, "[gateway-app-version-admin] 列表行解析不完整 total=%d parsed=%d rows=%d", total, len(items), len(rows))
	}
	r.Response.WriteJson(g.Map{
		"code":    0,
		"message": "ok",
		"items":   items,
		"total":   total,
		"maxId":   maxID,
	})
}

func gatewayAppVersionAdminGet(r *ghttp.Request) {
	if r.Method != http.MethodGet {
		r.Response.WriteStatusExit(http.StatusMethodNotAllowed)
		return
	}
	if !requireVersionAdminSession(r) {
		return
	}
	ctx := r.Context()
	id := r.Get("id").Int64()
	if id <= 0 {
		r.Response.Status = http.StatusBadRequest
		r.Response.WriteJson(g.Map{"code": 400, "message": "id 无效"})
		return
	}
	row, ok := loadAppVersionByID(ctx, id)
	if !ok {
		r.Response.Status = http.StatusNotFound
		r.Response.WriteJson(g.Map{"code": 404, "message": "版本记录不存在"})
		return
	}
	maxID := versionAdminMaxRowID(ctx)
	r.Response.WriteJson(g.Map{
		"code":    0,
		"message": "ok",
		"item":    appVersionItemJSON(row, maxID > 0 && row.Id == maxID),
	})
}

func gatewayAppVersionAdminUpdate(r *ghttp.Request) {
	if r.Method != http.MethodPost {
		r.Response.WriteStatusExit(http.StatusMethodNotAllowed)
		return
	}
	if !requireVersionAdminSession(r) {
		return
	}
	ctx := r.Context()
	var body struct {
		ID            int64   `json:"id"`
		LatestVersion *string `json:"latestVersion"`
		ReleaseNotes  *string `json:"releaseNotes"`
		ForceUpdate   *bool   `json:"forceUpdate"`
		MinVersion    *string `json:"minVersion"`
		ReleaseDate   *int64  `json:"releaseDate"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 64<<10)).Decode(&body); err != nil {
		r.Response.Status = http.StatusBadRequest
		r.Response.WriteJson(g.Map{"code": 400, "message": "JSON 解析失败"})
		return
	}
	if body.ID <= 0 {
		r.Response.Status = http.StatusBadRequest
		r.Response.WriteJson(g.Map{"code": 400, "message": "id 无效"})
		return
	}
	if _, ok := loadAppVersionByID(ctx, body.ID); !ok {
		r.Response.Status = http.StatusNotFound
		r.Response.WriteJson(g.Map{"code": 404, "message": "版本记录不存在"})
		return
	}
	data := g.Map{}
	if body.LatestVersion != nil {
		v := strings.TrimSpace(*body.LatestVersion)
		if v == "" {
			r.Response.Status = http.StatusBadRequest
			r.Response.WriteJson(g.Map{"code": 400, "message": "latestVersion 不能为空"})
			return
		}
		data[dao.AppVersion.Columns().LatestVersion] = v
	}
	if body.ReleaseNotes != nil {
		data[dao.AppVersion.Columns().ReleaseNotes] = strings.TrimSpace(*body.ReleaseNotes)
	}
	if body.ForceUpdate != nil {
		forceN := 0
		if *body.ForceUpdate {
			forceN = 1
		}
		data[dao.AppVersion.Columns().ForceUpdate] = forceN
	}
	if body.MinVersion != nil {
		data[dao.AppVersion.Columns().MinVersion] = strings.TrimSpace(*body.MinVersion)
	}
	if body.ReleaseDate != nil {
		if *body.ReleaseDate < 0 {
			r.Response.Status = http.StatusBadRequest
			r.Response.WriteJson(g.Map{"code": 400, "message": "releaseDate 无效"})
			return
		}
		data[dao.AppVersion.Columns().ReleaseDate] = *body.ReleaseDate
	}
	if len(data) == 0 {
		r.Response.Status = http.StatusBadRequest
		r.Response.WriteJson(g.Map{"code": 400, "message": "未提供可更新字段"})
		return
	}
	if _, err := dao.AppVersion.Ctx(ctx).Where(dao.AppVersion.Columns().Id, body.ID).Data(data).Update(); err != nil {
		glog.Warningf(ctx, "[gateway-app-version-admin] 更新版本行失败 id=%d err=%v", body.ID, err)
		r.Response.Status = http.StatusInternalServerError
		r.Response.WriteJson(g.Map{"code": 500, "message": "数据库更新失败"})
		return
	}
	gatewayapp.InvalidateAppVersionLatestCache(ctx)
	row, _ := loadAppVersionByID(ctx, body.ID)
	maxID := versionAdminMaxRowID(ctx)
	r.Response.WriteJson(g.Map{
		"code":    0,
		"message": "ok",
		"item":    appVersionItemJSON(row, maxID > 0 && row.Id == maxID),
	})
}

func gatewayAppVersionAdminDelete(r *ghttp.Request) {
	if r.Method != http.MethodPost {
		r.Response.WriteStatusExit(http.StatusMethodNotAllowed)
		return
	}
	if !requireVersionAdminSession(r) {
		return
	}
	ctx := r.Context()
	id := r.Get("id").Int64()
	if id <= 0 {
		var body struct {
			ID int64 `json:"id"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&body); err == nil {
			id = body.ID
		}
	}
	if id <= 0 {
		r.Response.Status = http.StatusBadRequest
		r.Response.WriteJson(g.Map{"code": 400, "message": "id 无效"})
		return
	}
	row, ok := loadAppVersionByID(ctx, id)
	if !ok {
		r.Response.Status = http.StatusNotFound
		r.Response.WriteJson(g.Map{"code": 404, "message": "版本记录不存在"})
		return
	}
	if _, err := dao.AppVersion.Ctx(ctx).Where(dao.AppVersion.Columns().Id, id).Delete(); err != nil {
		glog.Warningf(ctx, "[gateway-app-version-admin] 删除版本行失败 id=%d err=%v", id, err)
		r.Response.Status = http.StatusInternalServerError
		r.Response.WriteJson(g.Map{"code": 500, "message": "数据库删除失败"})
		return
	}
	tryRemoveApkForDownloadPath(ctx, row.DownloadUrl)
	gatewayapp.InvalidateAppVersionLatestCache(ctx)
	r.Response.WriteJson(g.Map{"code": 0, "message": "ok"})
}

func versionAdminMaxRowID(ctx context.Context) int64 {
	one, err := dao.AppVersion.Ctx(ctx).OrderDesc(dao.AppVersion.Columns().Id).Limit(1).Fields(dao.AppVersion.Columns().Id).One()
	if err != nil || one.IsEmpty() {
		return 0
	}
	return one[dao.AppVersion.Columns().Id].Int64()
}

func loadAppVersionByID(ctx context.Context, id int64) (entity.AppVersion, bool) {
	one, err := dao.AppVersion.Ctx(ctx).Where(dao.AppVersion.Columns().Id, id).One()
	if err != nil || one.IsEmpty() {
		return entity.AppVersion{}, false
	}
	row, ok := appVersionFromRecord(one)
	return row, ok
}

// appVersionFromRecord 从查询结果构造版本行，避免 Struct 映射失败导致管理列表为空。
func appVersionFromRecord(rec gdb.Record) (entity.AppVersion, bool) {
	if rec == nil || rec.IsEmpty() {
		return entity.AppVersion{}, false
	}
	c := dao.AppVersion.Columns()
	return entity.AppVersion{
		Id:            rec[c.Id].Int64(),
		LatestVersion: rec[c.LatestVersion].String(),
		ReleaseNotes:  rec[c.ReleaseNotes].String(),
		DownloadUrl:   rec[c.DownloadUrl].String(),
		ForceUpdate:   rec[c.ForceUpdate].Int(),
		MinVersion:    rec[c.MinVersion].String(),
		ReleaseDate:   rec[c.ReleaseDate].Int64(),
	}, true
}

func appVersionItemJSON(row entity.AppVersion, isLatest bool) g.Map {
	return g.Map{
		"id":            row.Id,
		"latestVersion": strings.TrimSpace(row.LatestVersion),
		"releaseDate":   row.ReleaseDate,
		"releaseNotes":  strings.TrimSpace(row.ReleaseNotes),
		"downloadUrl":   gatewayapp.NormalizeAssetPath(strings.TrimSpace(row.DownloadUrl)),
		"forceUpdate":   row.ForceUpdate != 0,
		"minVersion":    strings.TrimSpace(row.MinVersion),
		"isLatest":      isLatest,
	}
}

// tryRemoveApkForDownloadPath 删行后尽力删除约定目录下的 APK；失败仅记日志。
func tryRemoveApkForDownloadPath(ctx context.Context, dlPath string) {
	dlPath = gatewayapp.NormalizeAssetPath(strings.TrimSpace(dlPath))
	if !strings.HasPrefix(dlPath, versionAdminApkURLPrefix) {
		return
	}
	name := strings.TrimPrefix(dlPath, versionAdminApkURLPrefix)
	if name == "" || !apkFilenameSafe(name) || !strings.HasSuffix(strings.ToLower(name), ".apk") {
		return
	}
	dir := filepath.Clean(gatewayapp.ApkStorageDir(ctx))
	abs := filepath.Join(dir, name)
	dirClean := filepath.Clean(dir)
	absClean := filepath.Clean(abs)
	rel, err := filepath.Rel(dirClean, absClean)
	if err != nil || strings.HasPrefix(rel, "..") {
		return
	}
	if !gfile.Exists(abs) {
		return
	}
	if err := os.Remove(abs); err != nil {
		glog.Warningf(ctx, "[gateway-app-version-admin] 删除 APK 文件失败 path=%s err=%v", abs, err)
	}
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

