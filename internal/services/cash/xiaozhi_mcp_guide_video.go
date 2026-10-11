package cash

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	ucgclient "hello/internal/clients/ucg"
	"hello/internal/shared/featurelogo"
	"hello/internal/shared/mediacdn"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// 小智引导视频固定槽（与表 xiaozhi_mcp_guide_video.slot 一致）。
const (
	XiaozhiGuideSlotPromo    = "promo"
	XiaozhiGuideSlotTutorial = "tutorial"
)

// 与 UCG MaxMediaUploadBytes 对齐（Flutter 同限）。
const xiaozhiGuideVideoMaxBytes = 25 << 20

// XiaozhiGuideVideoSlot 单槽读模型（含 CDN 展示字段）。
type XiaozhiGuideVideoSlot struct {
	Slot         string `json:"slot"`
	Title        string `json:"title"`
	ObjectKey    string `json:"objectKey,omitempty"`
	VideoCdnURL  string `json:"videoCdnUrl,omitempty"`
	CoverCdnURL  string `json:"coverCdnUrl,omitempty"`
	Configured   bool   `json:"configured"`
	UpdatedAt    int64  `json:"updatedAt"`
}

// ListXiaozhiGuideVideosAdmin 返回两槽完整态（含未配置），供 Hub。
func ListXiaozhiGuideVideosAdmin(ctx context.Context) ([]XiaozhiGuideVideoSlot, error) {
	if err := ensureXiaozhiGuideSeed(ctx); err != nil {
		return nil, err
	}
	rows, err := g.DB().Model("xiaozhi_mcp_guide_video").Ctx(ctx).
		WhereIn("slot", []string{XiaozhiGuideSlotPromo, XiaozhiGuideSlotTutorial}).
		All()
	if err != nil {
		return nil, err
	}
	bySlot := map[string]XiaozhiGuideVideoSlot{}
	for _, row := range rows {
		slot := strings.TrimSpace(row["slot"].String())
		key := strings.TrimSpace(row["object_key"].String())
		item := XiaozhiGuideVideoSlot{
			Slot:      slot,
			Title:     strings.TrimSpace(row["title"].String()),
			ObjectKey: key,
			UpdatedAt: row["updated_at"].Int64(),
		}
		if key != "" {
			item.Configured = true
			item.VideoCdnURL = featurelogo.CdnURL(ctx, key)
			item.CoverCdnURL = featurelogo.CdnURL(ctx, mediacdn.VideoThumbObjectKey(key))
		}
		bySlot[slot] = item
	}
	out := make([]XiaozhiGuideVideoSlot, 0, 2)
	for _, slot := range []string{XiaozhiGuideSlotPromo, XiaozhiGuideSlotTutorial} {
		if it, ok := bySlot[slot]; ok {
			out = append(out, it)
			continue
		}
		out = append(out, XiaozhiGuideVideoSlot{Slot: slot})
	}
	return out, nil
}

// ListXiaozhiGuideVideosApp 仅返回已配置槽（含 CDN），供 Flutter。
func ListXiaozhiGuideVideosApp(ctx context.Context) ([]XiaozhiGuideVideoSlot, error) {
	all, err := ListXiaozhiGuideVideosAdmin(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]XiaozhiGuideVideoSlot, 0, len(all))
	for _, it := range all {
		if !it.Configured || strings.TrimSpace(it.ObjectKey) == "" {
			continue
		}
		out = append(out, XiaozhiGuideVideoSlot{
			Slot:        it.Slot,
			Title:       it.Title,
			VideoCdnURL: it.VideoCdnURL,
			CoverCdnURL: it.CoverCdnURL,
			Configured:  true,
			UpdatedAt:   it.UpdatedAt,
		})
	}
	return out, nil
}

// UpdateXiaozhiGuideVideoTitle 更新槽标题（可不依赖已上传视频）。
func UpdateXiaozhiGuideVideoTitle(ctx context.Context, slot, title string) error {
	slot, err := normalizeXiaozhiGuideSlot(slot)
	if err != nil {
		return err
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "标题不能为空")
	}
	if len([]rune(title)) > 128 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "标题过长")
	}
	if err := ensureXiaozhiGuideSeed(ctx); err != nil {
		return err
	}
	now := time.Now().Unix()
	_, err = g.DB().Model("xiaozhi_mcp_guide_video").Ctx(ctx).
		Where("slot", slot).
		Data(g.Map{"title": title, "updated_at": now}).
		Update()
	return err
}

