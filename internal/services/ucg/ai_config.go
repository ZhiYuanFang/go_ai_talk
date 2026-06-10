package ucg

import (
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const aiConfigSingletonID = 1

// defaultVisionModel 默认 vision 模型（DashScope Qwen3-VL-Plus；OpenAI 兼容 endpoint）。
const defaultVisionModel = "qwen3-vl-plus"

// defaultVisionEndpoint DashScope 北京区 OpenAI 兼容 chat completions。
const defaultVisionEndpoint = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"

// AllowedVisionModels Admin PUT 与 UI 下拉硬编码 allowlist。
var AllowedVisionModels = []string{
	"qwen3-vl-plus",
	"qwen3-vl-plus-2025-09-23",
	"qwen3-vl-flash",
	"qwen3-vl-flash-2025-10-15",
	"qwen-vl-plus",
	"qwen-vl-max",
	// 历史/自托管 provider 可选保留
	"deepseek-ai/deepseek-vl2",
	"deepseek-ai/deepseek-vl2-small",
	"deepseek-ai/deepseek-vl2-tiny",
	"deepseek-vl2",
}

// RuntimeAIConfig polish 运行时配置（含 DashScope 凭证）。
type RuntimeAIConfig struct {
	VisionModel         string
	MaxImagesPerRequest int
	DashScopeAPIKey     string
	VisionEndpoint      string
	TimeoutSeconds      int
}

type aiConfigCache struct {
	cfg       RuntimeAIConfig
	expiresAt time.Time
}

var (
	aiConfigMu    sync.RWMutex
	aiConfigEntry aiConfigCache
)

// LoadAIConfig reads DB singleton id=1 with YAML fallback and ~60s TTL cache.
func LoadAIConfig(ctx context.Context) RuntimeAIConfig {
	aiConfigMu.RLock()
	if time.Now().Before(aiConfigEntry.expiresAt) {
		cfg := aiConfigEntry.cfg
		aiConfigMu.RUnlock()
		return cfg
	}
	aiConfigMu.RUnlock()

	cfg := loadAIConfigFresh(ctx)
	aiConfigMu.Lock()
	aiConfigEntry = aiConfigCache{cfg: cfg, expiresAt: time.Now().Add(60 * time.Second)}
	aiConfigMu.Unlock()
	return cfg
}

// InvalidateAIConfigCache clears the in-memory AI config cache.
func InvalidateAIConfigCache() {
	aiConfigMu.Lock()
	aiConfigEntry = aiConfigCache{}
	aiConfigMu.Unlock()
}

func loadAIConfigFresh(ctx context.Context) RuntimeAIConfig {
	yamlVision := strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.ai.default_vision_model").String())
	yamlMax := g.Cfg().MustGet(ctx, "ucg.ai.default_max_images_per_request").Int()
	if yamlMax <= 0 {
		yamlMax = 9
	}
	if yamlVision == "" {
		yamlVision = defaultVisionModel
	}

	cfg := RuntimeAIConfig{
		VisionModel:         yamlVision,
		MaxImagesPerRequest: yamlMax,
		VisionEndpoint:      strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.ai.vision_endpoint").String()),
		DashScopeAPIKey:     strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.ai.dashscope_api_key").String()),
		TimeoutSeconds:      g.Cfg().MustGet(ctx, "ucg.ai.timeout_seconds").Int(),
	}
	if key := strings.TrimSpace(os.Getenv("UCG_DASHSCOPE_API_KEY")); key != "" {
		cfg.DashScopeAPIKey = key
	} else if key := strings.TrimSpace(os.Getenv("UCG_DEEPSEEK_API_KEY")); key != "" {
		// 兼容旧 env 名；新部署请用 UCG_DASHSCOPE_API_KEY
		cfg.DashScopeAPIKey = key
	}
	if cfg.VisionEndpoint == "" {
		cfg.VisionEndpoint = defaultVisionEndpoint
	}
	if cfg.TimeoutSeconds <= 0 {
		cfg.TimeoutSeconds = 60
	}

	var row entity.UcgAiConfig
	err := dao.UcgAiConfig.Ctx(ctx).Where(dao.UcgAiConfig.Columns().Id, aiConfigSingletonID).Scan(&row)
	if err == nil && row.Id == aiConfigSingletonID {
		if m := strings.TrimSpace(row.VisionModel); m != "" {
			cfg.VisionModel = m
		}
		if row.MaxImagesPerRequest > 0 {
			cfg.MaxImagesPerRequest = row.MaxImagesPerRequest
		}
	}
	return cfg
}

// AIConfigDTO Admin GET/PUT 对外字段。
type AIConfigDTO struct {
	VisionModel         string `json:"visionModel"`
	MaxImagesPerRequest int    `json:"maxImagesPerRequest"`
	UpdatedAt           int64  `json:"updatedAt"`
	UpdatedBy           string `json:"updatedBy"`
}

// GetAIConfigForAdmin returns current DB-backed AI config for Admin UI.
func GetAIConfigForAdmin(ctx context.Context) AIConfigDTO {
	var row entity.UcgAiConfig
	_ = dao.UcgAiConfig.Ctx(ctx).Where(dao.UcgAiConfig.Columns().Id, aiConfigSingletonID).Scan(&row)
	if row.Id != aiConfigSingletonID {
		fallback := loadAIConfigFresh(ctx)
		return AIConfigDTO{
			VisionModel:         fallback.VisionModel,
			MaxImagesPerRequest: fallback.MaxImagesPerRequest,
		}
	}
	return AIConfigDTO{
		VisionModel:         row.VisionModel,
		MaxImagesPerRequest: row.MaxImagesPerRequest,
		UpdatedAt:           row.UpdatedAt,
		UpdatedBy:           row.UpdatedBy,
	}
}

// UpdateAIConfigForAdmin validates and persists Admin PUT.
func UpdateAIConfigForAdmin(ctx context.Context, visionModel string, maxImages int, updatedBy string) error {
	visionModel = strings.TrimSpace(visionModel)
	if !isAllowedVisionModel(visionModel) {
		return gerror.NewCode(gcode.CodeInvalidParameter, "visionModel 不在 allowlist")
	}
	if maxImages <= 0 || maxImages > 9 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "maxImagesPerRequest 须为 1-9")
	}
	now := time.Now().Unix()
	cols := dao.UcgAiConfig.Columns()
	_, err := dao.UcgAiConfig.Ctx(ctx).Data(g.Map{
		cols.Id:                  aiConfigSingletonID,
		cols.VisionModel:         visionModel,
		cols.MaxImagesPerRequest: maxImages,
		cols.UpdatedAt:           now,
		cols.UpdatedBy:           strings.TrimSpace(updatedBy),
	}).Save()
	if err != nil {
		return err
	}
	InvalidateAIConfigCache()
	return nil
}

func isAllowedVisionModel(model string) bool {
	for _, m := range AllowedVisionModels {
		if m == model {
			return true
		}
	}
	return false
}

// UcgAdminPassword returns admin password from env UCG_ADMIN_PASSWORD or yaml ucg.admin.password.
func UcgAdminPassword(ctx context.Context) string {
	if v := strings.TrimSpace(os.Getenv("UCG_ADMIN_PASSWORD")); v != "" {
		return v
	}
	return strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.admin.password").String())
}

// VerifyUcgAdminPassword checks X-Admin-Password header value.
func VerifyUcgAdminPassword(ctx context.Context, password string) bool {
	expected := UcgAdminPassword(ctx)
	if expected == "" {
		return false
	}
	return strings.TrimSpace(password) == expected
}
