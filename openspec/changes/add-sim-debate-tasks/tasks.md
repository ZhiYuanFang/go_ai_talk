## 1. ucg posts/sample 辩论过滤

- [x] 1.1 `post_sample_internal.go`：实现 `excludeDebate`/`onlyDebate` SQL 过滤；`PostSampleItem` 增加 `DebateLeft`/`DebateRight`；select 字段含立场标签与 type
- [x] 1.2 `ucg_internal_posts_sample.go` + `api/v1/ucg_internal_posts_sample_http.go`：请求体字段、互斥校验 400、透传过滤参数

## 2. sim T2 排除辩论帖

- [x] 2.1 `RunCommentTask`：sample 请求增加 `excludeDebate:true`；若响应含非空 debate 标签则 skip

## 3. sim T7 辩论发帖（12h）

- [x] 3.1 `tasks.go`：`RunPostDebateTask`（LLM JSON → POST posts）
- [x] 3.2 `schema.go`：seed prompt `post_debate_text`；`task_llm_temp.go` 增加温度常量
- [x] 3.3 `runtime.go` / `runtime_config.go`：TaskPostDebate、IntervalPostDebate（默认 12h）及 DB/env 映射

## 4. sim T8 辩论论点（1h）

- [x] 4.1 `tasks.go`：`RunDebateCommentTask`（onlyDebate sample → vote → comment ≤10 字）
- [x] 4.2 `schema.go`：seed prompt `debate_comment`
- [x] 4.3 `runtime.go` / `runtime_config.go`：TaskDebateComment、IntervalDebateComment（默认 1h）

## 5. scheduler 与 admin

- [x] 5.1 `scheduler_manager.go`：注册 T7/T8 goroutine
- [x] 5.2 `config_admin.go`：taskScheduleDefs、hasAnyTaskConfigEnabled、默认 interval 回填
- [x] 5.3 `sim_admin_api.go` / `runtime_api.go` / `sim-admin.html`：taskAiModels、prompts、runtime 表单字段 T7/T8

## 6. 校验

- [x] 6.1 `go build ./cmd/sim-user-service/... ./cmd/ucg-service/...`
- [x] 6.2 `openspec validate add-sim-debate-tasks --strict`
