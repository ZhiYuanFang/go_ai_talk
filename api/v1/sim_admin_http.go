package v1

import "github.com/gogf/gf/v2/frame/g"

type SimAdminConfigGetReq struct {
	g.Meta `path:"/sim/admin/api/config" method:"get" tags:"sim-admin" summary:"读取模拟用户配置"`
}

type SimAdminConfigDTO struct {
	Enabled     bool   `json:"enabled"`
	MaxSimUsers int    `json:"maxSimUsers"`
	UpdatedAt   int64  `json:"updatedAt"`
	UpdatedBy   string `json:"updatedBy"`
}

type SimAdminConfigGetRes struct {
	Config SimAdminConfigDTO `json:"config"`
}

type SimAdminConfigPutReq struct {
	g.Meta      `path:"/sim/admin/api/config" method:"put" tags:"sim-admin" summary:"更新模拟用户配置"`
	Enabled     bool `json:"enabled"`
	MaxSimUsers int  `json:"maxSimUsers"`
}

type SimAdminConfigPutRes struct{}

type SimAdminPromptDTO struct {
	TaskType           string `json:"taskType"`
	SystemPrompt       string `json:"systemPrompt"`
	UserPromptTemplate string `json:"userPromptTemplate"`
	UpdatedAt          int64  `json:"updatedAt"`
	UpdatedBy          string `json:"updatedBy"`
}

type SimAdminPromptGetReq struct {
	g.Meta   `path:"/sim/admin/api/prompts/{taskType}" method:"get" tags:"sim-admin" summary:"读取 Prompt 模板"`
	TaskType string `json:"taskType" in:"path" p:"taskType"`
}

type SimAdminPromptGetRes struct {
	Prompt SimAdminPromptDTO `json:"prompt"`
}

type SimAdminPromptPutReq struct {
	g.Meta             `path:"/sim/admin/api/prompts/{taskType}" method:"put" tags:"sim-admin" summary:"更新 Prompt 模板"`
	TaskType           string `json:"taskType" in:"path" p:"taskType"`
	SystemPrompt       string `json:"systemPrompt"`
	UserPromptTemplate string `json:"userPromptTemplate" v:"required"`
}

type SimAdminPromptPutRes struct{}

type SimAdminTaskRunDTO struct {
	TaskName     string `json:"taskName"`
	LastRunAt    int64  `json:"lastRunAt"`
	SuccessCount int64  `json:"successCount"`
	FailCount    int64  `json:"failCount"`
	LastError    string `json:"lastError,omitempty"`
}

type SimAdminStatusDTO struct {
	Tasks            []SimAdminTaskRunDTO `json:"tasks"`
	PendingVideoJobs int                  `json:"pendingVideoJobs"`
}

type SimAdminStatusGetReq struct {
	g.Meta `path:"/sim/admin/api/status" method:"get" tags:"sim-admin" summary:"模拟任务运行状态"`
}

type SimAdminStatusGetRes struct {
	Status SimAdminStatusDTO `json:"status"`
}

// SimAdminRuntimeTaskSwitchesDTO 各任务 env 开关（只读）。
type SimAdminRuntimeTaskSwitchesDTO struct {
	Register  bool `json:"register"`
	Comment   bool `json:"comment"`
	PostImage bool `json:"postImage"`
	PostVideo bool `json:"postVideo"`
	Chat      bool `json:"chat"`
	Follow    bool `json:"follow"`
	VideoPoll bool `json:"videoPoll"`
}

// SimAdminRuntimeIntervalsDTO 各任务周期与相关 env 时长（只读）。
type SimAdminRuntimeIntervalsDTO struct {
	Register            string `json:"register"`
	Comment             string `json:"comment"`
	PostImage           string `json:"postImage"`
	PostVideo           string `json:"postVideo"`
	Chat                string `json:"chat"`
	Follow              string `json:"follow"`
	VideoPollIdle       string `json:"videoPollIdle"`
	VideoPollActive     string `json:"videoPollActive"`
	StartupStaggerMax   string `json:"startupStaggerMax"`
	EphemeralChatLoop   string `json:"ephemeralChatLoop"`
	EphemeralChatWindow string `json:"ephemeralChatWindow"`
}

// SimAdminRuntimeDTO 进程运行时配置快照（只读，不含密钥）。
type SimAdminRuntimeDTO struct {
	ServiceEnabled    bool                           `json:"serviceEnabled"`
	DbEnabled         bool                           `json:"dbEnabled"`
	DatabaseName      string                         `json:"databaseName"`
	SimUserCount      int                            `json:"simUserCount"`
	SimUserCountError string                         `json:"simUserCountError,omitempty"`
	MaxSimUsers       int                            `json:"maxSimUsers"`
	TaskSwitches      SimAdminRuntimeTaskSwitchesDTO `json:"taskSwitches"`
	Intervals         SimAdminRuntimeIntervalsDTO    `json:"intervals"`
	RateLimitRps      float64                        `json:"rateLimitRps"`
	RateLimitBurst    int                            `json:"rateLimitBurst"`
}

type SimAdminRuntimeGetReq struct {
	g.Meta `path:"/sim/admin/api/runtime" method:"get" tags:"sim-admin" summary:"读取模拟服务运行时配置（只读）"`
}

type SimAdminRuntimeGetRes struct {
	Runtime SimAdminRuntimeDTO `json:"runtime"`
}

// SimAdminTaskRunPostReq 管理页手动触发一次周期任务（异步执行）。
type SimAdminTaskRunPostReq struct {
	g.Meta   `path:"/sim/admin/api/tasks/{taskName}/run" method:"post" tags:"sim-admin" summary:"手动执行模拟任务一次"`
	TaskName string `json:"taskName" in:"path" p:"taskName" v:"required"`
}

// SimAdminTaskRunPostRes 手动触发已接受。
type SimAdminTaskRunPostRes struct {
	Accepted bool   `json:"accepted"`
	TaskName string `json:"taskName"`
	Message  string `json:"message"`
}