// UploadXiaozhiGuideVideo 上传/覆盖槽视频（经 clients/ucg，同 Flutter 验真与上限）。
func UploadXiaozhiGuideVideo(ctx context.Context, slot, originalFilename string, src io.Reader, size int64) (*XiaozhiGuideVideoSlot, error) {
	slot, err := normalizeXiaozhiGuideSlot(slot)
	if err != nil {
		return nil, err
	}
	if size > 0 && size > xiaozhiGuideVideoMaxBytes {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, fmt.Sprintf("视频过大，上限 %d 字节", xiaozhiGuideVideoMaxBytes))
	}
	lr := io.LimitReader(src, xiaozhiGuideVideoMaxBytes+1)
	data, err := io.ReadAll(lr)
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeInvalidParameter, err, "读取视频失败")
	}
	if len(data) == 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "视频文件为空")
	}
	if int64(len(data)) > xiaozhiGuideVideoMaxBytes {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, fmt.Sprintf("视频过大，上限 %d 字节", xiaozhiGuideVideoMaxBytes))
	}
	name := strings.TrimSpace(originalFilename)
	if name == "" {
		name = "guide.mp4"
	}
	ext := strings.ToLower(filepath.Ext(name))
	if ext != "" && ext != ".mp4" && ext != ".mpeg" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "请上传 mp4（须 H.264+AAC，与 App 视频同规）")
	}
	if err := ensureXiaozhiGuideSeed(ctx); err != nil {
		return nil, err
	}
	objectKey, _, err := ucgclient.UploadXiaozhiGuideVideo(ctx, name, data)
	if err != nil {
		return nil, err
	}
	objectKey = featurelogo.NormalizeObjectKey(objectKey)
	if objectKey == "" {
		return nil, gerror.NewCode(gcode.CodeInternalError, "上传未返回 objectKey")
	}
	now := time.Now().Unix()
	_, err = g.DB().Model("xiaozhi_mcp_guide_video").Ctx(ctx).
		Where("slot", slot).
		Data(g.Map{"object_key": objectKey, "updated_at": now}).
		Update()
	if err != nil {
		return nil, err
	}
	return slotView(ctx, slot)
}

// ClearXiaozhiGuideVideo 清空槽视频（保留标题）。
func ClearXiaozhiGuideVideo(ctx context.Context, slot string) error {
	slot, err := normalizeXiaozhiGuideSlot(slot)
	if err != nil {
		return err
	}
	if err := ensureXiaozhiGuideSeed(ctx); err != nil {
		return err
	}
	now := time.Now().Unix()
	_, err = g.DB().Model("xiaozhi_mcp_guide_video").Ctx(ctx).
		Where("slot", slot).
		Data(g.Map{"object_key": "", "updated_at": now}).
		Update()
	return err
}

func slotView(ctx context.Context, slot string) (*XiaozhiGuideVideoSlot, error) {
	one, err := g.DB().Model("xiaozhi_mcp_guide_video").Ctx(ctx).Where("slot", slot).One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotFound, "槽不存在")
	}
	key := strings.TrimSpace(one["object_key"].String())
	out := &XiaozhiGuideVideoSlot{
		Slot:      slot,
		Title:     strings.TrimSpace(one["title"].String()),
		ObjectKey: key,
		UpdatedAt: one["updated_at"].Int64(),
	}
	if key != "" {
		out.Configured = true
		out.VideoCdnURL = featurelogo.CdnURL(ctx, key)
		out.CoverCdnURL = featurelogo.CdnURL(ctx, mediacdn.VideoThumbObjectKey(key))
	}
	return out, nil
}

func normalizeXiaozhiGuideSlot(slot string) (string, error) {
	slot = strings.ToLower(strings.TrimSpace(slot))
	switch slot {
	case XiaozhiGuideSlotPromo, XiaozhiGuideSlotTutorial:
		return slot, nil
	default:
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "slot 须为 promo 或 tutorial")
	}
}

// ensureXiaozhiGuideSeed 保证两槽行存在（兼容旧库未跑种子）。
func ensureXiaozhiGuideSeed(ctx context.Context) error {
	now := time.Now().Unix()
	_, err := g.DB().Exec(ctx, `
INSERT IGNORE INTO xiaozhi_mcp_guide_video (slot, title, object_key, updated_at)
VALUES (?, '能力介绍', '', ?), (?, '绑定教学', '', ?)`,
		XiaozhiGuideSlotPromo, now, XiaozhiGuideSlotTutorial, now)
	return err
}
